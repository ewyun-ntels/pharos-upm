# Simulator & Collector

이 프로젝트는 저장된 로그 데이터 파일을 읽어 실제 시간 흐름에 맞춰 재생(Replay)하고, 이를 Kafka, OpenSearch, Console 등으로 전송하는 시뮬레이터 및 수집기입니다.

## 1. 빌드 (Build)

프로젝트 루트 디렉토리에서 다음 명령어를 실행하여 빌드합니다.

```bash
# Windows
go build -o collector.exe ./cmd/collector

# Linux/Mac
go build -o collector ./cmd/collector
```

## 2. 실행 (Run)

설정 파일을 지정하여 실행합니다.

```bash
./collector.exe -config cfg/col-sim.yaml
```

### 주요 옵션
- `-config`: 설정 파일 경로 (필수)
- `-log`: 로그 파일 경로 (기본값: stdout)
- `-timezone`: 타임존 설정 (예: "Asia/Seoul")
- `-version`: 버전 정보 출력

## 3. 설정 (Configuration)

설정 파일(`yaml`)은 크게 `writer`와 `collector` 두 부분으로 나뉩니다.

### 예시 (`cfg/col-sim.yaml`)

```yaml
# 1. Writer 설정: 데이터를 보낼 목적지 정의
writer:
  stdout:             # Writer ID
    type: console     # 타입: console, kafka, open-search
  kaf1:
    type: kafka
    bootstrap-servers:
      - 192.168.15.99:30300
    topic: st

# 2. Collector 설정: 데이터 수집(시뮬레이션) 소스 정의
collector:
  simulator:
    - writers:        # 데이터를 보낼 Writer ID 목록
        - stdout
        - kaf1
      data-file: cfg/sim-data  # 재생할 데이터 파일 경로 (JSON)
      repeat: true             # 파일 끝 도달 시 반복 여부
      repeat-interval: 10m     # 반복 시 대기 시간
      time-field: timestamp    # 데이터 내 시간 필드명
      time-format: 2006-01-02T15:04:05Z07:00 # 시간 포맷 (Go layout)
```

## 4. 데이터 파일 형식

시뮬레이터가 읽을 데이터 파일은 **JSON Lines** (줄바꿈으로 구분된 JSON) 형식이어야 합니다.
설정된 `time-field`와 `time-format`에 맞는 시간 값이 반드시 포함되어야 합니다.

**예시:**
```json
{"timestamp": "2025-11-19T10:00:00+09:00", "level": "info", "msg": "test log 1"}
{"timestamp": "2025-11-19T10:00:05+09:00", "level": "error", "msg": "test log 2"}
```

## 5. 부록: UDP Forwarder

UDP 패킷을 받아 다른 주소로 포워딩하는 간단한 유틸리티입니다.

**빌드:**
```bash
go build -o udp-forward.exe ./cmd/udp-forward
```

**사용법:**
```bash
# Usage: udp-forward <local port> <remote address> [remote port]
./udp-forward.exe 162 192.168.1.100 162
```
