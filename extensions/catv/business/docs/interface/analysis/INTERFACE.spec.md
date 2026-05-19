# SKB CATV Log Agent 단말 인터페이스 정의서 v2.8

> **⚠️ 자동 생성된 문서입니다**  
> 이 파일은 엑셀 원본에서 자동 생성되었습니다. 직접 수정하지 마세요.  
> 수정이 필요한 경우 엑셀 파일을 업데이트한 후 `python generate_spec.py`를 실행하세요.

---

## 문서 정보

| 항목 | 내용 |
|------|------|
| 문서 버전 | v2.8 |
| 최종 수정일 | 2024년 12월 8일 |
| 생성일 | 2026년 03월 11일 11:58:40 |
| 원본 파일 | 1_COMS_SKB_CATV_Log Agent 단말 인터페이스 정의서_v2.8(251208)_해제.xlsx |

---

## 개정이력

| 버전 | 날짜 | 수정 내역 | 작성자 |
|:----:|:--------:|----------|:------:|
| v1.0 | 2021-08-05 | initial draft | Altimedia |
| v1.1 | 2021-10-22 | - 성인 컨텐츠 표시 여부 설정 (limitContents) 추가<br>   • 일일전송<br>   • 전송요청<br>   • 단말제어 | Altimedia |
| v1.2 | 2021-11-08 | - VOD 시청 시, chName 수집 데이터 추가.<br>   • chName: VOD<br>- 안드로이드의 경우, App 사용 시, chName 에 App title 전달. | Altimedia |
| v1.3 | 2021-11-12 | - 모닝 알람 설정 수집/전달 param 변경<br>- 전원 ON/OFF 추가(안드로이드 only) | Altimedia |
| v2.0 | 2025-08-12 | 품질계측전송, 자가진단전송, 망품질전환전송, 스마트리부팅 제어 항목 추가 | Altimedia, SKB |
| v2.1 | 2025-09-29 | 상용, 개발 연동 시스템 정보 수정 | SKB |
| v2.2 | 2025-09-30 | - 오탈자 수정<br>- 망품질전환전송 데이터 삭제<br>   • qamChPwrLvl<br>   • qamChSnr | Altimedia |
| v2.3 | 2025-10-22 | - 원격 요청에 의한 자가진단전송 항목 추가 | Altimedia |
| v2.4 | 2025-10-22 | - work type 수정, UDP Port 변경 | SKB |
| v2.5 | 2025-11-12 | - 망품질전환전송 key 명칭 수정 | SKB |
| v2.6 | 2025-12-01 | - STB 리셋 기능 규격 추가 | SKB |
| v2.7 | 2025-12-02 | - 스마트 리부팅, STB재시작 worktype 변경 | SKB |
| v2.8 | 2025-12-08 | - 품질계측전송 HTTP -> UDP로 변경(안정적으로 동작함) | SKB |

---

## 주기적 전송 인터페이스

**연동구분**: UDP
**연동**: 상용 : ADAMS 전송, 개발 : IP 172.18.67.25 PORT 50000
**플랫폼**: Android, OCAP

**전송조건**: 주기적 전송
(10분)

| 연동구분 | 연동 | KEY명1 | 타입 | 필수여부 | 항목 | 내용 | Default값 | 예제값 | 비고 | 검토 (Android) | 검토 (OCAP) | json 샘플 |
|------|------|------|:----:|:----:|------|------|------|------|------|------|------|------|
| 주기적 전송<br>(10분) | Request | `hostId` | string | ✓ | HOST ID | 10자리 HOST 구분 ID 정보  |  | 1A8020DF41 |  |  |  | <details><summary>JSON 샘플</summary><pre>{<br>  "hostId": "string",<br>  "macAddr": "string",<br>  "cmMac": "string",<br>  "stbIp": "string",<br>  "cmIp": "string",<br>  "stbModel": "string",<br>  "mwVer": "string",<br>  "localVer": "string",<br>  "cloudVer": "string",<br>  "loggingTime": "string",<br>  "sendingTime": "string",<br>  "chSid": "string",<br>  "chNum": "string",<br>  "chName": "string",<br>  "chPrg": "string",<br>  "chFreq": "string",<br>  "chMode": "string",<br>  "pwrLvl": "string",<br>  "snr": "string",<br>  "sigWeak": "string",<br>  "sigWeekCnt": "string"<br>  "stbState": "string"<br>  "runningTime": "string"<br>}</pre></details> |
|  |  | `macAddr` | string | ✓ | STB MAC | 12자리 표준 MAC Address 정보 |  | a0:72:2c:b6:ae:2b | STB MAC은 안드로이드는 고민 필요 |  |  |  |
|  |  | `cmMac` | string | ✓ | CM MAC | 12자리 표준 MAC Address 정보 |  | a0:72:2c:b6:ae:2a |  |  |  |  |
|  |  | `stbIp` | string | ✓ | STB IP | IP 정보 |  | 10.43.81.14 |  |  |  |  |
|  |  | `cmIp` | string | ✓ | CM IP | IP 정보 |  | 10.4.58.208 |  |  |  |  |
|  |  | `stbModel` | string | ✓ | STB 모델명 | STB 구분 모델명 |  | THX-U300 |  |  |  |  |
|  |  | `mwVer` | string | ✓ | 서비스 버전 | STB 미들웨어 버전 정보 |  | 3.1.46 |  |  |  |  |
|  |  | `localVer` | string | ✓ | Local UI 버전 | STB Local 앱 버전 정보 |  | 1.0.6.03 |  |  |  |  |
|  |  | `cloudVer` | string | ✓ | Cloud UI 버전 | 서비스 CloudUI 앱 버전 정보 |  | 1.6.12 |  |  |  |  |
|  |  | `loggingTime` | string | ✓ | 정보 수집 시간 | 정보 수집한 날짜/시간 | yyyy/mm/dd hh:mm | yyyy/mm/dd hh:mm |  |  |  |  |
|  |  | `sendingTime` | string | ✓ | 정보 전송 시간 | 수집된 정보를 전송한 날짜/시간 | yyyy/mm/dd hh:mm | yyyy/mm/dd hh:mm |  |  |  |  |
|  |  | `chSid` | string |  | 채널 Source ID | - 채널: 채널에 할당되는 Unique ID 정보 | 0 ~ 999 | - 채널: 185 |  |  |  |  |
|  |  | `chNum` | string |  | 채널 번호 | - 채널 : STB에 표시되는 채널 번호<br>- VOD : VOD 로 전달<br>- Data 방송 : DATA 로 전달 |  | - 채널: 3<br>- VOD 시청 시: VOD<br>- Data 방송 이용 시: DATA | 안드로이드는 Apps 에 표시되는 App 을 사용 시, DATA 로 전달. |  |  |  |
|  |  | `chName` | string |  | 채널명 | - 채널: 해당 채널 명<br>- VOD: VOD 로 전달<br>- Data 방송: App title (안드로이드 only) |  | - 채널: EBS HD<br>- VOD 시청 시: VOD<br>- Data 방송: Youtube | [2021.11.08]<br>- SKB 요청에 의해 VOD 시청중일 경우에 VOD string 으로 전달<br>- SKB 요청에 의해 안드로이드 단말의 경우, App 사용 시, App title 명을 전달 |  |  |  |
|  |  | `chPrg` | string |  | 프로그램 명 | - 채널 : 수집 시간에 방영되는 프로그램 명<br>- VOD : VOD title |  | 일단 해봐요 생방송 오후 1시<지긋지긋한 무릎 |  |  |  |  |
|  |  | `chFreq` | string |  | 채널 주파수 | 채널 주파수 정보 (VOD 포함) | MHz | MHz |  |  |  |  |
|  |  | `chMode` | string |  | 변조방식 | 채널 변조 방식 정보 | 8VSB/256QAM | 8VSB/256QAM |  |  |  |  |
|  |  | `pwrLvl` | string |  | Power Level | STB에서 확인되는 신호 세기 | 12 ~ -12dBmV | 12 ~ -12dBmV |  |  |  |  |
|  |  | `snr` | string |  | SNR | STB에서 확인되는 신호 품질 | 33dB 이상 | 33dB 이상 |  |  |  |  |
|  |  | `sigWeak` | string |  | 신호미약 팝업 발생 | 신호 미약 팝업이 STB에서 발생 여부 | Y/N | Y/N |  |  |  |  |
|  |  | `sigWeakCnt` | string |  | 팝업 발생 Count | 신호 미약 팝업 발생 수 |  |  |  |  |  |  |
|  |  | `stbState` | string |  | STB 전원 상태 | 시청 중 or 대기 상태 정보 확인 | watching/standby | watching | STB 전원 상태 확인 요청 |  |  |  |
|  |  | `runningTime` | string |  | STB 구동 시간 정보 | STB 부팅 후 구동 시간 정보 |  | 1day 1hour 1min 1sec |  |  |  |  |

