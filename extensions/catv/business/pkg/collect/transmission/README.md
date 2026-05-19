# UDP Transmission Package

CATV STB로부터 UDP를 통해 다양한 유형의 데이터를 수집하는 패키지입니다.

---

## 📁 패키지 구조

```
transmission/
├── load.go          # UDP 서버 등록 (5개)
└── udp/
    ├── daily.go                       # 일일 통계
    ├── diagnostic.go                  # 자가 진단
    ├── network_quality_transition.go  # 네트워크 품질 변화 이벤트
    ├── periodic.go                    # 주기적 모니터링
    └── quality_measurement.go         # 채널별 품질 측정
```

---

## 📋 개요

`load.go`의 `Load()` 함수가 설정에서 포트를 읽어 5개의 gnet UDP 서버를 등록합니다.

모든 서버는 동일한 패턴으로 동작합니다:

```
UDP 패킷 수신
    ↓
c.Next(-1)로 버퍼 전체 읽기
    ↓
JSON Unmarshal
    ↓
collector.InstrumentUDPHandler() — Prometheus 메트릭 계측
    ↓
ClickHouse 동기 삽입
```

---

## 🖧 UDP 서버별 상세

### 1. Daily — 일일 통계

STB가 하루 1회 전송하는 일일 사용 통계 및 설정 정보입니다 (41개 필드).

주요 필드: `hostId`, `macAddr`, `cmMac`, `stbIp`, `stbModel`, `mwVer`, `limitAge`, `resolution`, `audioMode` 등

### 2. Diagnostic — 자가 진단

사용자가 리모컨으로 핫키 `*106OK`를 입력하면 STB가 자가 진단 데이터를 전송합니다.

주요 필드: `hostId`, `macAddr`, `chSid`, `chNum`, `chFreq`, `chMode`, `pwrLvl`, `snr` 등

> TCP를 통한 원격 자가 진단 제어는 별도 HTTP 핸들러(`http` 패키지)로 처리됩니다.

### 3. Network Quality Transition — 네트워크 품질 변화

채널 품질이 임계값을 넘어 변화할 때 이벤트 형태로 전송됩니다 (v2.5에서 필드명 변경됨).

### 4. Periodic — 주기적 모니터링

STB가 주기적으로 전송하는 실시간 모니터링 데이터입니다 (23개 필드).

주요 필드: `hostId`, `macAddr`, `chSid`, `chNum`, `chMode`, `pwrLvl`, `snr` 등

### 5. Quality Measurement — 채널별 품질 측정

복수 채널의 품질을 한 번에 측정하여 전송합니다.

`channels` 필드는 JSON 배열로 수신되며, DB 저장 시에는 JSON 직렬화 문자열로 변환합니다 (`ChannelsStr`).

```
수신 JSON:
  "channels": [{"chSid":"...", "chNum":"...", "pwrLvl":"...", "snr":"..."}]

DB 저장:
  channels VARCHAR  ← JSON 배열을 문자열로 직렬화
```

---

## ⚙️ 주요 설정

```toml
[catv.collect.transmission.daily]
port = "9001"

[catv.collect.transmission.diagnostic]
port = "9002"

[catv.collect.transmission.network_quality_transition]
port = "9003"

[catv.collect.transmission.periodic]
port = "9004"

[catv.collect.transmission.quality_measurement]
port = "9005"
```
