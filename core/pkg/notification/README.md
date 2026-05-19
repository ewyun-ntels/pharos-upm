# Notification 서브시스템

이 패키지는 알림(경보) 시스템에서 사용하는 플러그형(Pluggable) 알림 규칙을 제공합니다. 현재 지원되는 규칙은 다음과 같습니다.

- pkg/notification/rule 아래의 SNMP 트랩 발신기(v2c 및 v3)

## 아키텍처 및 구조

### 지원하는 Notification 타입

현재 시스템에서 지원하는 Notification 타입은 다음과 같습니다:

- **Email**: 이메일 알림 (SMTP 설정 지원)
- **Slack**: Slack 채널 알림 (웹훅, 채널, 멘션 지원)  
- **Teams**: Microsoft Teams 알림
- **Webhook**: HTTP 웹훅 알림 (커스텀 헤더, 페이로드 지원)
- **SNMP**: SNMP 트랩 발신 (v2c, v3 지원)

### Adapter 계층 구조

Notification 시스템은 다계층 adapter 패턴을 통해 데이터 변환을 수행합니다:

```
API Request (shared/types) → Resource (resources/) → Model (model/) → Database
                          ←                       ←                 ←
API Response              ← Resource              ← Model           ← Database
```

#### 1. RuleAdapter (`adapters/rule_adapter.go`)
- **목적**: 내부 모델과 API 리소스 간 변환 및 JSON 처리
- **지원 타입**: Email, Slack, Teams, Webhook, SNMP 모든 타입
- **주요 메서드**:
  - `ModelToResource()`: 데이터베이스 모델 → API 리소스 변환 (JSON 파싱 포함)
  - `ResourceToModel()`: API 리소스 → 데이터베이스 모델 변환 (JSON 마샬링 포함)
  - `ModelListToResourceList()`: 배치 변환 지원

#### 2. ApiAdapter (`adapters/api_adapter.go`)
- **목적**: shared/types API와 내부 리소스 타입 간 변환
- **주요 메서드**:
  - `ToAPI()`: 내부 모델 → API 응답 리소스
  - `FromAPI()`: API 요청 리소스 → 내부 모델
  - `FromAPINotificationRule()`: shared/types API → 내부 리소스
  - `ToAPINotificationRuleResponse()`: 내부 리소스 → shared/types API 응답

### 테스트 커버리지

Notification adapter는 **100% 테스트 커버리지**를 달성했습니다:

#### 기본 Adapter 테스트
- 기본 변환 테스트 (5개)
- 통합 테스트 (2개)

#### CRUD 통합 테스트  
- **모든 Notification 타입 CRUD 테스트**: Email, Slack, Teams, Webhook (4개)
- **전체 변환 체인 검증**: API → Resource → Model → Resource → API
- **복잡한 설정 테스트**: 각 타입별 고급 설정 옵션 (3개)
- **Edge Case 테스트**: nil 처리, 빈 값, 알 수 없는 타입, 라운드트립 무결성 (4개)

**총 테스트**: 22개 모든 테스트 통과 ✅

#### 설정 예시

각 notification 타입의 복잡한 설정 예시:

**Email 설정:**
```json
{
  "recipients": ["test1@example.com", "test2@example.com"],
  "subject": "Alert Notification", 
  "template": "alert_template",
  "smtp": {
    "server": "smtp.example.com",
    "port": 587
  }
}
```

**Slack 설정:**
```json
{
  "channel": "#alerts",
  "username": "pharos-bot", 
  "icon": ":warning:",
  "webhook": "https://hooks.slack.com/services/xxx/yyy/zzz",
  "mentions": ["@channel", "@here"]
}
```

**Webhook 설정:**
```json
{
  "url": "https://api.example.com/webhooks/alerts",
  "method": "POST",
  "headers": {
    "Content-Type": "application/json",
    "Authorization": "Bearer token123"
  },
  "payload": {
    "alert_type": "system", 
    "priority": "high"
  }
}
```

## 개발 및 테스트

### Adapter 테스트 실행

notification adapter의 모든 테스트를 실행하려면:

```bash
cd core/pkg/notification/adapters
go test -v
```

**예상 결과**: 22개 테스트 모두 통과 (PASS)

개별 테스트 그룹 실행:
```bash
# CRUD 테스트만 실행
go test -v -run "TestNotificationRuleCRUD"

# 복잡한 설정 테스트만 실행  
go test -v -run "TestNotificationRuleComplexConfiguration"

# Edge case 테스트만 실행
go test -v -run "TestNotificationRuleEdgeCases"
```

### 데이터 변환 검증

시스템의 데이터 무결성을 확인하려면 다음과 같이 라운드트립 테스트를 실행할 수 있습니다:

1. **API → Resource → Model → Resource → API** 전체 체인 검증
2. **JSON 마샬링/언마샬링** 정확성 확인
3. **모든 notification 타입**에 대한 설정 보존 검증

## SNMP 버전 선택과 공통 필드