---

## 일일 전송 인터페이스

**연동구분**: UDP
**연동**: 상용 : ADAMS 전송, 개발 : IP 172.18.67.25 PORT 50001
**플랫폼**: Andriod, OCAP

**전송조건**: 일일전송

| 연동구분 | 연동 | KEY명1 | 타입 | 필수여부 | 항목 | 내용 | Default값 | 예제값 | 비고 | 검토 (Andriod) | 검토 (OCAP) | json 샘플 |
|------|------|------|:----:|:----:|------|------|------|------|------|------|------|------|
| 일일전송 | Request | `hostId` | string | ✓ | HOST ID | 10자리 HOST 구분 ID 정보  |  | 1A8020DF41 |  |  |  | <details><summary>JSON 샘플</summary><pre>{<br>  "hostId": "string",<br>  "macAddr": "string",<br>  "cmMac": "string",<br>  "stbIp": "string",<br>  "cmIp": "string",<br>  "stbModel": "string",<br>  "mwVer": "string",<br>  "localVer": "string",<br>  "cloudVer": "string",<br>  "loggingTime": "string",<br>  "sendingTime": "string",<br>  "limitAge": "string",<br>  "tvLock": "string",<br>  "skipCh": "string",<br>  "easyBuying": "string",<br>  "favCh": "string",<br>  "zappingAd": "string",<br>  "miniEpg": "string",<br>  "miniEpgAd": "string",<br>  "tvCaption": "string",<br>  "tvImpaired": "string",<br>  "barkerCh": "string",<br>  "vodView": "string",<br>  "vodRelay": "string",<br>  "resolution": "string",<br>  "audioMode": "string",<br>  "hdmiCec": "string",<br>  "hdr": "string",<br>  "mobilePay": "string",<br>  "morningAlarm": "string",<br>  "bootMenu": "string",<br>  "pmsOn": "string",<br>  "oneAdOn": "string",<br>  "audioLang": "string"<br>  "standbyMode": "string",<br>  "savePwr": "string",<br>  "voiceGuide": "string",<br>  "runningTime":"string",<br>  "limitContents":"string"<br>}</pre></details> |
|  |  | `macAddr` | string | ✓ | STB MAC | 12자리 표준 MAC Address 정보 |  | a0:72:2c:b6:ae:2b |  |  |  |  |
|  |  | `cmMac` | string | ✓ | CM MAC | 12자리 표준 MAC Address 정보 |  | a0:72:2c:b6:ae:2a |  |  |  |  |
|  |  | `stbIp` | string | ✓ | STB IP | IP 정보 |  | 10.43.81.14 |  |  |  |  |
|  |  | `cmIp` | string | ✓ | CM IP | IP 정보 |  | 10.4.58.208 |  |  |  |  |
|  |  | `stbModel` | string | ✓ | STB 모델명 | STB 구분 모델명 |  | THX-U300 |  |  |  |  |
|  |  | `mwVer` | string | ✓ | 서비스 버전 | STB 미들웨어 버전 정보 |  | 3.1.46 |  |  |  |  |
|  |  | `localVer` | string | ✓ | Local UI 버전 | STB Local 앱 버전 정보 |  | 1.0.6.03 |  |  |  |  |
|  |  | `cloudVer` | string | ✓ | Cloud UI 버전 | 서비스 CloudUI 앱 버전 정보 |  | 1.6.12 |  |  |  |  |
|  |  | `loggingTime` | string | ✓ | 정보 수집 시간 | 정보 수집한 날짜/시간 | yyyy/mm/dd hh:mm | yyyy/mm/dd hh:mm |  |  |  |  |
|  |  | `sendingTime` | string | ✓ | 정보 전송 시간 | 수집된 정보를 전송한 날짜/시간 | yyyy/mm/dd hh:mm | yyyy/mm/dd hh:mm |  |  |  |  |
|  |  | `limitAge` | string |  | 시청 연령 제한 | 설정된 시청 등급 이상의 방송 시청 연령 제한 설정 값 | 0(사용안함), 1(7세), 2(12세), 3(15세), 4(19세) |  |  |  |  |  |
|  |  | `tvLock` | string |  | TV 잠금 | TV 잠금 설정 값 | 사용안함/바로잠금/1시간뒤/2시간뒤<br>Off (사용안함), 0(바로잠금), 1(1시간 후 잠금), 2(2시간 후 잠금) | Off | 항목  추가 | - 지원 불가(해당 기능 없음) |  |  |
|  |  | `skipCh` | string |  | 차단 채널 | 채널 이동 시 차단된 채널 설정 값 | ,(쉼표)를 구분자로한 채널번호 목록 | 11,10,15 | 차단 채널 정보 전송이 필요성 검토 | - 채널 번호의 목록 전송 가능<br>- 구분자는 쉽표(,)를 사용 | - 채널 번호의 목록 전송 가능<br>- 구분자는 쉽표(,)를 사용 |  |
|  |  | `easyBuying` | string |  | 간편 구매 | VOD 구매 시 인증번호 없이 구매 설정 값 | On(전체VOD 적용), Off(지상파 VOD만 적용), NoOpt(사용안함) | NoOpt |  |  |  |  |
|  |  | `favCh` | string |  | 선호 채널 | 선호 채널 등록 값 | ,(쉼표)를 구분자로한 채널번호 목록 | 11,10,15 | 선호채널 정보 전송이 필요성 검토 | - 채널 번호의 목록 전송 가능<br>- 쉽표(,)를 구분자로 사용 | - 채널 번호의 목록 전송 가능<br>- 쉽표(,)를 구분자로 사용 |  |
|  |  | `zappingAd` | string |  | 채널 전환 홍보 | 채널 전환 시 홍보 이미지 표시 여부 설정 값 | 1(사용), 0(사용안함) | 1 | 항목 추가 | - 지원 불가 (해당 기능 없음) |  |  |
|  |  | `miniEpg` | string |  | 채널 가이드 표시 | 채널 이동 시 표시되는 가이드 노출 시간 | 0(사용안함), 3(3초),5 (5초), 10(10초) | 3 |  |  |  |  |
|  |  | `miniEpgAd` | string |  | 채널 가이드 배너 노출 | 채널 가이드 후 표시 배너 노출 여부 | On(사용), Off(사용안함) | On |  |  |  |  |
|  |  | `tvCaption` | string |  | 자막 방송 | 실시간 TV 자막 방송 | 1(사용-디지털자막1),<br> 2(사용-디지털자막2),<br> 3(사용-디지털자막3),<br> 4(사용-디지털자막4),<br> 5(사용-디지털자막5), <br>6(사용-디지털자막6),<br>0(사용안함) |  |  |  |  |  |
|  |  | `tvImpaired` | string |  | 화면 해설 방송 | 실시간 TV 화면해설 방송 | On(사용), Off(사용안함) | Off |  |  |  |  |
|  |  | `barkerCh` | string |  | TV 시작 채널 | TV를 켰을 때 보이는 첫 채널 선택 | N(마지막 시청채널), Y(기본채널) | Y |  |  |  |  |
|  |  | `vodView` | string |  | VOD 확인 방식 | VOD 메뉴에서 콘텐츠 확인 방식 | Poster(포스터), Text(텍스트) | Poster |  |  |  |  |
|  |  | `vodRelay` | string |  | 회차 이어보기 | 시청 중인 회차가 끝나면 다음회차 재생 여부 | On(사용), Off(사용안함) | On |  |  |  |  |
|  |  | `resolution` | string |  | 화면 비율 | TV 화면 비율 설정 | 0(16:9 와이드 모드),<br> 1(4:3 중앙모드), <br>2(16:9표준모드),<br>3 (4:3 전체모드),<br>4(16:9줌모드), <br>5(4:3 시네마모드) |  |  |  |  |  |
|  |  | `audioMode` | string |  | 오디오 출력 | 오디오 출력 방식 설정 | 2(PCM), 3(Dolby AC3) | 2 |  |  | - U300, HC100 만 지원 가능 |  |
|  |  | `hdmiCec` | string |  | 전원 동기화 | HDMI-CEC 설정 값 | true(사용), false(사용안함) | false | 사용 가능한 대상 STB 만 수집 필요 | 지원 검토 요청 | <br>- U300, HC100 만 지원 가능 |  |
|  |  | `hdcp` | string |  | HDCP | HDCP 설정 값 제어 | on(사용), off(사용안함) | on | 안드로이드 STB에서 해당 기능 동작하기 위한 검토 요청 | 지원 검토 요청 |  |  |
|  |  | `hdr` | string |  | HDR | HDR 기능 설정 | 1(사용), 0(사용안함) | 1 | U300만 가능 |  | - U300 만 지원 가능 |  |
|  |  | `mobilePay` | string |  | 결제 방식 추가 | 콘텐츠 구매시 사용할 결제 방식 추가 설정 | Y(사용), N(사용안함) | N |  |  |  |  |
|  |  | `morningAlarm` | string |  | 모닝 알람 | 설정한 시간에 셋톱박스 전원 켜는 기능 설정 | repeatSetting:<br>0 - 설정안함,<br>1 - 한번만,<br>2 - 매일,<br>3 – 평일,<br>4 – 주말<br>channelSetting:sid<br>channelNum: 채널번호<br>channelName: 채널명<br>timeSetting:0000  | repeatSetting:1 ,<br>channelSetting:111,<br>channelNum: 11,<br>channelName: MBC<br>timeSetting:0700  | 기능 제공 필요성 검토  | - 지원 불가 (해당 기능 없음) |  |  |
|  |  | `bootMenu` | string |  | 홈 메뉴 노출 | TV를 켰을 때 홈 메뉴 노출 설정 | On(사용), Off(사용안함) | On |  |  |  |  |
|  |  | `pmsOn` | string |  | 실시간 혜택 정보 제공 | 실시간 혜택 정보 제공 여부 설정 | On(사용), Off(사용안함) | On |  |  |  |  |
|  |  | `oneAdOn` | string |  | 맞춤형 광고 보기 | 맞춤형 광고 보기 설정 | true(사용), false(사용안함) |  true |  |  |  |  |
|  |  | `audioLang` | string |  | 음성 언어 | 채널 기본 음성언어 설정 | kor(한국어), eng(영어), jpn(기타-일본어), chi(기타-중국어), fre(기타-프랑스어), ger(기타-독일어), spa(기타-스페인어), ara(기타-아랍어), por(기타-포르투갈어), ita(기타-이탈리아어), rus(기타-러시아어) | kor |  |  |  |  |
|  |  | `standbyMode` | string |  | 대기모드 전환 | 3시간 리모컨 입력 없으면 대기모드 전환 기능 | On(사용), Off(사용안함) | Off |  | - 지원 불가 (해당 기능 없음) |  |  |
|  |  | `savePwr` | string |  | 저전력 모드 | 대기모드에서 저전력 모드로 전환 설정 | 0(사용안함), 300000(5분), 10800000(3시간) | 300000 | 저전력 모드 : STB마다 상이함 <br>현재 메뉴에서 설정 가능한 STB은thxu300,uc2000,uc2600,sx730C 뿐임 (부분적 지원 가능) |  |  |  |
|  |  | `voiceGuide` | string |  | 음성 안내 | 시각장애인을 위한 음성 안내 설정 | 음성안내설정|음성안내속도 <br>음성안내설정 : On(사용) , Off(사용안함) <br>음성안내속도 : 1(매우느림), 2(느림), 3(기본), 4(빠름), 5(매우빠름) | On|1 | 안드로이드 단말 대상 (U400) |  |  |  |
|  |  | `runningTime` | string |  | STB 구동 시간 정보 | STB 부팅 후 구동 시간 정보 |  | 1day 1hour 1min 1sec |  |  |  |  |
|  |  | `limitContents` | string |  | 성인 콘텐츠 표시 | 자녀 안심 설정을 통한 성인 콘텐츠 노출 여부 관련 설정 | Protect(청소년 보호)<br>Hide(콘텐츠 숨김)<br>Show(콘텐츠 표시) | Hide | [2021.10.22]<br>SKB 요청에 의해 항목 추가. |  |  |  |

