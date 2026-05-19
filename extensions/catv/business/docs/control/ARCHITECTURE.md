# STB 제어 시스템 아키텍처

**최종 업데이트**: 2026-03-31

---

## 목차

- [시스템 구조](#시스템-구조)
- [🎯 핵심 결정](#-핵심-결정)
  - [실행 방식: Kubernetes Job (단발성)](#실행-방식-kubernetes-job-단발성)
  - [통신 프로토콜: TLS over TCP (직접 다이얼)](#통신-프로토콜-tls-over-tcp-직접-다이얼)
  - [결과 저장: AsyncBatchWriter (3단계 보호)](#결과-저장-asyncbatchwriter-3단계-보호)
  - [종료 순서: WorkerPool → AsyncBatchWriter](#종료-순서-workerpool--asyncbatchwriter)
- [🏗️ 컴포넌트 역할](#️-컴포넌트-역할)
  - [WorkerPool](#workerpool)
  - [Rate Limiter](#rate-limiter)
  - [StbControlHandler](#stbcontrolhandler)
  - [AsyncBatchWriter](#asyncbatchwriter)
- [📊 성능 수치](#-성능-수치)

---

## 시스템 구조

```
┌─────────────────────────────────────────────────────────────┐
│                   Controller Command System                 │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐    ┌─────────────┐    ┌────────────────┐  │
│  │   Command   │───▶│   Service   │───▶│  WorkerPool    │  │
│  │  (Cobra)    │    │  (Loader)   │    │  6,000 workers │  │
│  └─────────────┘    └─────────────┘    └────────────────┘  │
│                                                │            │
│                                                ▼            │
│                                      ┌─────────────────┐   │
│                                      │StbControlHandler│   │
│                                      │ - Direct TLS    │   │
│                                      │ - Buffer pool   │   │
│                                      └─────────────────┘   │
│                                                │            │
│                         ┌──────────────────────┘           │
│                         ▼                                   │
│               ┌──────────────────┐                         │
│               │   Rate Limiter   │                         │
│               │  333 conn/sec    │                         │
│               └──────────────────┘                         │
│                         │                                   │
│                         ▼                                   │
│               ┌──────────────────┐                         │
│               │  Settopbox (TLS) │                         │
│               └──────────────────┘                         │
│                         │                                   │
│         ┌───────────────┴──────────────────┐               │
│         │     Result Path (3단계 보호)      │               │
│         ▼                                   │               │
│  ┌──────────────────────────────────┐       │               │
│  │  Channel (10s timeout)           │       │               │
│  │  ├─ Primary: resultChannel <- req│       │               │
│  │  └─ Timeout: Direct DB insert    │       │               │
│  └──────────────────────────────────┘       │               │
│                         │                   │               │
│                         ▼                   │               │
│  ┌──────────────────────────────────────┐   │               │
│  │  AsyncBatchWriter                    │   │               │
│  │  - Channel: 20,000 / Batch: 10,000   │   │               │
│  │  - Flush: 5s                         │   │               │
│  └──────────────────────────────────────┘   │               │
│                         │                   │               │
│                         ▼                   │               │
│  ┌──────────────────────────────────────┐   │               │
│  │  ClickHouse Batch Insert             │   │               │
│  └──────────────────────────────────────┘   │               │
│                                             │               │
└─────────────────────────────────────────────┴───────────────┘
```

---

## 🎯 핵심 결정

### 실행 방식: Kubernetes Job (단발성)

| 비교 항목 | K8s Job (선택)                        | Daemon / CronJob         |
| --------- | ------------------------------------- | ------------------------ |
| 실행 단위 | 스케줄 1건 = Job 1개                  | 단일 프로세스 공유       |
| 격리성    | 완전 격리 (실패해도 타 스케줄 무영향) | 공유 메모리 경합         |
| 추적      | Job 이름으로 로그/결과 추적           | 스케줄 ID 별도 관리 필요 |
| 확장성    | 병렬 스케줄 실행 자연스러움           | 멀티스레딩 복잡도 증가   |

**결론**: 스케줄마다 독립된 K8s Job을 생성 → k8s_job_name을 결과 테이블의 파티션 키로 활용

---

### 통신 프로토콜: TLS over TCP (직접 다이얼)

#### TLS 필요성

최초 plain TCP로 접근 시 STB가 TLS Alert를 반환. STB는 **TLS 서버**로 동작.

#### Go 기본값으로는 동작하지 않는 이유

| 문제                  | 원인                                            | 해결                                           |
| --------------------- | ----------------------------------------------- | ---------------------------------------------- |
| TLS handshake failure | STB가 TLS 1.0만 지원, Go 기본 MinVersion=TLS1.2 | `MinVersion: tls.VersionTLS10`                 |
| TLS handshake failure | STB가 ECDHE 계열 미지원, RSA 키 교환만 지원     | `InsecureCipherSuites` 포함 전체 cipher 활성화 |
| 인증서 검증 실패      | STB 자체 서명 인증서 사용                       | `InsecureSkipVerify: true`                     |

#### 응답 방식

```
Client → Write(JSON) → CloseWrite(TLS close_notify) → STB가 응답 전송
```

STB는 단순 데이터 수신만으로는 응답을 보내지 않고, **write half-close 신호를 받은 후에만** 응답을 전송하는 프로토콜을 사용한다. `tls.Conn.CloseWrite()`는 read 방향을 열어두므로 응답 수신이 가능하다.

자세한 모델별 호환성 분석은 [STB_COMPATIBILITY.md](STB_COMPATIBILITY.md) 참조.

#### Connection Pool 미사용 이유

각 STB가 고유 IP를 가지므로 Connection Pool의 재사용률이 **0%**. Pool 유지 비용(hash 계산, lock, map 조회, singleflight) 대비 이득이 없어 직접 `tls.Dialer.DialContext()`를 사용.

---

### 결과 저장: AsyncBatchWriter (3단계 보호)

단순 INSERT 대신 배치 처리를 선택한 이유:

- STB 수천 대가 동시에 완료 → ClickHouse에 단건 INSERT 수천 번 발생 시 부하 집중
- 배치 INSERT로 I/O 횟수 대폭 감소

```
Job 완료
    │
    ├─ [1차] resultChannel <- result (10초 타임아웃)
    │         └─ AsyncBatchWriter → ClickHouse Batch INSERT
    │
    ├─ [2차] 채널 타임아웃 시 → context.Background()로 DB 직접 INSERT
    │
    └─ [3차] ctx.Done() 감지 시 → DB 직접 INSERT 후 ctx.Err() 반환
```

`context.Background()`를 사용하는 이유: 원본 ctx가 취소돼도 결과는 반드시 저장해야 하기 때문.

---

### 종료 순서: WorkerPool → AsyncBatchWriter

```
Stop() 호출
    │
    ├─ 1. cancel()        — 신규 TLS 연결 차단
    ├─ 2. rateLimiter.Close()
    ├─ 3. workerPool.Stop()   — 실행 중 Job이 모두 완료될 때까지 대기
    └─ 4. resultWriter.Stop() — 채널에 남은 레코드 모두 flush 후 종료
```

WorkerPool이 완전히 종료된 후 BatchWriter를 종료하므로 in-flight 결과가 유실되지 않는다.

---

## 🏗️ 컴포넌트 역할

### WorkerPool

- `WorkerCount`: 6,000 (= max_connections_per_second × avg_processing_time)
- `JobQueueSize`: 1,200,000
- 실제 처리는 `StbControlHandler.Handle()`에 위임

### Rate Limiter

포트 고갈(port exhaustion) 방지. TIME_WAIT 소켓이 60초간 유지되므로:

```
사용 가능한 포트 20,000 ÷ TIME_WAIT 60s = 333 conn/s (이론적 최대)
```

SERVICE 인스턴스마다 독립적인 Rate Limiter를 생성하므로, 동시 서비스 수가 많아지면 실제 연결 속도가 배수로 증가함에 유의.

### StbControlHandler

- TLS dial → JSON write → CloseWrite → response read
- Port exhaustion 감지 시 progressive backoff (20s, 40s)
- [STB_COMPATIBILITY.md](STB_COMPATIBILITY.md)에 명시된 모델은 API 미지원으로 `response_length=0` 반환

### AsyncBatchWriter

- Channel: `max_count × 2` (기본 20,000)
- Batch: `max_count` (기본 10,000)
- Flush: 5s 또는 배치 크기 도달 시

---

## 📊 성능 수치

| 항목               | 값         | 비고                        |
| ------------------ | ---------- | --------------------------- |
| Rate limit         | 333 conn/s | per service instance        |
| Worker count       | 6,000      | = 333 × ~18s avg processing |
| Connection timeout | 5s         | TLS handshake 포함          |
| Send timeout       | 30s        | write + CloseWrite + read   |
| Batch size         | 10,000     | ClickHouse INSERT 단위      |
| Channel size       | 20,000     | 배압(backpressure) 버퍼     |
