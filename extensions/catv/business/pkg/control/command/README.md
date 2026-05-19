# Controller Command Package

CATV STB 제어를 위한 배치 작업 처리 패키지입니다. Kubernetes Job으로 실행되어 TLS over TCP 통신으로 셋톱박스에 명령을 전송하고 결과를 ClickHouse에 저장합니다.

---

## 📁 목차

- [패키지 구조](#-패키지-구조)
- [STB TLS 통신 프로토콜](#-stb-tls-통신-프로토콜)
- [제어 명령 타입](#-제어-명령-타입)
- [실행 흐름](#-실행-흐름)
- [사용 방법](#-사용-방법)
- [시뮬레이션 모드](#-시뮬레이션-모드)
- [주요 설정](#️-주요-설정)

---

## 📁 패키지 구조

```
command/
├── command.go              # Cobra CLI 진입점 (catv-schedule-settopbox)
├── service.go              # 작업 로드·실행·종료 생명주기 (Service 구조체)
├── types.go                # 제어 명령 상수 (40+개), AllWorkTypes, StbControlResponse
├── settopbox.go            # TLS 연결·명령 전송·응답 수신 (Settopbox 구조체)
├── worker_pool_handler.go  # 워커풀 핸들러 (StbControlHandler) — Settopbox 호출 및 결과 DB 저장
└── async_batch_writer.go   # 비동기 배치 ClickHouse 쓰기 (AsyncBatchWriter)
```

도메인 수준 문서: [docs/control/](../../../../../docs/control/)

> 아키텍처 결정, 컴포넌트 역할, 결과 저장 설계, 종료 순서는 [docs/control/ARCHITECTURE.md](../../../../../docs/control/ARCHITECTURE.md)를 참조하세요.

---

## 🔐 STB TLS 통신 프로토콜

STB 장비는 일반 TCP가 아닌 **TLS 서버**로 동작합니다.

### TLS 설정

STB 장비는 구형 cipher suite(RSA 키 교환 계열)만 지원하므로 다음 설정이 필요합니다:

```go
tls.Config{
    InsecureSkipVerify: true,             // STB 자체 서명 인증서
    MinVersion:         tls.VersionTLS10, // STB는 TLS 1.0 지원
    CipherSuites:       allCiphers,       // 구형 RSA cipher 포함 전체 활성화
}
```

- `InsecureSkipVerify`: STB가 자체 서명 인증서를 사용하므로 검증 생략
- `MinVersion TLS1.0`: Go 기본값(TLS 1.2)으로는 handshake 실패
- `InsecureCipherSuites`: Go 기본 cipher는 ECDHE 계열만 포함 → handshake 실패

### 프로토콜 시퀀스

```
Client (Pharos)                    STB (TLS Server)
      │                                   │
      │──── TLS Handshake ───────────────▶│
      │◀─── TLS Handshake ────────────────│
      │                                   │
      │──── Write(JSON command) ─────────▶│
      │──── CloseWrite (TLS close_notify)▶│  ← STB는 이 FIN을 받아야 응답
      │                                   │
      │◀─── Read(JSON response) ──────────│
      │◀─── EOF ──────────────────────────│
```

STB는 클라이언트의 **write half-close(TLS close_notify)** 를 수신한 후에만 응답을 전송합니다.
`tls.Conn.CloseWrite()`는 read 방향은 열어두므로 응답 수신이 가능합니다.

모델별 호환성 분석: [docs/control/STB_COMPATIBILITY.md](../../../../../docs/control/STB_COMPATIBILITY.md)

---

## 🔧 제어 명령 타입

`types.go`의 `AllWorkTypes`에 40+개 명령 상수가 정의되어 있습니다.

| 카테고리          | 명령                                                                                                                                                                                                                                                                                                                                                                                                                |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **정보 수집**     | `sysCheck`, `stb_request_info`                                                                                                                                                                                                                                                                                                                                                                                      |
| **재부팅/재시작** | `smartReboot`, `stbRestart`                                                                                                                                                                                                                                                                                                                                                                                         |
| **SNMP 제어**     | `stb_control_restart`, `stb_control_reset`                                                                                                                                                                                                                                                                                                                                                                          |
| **단말 제어**     | `limitAge`, `tvLock`, `skipCh`, `easyBuying`, `resetPin`, `favCh`, `zappingAd`, `miniEpg`, `miniEpgAd`, `tvCaption`, `tvImpaired`, `barkerCh`, `vodView`, `vodRelay`, `resolution`, `audioMode`, `hdmiCec`, `hdcp`, `hdr`, `mobilePay`, `morningAlarm`, `bootMenu`, `pmsOn`, `oneAdOn`, `audioLang`, `standbyMode`, `savePwr`, `voiceGuide`, `chUpDown`, `chDca`, `volume`, `showMenu`, `limitContents`, `stbPower` |

`Settopbox.Control()`은 `AllWorkTypes` 슬라이스로 유효성을 검사하여 지원하지 않는 `work_type`은 즉시 에러를 반환합니다.

---

## 🔄 실행 흐름

```
K8s Job 시작
    ↓
command.go: --k8s-job-name 플래그 파싱
    ↓
service.go: NewService() — ClickHouse에서 해당 k8s_job_name의 레코드 목록 조회
    ↓
service.go: load() — RateLimiter / WorkerPool / AsyncBatchWriter 초기화
    ↓
service.go: Start() — 레코드마다 RateLimiter.Wait() → BatchSubmitter.Add(Job)
    ↓
worker_pool_handler.go: StbControlHandler.Handle() — Settopbox.Control() 호출
    ↓
settopbox.go: TLS 연결 → JSON 전송 → CloseWrite → 응답 수신 → 파싱
    ↓
worker_pool_handler.go: 결과를 resultChannel로 전송
    ↓
async_batch_writer.go: 배치 단위로 ClickHouse에 upsert
    ↓
service.go: Stop() — ctx 취소 → RateLimiter 종료 → WorkerPool.Stop() → AsyncBatchWriter.Stop()
```

**종료 순서가 중요합니다**: WorkerPool이 완전히 종료된 후에 AsyncBatchWriter를 종료해야 채널에 남은 결과가 유실되지 않습니다.

---

## 🚀 사용 방법

```bash
# Kubernetes Job에서 자동 호출
./pharos control-schedule --k8s-job-name <k8s-job-name>

# 설정 파일 지정
./pharos control-schedule --config /etc/pharos/config.toml --k8s-job-name <k8s-job-name>

# 데몬 모드로 실행
./pharos control-schedule --daemon --pid-file /var/run/control-schedule.pid --log-file /var/log/control-schedule.log --k8s-job-name <k8s-job-name>
```

### 플래그

| 플래그           | 기본값                                 | 설명                            |
| ---------------- | -------------------------------------- | ------------------------------- |
| `--config`, `-c` | `./config/config.toml`                 | 설정 파일 경로 (복수 지정 가능) |
| `--k8s-job-name` | (필수)                                 | 처리할 K8s Job 이름             |
| `--daemon`, `-d` | `false`                                | 백그라운드 데몬 모드 실행       |
| `--pid-file`     | `/var/run/catv-schedule-settopbox.pid` | 데몬 모드 PID 파일 경로         |
| `--log-file`     | `/var/log/catv-schedule-settopbox.log` | 데몬 모드 로그 파일 경로        |

---

## 🧪 시뮬레이션 모드

실제 셋톱박스 없이 동작을 확인할 수 있는 개발/테스트 모드입니다.

```toml
[catv.control]
simulation = true
```

활성화 시:

- 실제 TLS 통신 없이 Mock 응답 생성
- `stb_request_info` 명령은 스펙에 정의된 유효한 값 중 랜덤 선택
- 랜덤 성공(`1`) / 기기 오류(`0`) / 시스템 오류(`-1`) 코드 반환
- 시스템 오류 메시지: 연결 타임아웃, TLS illegal parameter, connection reset 등 실제 에러 패턴 포함

---

## ⚙️ 주요 설정

```toml
[catv.control]
simulation                  = false
port                        = 8801
max_connections_per_second  = 300   # 기본값. 포트 고갈 방지 (TIME_WAIT 60s 기준 이론치: 20,000 ÷ 60 ≒ 333)

[catv.control.batch]
max_count      = 10000
max_size       = 52428800  # 50MB
flush_interval = "5s"

[catv.control.timeout]
connection = "5s"   # TLS 연결 수립 타임아웃
send       = "30s"  # 명령 전송 및 응답 수신 타임아웃

[catv.control.retry.total]
count = 1      # 전체 재시도 횟수 (0이면 재시도 없음)
delay = "5s"   # 재시도 간 backoff: attempt × delay (progressive)

[catv.control.retry.port_exhaustion]
count = 3      # 포트 고갈(EADDRNOTAVAIL) 시 재시도 횟수
delay = "20s"  # 재시도 간 backoff: 20s, 40s, 60s

[pool.workers]
worker_count    = 6000
job_queue_size  = 1200000  # 권장: max_connections_per_second × avg_processing_time_seconds
batch_size      = 333
enable_metrics  = false
```

### 재시도 정책

`isRetryableError()`는 다음 에러 패턴을 재시도 가능으로 판정합니다:

- `EADDRNOTAVAIL` (포트 고갈) → `retry.port_exhaustion` 설정 적용
- 그 외 네트워크 오류 → `retry.total` 설정 적용

`connection refused`는 **재시도 불가**로 판정합니다 (기기가 꺼진 경우).

### 포트 고갈 계산

각 서비스 인스턴스는 독립적인 RateLimiter를 생성합니다. 동시에 여러 K8s Job이 실행 중이라면 실제 연결 속도는 `max_connections_per_second × 활성 Job 수`가 됩니다.