- version: 2(v2c) 또는 3(v3)을 지정합니다. 필수.
- target: 트랩을 전송할 호스트/IP. 필수.
- port: 트랩 수신 포트(기본 162). 선택.
- transport: udp(기본) 또는 tcp. 선택.
- trap_pdu_rule/append_pdu, value_oid/description_oid/status_oid 등은 공통 규칙 구성에 사용됩니다.
- 버전에 따라 사용하는 하위 설정 블록이 달라집니다.
    - v2c: snmpv2_config 사용
    - v3: snmpv3_config 사용

## SNMP v2c 구성

REST API 또는 마이그레이션으로 "snmp" 규칙을 등록할 때, v2c는 community 문자열만 있으면 됩니다.

예시:

```json
{
  "notification_type": "snmp",
  "rule": {
    "name": "snmp-v2c-sender",
    "target": "127.0.0.1",
    "port": 162,
    "version": 2,
    "transport": "udp",
    "snmpv2_config": {
      "community": "public"
    }
  }
}
```

메모:

- snmpv2_config.community는 필수입니다.
- snmpv3_config는 v2c에서 무시됩니다.
- 기본 타임아웃 1초, 재시도 3회(지수 백오프)를 사용합니다.

## SNMP v3 구성

REST API 또는 마이그레이션을 통해 "snmp" 유형의 알림 규칙을 등록할 때, SNMPv3 관련 설정 블록은 아래와 같습니다(v3에 필요한 필드만 예시):

```json
{
  "notification_type": "snmp",
  "rule": {
    "name": "snmp-sender",
    "target": "127.0.0.1",
    "port": 162,
    "version": 3,
    "transport": "udp",
    "snmpv3_config": {
      "authoritative_engine_id": "80:00:13:70:03",
      "username": "user1",
      "auth_protocol": "sha256",
      "auth_passphrase": "authpass",
      "priv_protocol": "aes",
      "priv_passphrase": "privpass"
    }
  }
}
```

### authoritative_engine_id (16진수 HEX)

SNMPv3의 authoritativeEngineID 값은 16진수(HEX) 문자열로 제공해야 합니다. 다음과 같은 형식을 허용합니다.

- 접두사 없는 순수 HEX: `8000137003`
- `0x` 접두사 포함: `0x8000137003`
- 바이트 단위 구분(콜론 또는 공백): `80:00:13:70:03` 또는 `80 00 13 70 03`

제공된 값은 파싱되어 SNMPv3에서 요구하는 원시 바이트(Octets)로 변환됩니다. 해당 필드를 비워두면 표준 SNMPv3 Discovery 절차를 통해 엔진 ID를 자동으로 학습합니다.

### 추가 메모(프로토콜 옵션)

- `auth_protocol`: md5, sha, sha224, sha256, sha384, sha512 중 하나. 비우면 기본값은 `noauth`이며 인증 없음 모드가 됩니다.
- `priv_protocol`: des, aes, aes192, aes256 중 하나. 비우면 기본값은 `nopriv`이며 암호화 없음 모드가 됩니다.
- USM(사용자 기반 보안 모델)에서 `noauth` 또는 `nopriv` 모드 사용 시 보안 요건을 만족하는지 사전 검토가 필요합니다.

## 개발(Development) 참고

- SNMP 규칙에 대한 단위 테스트는 pkg/notification/rule/snmp/snmp_test.go에 있습니다. 특히 EngineID의 HEX 파싱 로직을 검증합니다.
- SNMP 규칙은 Load() 시 설정된 타깃으로 실제 네트워크 연결을 시도할 수 있으므로, 테스트에서는 순수 함수만을 대상으로 하여 네트워크 의존성을 회피합니다.

### 예시: 특정 EngineID를 설정하고 tcpdump로 검증하기

SNMPv3 msgAuthoritativeEngineID가 네트워크 상에서 정확히 바이트 `80 00 1f 88 80 59 dc 48 61`로 보이게 하고 싶다면, `authoritative_engine_id` 값을
16진수 문자열 "80001f888059dc4861"로 설정합니다.

예시 규칙 스니펫:

```json
{
  "notification_type": "snmp",
  "rule": {
    "name": "snmp-sender",
    "target": "127.0.0.1",
    "port": 162,
    "version": 3,
    "transport": "udp",
    "snmpv3_config": {
      "authoritative_engine_id": "80001f888059dc4861",
      "username": "user1",
      "auth_protocol": "sha256",
      "auth_passphrase": "authpass",
      "priv_protocol": "aes",
      "priv_passphrase": "privpass"
    }
  }
}
```

메모:

- 본 코드는 HEX 문자열을 파싱하여 gosnmp의 `UsmSecurityParameters.AuthoritativeEngineID`에 원시 바이트로 설정합니다. 패킷 캡처 시
  msgAuthoritativeEngineID 필드가 `80 00 1f 88 80 59 dc 48 61`로 보입니다.
- 허용되는 형식: `80001f888059dc4861`, `0x80001f888059dc4861`, 또는 `80:00:1f:88:80:59:dc:48:61`와 같은 구분자 포함 형식.
- SNMPv3 EngineID는 RFC 3411/3414에 따라 일반적으로 5–32 옥텟입니다. 수신 측(트랩 리시버)이 해당 값 범위와 형식을 허용하는지 확인하세요.

