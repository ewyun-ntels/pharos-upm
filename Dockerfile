FROM docker.io/library/node:24-alpine AS builder_pharos_frontend

# [frontend_source_collector]
# Go 소스 파일을 제거하여 Go 변경 시 프론트엔드 캐시가 깨지지 않도록 한다.
# gomod_collector와 동일한 content-addressable 캐시 원리 적용:
#   - Go 파일만 변경 → 이 스테이지는 재실행되지만 /src 내용이 동일 → COPY --from 캐시 히트
#   - 프론트엔드 파일 변경 → /src 내용 변경 → 하위 레이어 캐시 무효화
FROM alpine AS frontend_source_collector
WORKDIR /src
COPY . .
RUN find . -name "*.go" \
        -not -path "*/.git/*" \
        -not -path "*/node_modules/*" \
    | xargs -r rm -f

FROM builder_pharos_frontend AS deps

RUN apk add --no-cache libc6-compat

WORKDIR /app

#COPY core/frontend/package.json ./core/ui/yarn.lock* ./core/ui/package-lock.json* ./core/ui/pnpm-lock.yaml* ./core/ui/.npmrc* ./

COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./

COPY --from=frontend_source_collector /src/core ./core
COPY --from=frontend_source_collector /src/extensions ./extensions
COPY --from=frontend_source_collector /src/shared ./shared
COPY --from=frontend_source_collector /src/meta ./meta

#RUN \
#    if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
#    elif [ -f package-lock.json ]; then npm ci; \
#    elif [ -f pnpm-lock.yaml ]; then yarn global add pnpm && pnpm i --frozen-lockfile; \
#    else echo "Lockfile not found." && exit 1; \
#    fi

#COPY ./core/frontend ./

RUN --mount=type=cache,target=/root/.npm \
    --mount=type=cache,target=/root/.local/share/pnpm/store \
    npm install -g pnpm \
    && pnpm install --frozen-lockfile

ARG SITE_MODE="${SITE_MODE}"
# Extension auto-import 스크립트 사용 (SITE_MODE 기반 필터링)
# Frontend: meta/scripts/generate-metadata.ts
# Backend: meta/scripts/generate-metadata.go
WORKDIR /app/core/frontend
RUN --mount=type=cache,target=/app/core/frontend/.next/cache \
    --mount=type=cache,target=/app/node_modules/.cache \
    SITE_MODE=${SITE_MODE} pnpm build


# [gomod_collector]
# go.mod / go.sum 파일을 동적으로 수집하는 중간 스테이지
#
# 목적: [2/4] COPY go.mod/go.sum 경로 하드코딩 제거
#   - go.work에 모듈 추가/이동 시 Dockerfile 수정 불필요
#   - find로 전체 소스에서 go.mod/go.sum을 자동 발견
#
# BuildKit 캐시 동작:
#   - 소스 변경 → 이 스테이지는 재실행되지만 /out 내용이 동일하면
#     COPY --from=gomod_collector 레이어는 캐시 히트 (content-addressable)
#   - go.mod/go.sum 변경 → /out 내용 변경 → 하위 레이어 캐시 무효화
FROM alpine AS gomod_collector
WORKDIR /src
COPY . .
RUN find . \( -name "go.mod" -o -name "go.sum" \) \
        -not -path "*/.git/*" \
        -not -path "*/node_modules/*" \
        -not -path "*/vendor/*" \
    | sort \
    | while IFS= read -r f; do \
        install -D "$f" "/out/$f"; \
      done && \
    cp go.work go.work.sum /out/

FROM docker.io/library/golang:1.26.2-trixie AS builder_pharos_backend

ARG SITE_MODE="${SITE_MODE}"

# BUILD GO PKG
WORKDIR /SRC

# [1/4] 시스템 패키지 설치 — 소스 변경과 무관하게 캐시 유지
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    upx-ucl unixodbc-dev

# [2/4] go.mod / go.sum 동적 수집 (gomod_collector 참조)
#   - go.mod/go.sum 내용이 같으면 소스가 변경돼도 캐시 히트
COPY --from=gomod_collector /out/ .

# [3/4] 외부 모듈 다운로드 — [2/4] 캐시 히트 시 건너뜀
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# [4/4] 전체 소스 복사 후 빌드
COPY Makefile Makefile
COPY shared/ shared/
COPY core/ core/
COPY extensions/ extensions/
COPY third_party/ third_party/
COPY meta/ meta/
COPY --from=deps /app/core/frontend/out core/frontend/out

# Version/build-time arguments passed from build system (Jenkins/Kaniko)
ARG GIT_COMMIT
ARG GIT_TAG
ARG BUILD_TIME

# go work sync / go mod download는 build_prepare에서 실행됨
# go mod download는 [3/4] 레이어에서도 실행되나 캐시 히트로 빠름
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    if [ -z "$BUILD_TIME" ]; then BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"; fi; \
    ARGS=""; \
    if [ -n "$GIT_COMMIT" ]; then ARGS="$ARGS GIT_COMMIT=$GIT_COMMIT"; fi; \
    if [ -n "$GIT_TAG" ]; then ARGS="$ARGS GIT_TAG=$GIT_TAG"; fi; \
    if [ -n "$BUILD_TIME" ]; then ARGS="$ARGS BUILD_TIME=$BUILD_TIME"; fi; \
    BIN_PATH=/APP/bin make $ARGS build_pharos

#
# DEPLOY
#
FROM docker.io/library/debian:stable-slim

LABEL maintainer="ntels"
LABEL title="pharos"
LABEL version="0.1"
LABEL description="pharos"

# install curl
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
    apt-get update && \
    apt-get install -y --no-install-recommends curl tzdata unixodbc-dev

ARG APP_DEST="/opt/pharos"
ARG APP_BIN_DEST="${APP_DEST}/bin"
ARG APP_CONFIG_DEST="${APP_DEST}/config"

RUN mkdir -p ${APP_BIN_DEST} ${APP_CONFIG_DEST}

# COPY from host
COPY core/bin/ ${APP_BIN_DEST}
COPY core/config/ ${APP_CONFIG_DEST}

# COPY from build
COPY --from=builder_pharos_backend /APP/bin/ ${APP_BIN_DEST}

ENV PATH="${PATH}:${APP_BIN_DEST}"
ENV MASTER_STATISTICS_ALTIBASE_DRIVER=${APP_CONFIG_DEST}/shared-library/altibase/linux/7.3.0/libaltibase_odbc-64bit-ul64.so
ENV SHARED-LIBRARY_ALTIBASE_6_DRIVER=${APP_CONFIG_DEST}/shared-library/altibase/linux/6.5.1.10.3/libaltibase_odbc-64bit-ul64.so
ENV SHARED-LIBRARY_ALTIBASE_7_DRIVER=${APP_CONFIG_DEST}/shared-library/altibase/linux/7.3.0/libaltibase_odbc-64bit-ul64.so
ENV PLUGINS_PATH=${APP_BIN_DEST}/plugins

## Add non-root user and set ownership (여기에 사용자 추가)
#RUN groupadd -g 10001 pharos && useradd -m -u 10001 -g pharos -s /usr/sbin/nologin pharos && chown -R pharos:pharos ${APP_DEST}
#USER pharos

ENTRYPOINT ["/opt/pharos/bin/run.sh"]
