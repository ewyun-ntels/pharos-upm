# SNMP Notification Rule Package

이 패키지는 Pharos 시스템에서 SNMP Trap을 통한 알림 발송을 담당하는 핵심 모듈입니다. SNMPv2c와 SNMPv3를 모두 지원하며, 알림 이벤트를 SNMP 트랩으로 변환하여 네트워크 관리 시스템으로
전송합니다.

## 주요 기능

- **SNMPv2c 및 SNMPv3 지원**: 커뮤니티 기반 v2c와 인증/암호화가 가능한 v3 모두 지원
- **유연한 PDU 구성**: Alert의 Label을 SNMP PDU로 동적 매핑
- **다양한 데이터 타입 지원**: OctetString, Counter32/64, Gauge32, Float/Double 등
- **필터링 기능**: 특정 Label 값을 제외하여 불필요한 트랩 방지
- **추가 PDU 지원**: 고정값을 가진 추가 PDU 자동 삽입
- **강력한 유효성 검사**: 설정 오류 사전 차단 및 상세한 오류 메시지

## 아키텍처

### 핵심 구조체

#### 1. `SNMP` (메인 구조체)

```go
type SNMP struct {
Id             string  // 규칙 고유 식별자
Name           string  // 규칙 이름 (필수)
ValueOid       *string // 알림 값을 담을 OID
DescriptionOid *string // 알림 설명을 담을 OID  
StatusOid      *string // 알림 상태를 담을 OID
ExcludeFilter  map[string]interface{} // 제외할 Label 필터
TrapPduRule    []TrapPduRule          // Label → PDU 매핑 규칙
Port           uint16                 // SNMP 포트 (기본: 162)
Target         string // 대상 호스트
Version        uint8  // SNMP 버전 (1=v2c, 3=v3)
Transport      string // 전송 프로토콜 (udp/tcp, 기본: udp)
SNMPv2Config   *SNMPv2Config // v2c 설정
SNMPv3Config   *SNMPv3Config // v3 설정
AppendPdu      []AppendPdu           // 추가 고정 PDU
}
```

#### 2. `TrapPduRule` (Label-PDU 매핑)

```go
type TrapPduRule struct {
LabelName string  // Alert Label 이름
Oid       string  // 매핑할 SNMP OID
ValueType *string // SNMP 데이터 타입 (선택사항)
}
```

#### 3. `AppendPdu` (추가 PDU)

```go
type AppendPdu struct {
Oid       string // SNMP OID
ValueType string // SNMP 데이터 타입 (필수)
Value     interface{} // 고정 값
}
```

### 버전별 설정

#### SNMPv2c 설정

```go
type SNMPv2Config struct {
Community string // SNMP 커뮤니티 문자열
}
```

#### SNMPv3 설정

```go
type SNMPv3Config struct {
AuthoritativeEngineIDHex string // 엔진 ID (hex 형식)
Username                 string // 사용자명
AuthProtocol             string // 인증 프로토콜 (SHA/SHA256/MD5 등)
AuthPassphrase           string // 인증 암호
PrivProtocol             string // 암호화 프로토콜 (AES/DES 등)  
PrivPassphrase           string // 암호화 암호
}
```

## 지원하는 SNMP 데이터 타입

| 상수명                         | SNMP 타입          | 설명             |
|-----------------------------|------------------|----------------|
| `ValueTypeOctetString`      | OctetString      | 문자열 데이터        |
| `ValueTypeCounter32`        | Counter32        | 32비트 카운터       |
| `ValueTypeGauge32`          | Gauge32          | 32비트 게이지       |
| `ValueTypeCounter64`        | Counter64        | 64비트 카운터 (기본값) |
| `ValueTypeUinteger32`       | Uinteger32       | 32비트 부호없는 정수   |
| `ValueTypeOpaqueFloat`      | OpaqueFloat      | 32비트 부동소수점     |
| `ValueTypeOpaqueDouble`     | OpaqueDouble     | 64비트 부동소수점     |
| `ValueTypeObjectIdentifier` | ObjectIdentifier | OID            |