## 자주 묻는 질문(FAQ) 및 트러블슈팅

- 엔진 ID를 비워두었는데 통신이 되지 않아요.
    - v3 Discovery가 상대 수신기와 정상적으로 상호작용해야 합니다. 방화벽/ACL에서 SNMP 트래픽(기본 UDP/162 또는 설정 포트)이 허용되는지 확인하세요.
- auth/priv 프로토콜을 설정했는데 인증 실패가 납니다.
    - 수신기 측 USM 사용자, 프로토콜, 패스프레이즈가 정확히 일치해야 합니다. 패스프레이즈 길이/정책(예: 최소 길이 8바이트 등)을 점검하세요.
- TCP 대신 UDP를 쓰나요?
    - 대부분의 트랩은 UDP를 사용합니다. 환경적 요구로 TCP가 필요하다면 `transport`를 `tcp`로 바꾸되, 수신기도 TCP를 수용하는지 확인하세요.
- EngineID 형식을 잘못 넣은 것 같습니다.
    - 공백/콜론 구분 형식, `0x` 접두 형식, 순수 HEX 모두 허용됩니다. 유효하지 않은 문자가 포함되지 않았는지 확인하세요.

## 검증 팁

- 로컬에서 tcpdump나 Wireshark로 트래픽을 캡처하여 msgAuthoritativeEngineID, auth/priv 설정 적용 여부를 확인할 수 있습니다.
- 구성 변경 후에는 규칙을 재적용(또는 서비스 재시작)하여 최신 설정이 반영되었는지 점검하세요.

## 알림 규칙 CRUD API

다음 엔드포인트는 알림 규칙(Notification Rule)의 생성/조회/수정/삭제를 제공합니다. 기본 prefix는 내부 라우팅에 따라 다를 수 있으나, 핸들러는
pkg/notification/handler.go의 RegisterRoutes를 통해 다음 경로를 등록합니다.

- POST /rule (Create)
- GET /rule (Read: List)
- GET /rule/:id (Read: Get by ID)
- PUT /rule/:id (Update)
- DELETE /rule/:id (Delete)

권한

- 대부분의 작업은 인증이 필요하며, 생성/수정/삭제는 SuperAdmin 또는 ManageNotification 권한이 필요합니다.

요청/응답 포맷

- 공통 Request Body 스키마(resources.Rule):

```shell
  {
    "notification_type": "snmp",
    "rule": { ... 구체 규칙 설정 ... }
  }
```

- 응답은 상황에 따라 다음과 같습니다.
    - 생성 성공(POST /rule): {"id": "<생성된 규칙 ID>"}
    - 목록 조회(GET /rule): 권한에 따라 전체 규칙 또는 축약 정보([{ notification_type, rule(간단), timestamp, updated_at }])
    - 단건 조회(GET /rule/:id): { notification_type, rule(전체 설정), timestamp, updated_at }
    - 수정 성공(PUT /rule/:id): HTTP 200 OK (바디 없음)
    - 삭제 성공(DELETE /rule/:id): HTTP 200 OK (바디 없음)

예시 cURL

1) 규칙 생성 (SNMP v2c)

```shell
curl -X POST "$HOST/notification/rule" \
 -H "Authorization: Bearer $TOKEN" \
 -H "Content-Type: application/json" \
 -d '{
   "notification_type": "snmp",
   "rule": {
     "name": "snmp-v2c-sender",
     "target": "127.0.0.1",
     "port": 162,
     "version": 2,
     "transport": "udp",
     "snmpv2_config": { "community": "public" }
   }
 }'
```

응답: {"id":"<uuid>"}

2) 규칙 목록 조회

```shell
curl -X GET "$HOST/notification/rule" -H "Authorization: Bearer $TOKEN"
```

3) 규칙 단건 조회

```shell
curl -X GET "$HOST/notification/rule/<id>" -H "Authorization: Bearer $TOKEN"
```

4) 규칙 수정 (예: v3로 변경)

```shell
curl -X PUT "$HOST/notification/rule/<id>" \
 -H "Authorization: Bearer $TOKEN" \
 -H "Content-Type: application/json" \
 -d '{
   "notification_type": "snmp",
   "rule": {
     "name": "snmp-sender",
     "target": "127.0.0.1",
     "port": 162,
     "version": 3,
     "transport": "udp",
     "snmpv3_config": {
       "authoritative_engine_id": "80001f888059dc4861",
       "username": "user1",
       "auth_protocol": "sha256",
       "auth_passphrase": "authpass",
       "priv_protocol": "aes",
       "priv_passphrase": "privpass"
     }
   }
 }'
```

응답: HTTP/1.1 200 OK

5) 규칙 삭제

```shell
curl -X DELETE "$HOST/notification/rule/<id>" -H "Authorization: Bearer $TOKEN"
```

응답: HTTP/1.1 200 OK