---

## 품질계측전송 인터페이스

**연동구분**: UDP
**연동**: 상용 : COQP 전송(신규 시스템 구축 예정), 개발 : IP 172.18.67.25 PORT 50005
**플랫폼**: Android, OCAP

**전송조건**: Sleep 모드 진입

| 연동구분 | 연동 | KEY명1 | 타입 | 필수여부 | 항목 | 내용 | Default값 | 예제값 | 비고 | 검토 (Android) | 검토 (OCAP) | json 샘플 |
|------|------|------|:----:|:----:|------|------|------|------|------|------|------|------|
| Sleep 모드 진입 | Request | `hostId` | string | ✓ | HOST ID | 10자리 HOST 구분 ID 정보  |  | 1A8020DF41 |  |  |  |  |
|  |  | `macAddr` | string | ✓ | STB MAC | 12자리 표준 MAC Address 정보 |  | a0:72:2c:b6:ae:2b |  |  |  |  |
|  |  | `cmMac` | string | ✓ | CM MAC | 12자리 표준 MAC Address 정보 |  | a0:72:2c:b6:ae:2a |  |  |  |  |
|  |  | `stbIp` | string | ✓ | STB IP | IP 정보 |  | 10.43.81.14 |  |  |  |  |
|  |  | `cmIp` | string | ✓ | CM IP | IP 정보 |  | 10.4.58.208 |  |  |  |  |
|  |  | `stbModel` | string | ✓ | STB 모델명 | STB 구분 모델명 |  | THX-U300 |  |  |  |  |
|  |  | `mwVer` | string | ✓ | 서비스 버전 | STB 미들웨어 버전 정보 |  | 3.1.46 |  |  |  |  |
|  |  | `localVer` | string | ✓ | Local UI 버전 | STB Local 앱 버전 정보 |  | 1.0.6.03 |  |  |  |  |
|  |  | `cloudVer` | string | ✓ | Cloud UI 버전 | 서비스 CloudUI 앱 버전 정보 |  | 1.6.12 |  |  |  |  |
|  |  | `loggingTime` | string | ✓ | 정보 수집 시간 | 정보 수집한 날짜/시간 | yyyy/mm/dd hh:mm | yyyy/mm/dd hh:mm |  |  |  |  |
|  |  | `channels` |  array | ✓ | 채널 품질 수집 목록 | 채널 품질 수집 목록 |  |  |  |  |  |  |
|  |  | `channel 객체 구조` |  |  |  |  |  |  |  |  |  |  |
|  |  | `chSid` |  string | ✓ | 채널 sid | 채널 sid |  | 251 |  |  |  |  |
|  |  | `chNum` |  string | ✓ | 채널 number | 채널 number |  |  |  |  |  |  |
|  |  | `chFreq` |  string | ✓ | frequency | frequency  정보 |  | 741 |  |  |  |  |
|  |  | `chMode` |  string | ✓ | modulator | modulator 정보 |  | 256QAM |  |  |  |  |
|  |  | `pwrLvl` |  string | ✓ | powerLevel | powerLevel 정보 |  |  |  |  |  |  |
|  |  | `snr` |  string | ✓ | snr | snr 정보 |  | 40 |  |  |  |  |