## 핵심 함수

### 1. 설정 및 초기화

#### `Load(config common.Config, data map[string]interface{}) error`

- 설정 데이터를 파싱하여 SNMP 구조체 초기화
- SNMP 연결 설정 및 커넥션 생성
- 버전별 설정 검증 및 적용

#### `Validate() error`

- 전송 프로토콜 검증 (udp/tcp만 허용)
- TrapPduRule의 중복 Label명/OID 검사
- AppendPdu 유효성 검증

### 2. 알림 전송

#### `Send(value *notification_common.AlertValue) error`

핵심 알림 전송 로직:

1. **Label 처리**: Alert의 Label을 TrapPduRule에 따라 PDU로 변환
2. **필터링**: ExcludeFilter에 해당하는 Label 제외
3. **메타데이터 추가**: Description, Value, Status OID 추가
4. **추가 PDU**: AppendPdu 목록의 고정값 PDU 추가
5. **트랩 전송**: 구성된 PDU로 SNMP Trap 전송
6. **이력 저장**: 알림 이력을 데이터베이스에 기록

### 3. 유틸리티 함수

#### `parseEngineIDHex(s string) []byte`

다양한 hex 형식을 SNMPv3 Engine ID로 변환:

- `"8000137003"` → `[]byte{0x80, 0x00, 0x13, 0x70, 0x03}`
- `"0x8000137003"` (0x 접두사 포함)
- `"80:00:13:70:03"` (콜론 구분)
- `"80 00 13 70 03"` (공백 구분)

#### `getAsnBER(name string) (g.Asn1BER, error)`

문자열 타입명을 gosnmp ASN.1 BER 타입으로 변환

#### `getAuthProtocol(protocol string) g.SnmpV3AuthProtocol`

인증 프로토콜 문자열을 gosnmp 상수로 변환:

- `"sha"`, `"sha256"`, `"sha384"`, `"sha512"`, `"md5"` 지원

#### `getPrivProtocol(protocol string) g.SnmpV3PrivProtocol`

암호화 프로토콜 문자열을 gosnmp 상수로 변환:

- `"des"`, `"aes"`, `"aes192"`, `"aes256"` 지원

#### `checkSnmpTypeMatch(t g.Asn1BER, v interface{}) (interface{}, error)`

JSON 파싱된 값을 SNMP 타입에 맞게 변환 및 검증

## 사용 예시

### SNMPv2c 설정 예시

```json
{
  "name": "network-alerts",
  "target": "192.168.1.100",
  "port": 162,
  "version": 1,
  "transport": "udp",
  "snmpv2_config": {
    "community": "public"
  },
  "value_oid": "1.3.6.1.4.1.12345.1.1",
  "description_oid": "1.3.6.1.4.1.12345.1.2",
  "status_oid": "1.3.6.1.4.1.12345.1.3",
  "trap_pdu_rule": [
    {
      "label_name": "severity",
      "oid": "1.3.6.1.4.1.12345.2.1",
      "value_type": "OctetString"
    },
    {
      "label_name": "instance",
      "oid": "1.3.6.1.4.1.12345.2.2"
    }
  ],
  "exclude_filter": {
    "internal": "true"
  },
  "append_pdu": [
    {
      "oid": "1.3.6.1.6.3.1.1.4.1.0",
      "value_type": "objectIdentifier",
      "value": "1.3.6.1.4.1.12345.0.1"
    }
  ]
}
```

### SNMPv3 설정 예시

```json
{
  "name": "secure-alerts",
  "target": "snmp-server.example.com",
  "port": 162,
  "version": 3,
  "transport": "udp",
  "snmpv3_config": {
    "authoritative_engine_id": "80:00:13:70:03",
    "username": "pharos-user",
    "auth_protocol": "sha256",
    "auth_passphrase": "auth-password-123",
    "priv_protocol": "aes",
    "priv_passphrase": "priv-password-456"
  },
  "value_oid": "1.3.6.1.4.1.9999.1.1",
  "trap_pdu_rule": [
    {
      "label_name": "alertname",
      "oid": "1.3.6.1.4.1.9999.2.1",
      "value_type": "OctetString"
    }
  ]
}
```

