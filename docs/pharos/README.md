# Pharos 프로젝트 분석 문서

이 디렉토리는 Pharos/CATV 프로젝트 전체에 대한 상세 분석 문서를 포함합니다.

## 문서 목록

| 파일                                         | 내용                                                       |
| -------------------------------------------- | ---------------------------------------------------------- |
| [01-overview.md](01-overview.md)             | 프로젝트 개요, 목적, 기술 스택 전체 요약                   |
| [02-architecture.md](02-architecture.md)     | 시스템 전체 아키텍처 및 컴포넌트 관계                      |
| [03-backend.md](03-backend.md)               | Go 백엔드 구조, 패키지 상세, 설정 체계                     |
| [04-frontend.md](04-frontend.md)             | React 프론트엔드 구조, Extension 레지스트리, 주요 컴포넌트 |
| [05-catv-extension.md](05-catv-extension.md) | CATV Extension 상세 (수집·제어·ETL·메트릭·마이그레이션)    |
| [06-database.md](06-database.md)             | ClickHouse 스키마 및 테이블 설계                           |
| [07-cicd.md](07-cicd.md)                     | CI/CD 파이프라인 및 배포 방법                              |
| [08-config.md](08-config.md)                 | 설정 파일 구조 및 env 변수 레퍼런스                        |
| [09-prepare-role-model.md](09-prepare-role-model.md) | prepare/role/attribute 관계와 UI 표시 개선 계획            |

## 빠른 참고

- **빌드**: `SITE_MODE=catv make build`
- **개발 서버**: `SITE_MODE=catv pnpm dev` (Frontend) + `SITE_MODE=catv make build_pharos && ./bin/pharos server` (Backend)
- **Docker**: `SITE_MODE=catv make docker`
- **SITE_MODE 옵션**: `catv` / `demo` / `default` / `example` / `clickhouse` / `all`