---

## 자가진단전송 인터페이스

**연동구분**: UDP
**연동**: 상용 : COQP 전송(신규 시스템 구축 예정), 개발 : IP 172.18.67.25 PORT 50003
**플랫폼**: Android, OCAP

**전송조건**: 핫키(*106OK) 입력 시


| 연동구분 | 연동 | KEY명1 | 타입 | 필수여부 | 항목 | 내용 | Default값 | 예제값 | 비고 | 검토 (Android) | 검토 (OCAP) | json 샘플 |
|------|------|------|:----:|:----:|------|------|------|------|------|------|------|------|
| 핫키(*106OK) 입력 시<br> | Request | `hostId` | string | ✓ | HOST ID | 10자리 HOST 구분 ID 정보  |  | 1A8020DF41 |  |  |  |  |
|  |  | `macAddr` | string | ✓ | STB MAC | 12자리 표준 MAC Address 정보 |  | a0:72:2c:b6:ae:2b |  |  |  |  |
|  |  | `cmMac` | string | ✓ | CM MAC | 12자리 표준 MAC Address 정보 |  | a0:72:2c:b6:ae:2a |  |  |  |  |
|  |  | `stbIp` | string | ✓ | STB IP | IP 정보 |  | 10.43.81.14 |  |  |  |  |
|  |  | `cmIp` | string | ✓ | CM IP | IP 정보 |  | 10.4.58.208 |  |  |  |  |
|  |  | `stbModel` | string | ✓ | STB 모델명 | STB 구분 모델명 |  | THX-U300 |  |  |  |  |
|  |  | `mwVer` | string | ✓ | 서비스 버전 | STB 미들웨어 버전 정보 |  | 3.1.46 |  |  |  |  |
|  |  | `localVer` | string | ✓ | Local UI 버전 | STB Local 앱 버전 정보 |  | 1.0.6.03 |  |  |  |  |
|  |  | `cloudVer` | string | ✓ | Cloud UI 버전 | 서비스 CloudUI 앱 버전 정보 |  | 1.6.12 |  |  |  |  |
|  |  | `loggingTime` | string | ✓ | 정보 수집 시간 | 정보 수집한 날짜/시간 | yyyy/mm/dd hh:mm | yyyy/mm/dd hh:mm |  |  |  |  |
|  |  | `chSid` |  string | ✓ | 채널 sid | 채널 sid |  | 251 |  |  |  |  |
|  |  | `chNum` |  string | ✓ | 채널 number | 채널 number |  |  |  |  |  |  |
|  |  | `chFreq` |  string | ✓ | frequency | frequency  정보 |  | 741 |  |  |  |  |
|  |  | `chMode` |  string | ✓ | modulator | modulator 정보 |  | 256QAM |  |  |  |  |
|  |  | `pwrLvl` |  string | ✓ | powerLevel | powerLevel 정보 |  |  |  |  |  |  |
|  |  | `snr` |  string | ✓ | snr | snr 정보 |  | 40 |  |  |  |  |
|  | 소켓 (TCP) | `serverIp: 셋탑아이피, port: 8801` |  |  |  |  |  |  |  |  | Android | OCAP |
| 제어 | Request | `work_type` | sysCheck | string |  | 자가진단 전송 요청 | 자가진단 내용 전송요청 |  |  |  |  |  |
|  | Response | `code` |  | string |  |  | 1: 성공, 0:실패 |  |  |  |  |  |
|  |  | `hostId` |  | string | O | HOST ID | 10자리 HOST 구분 ID 정보  |  | 1A8020DF41 |  |  |  |
|  |  | `macAddr` |  | string | O | STB MAC | 12자리 표준 MAC Address 정보 |  | a0:72:2c:b6:ae:2b |  |  |  |
|  |  | `cmMac` |  | string | O | CM MAC | 12자리 표준 MAC Address 정보 |  | a0:72:2c:b6:ae:2a |  |  |  |
|  |  | `stbIp` |  | string | O | STB IP | IP 정보 |  | 10.43.81.14 |  |  |  |
|  |  | `cmIp` |  | string | O | CM IP | IP 정보 |  | 10.4.58.208 |  |  |  |
|  |  | `stbModel` |  | string | O | STB 모델명 | STB 구분 모델명 |  | THX-U300 |  |  |  |
|  |  | `mwVer` |  | string | O | 서비스 버전 | STB 미들웨어 버전 정보 |  | 3.1.46 |  |  |  |
|  |  | `localVer` |  | string | O | Local UI 버전 | STB Local 앱 버전 정보 |  | 1.0.6.03 |  |  |  |
|  |  | `cloudVer` |  | string | O | Cloud UI 버전 | 서비스 CloudUI 앱 버전 정보 |  | 1.6.12 |  |  |  |
|  |  | `loggingTime` |  | string | O | 정보 수집 시간 | 정보 수집한 날짜/시간 | yyyy/mm/dd hh:mm | yyyy/mm/dd hh:mm |  |  |  |
|  |  | `chSid` |  |  string | O | 채널 sid | 채널 sid |  | 251 |  |  |  |
|  |  | `chNum` |  |  string | O | 채널 number | 채널 number |  |  |  |  |  |
|  |  | `chFreq` |  |  string | O | frequency | frequency  정보 |  | 741 |  |  |  |
|  |  | `chMode` |  |  string | O | modulator | modulator 정보 |  | 256QAM |  |  |  |
|  |  | `pwrLvl` |  |  string | O | powerLevel | powerLevel 정보 |  |  |  |  |  |
|  |  | `snr` |  |  string | O | snr | snr 정보 |  | 40 |  |  |  |

---

## 망품질전환전송 인터페이스

**연동구분**: UDP
**연동**: 상용 : COQP 전송(신규 시스템 구축 예정), 개발 : IP 172.18.67.25 PORT 50004
**플랫폼**: Android, OCAP

**전송조건**: 망품질 전환 시
(QAM > 8VSB)

| 연동구분 | 연동 | KEY명1 | 타입 | 필수여부 | 항목 | 내용 | Default값 | 예제값 | 비고 | 검토 (Android) | 검토 (OCAP) | json 샘플 |
|------|------|------|:----:|:----:|------|------|------|------|------|------|------|------|
| 망품질 전환 시<br>(QAM > 8VSB) | Request | `hostId` | string | ✓ | HOST ID | 10자리 HOST 구분 ID 정보  |  | 1A8020DF41 |  |  |  |  |
|  |  | `macAddr` | string | ✓ | STB MAC | 12자리 표준 MAC Address 정보 |  | a0:72:2c:b6:ae:2b |  |  |  |  |
|  |  | `cmMac` | string | ✓ | CM MAC | 12자리 표준 MAC Address 정보 |  | a0:72:2c:b6:ae:2a |  |  |  |  |
|  |  | `stbIp` | string | ✓ | STB IP | IP 정보 |  | 10.43.81.14 |  |  |  |  |
|  |  | `cmIp` | string | ✓ | CM IP | IP 정보 |  | 10.4.58.208 |  |  |  |  |
|  |  | `stbModel` | string | ✓ | STB 모델명 | STB 구분 모델명 |  | THX-U300 |  |  |  |  |
|  |  | `mwVer` | string | ✓ | 서비스 버전 | STB 미들웨어 버전 정보 |  | 3.1.46 |  |  |  |  |
|  |  | `localVer` | string | ✓ | Local UI 버전 | STB Local 앱 버전 정보 |  | 1.0.6.03 |  |  |  |  |
|  |  | `cloudVer` | string | ✓ | Cloud UI 버전 | 서비스 CloudUI 앱 버전 정보 |  | 1.6.12 |  |  |  |  |
|  |  | `loggingTime` | string | ✓ | 정보 수집 시간 | 정보 수집한 날짜/시간 | yyyy/mm/dd hh:mm | yyyy/mm/dd hh:mm |  |  |  |  |
|  |  | `chSid` |  string | ✓ | 채널 sid | 채널 sid |  | 251 |  |  |  |  |
|  |  | `chNum` |  string | ✓ | 채널 number | 채널 number |  |  |  |  |  |  |
|  |  | `chName` | string |  | 채널명 | - 채널: 해당 채널 명<br>- VOD: VOD 로 전달<br>- Data 방송: App title (안드로이드 only) |  | - 채널: EBS HD<br>- VOD 시청 시: VOD<br>- Data 방송: Youtube | [2021.11.08]<br>- SKB 요청에 의해 VOD 시청중일 경우에 VOD string 으로 전달<br>- SKB 요청에 의해 안드로이드 단말의 경우, App 사용 시, App title 명을 전달 |  |  |  |
|  |  | `chQamFreq` |  string | ✓ | frequency | frequency  정보 |  | 741 |  |  |  |  |
|  |  | `chQamMode` |  string | ✓ | modulator | modulator 정보 |  | 256QAM |  |  |  |  |
|  |  | `ch8vsbFreq` |  string | ✓ | frequency | frequency  정보 |  | 240 |  |  |  |  |
|  |  | `ch8vsbMode` |  string | ✓ | modulator | modulator 정보 |  | 8vsb |  |  |  |  |
|  |  | `ch8vsbPwrLvl` |  string | ✓ | powerLevel | powerLevel 정보 |  |  |  |  |  |  |
|  |  | `ch8vsbSnr` |  string | ✓ | snr | snr 정보 |  | 40 |  |  |  |  |

