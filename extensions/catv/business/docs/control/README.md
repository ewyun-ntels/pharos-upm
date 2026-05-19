# STB 제어 시스템

> CATV 셋톱박스에 제어 명령을 전송하고 결과를 저장하는 배치 처리 시스템

**최종 업데이트**: 2026-03-31  
**상태**: 🟢 운영 중

---

## 목차

- [🎯 개요](#-개요)
- [📡 대상 규모](#-대상-규모)
- [📁 문서 구조](#-문서-구조)
- [🔧 제어 명령 타입](#-제어-명령-타입)
- [📊 호환성 현황](#-호환성-현황)
- [⚙️ 주요 설정](#️-주요-설정)

---

## 🎯 개요

### 빠른 정보

| 항목          | 내용                                  |
| ------------- | ------------------------------------- |
| **실행 방식** | Kubernetes Job (스케줄 1건 = Job 1개) |
| **통신 방식** | TLS over TCP (port 8801)              |
| **대상**      | CATV STB (셋톱박스)                   |
| **결과 저장** | ClickHouse (AsyncBatchWriter)         |
| **처리량**    | 333 conn/s per service                |

### 처리 흐름

```
K8s Job 생성
    ↓
ClickHouse에서 처리 대상 STB 목록 조회 (k8s_job_name 기준)
    ↓
WorkerPool (6,000 workers) + RateLimiter (333/s)로 병렬 처리
    ↓
STB별 TLS 연결 → JSON 명령 전송 → CloseWrite → 응답 수신
    ↓
AsyncBatchWriter → ClickHouse 배치 저장
```

---

## � 대상 규모

### 쿼리

```sql
SELECT
    SCRBR_SO_NM,
    count(DISTINCT CM_MAC_ADDR) AS count
FROM "catv"."dist_stb_information" final
WHERE SCRBR_SO_NM != ''
GROUP BY SCRBR_SO_NM
ORDER BY count DESC;
```

### 결과

SO 24개, STB 763619대

| SO 이름    | STB 수 |
| ---------- | ------ |
| 수원방송   | 85348  |
| 기남방송   | 77565  |
| 한빛방송   | 57496  |
| 도봉강북   | 48300  |
| ABC방송    | 43475  |
| 중부방송   | 41754  |
| 낙동방송   | 40709  |
| 서부산방송 | 40561  |
| 전주방송   | 39438  |
| 티씨엔방송 | 34335  |
| 동남방송   | 30479  |
| 강서방송   | 26045  |
| 남동방송   | 25447  |
| 새롬방송   | 20095  |
| 광진성동   | 19557  |
| 대경방송   | 19423  |
| 대구방송   | 18441  |
| 동대문방송 | 13499  |
| 서해방송   | 13336  |
| 노원방송   | 13235  |
| 서대문방송 | 10298  |
| 종로중구   | 9195   |
| 세종방송   | 8587   |
| testbed    | 64     |

---

## �📁 문서 구조

| 문서                                               | 내용                        | 대상             |
| -------------------------------------------------- | --------------------------- | ---------------- |
| **[ARCHITECTURE.md](./ARCHITECTURE.md)**           | 핵심 설계 결정 및 구현 근거 | 개발자, 아키텍트 |
| **[STB_COMPATIBILITY.md](./STB_COMPATIBILITY.md)** | 모델별 TLS 호환성 분석 결과 | 운영자, 개발자   |

### 코드 문서

- [pkg/control/command/README.md](../../pkg/control/command/README.md) — 패키지 사용법 및 설정
- [pkg/control/command/](../../pkg/control/command/) — Go 구현 코드

---

## 🔧 제어 명령 타입

`types.go`에 40+개 명령 상수 정의. 주요 카테고리:

| 카테고리  | 명령 예시                                       |
| --------- | ----------------------------------------------- |
| 정보 수집 | `stb_request_info`, `sysCheck`                  |
| 재부팅    | `smartReboot`, `stbRestart`                     |
| 단말 제어 | `limitAge`, `tvLock`, `resolution`, `audioMode` |

---

## 📊 호환성 현황

정상 동작 모델: 13개 (THX-U300, BHX-HC100, UC1600, SMT-C5010, UC2600, SMT-C5012, UC2000, SX730C-CT, UC1000, SMT-C3022, SMT-C5011, LSC630-8DTB, GX-KD630CH)  
미지원 모델: BKO-UC500 — `stb_request_info` API 미지원 확인  
미확인 모델: TMA-U400 — 테스트 환경에 해당 장비 없음

자세한 내용은 [STB_COMPATIBILITY.md](./STB_COMPATIBILITY.md) 참조.

---

## ⚙️ 주요 설정

```toml
[catv.control]
port                       = 8801
max_connections_per_second = 333

[catv.control.timeout]
connection = "5s"
send       = "30s"

[catv.control.retry.total]
count = 1
delay = "5s"

[pool.workers]
worker_count   = 6000
job_queue_size = 1200000
```

전체 설정은 [pkg/control/command/README.md](../../pkg/control/command/README.md#️-주요-설정) 참조.
