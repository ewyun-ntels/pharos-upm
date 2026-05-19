# SKB CATV STB UDP Simulator

## 개요

SKB CATV 셋톱박스(STB)의 UDP 데이터 전송을 시뮬레이션하는 도구입니다.

## 지원 인터페이스

1. **주기전송** (Periodic Transmission)
   - 전송 주기: 10분마다
   - 포트: 50000 (기본값)
   - 23개 필드

2. **일일전송** (Daily Transmission)
   - 전송 주기: 하루 1회 (자정)
   - 포트: 50001 (기본값)
   - 40개 필드

3. **품질계측전송** (Quality Measurement)
   - 전송 주기: 1분마다 (시뮬레이션)
   - 포트: 50005 (기본값, v2.8: UDP로 변경)
   - 18개 필드 + channels 배열

4. **자가진단전송** (Diagnostic)
   - 전송 주기: 1분마다 (시뮬레이션)
   - 포트: 50003 (기본값)
   - 18개 필드

5. **망품질전환전송** (Network Quality Transition)
   - 전송 주기: 1분마다 (시뮬레이션)
   - 포트: 50004 (기본값)
   - 21개 필드

## 설치 및 빌드

```bash
go build -o simulator main.go
```

## 설정 파일

`config.yml` 예제:

```yaml
clickhouse:
  database: 'default'
  host: '192.168.15.102'
  port: 30900
  username: 'default'
  password: 'default'
  max_open_connection: 10
  max_lifetime: 0

server:
  host: '127.0.0.1'

port:
  daily: 50001
  diagnostic: 50003
  network_quality_transition: 50004
  periodic: 50000
  quality_measurement: 50005

devices:
  - hostId: '1A8020DF41'
    macAddr: 'a0:72:2c:b6:ae:2b'
    cmMac: 'a0:72:2c:b6:ae:2a'
    stbIp: '10.43.81.14'
    cmIp: '10.4.58.208'
    stbModel: 'THX-U300'
    mwVer: '3.1.46'
    localVer: '1.0.6.03'
    cloudVer: '1.6.12'
  - hostId: '2B9031EF52'
    macAddr: 'b1:83:3d:c7:bf:3c'
    cmMac: 'b1:83:3d:c7:bf:3b'
    stbIp: '10.43.82.15'
    cmIp: '10.4.59.209'
    stbModel: 'THX-U400'
    mwVer: '3.2.10'
    localVer: '1.1.2.05'
    cloudVer: '1.7.5'
```

### 설정 항목 설명

- **server**: 서버 설정
  - **host**: UDP 데이터를 전송할 서버 IP 주소
- **port**: 각 인터페이스별 UDP 포트 번호
  - **periodic**: 주기전송 포트
  - **daily**: 일일전송 포트
  - **quality_measurement**: 품질계측전송 포트
  - **diagnostic**: 자가진단전송 포트
  - **network_quality_transition**: 망품질전환전송 포트
- **clickhouse**: ClickHouse 데이터베이스 설정 (시뮬레이터에서는 미사용)
- **devices**: 시뮬레이션할 STB 장치 목록 (여러 대 등록 가능)
  - **hostId**: 10자리 HOST ID
  - **macAddr**: STB MAC 주소
  - **cmMac**: CM MAC 주소
  - **stbIp**: STB IP 주소
  - **cmIp**: CM IP 주소
  - **stbModel**: STB 모델명
  - **mwVer**: 미들웨어 버전
  - **localVer**: Local UI 버전
  - **cloudVer**: Cloud UI 버전

## 실행

```bash
# 설정 파일 지정하여 실행
./simulator -config config.yml

# 또는 직접 Go로 실행
go run main.go -config config.yml
```

## 실행 예제 로그

```
2025/12/10 15:07:42 Loaded configuration with 2 devices
2025/12/10 15:07:42 Server IP: 127.0.0.1
2025/12/10 15:07:42 Interfaces: Periodic=50000, Daily=50001, QualityMeasurement=50002, Diagnostic=50003, NetworkQualityTransition=50004
2025/12/10 15:07:42 [1A8020DF41] Starting device simulator
2025/12/10 15:07:42 [2B9031EF52] Starting device simulator
2025/12/10 15:07:42 All device simulators started. Press Ctrl+C to stop.
2025/12/10 15:07:42 [1A8020DF41] Sent periodic data to 127.0.0.1:50000
2025/12/10 15:07:42 [1A8020DF41] Sent quality measurement data to 127.0.0.1:50002
2025/12/10 15:07:42 [1A8020DF41] Sent diagnostic data to 127.0.0.1:50003
2025/12/10 15:07:42 [1A8020DF41] Sent network quality transition data to 127.0.0.1:50004
```

## 전송 주기

| 인터페이스     | 실제 전송 조건            | 시뮬레이터 전송 주기 |
| -------------- | ------------------------- | -------------------- |
| 주기전송       | 10분마다                  | 10분마다             |
| 일일전송       | 하루 1회 (자정)           | 하루 1회 (자정)      |
| 품질계측전송   | Sleep 모드 진입 시        | 1분마다              |
| 자가진단전송   | 핫키(\*106OK) 입력 시     | 1분마다              |
| 망품질전환전송 | 망품질 전환 시 (QAM→8VSB) | 1분마다              |

## 데이터 필드

모든 데이터 필드는 `INTERFACE.md`와 ClickHouse SQL 스키마를 기반으로 구현되었습니다.

### 주요 필드 예제

- **공통 필드**: hostId, macAddr, cmMac, stbIp, cmIp, stbModel, mwVer, localVer, cloudVer, loggingTime
- **채널 정보**: chSid, chNum, chName, chPrg, chFreq, chMode
- **신호 품질**: pwrLvl (Power Level), snr (SNR)
- **설정 정보**: limitAge, tvLock, skipCh, resolution, audioMode 등

## 데이터 생성 방식

시뮬레이터는 다음과 같이 샘플 데이터를 생성합니다:

- **Power Level**: -12 ~ 12 dBmV 범위의 랜덤 값
- **SNR**: 30 ~ 40 dB 범위의 랜덤 값
- **채널 정보**: 고정된 샘플 채널 데이터 (EBS HD, MBC 등)
- **설정 값**: 실제 STB 설정에 맞는 샘플 값

## 참고 문서

- `INTERFACE.md`: SKB CATV 인터페이스 정의서 완전 분석
- SQL 스키마: `pkg/migrations/databases/master/statistics/clickhouse/`
  - `20251210000000_create_daily.sql`
  - `20251210000001_create_diagnostic.sql`
  - `20251210000002_create_network_quality_transition.sql`
  - `20251210000003_create_periodic.sql`
  - `20251210000004_create_quality_measurement.sql`

## 종료

`Ctrl+C`를 눌러 시뮬레이터를 종료합니다.