---

## 전송 요청에 의한 정보 전달 인터페이스

**연동구분**: 소켓 (TCP)
**연동**: serverIp: 셋탑아이피, port: 8801
**플랫폼**: Android, OCAP

| 연동구분 | 연동 | KEY명1 | KEY명2 | 타입 | 필수여부 | 항목 | 내용 | Default값 | 예제값 | 비고 | 검토 (Android) | 검토 (OCAP) | json 샘플 |
|------|------|------|------|:----:|:----:|------|------|------|------|------|------|------|------|
| 정보요청 | request | `work_type` | `stb_request_info` | string | ✓ |  |  |  |  |  |  |  | <details><summary>JSON 샘플</summary><pre>{ 'work_type': 'stb_request_info' }</pre></details> |
|  | Response |  | `hostId` | string | ✓ | HOST ID | 10자리 HOST 구분 ID 정보  |  | 1A8020DF41 |  |  |  | <details><summary>JSON 샘플</summary><pre>{<br>  "hostId": "string",<br>  "macAddr": "string",<br>  "cmMac": "string",<br>  "stbIp": "string",<br>  "cmIp": "string",<br>  "stbModel": "string",<br>  "mwVer": "string",<br>  "localVer": "string",<br>  "cloudVer": "string",<br>  "loggingTime": "string",<br>  "sendingTime": "string",<br>  "limitAge": "string",<br>  "tvLock": "string",<br>  "skipCh": "string",<br>  "easyBuying": "string",<br>  "favCh": "string",<br>  "zappingAd": "string",<br>  "miniEpg": "string",<br>  "miniEpgAd": "string",<br>  "tvCaption": "string",<br>  "tvImpaired": "string",<br>  "barkerCh": "string",<br>  "vodView": "string",<br>  "vodRelay": "string",<br>  "resolution": "string",<br>  "audioMode": "string",<br>  "hdmiCec": "string",<br>  "hdr": "string",<br>  "mobilePay": "string",<br>  "morningAlarm": "string",<br>  "bootMenu": "string",<br>  "pmsOn": "string",<br>  "oneAdOn": "string",<br>  "service_env_scnratio": "string",<br>  "service_env_audio": "string",<br>  "service_env_hdr": "string",<br>  "audioLang": "string",<br>  "standbyMode": "string",<br>  "savePwr": "string",<br>  "voiceGuide": "string",<br>  "chNum": "string",<br>  "volume": "string",<br>  "homeState": "string",<br>  "stbState": "string",<br>  "runningTime": "string",<br>  "limitContents":"string"<br>}</pre></details> |
|  |  |  | `macAddr` | string | ✓ | STB MAC | 12자리 표준 MAC Address 정보 |  | a0:72:2c:b6:ae:2b |  |  |  |  |
|  |  |  | `cmMac` | string | ✓ | CM MAC | 12자리 표준 MAC Address 정보 |  | a0:72:2c:b6:ae:2a |  |  |  |  |
|  |  |  | `stbIp` | string | ✓ | STB IP | IP 정보 |  | 10.43.81.14 |  |  |  |  |
|  |  |  | `cmIp` | string | ✓ | CM IP | IP 정보 |  | 10.4.58.208 |  |  |  |  |
|  |  |  | `stbModel` | string | ✓ | STB 모델명 | STB 구분 모델명 |  | THX-U300 |  |  |  |  |
|  |  |  | `mwVer` | string | ✓ | 서비스 버전 | STB 미들웨어 버전 정보 |  | 3.1.46 |  |  |  |  |
|  |  |  | `localVer` | string | ✓ | Local UI 버전 | STB Local 앱 버전 정보 |  | 1.0.6.03 |  |  |  |  |
|  |  |  | `cloudVer` | string | ✓ | Cloud UI 버전 | 서비스 CloudUI 앱 버전 정보 |  | 1.6.12 |  |  |  |  |
|  |  |  | `loggingTime` | string | ✓ | 정보 수집 시간 | 정보 수집한 날짜/시간 | yyyy/mm/dd hh:mm | yyyy/mm/dd hh:mm |  |  |  |  |
|  |  |  | `sendingTime` | string | ✓ | 정보 전송 시간 | 수집된 정보를 전송한 날짜/시간 | yyyy/mm/dd hh:mm | yyyy/mm/dd hh:mm |  |  |  |  |
|  |  |  | `limitAge` | string |  | 시청 연령 제한 | 설정된 시청 등급 이상의 방송 시청 연령 제한 | 0(사용안함), 1(7세), 2(12세), 3(15세), 4(19세) |  |  |  |  |  |
|  |  |  | `tvLock` | string |  | TV 잠금 | TV 잠금 설정 값 | 사용안함/바로잠금/1시간뒤/2시간뒤<br>Off (사용안함), 0(바로잠금), 1(1시간 후 잠금), 2(2시간 후 잠금)  | Off | 항목  추가 | - 지원 불가 (해당 기능 없음) |  |  |
|  |  |  | `skipCh` | string |  | 차단 채널 | 채널 이동 시 차단된 채널 설정 값 | ,(쉼표)를 구분자로한 채널번호 목록 | 11,10,15 | 차단 채널 정보 전송이 필요성 검토 | - 채널 번호의 목록 전송 가능<br>- 구분자는 쉽표(,)를 사용 | - 채널 번호의 목록 전송 가능<br>- 구분자는 쉽표(,)를 사용 |  |
|  |  |  | `easyBuying` | string |  | 간편 구매 | VOD 구매 시 인증번호 없이 구매 | On(전체VOD 적용), Off(지상파 VOD만 적용), NoOpt(사용안함) | NoOpt |  |  |  |  |
|  |  |  | `favCh` | string |  | 선호 채널 | 선호 채널 등록  | ,(쉼표)를 구분자로한 채널번호 목록 | 11,10,15 | 선호채널 정보 전송이 필요성 검토 | - 채널 번호의 목록 전송 가능<br>- 구분자는 쉽표(,)를 사용 | - 채널 번호의 목록 전송 가능<br>- 구분자는 쉽표(,)를 사용 |  |
|  |  |  | `zappingAd` | string |  | 채널 전환 홍보 | 채널 전환 시 홍보 이미지 표시 여부 설정 값 | 1(사용), 0(사용안함) | 1 | 항목 추가 | - 지원 불가 (해당 기능 없음) |  |  |
|  |  |  | `miniEpg` | string |  | 채널 가이드 표시 | 채널 이동 시 표시되는 가이드 노출 시간 | 0(사용안함), 3(3초),5 (5초), 10(10초) | 3 |  |  |  |  |
|  |  |  | `miniEpgAd` | string |  | 채널 가이드 배너 노출 | 채널 가이드 후 표시 배너 노출 여부 | On(사용), Off(사용안함) | On |  |  |  |  |
|  |  |  | `tvCaption` | string |  | 자막 방송 | 실시간 TV 자막 방송 | 1(사용-디지털자막1),<br> 2(사용-디지털자막2),<br> 3(사용-디지털자막3),<br> 4(사용-디지털자막4),<br> 5(사용-디지털자막5), <br>6(사용-디지털자막6),<br>0(사용안함) |  |  |  |  |  |
|  |  |  | `tvImpaired` | string |  | 화면 해설 방송 | 실시간 TV 화면해설 방송 | On(사용), Off(사용안함) | Off |  |  |  |  |
|  |  |  | `barkerCh` | string |  | TV 시작 채널 | TV를 켰을 때 보이는 첫 채널 선택 | N(마지막 시청채널), Y(기본채널) | Y |  |  |  |  |
|  |  |  | `vodView` | string |  | VOD 확인 방식 | VOD 메뉴에서 콘텐츠 확인 방식 | Poster(포스터), Text(텍스트) | Poster |  |  |  |  |
|  |  |  | `vodRelay` | string |  | 회차 이어보기 | 시청 중인 회차가 끝나면 다음회차 재생 여부 | On(사용), Off(사용안함) | Off |  |  |  |  |
|  |  |  | `resolution` | string |  | 화면 비율 | TV 화면 비율 설정 | 0(16:9 와이드 모드),<br> 1(4:3 중앙모드), <br>2(16:9표준모드),<br>3 (4:3 전체모드),<br>4(16:9줌모드), <br>5(4:3 시네마모드) |  |  |  |  |  |
|  |  |  | `audioMode` | string |  | 오디오 출력 | 오디오 출력 방식 설정 | 2(PCM), 3(Dolby AC3) | 2 |  |  | - U300, HC100 만 지원 |  |
|  |  |  | `hdmiCec` | string |  | 전원 동기화 | HDMI-CEC 설정 값 | true(사용), false(사용안함) | false  | 안드로이드 STB에서 해당 기능 동작하기 위한 검토 요청 |  | - U300, HC100 만 지원 |  |
|  |  |  | `hdcp` | String |  | HDCP | HDCP 설정 값 제어 | on(사용), off(사용안함) | on | 안드로이드 STB에서 해당 기능 동작하기 위한 검토 요청 |  |  |  |
|  |  |  | `hdr` | string |  | HDR | HDR 기능 설정 | 1(사용), 0(사용안함) | 1 | U300만 가능 |  | - U300 만 지원 |  |
|  |  |  | `mobilePay` | string |  | 결제 방식 추가 | 콘텐츠 구매시 사용할 결제 방식 추가 설정 | Y(사용), N(사용안함) | N |  |  |  |  |
|  |  |  | `morningAlarm` | string |  | 모닝 알람 | 설정한 시간에 셋톱박스 전원 켜는 기능 설정 | repeatSetting:<br>0 - 설정안함,<br>1 - 한번만,<br>2 - 매일,<br>3 – 평일,<br>4 – 주말<br>channelSetting:sid<br>channelNum: 채널번호<br>channelName: 채널명<br>timeSetting:0000  | repeatSetting:1 ,<br>channelSetting:111,<br>channelNum: 11,<br>channelName: MBC<br>timeSetting:0700  | 기능 제공 필요성 검토  | - 지원 불가 (해당 기능 없음) |  |  |
|  |  |  | `bootMenu` | string |  | 홈 메뉴 노출 | TV를 켰을 때 홈 메뉴 노출 설정 | On(사용), Off(사용안함) | On |  |  |  |  |
|  |  |  | `pmsOn` | string |  | 실시간 혜택 정보 제공 | 실시간 혜택 정보 제공 여부 설정 | On(사용), Off(사용안함) | On |  |  |  |  |
|  |  |  | `oneAdOn` | string |  | 맞춤형 광고 보기 | 맞춤형 광고 보기 설정 | true(사용), false(사용안함) |  true |  |  |  |  |
|  |  |  | `audioLang` | string |  | 음성 언어 | 채널 기본 음성언어 설정 | kor(한국어), eng(영어), jpn(기타-일본어), chi(기타-중국어), fre(기타-프랑스어), ger(기타-독일어), spa(기타-스페인어), ara(기타-아랍어), por(기타-포르투갈어), ita(기타-이탈리아어), rus(기타-러시아어) | kor |  |  |  |  |
|  |  |  | `standbyMode` | string |  | 대기모드 전환 | 3시간 리모컨 입력 없으면 대기모드 전환 기능 | On(사용), Off(사용안함) | Off |  | - 지원 불가 (해당 기능 없음) |  |  |
|  |  |  | `savePwr` | string |  | 저전력 모드 | 대기모드에서 저전력 모드로 전환 설정 | 0(사용안함), 300000(5분), 10800000(3시간) | 300000 | 저전력 모드 : STB마다 상이함 <br>현재 메뉴에서 설정 가능한 STB은thxu300,uc2000,uc2600,sx730C 뿐임 (부분적 지원 가능) |  |  |  |
|  |  |  | `voiceGuide` | string |  | 음성 안내 | 시각장애인을 위한 음성 안내 설정 | 음성안내설정|음성안내속도 <br>음성안내설정 : 0(사용) , 1(사용안함) <br>음성안내속도 : 0(매우느림), 1(느림), 2(기본), 3(빠름), 4(매우빠름) | 1|1 | 안드로이드 단말 대상 (U400) |  | - 지원 불가 (해당 기능 없음) |  |
|  |  |  | `chNum` | string |  | 현재채널 | 현재 채널 번호 | - 채널 : 현재채널번호<br>- VOD : VOD 로 전달<br>- Data 방송 : DATA 로 전달 | - 채널: 15<br>- VOD 시청 시: VOD<br>- Data 방송 이용 시: DATA | 안드로이드는 Apps 에 표시되는 App 을 사용 시, DATA 로 전달. |  |  |  |
|  |  |  | `chSid` | string |  | 채널 Source ID | - 채널: 채널에 할당되는 Unique ID 정보 | 0 ~ 999 | - 채널: 185 |  |  |  |  |
|  |  |  | `chName` | string |  | 채널명 | - 채널: 해당 채널 명<br>- VOD: VOD 로 전달<br>- Data 방송: App title (안드로이드 only) |  | - 채널: EBS HD<br>- VOD 시청 시: VOD<br>- Data 방송: Youtube | [2021.11.08]<br>- SKB 요청에 의해 VOD 시청중일 경우에 VOD string 으로 전달<br>- SKB 요청에 의해 안드로이드 단말의 경우, App 사용 시, App title 명을 전달 |  |  |  |
|  |  |  | `chPrg` | string |  | 프로그램 명 | - 채널 : 수집 시간에 방영되는 프로그램 명<br>- VOD : VOD title |  | 일단 해봐요 생방송 오후 1시<지긋지긋한 무릎 |  |  |  |  |
|  |  |  | `chFreq` | string |  | 채널 주파수 | 채널 주파수 정보 (VOD 포함) | MHz | MHz |  |  |  |  |
|  |  |  | `chMode` | string |  | 변조방식 | 채널 변조 방식 정보 | 8VSB/256QAM | 8VSB/256QAM |  |  |  |  |
|  |  |  | `pwrLvl` | string |  | Power Level | STB에서 확인되는 신호 세기 | 12 ~ -12dBmV | 12 ~ -12dBmV |  |  |  |  |
|  |  |  | `snr` | string |  | SNR | STB에서 확인되는 신호 품질 | 33dB 이상 | 33dB 이상 |  |  |  |  |
|  |  |  | `sigWeak` | string |  | 신호미약 팝업 발생 | 신호 미약 팝업이 STB에서 발생 여부 | Y/N | Y/N |  |  |  |  |
|  |  |  | `sigWeakCnt` | string |  | 팝업 발생 Count | 신호 미약 팝업 발생 수 |  |  |  |  |  |  |
|  |  |  | `volume` | string |  | 현재볼륨 | 현재볼륨 | 00 ~ 20 | 10 | 현재 볼륨 요청 |  |  |  |
|  |  |  | `homeState` | string |  | 홈메뉴 노출 여부 | 홈메뉴 노출 여부 | show/hide | hide | 홈 메뉴 실행 상태 요청 |  |  |  |
|  |  |  | `stbState` | string |  | STB 전원 상태 | 시청 중 or 대기 상태 정보 확인 | watching/standby | watching | STB 전원 상태 확인 요청 |  |  |  |
|  |  |  | `runningTime` | string |  | STB 구동 시간 정보 | STB 부팅 후 구동 시간 정보 |  | 1day 1hour 1min 1sec |  |  |  |  |
|  |  |  | `limitContents` | string |  | 성인 콘텐츠 표시 | 자녀 안심 설정을 통한 성인 콘텐츠 노출 여부 관련 설정 | Protect(청소년 보호)<br>Hide(콘텐츠 숨김)<br>Show(콘텐츠 표시) | Hide | [2021.10.22]<br>SKB 요청에 의해 항목 추가. |  |  |  |

