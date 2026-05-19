# STB 제어 호환성 분석

STB 장비와의 TLS 통신 프로토콜 탐지 과정 및 모델별 호환성 분석 결과를 기록합니다.

---

## 목차

- [프로토콜 확정 과정](#프로토콜-확정-과정)
  - [확정 프로토콜](#확정-프로토콜)
- [모델별 호환성 현황](#모델별-호환성-현황)
  - [정상 동작 모델](#정상-동작-모델)
  - [미확인 모델](#미확인-모델)
  - [비정상 모델](#비정상-모델)

---

## 프로토콜 확정 과정

| 시도                | 변경 내용                               | 결과                           |
| ------------------- | --------------------------------------- | ------------------------------ |
| Plain TCP           | `net.Dial` 직접 연결                    | TLS Alert 수신 → TLS 필요 확인 |
| TLS 기본값          | `tls.Dialer`, Go 기본 설정              | TLS handshake failure          |
| MinVersion 추가     | `MinVersion: tls.VersionTLS10`          | TLS handshake failure          |
| Cipher 전체 활성화  | `InsecureCipherSuites` 포함 전체 cipher | 연결 성공, 응답 없음           |
| 종단자 테스트       | `\r\n` / `\n` / 종단자 없음             | 모두 무응답                    |
| **CloseWrite 추가** | **JSON 전송 후 `tls.CloseWrite()`**     | **응답 수신 성공** ✅          |

### 확정 프로토콜

STB는 클라이언트의 **write half-close(TLS close_notify)** 를 수신한 후에만 응답을 전송한다.
종단자(`\r\n`, `\n`) 없이 JSON을 전송하고 `tls.Conn.CloseWrite()`를 호출하는 것이 올바른 방식이다.

---

## 모델별 호환성 현황

### 정상 동작 모델

| 모델명      | 비고                                             |
| ----------- | ------------------------------------------------ |
| THX-U300    |                                                  |
| BHX-HC100   |                                                  |
| UC1600      |                                                  |
| SMT-C5010   |                                                  |
| UC2600      |                                                  |
| SMT-C5012   |                                                  |
| UC2000      |                                                  |
| SX730C-CT   |                                                  |
| UC1000      |                                                  |
| SMT-C3022   | 일부 개체 connection reset (STB 일시적 오프라인) |
| SMT-C5011   | 일부 개체 i/o timeout (STB 일시적 오프라인)      |
| LSC630-8DTB |                                                  |
| GX-KD630CH  | 일부 개체 예외 (아래 별도 항목 참조)             |

### 미확인 모델

| 모델명   | 사유                         |
| -------- | ---------------------------- |
| TMA-U400 | 테스트 환경에 해당 장비 없음 |

### 비정상 모델

#### BKO-UC500 (OUI: `24:E4:CE`)

**결론: `stb_request_info` API 미지원 기종으로 판단. 코드 레벨에서 대응 불가.**

시도한 모든 방식과 결과:

| 방식                                 | 결과                                 |
| ------------------------------------ | ------------------------------------ |
| none + CloseWrite                    | `response_length=0`, 즉시 EOF (~2초) |
| `\r\n` 종단자, CloseWrite 없음       | `i/o timeout` 30초 소진              |
| `\n` 종단자, CloseWrite 없음         | `i/o timeout` 30초 소진              |
| Read-first probe (server-first 가설) | `bytes_read=0`, `i/o timeout` 2초    |

**분석**:

- TLS handshake: 성공
- JSON Write: 성공 (에러 없음)
- CloseWrite 시 → 데이터 없이 즉시 EOF 반환
- 종단자 전송 시 → 30초 완전 무응답

STB가 살아있음은 확인됨 (연결 수립 및 EOF 반환). 펌웨어에 제어 에이전트가 없거나 해당 API를 지원하지 않는 기종으로 결론.

#### GX-KD630CH 일부 개체 (`54:FA:3E:C2:0A:21`)

**결론: 해당 개체의 TLS 설정 이상. 동일 모델 다수는 정상 동작.**

- 에러: `remote error: tls: illegal parameter`
- 동일 모델(GX-KD630CH) 9개 중 1개만 발생
- STB 측 TLS 설정 또는 펌웨어 문제로 판단