### 전송되는 Alert 예시

```go
alertValue := &notification_common.AlertValue{
Description: "High CPU usage detected",
Value:       85.5,
Status:      "alerting",
Labels: map[string]string{
"severity": "critical",
"instance": "web-server-01",
"internal": "true", // ExcludeFilter로 제외됨
},
}
```

## 에러 처리

### 설정 검증 에러

- `"invalid transport type"`: udp/tcp 외 프로토콜 사용 시
- `"duplicate label name"`: TrapPduRule에서 Label명 중복
- `"duplicate OID"`: TrapPduRule에서 OID 중복
- `"SNMPv2Config is nil"`: v2c 설정 누락
- `"SNMPv3Config is nil"`: v3 설정 누락

### 런타임 에러

- `"value is nil"`: 널 AlertValue 전달
- `"pdu get error"`: Label 값의 타입 변환 실패
- `"pdu append error"`: AppendPdu 값 타입 불일치
- `"snmp connect failed"`: SNMP 연결 실패

## 테스트 범위

### 단위 테스트 (`snmp_test.go`)

- **Engine ID 파싱**: 다양한 hex 형식 변환 검증
- **타입 변환**: ASN.1 BER 타입 매핑 정확성
- **설정 검증**: 잘못된 설정에 대한 에러 처리
- **PDU 생성**: Label 값의 PDU 변환 로직
- **트랩 전송**: Mock 함수를 통한 트랩 생성 검증
- **필터링**: ExcludeFilter 동작 확인
- **프로토콜 매핑**: 인증/암호화 프로토콜 변환

### 테스트 커버리지

- Engine ID hex 파싱 (다양한 형식)
- SNMP 타입 매핑 및 검증
- 설정 유효성 검사 (필수 필드, 중복 체크)
- PDU 생성 로직 (기본/명시적 타입)
- 트랩 전송 시뮬레이션
- 에러 상황별 예외 처리

## 의존성

- **gosnmp**: SNMP 프로토콜 구현 (`github.com/gosnmp/gosnmp`)
- **mapstructure**: 설정 데이터 파싱 (`github.com/go-viper/mapstructure/v2`)
- **slog**: 구조화된 로깅 (Go 표준 라이브러리)
- **pharos 내부**: `common`, `notification/common`, `notification/rule/common`

## 성능 특징

- **연결 풀링**: SNMP 연결을 재사용하여 오버헤드 최소화
- **비동기 처리**: 이력 저장 실패가 트랩 전송을 블록하지 않음
- **타입 최적화**: JSON 파싱 결과를 SNMP 타입에 맞게 효율적 변환
- **에러 복구**: 개별 PDU 생성 실패 시 전체 트랩 전송이 중단되지 않음

## 보안 고려사항

### SNMPv3 보안

- **인증**: SHA/SHA256/SHA384/SHA512/MD5 해시 알고리즘 지원
- **암호화**: AES/AES192/AES256/DES 암호화 프로토콜 지원
- **Engine ID**: Hex 형식의 고유 식별자로 보안 컨텍스트 생성

### 데이터 보호

- 인증/암호화 키는 메모리에서만 사용, 로그에 기록하지 않음
- Transport 계층에서 UDP/TCP 선택 가능
- 타임아웃 및 재시도 로직으로 DoS 공격 방지

## 모니터링 및 디버깅

### 로깅

- 구조화된 로깅 (slog) 사용
- 연결 실패, PDU 생성 오류, 이력 저장 실패 등 상세 기록
- 운영 환경에서 민감 정보 (인증키 등) 로깅하지 않음

### 메트릭스

- 트랩 전송 성공/실패 카운트
- SNMP 연결 상태 모니터링
- PDU 생성 오류 통계

이 패키지는 Pharos 시스템의 알림을 표준 SNMP 트랩으로 변환하여 기존 네트워크 관리 시스템과의 통합을 용이하게 하는 핵심 역할을 담당합니다.