---

## 단말 제어 인터페이스

**연동구분**: 소켓 (TCP)
**연동**: serverIp: 셋탑아이피, port: 8801
**플랫폼**: Android, OCAP

| 연동구분 | 연동 | KEY명1 | KEY명2 | 타입 | 필수여부 | 항목 | 내용 | Default값 | 예제값 | 비고 | 검토 (Android) | 검토 (OCAP) | json 샘플 |
|------|------|------|------|:----:|:----:|------|------|------|------|------|------|------|------|
| 제어 | Request | `work_type` | `limitAge` | string |  | 시청 연령 제한 | 설정된 시청 등급 이상의 방송 시청 연령 제한 | 0(사용안함), 1(7세), 2(12세), 3(15세), 4(19세) |  |  |  |  | <details><summary>JSON 샘플</summary><pre>{<br>  "work_type" : "limitAge",<br>  "work_value" : "4"<br>}</pre></details> |
|  |  |  | `tvLock` | string |  | TV 잠금 | TV 잠금 설정 값 | 사용안함/바로잠금/1시간뒤/2시간뒤<br>Off (사용안함), 0(바로잠금), 1(1시간 후 잠금), 2(2시간 후 잠금)  | Off | 항목  추가 | - 지원 불가 (해당 기능 없음) |  |  |
|  |  |  | `skipCh` | string |  | 차단 채널 | 차단채널 목록 초기화 | N/A |  | 차단 채널 정보 전송이 필요성 검토 | - 차단 채널 목록 설정 초기화 기능 추가 가능<br>- 특정 채널만 삭제 기능은 불가 | - 차단 채널 목록 설정 초기화 기능 추가 가능<br>- 특정 채널만 삭제 기능은 불가 |  |
|  |  |  | `easyBuying` | string |  | 간편 구매 | VOD 구매 시 인증번호 없이 구매 | On(전체VOD 적용), Off(지상파 VOD만 적용), NoOpt(사용안함) | NoOpt |  |  |  |  |
|  |  |  | `resetPin` | string |  | 인증 번호 | 제한/구매 인증번호 설정 초기화 | 제한인증번호, 구매인증번호 |  | 초기화 기능만 제공 |  |  |  |
|  |  |  | `favCh` | string |  | 선호 채널 | 선호채널 목록 초기화 | N/A |  | 선호채널 정보 전송이 필요성 검토 | - 선호 채널 목록 설정 초기화 기능은 추가 가능<br>- 특정 채널만 삭제 기능은 불가 | - 선호 채널 목록 설정 초기화 기능은 추가 가능<br>- 특정 채널만 삭제 기능은 불가 |  |
|  |  |  | `zappingAd` | string |  | 채널 전환 홍보 | 채널 전환 시 홍보 이미지 표시 여부 설정 값 | 1(사용), 0(사용안함) | 1 | 항목 추가 | - 지원 불가 (해당 기능 없음) |  |  |
|  |  |  | `miniEpg` | string |  | 채널 가이드 표시 | 채널 이동 시 표시되는 가이드 노출 시간 | 0(사용안함), 3(3초),5 (5초), 10(10초) | 3 |  |  |  |  |
|  |  |  | `miniEpgAd` | string |  | 채널 가이드 배너 노출 | 채널 가이드 후 표시 배너 노출 여부 | On(사용), Off(사용안함) | On |  |  |  |  |
|  |  |  | `tvCaption` | string |  | 자막 방송 | 실시간 TV 자막 방송 | 1(사용-디지털자막1),<br> 2(사용-디지털자막2),<br> 3(사용-디지털자막3),<br> 4(사용-디지털자막4),<br> 5(사용-디지털자막5), <br>6(사용-디지털자막6),<br>0(사용안함) |  |  |  |  |  |
|  |  |  | `tvImpaired` | string |  | 화면 해설 방송 | 실시간 TV 화면해설 방송 | On(사용), Off(사용안함) | Off |  |  |  |  |
|  |  |  | `barkerCh` | string |  | TV 시작 채널 | TV를 켰을 때 보이는 첫 채널 선택 | N(마지막 시청채널), Y(기본채널) | Y |  |  |  |  |
|  |  |  | `vodView` | string |  | VOD 확인 방식 | VOD 메뉴에서 콘텐츠 확인 방식 | Poster(포스터), Text(텍스트) | Poster |  |  |  |  |
|  |  |  | `vodRelay` | string |  | 회차 이어보기 | 시청 중인 회차가 끝나면 다음회차 재생 여부 | On(사용), Off(사용안함) | Off |  |  |  |  |
|  |  |  | `resolution` | string |  | 화면 비율 | TV 화면 비율 설정 | 0(16:9 와이드 모드),<br> 1(4:3 중앙모드), <br>2(16:9표준모드),<br>3 (4:3 전체모드),<br>4(16:9줌모드), <br>5(4:3 시네마모드) |  |  |  |  |  |
|  |  |  | `audioMode` | string |  | 오디오 출력 | 오디오 출력 방식 설정 | 2(PCM), 3(Dolby AC3) | 2 |  |  | - U300, HC-100 만 지원 |  |
|  |  |  | `hdmiCec` | string |  | 전원 동기화 | HDMI-CEC 설정 값 | true(사용), false(사용안함) | false  | 안드로이드 STB에서 해당 기능 동작하기 위한 검토 요청 |  | - U300, HC-100 만 지원 |  |
|  |  |  | `hdcp` |  |  | HDCP | HDCP 설정 값 제어 | on(사용),off(사용안함) | on | 안드로이드 STB에서 해당 기능 동작하기 위한 검토 요청 |  |  |  |
|  |  |  | `hdr` | string |  | HDR | HDR 기능 설정 | 1(사용), 0(사용안함) | 1 | U300만 가능 |  | - U300 만 지원 |  |
|  |  |  | `mobilePay` | string |  | 결제 방식 추가 | 콘텐츠 구매시 사용할 결제 방식 추가 설정 | Y(사용), N(사용안함) | N |  |  |  |  |
|  |  |  | `morningAlarm` | string |  | 모닝 알람 | 설정한 시간에 셋톱박스 전원 켜는 기능 설정 | 0 (설정 안함) |  | 기능 제공 필요성 검토  | - 지원 불가 (해당 기능 없음) | - 모닝 알람을 사용 안함 설정으로 변경 기능만 제공 |  |
|  |  |  | `bootMenu` | string |  | 홈 메뉴 노출 | TV를 켰을 때 홈 메뉴 노출 설정 | On(사용), Off(사용안함) | On |  |  |  |  |
|  |  |  | `pmsOn` | string |  | 실시간 혜택 정보 제공 | 실시간 혜택 정보 제공 여부 설정 | On(사용), Off(사용안함) | On |  |  |  |  |
|  |  |  | `oneAdOn` | string |  | 맞춤형 광고 보기 | 맞춤형 광고 보기 설정 | true(사용), false(사용안함) |  true |  |  |  |  |
|  |  |  | `audioLang` | string |  | 음성 언어 | 채널 기본 음성언어 설정 | kor(한국어), eng(영어), jpn(기타-일본어), chi(기타-중국어), fre(기타-프랑스어), ger(기타-독일어), spa(기타-스페인어), ara(기타-아랍어), por(기타-포르투갈어), ita(기타-이탈리아어), rus(기타-러시아어) | kor |  |  |  |  |
|  |  |  | `standbyMode` | string |  | 대기모드 전환 | 3시간 리모컨 입력 없으면 대기모드 전환 기능 | On(사용), Off(사용안함) | Off |  | - 지원 불가 (해당 기능 없음) |  |  |
|  |  |  | `savePwr` | string |  | 저전력 모드 | 대기모드에서 저전력 모드로 전환 설정 | 0(사용안함), 300000(5분), 10800000(3시간) | 300000 | 저전력 모드 : STB마다 상이함 <br>현재 메뉴에서 설정 가능한 STB은thxu300,uc2000,uc2600,sx730C 뿐임 (부분적 지원 가능) |  |  |  |
|  |  |  | `voiceGuide` | string |  | 음성 안내 | 시각장애인을 위한 음성 안내 설정 | 음성안내설정|음성안내속도 <br>음성안내설정 : On(사용) , Off(사용안함) <br>음성안내속도 : 1(매우느림), 2(느림), 3(기본), 4(빠름), 5(매우빠름) | On|1 | 안드로이드 단말 대상 (U400) |  | - 지원 불가 (해당 기능 없음) |  |
|  |  |  | `chUpDown` | string |  | 채널 변경 | 채널 +, - 통한 채널 이동 | up(채널업),<br>down(채널다운) | up | 현재 채널에서 채널 업다운 키를 통해 채널 변경<br> |  |  |  |
|  |  |  | `chDca` | string |  | 채널 이동 | 채널 번호를 통한 채널 이동 | 채널번호 | 11 | 채널 번호 이동을 통한 채널 변경 |  |  |  |
|  |  |  | `volume` | string |  | 볼륨 변경 | STB 볼륨 조절 | 0~20 사이의 수 | 5 | 볼륨 업다운 키를 통해 볼륨 조절 | - 시스템 UI 를 사용 하는 관계로 제외 필요 |  |  |
|  |  |  | `showMenu` | string |  | 홈 메뉴 실행 | 홈 메뉴 실행 | N/A |  | 홈 메뉴 실행 |  |  |  |
|  |  |  | `limitContents` | string |  | 성인 콘텐츠 표시 | 자녀 안심 설정을 통한 성인 콘텐츠 노출 여부 관련 설정 | Protect(청소년 보호)<br>Hide(콘텐츠 숨김)<br>Show(콘텐츠 표시) | Hide | [2021.10.22]<br>SKB 요청에 의해 항목 추가. |  |  |  |
|  |  |  | `stbPower` | string |  | 전원 ON/OFF | 전원 상태 변경 (시청 or 대기 상태) | N/A |  | [2021.11.12]<br>SKB 요청에 의해 안드로이드만 적용.<br><br>기본적으로 Local App에서 변경은 가능하나 전원키 입력시 제조사에서 스스로 작업하는 일이 있어서 제조사 지원없이 작업시 side effect 존재 가능성 있음 (상태가 꼬이는 등)<br>- 구현 후 상황 보고 결정 필요<br><br>[추가] standby 모드에서 제어를 위해 socket을 열어두게 될 경우 고려해야 할 사항이 많고, 저전력 케이스 등까지 고려했을 때 가능하면 기존 mib을 사용하는 방향으로 했으면 함 | [2021.11.12]<br>리모컨 전원키에 해당하는 동작 (Screen On/Off) 이며 절전 모드 시, 동작 불가. |  |  |
|  |  |  | `stb_control_restart` | string |  | STB 리셋 | STB 강제 재시작 | SNMP로처리 |  |  |  |  |  |
|  |  |  | `stb_control_reset` | string |  | STB 공장 초기화 | STB 원격 공장 초기화 동작 | SNMP로처리 |  | 앱에서 공장초기화를 부를 수 있는 api 미존재 |  |  |  |
|  |  | `work_value` |  |  |  | 제어요청값 | 해당 항목의 설정 요청값 |  |  |  |  |  |  |
|  | Response | `code` |  | string | ✓ |  | 1: 성공,  0:실패 |  |  |  |  |  | <details><summary>JSON 샘플</summary><pre>{ 'code' : '1', 'message': '' }</pre></details> |
|  |  | `message` |  | string |  |  |  |  |  |  |  |  |  |

---

## 스마트리부팅 인터페이스

**연동구분**: 소켓 (TCP)
**연동**: serverIp: 셋탑아이피, port: 8801
**플랫폼**: Android, OCAP

| 연동구분 | 연동 | KEY명1 | KEY명2 | 타입 | 필수여부 | 항목 | 내용 | Default값 | 예제값 | 비고 | 검토 (Android) | 검토 (OCAP) | json 샘플 |
|------|------|------|------|:----:|:----:|------|------|------|------|------|------|------|------|
| 제어 | Request | `work_type` | `smartReboot` | string |  | stb 스마트 리부팅 | 자동 리부팅(Stablizer) 기능을 통한 스마트 리부팅 |  |  |  |  |  |  |
|  |  | `work_value` |  |  |  |  |  |  |  |  |  |  |  |
|  | Response | `code` |  | string |  |  | 1: 성공, 0:실패 |  |  |  |  |  |  |
|  |  | `message` |  | string |  |  |  |  |  |  |  |  |  |

---

## STB 리셋 인터페이스

**연동구분**: 소켓 (TCP)
**연동**: serverIp: 셋탑아이피, port: 8801
**플랫폼**: Android, OCAP

| 연동구분 | 연동 | KEY명1 | KEY명2 | 타입 | 필수여부 | 항목 | 내용 | Default값 | 예제값 | 비고 | 검토 (Android) | 검토 (OCAP) | json 샘플 |
|------|------|------|------|:----:|:----:|------|------|------|------|------|------|------|------|
| 제어 | Request | `work_type` | `stbRestart` | string |  | stb 재시작 | STB 재시작 기능으로 대기상태 재시작 시 대기상태 부팅, 시청 상태 재시작 시 즉시 재시작 후 시청 채널로 이동 기능 |  |  | 2025-12-01 추가 요구사항 |  |  |  |
|  |  | `work_value` |  |  |  |  |  |  |  |  |  |  |  |
|  | Response | `code` |  | string |  |  | 1: 성공, 0:실패 |  |  |  |  |  |  |
|  |  | `message` |  | string |  |  |  |  |  |  |  |  |  |

---

## 제어 응답 메시지 종류

| code | message | 설명 |
|:----:|------|------|
| 1 | success | 성공 |
|  | fail | 실패 |
|  | wrong parameter | 해당 제어 요청의 설정값(parameter)이 유효하지 않음 |
|  | wrong work type | 해당 제어 요청의 명령어(work type)가 유효하지 않음 |
|  | no data | 데이터가 존재하지 않음 |
|  | no action | 해당 제어 요청에 대해 동작하지 않음 |
|  | not supported | 해당 제어는 지원하지 않음 |

---

## 문서 생성 정보

- 생성 시각: 2026-03-11 11:58:40
- 생성 스크립트: `generate_spec.py`
- 원본 데이터: `interface_data.json`
