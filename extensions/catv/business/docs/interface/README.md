# SKB CATV 인터페이스 문서 관리

> 엑셀 기반 인터페이스 정의서를 Markdown 문서로 자동 생성하는 시스템

---

## 목차

- [빠른 시작](#빠른-시작)
- [📘 문서 선택 가이드](#-문서-선택-가이드)
- [현재 버전 정보](#현재-버전-정보)
- [📁 디렉토리 구조](#-디렉토리-구조)
- [🚀 사용 방법](#-사용-방법)
- [📝 스크립트 설명](#-스크립트-설명)
- [🔧 개발 워크플로우](#-개발-워크플로우)
- [⚠️ 주의사항](#️-주의사항)
- [📚 문서 이용 안내](#-문서-이용-안내)
- [🛠️ 트러블슈팅](#️-트러블슈팅)
- [📊 변환 통계](#-변환-통계)
- [인터페이스 개요](#인터페이스-개요)
- [🛠️ 개발 참고 사항](#️-개발-참고-사항)
- [🔧 문서 생성 시스템](#-문서-생성-시스템)

---

## 빠른 시작

### 문서 찾기

목적에 맞는 문서를 선택하세요:

| 문서                                                             | 용도                         | 위치        | 대상 독자                    |
| ---------------------------------------------------------------- | ---------------------------- | ----------- | ---------------------------- |
| **[INTERFACE.spec.md](analysis/INTERFACE.spec.md)**              | 공식 규격서 (엑셀 100% 일치) | `analysis/` | QA, 법무, 규격 검증          |
| **[IMPLEMENTATION_VALIDATION.md](IMPLEMENTATION_VALIDATION.md)** | 구현 검증 보고서 (코드↔SPEC) | `docs/`     | 개발 리더, QA 리더, 아키텍트 |
| **[VALIDATION_REPORT.md](VALIDATION_REPORT.md)**                 | 문서 검증 보고서 (엑셀↔SPEC) | `docs/`     | QA, 문서 관리자              |

자세한 비교는 [📘 문서 선택 가이드](#-문서-선택-가이드)를 참조하세요.

### 엑셀 업데이트 시

```bash
# docs 디렉토리에서 실행
python3 parse_excel.py && python3 generate_spec.py
```

---

## 📘 문서 선택 가이드

### **[INTERFACE.spec.md](analysis/INTERFACE.spec.md)** - 공식 규격서 ⚡

엑셀 원본과 100% 일치하는 공식 규격서입니다.

**이런 경우 사용하세요:**

- ✅ 엑셀 파일과 정확히 일치하는 내용 확인 필요
- ✅ 법적/계약적 레퍼런스 필요
- ✅ QA 테스트 케이스 작성
- ✅ 원본 데이터 추적 및 검증
- ✅ Git diff로 정확한 변경 내역 확인

**특징:**

- 🤖 자동 생성 (직접 수정 금지)
- 📊 엑셀 시트를 Markdown 테이블로 변환
- 📝 최소한의 구조화
- 🎯 정확성 중심
- 📏 405줄, 29KB

**대상 독자:** QA 엔지니어, 법무팀, 정확한 규격이 필요한 개발자

### **[IMPLEMENTATION_VALIDATION.md](IMPLEMENTATION_VALIDATION.md)** - 구현 검증 보고서 ✅

INTERFACE.spec.md v2.8과 실제 구현 코드 간의 일치 여부를 검증한 보고서입니다.

**이런 경우 사용하세요:**

- ✅ 스펙 준수 확인 (개발 완료 후)
- ✅ 코드 리뷰 및 감사
- ✅ 인터페이스 변경 영향도 분석
- ✅ 아키텍처 품질 평가
- ✅ 신규 개발자 코드베이스 이해

**특징:**

- 📊 **9개 인터페이스 전체 검증**: 100% 스펙 준수 확인
- 🔍 **필드별 매핑 검증**: 213개 필드 정확도 분석
- 📈 **버전 변경사항 추적**: v2.2, v2.5, v2.6, v2.7, v2.8 반영 확인
- 💻 **구현 파일 매핑**: 각 인터페이스별 구현 파일 명시
- ⚙️ **코드 품질 평가**: 아키텍처, 에러 처리, 최적화
- ⚠️ **특이사항 보고**: v2.2 삭제 필드, v2.8 프로토콜 변경 등
- 📏 **문서 통계**: 파일별/인터페이스별 복잡도 분석
- 🧪 **시뮬레이터 검증**: 데이터 전송 시뮬레이터 구현 확인

**검증 결과:**

- ✅ 필드 매핑 정확도: **100%** (213/213)
- ✅ 프로토콜 준수: **100%**
- ✅ 버전 반영: **100%** (v2.2 삭제 필드 제거 완료)
- ✅ 코드 품질: **우수**

**대상 독자:** 개발 리더, 아키텍트, QA 리더, 코드 감사자

### **[VALIDATION_REPORT.md](VALIDATION_REPORT.md)** - 문서 검증 보고서

엑셀 원본과 INTERFACE.spec.md 간의 일치 여부를 검증한 보고서입니다.

**이런 경우 사용하세요:**

- ✅ 문서 생성 스크립트 검증
- ✅ 엑셀 → Markdown 변환 품질 확인
- ✅ 누락된 필드 탐지
- ✅ 문서 자동화 개선

**특징:**

- 📝 시트별 Request/Response 필드 대조
- 🔧 스크립트 개선 이력
- ✅ 수정 완료 확인

**대상 독자:** QA 담당자, 문서 관리자, 스크립트 개발자

---

## 현재 버전 정보

| 항목             | 내용                                                                     |
| ---------------- | ------------------------------------------------------------------------ |
| 엑셀 버전        | v2.8                                                                     |
| 엑셀 최종 수정일 | 2024년 12월 8일                                                          |
| SPEC 생성일      | 2026년 03월 11일                                                         |
| 구현 검증 완료일 | 2026년 03월 11일                                                         |
| 원본 파일명      | 1*COMS_SKB_CATV_Log Agent 단말 인터페이스 정의서\_v2.8(251208)*해제.xlsx |

### 주요 변경사항 (v2.7 → v2.8)

- **품질계측전송**: HTTP PORT 50002 → UDP PORT 50005 (안정적으로 동작함)
- **프로토콜 변경**: HTTP 기반 → UDP 기반 전송
- **구현 파일**: `http/handler.go` → `udp/quality_measurement.go`
- **시뮬레이터**: 품질계측전송 포트 50005로 업데이트

---

## 📁 디렉토리 구조

```
docs/
├── parse_excel.py                   # 엑셀 → JSON 변환 스크립트
├── generate_spec.py                 # JSON → SPEC 생성 스크립트
├── interface_data.json              # 중간 데이터 (자동 생성)
├── IMPLEMENTATION_VALIDATION.md     # 구현 검증 보고서 (코드↔SPEC)
├── VALIDATION_REPORT.md             # 문서 검증 보고서 (엑셀↔SPEC)
├── README.md                        # 문서 관리 가이드 (이 파일)
├── reference/                       # 원본 엑셀 파일
│   └── 1_COMS_SKB_CATV_Log Agent 단말 인터페이스 정의서_v2.8(251208)_해제.xlsx
└── analysis/                        # 생성된 문서
    └── INTERFACE.spec.md            # 공식 규격서 (자동 생성)
```

## 🚀 사용 방법

### 1단계: 엑셀 파일을 JSON으로 변환

```bash
python3 parse_excel.py
```

**출력:**

- `interface_data.json` 생성 (약 16MB)
- 12개 시트, 9000+ 행의 모든 셀 데이터 포함

### 2단계: JSON에서 SPEC 문서 생성

```bash
python3 generate_spec.py
```

**출력:**

- `analysis/INTERFACE.spec.md` 생성 (약 29KB)
- 엑셀 원본과 100% 일치하는 공식 규격서

### 전체 프로세스 (엑셀 업데이트 시)

```bash
# docs 디렉토리에서 실행
python3 parse_excel.py && python3 generate_spec.py
```

## 📝 스크립트 설명

### `parse_excel.py`

**목적:** 엑셀 파일을 구조화된 JSON으로 변환

**기능:**

- `reference/` 폴더에서 최신 엑셀 파일 자동 탐색
- 모든 시트의 모든 셀 데이터 파싱
- 병합 셀 정보 보존
- 셀 값의 타입 보존 (숫자, 문자열)

**의존성:**

```bash
pip install openpyxl
```

### `generate_spec.py`

**목적:** JSON 데이터를 Markdown 규격서로 변환

**기능:**

- 개정이력 테이블 생성
- 인터페이스별 섹션 생성
- 플랫폼 정보 (Android/OCAP) 구분
- 검토 컬럼 플랫폼별 분리
- JSON 샘플 접기 가능한 형태로 포함

## 🔧 개발 워크플로우

### 신규 엑셀 버전 반영

1. **엑셀 파일 업데이트**

   ```bash
   # reference/ 폴더에 새 버전 엑셀 파일 복사
   cp "새_엑셀_파일.xlsx" reference/
   ```

2. **자동 변환 실행**

   ```bash
   python3 parse_excel.py
   python3 generate_spec.py
   ```

3. **검증**

   ```bash
   # SPEC 파일 통계 확인
   wc -l analysis/INTERFACE.spec.md

   # 검토 컬럼 개수 확인
   grep -c "검토 (Android)" analysis/INTERFACE.spec.md
   ```

## ⚠️ 주의사항

### 자동 생성 파일 (직접 수정 금지)

- ❌ `interface_data.json` - 엑셀 변경 시 재생성됨
- ❌ `analysis/INTERFACE.spec.md` - 스크립트로만 생성

### 수동 관리 파일

- ✅ `README.md` - 문서 관리 가이드
- ✅ `IMPLEMENTATION_VALIDATION.md` - 구현 검증 보고서
- ✅ `VALIDATION_REPORT.md` - 문서 검증 보고서

### Git 관리

```bash
# .gitignore에 추가 권장
interface_data.json   # 크기가 크고 자동 생성됨
```

## 📚 문서 이용 안내

**SPEC 문서**: 정확한 규격 확인, QA 테스트, 계약/법무 레퍼런스
**IMPLEMENTATION_VALIDATION**: 구현 검증, 코드 리뷰, 아키텍처 품질 평가
**VALIDATION_REPORT**: 엑셀↔SPEC 일치 여부, 문서 생성 스크립트 검증

## 🛠️ 트러블슈팅

### 엑셀 파일을 찾을 수 없음

```bash
# reference/ 폴더에 엑셀 파일이 있는지 확인
ls -la reference/*.xlsx
```

### openpyxl 설치 오류

```bash
# Python 가상환경 활성화 후
source ~/.venv/bin/activate
pip install openpyxl
```

### JSON 파일이 너무 큼

`interface_data.json`은 모든 셀 데이터를 포함하므로 크기가 큽니다 (약 16MB).  
이는 정상이며 Git에서 제외하는 것을 권장합니다.

## 📊 변환 통계

**엑셀 → JSON:**

- 입력: v2.8 엑셀 파일 (12개 시트)
- 출력: 15.8 MB JSON (9,001 행)

**JSON → SPEC:**

- 입력: interface_data.json
- 출력: 29 KB Markdown (405 줄)

## 인터페이스 개요

### 데이터 전송 인터페이스 (STB → Server)

| 인터페이스     | 전송 조건       | 프로토콜 | 포트         | 링크                                                                     |
| -------------- | --------------- | -------- | ------------ | ------------------------------------------------------------------------ |
| 주기전송       | 10분마다        | UDP      | 50000        | [SPEC](analysis/INTERFACE.spec.md#주기적-전송-인터페이스)                |
| 일일전송       | 1일 1회         | UDP      | 50001        | [SPEC](analysis/INTERFACE.spec.md#일일-전송-인터페이스)                  |
| 품질계측전송   | Sleep 모드 진입 | **UDP**  | **50005** ⭐ | [SPEC](analysis/INTERFACE.spec.md#품질계측전송-인터페이스)               |
| 자가진단전송   | 핫키/원격 요청  | UDP/TCP  | 50003/8801   | [SPEC](analysis/INTERFACE.spec.md#자가진단전송-인터페이스)               |
| 망품질전환전송 | QAM→8VSB 전환   | UDP      | 50004        | [SPEC](analysis/INTERFACE.spec.md#망품질전환전송-인터페이스)             |
| 전송요청       | 서버 요청       | TCP      | 8801         | [SPEC](analysis/INTERFACE.spec.md#전송-요청에-의한-정보-전달-인터페이스) |

### 단말 제어 인터페이스 (Server → STB)

| 인터페이스   | 프로토콜 | 포트 | 링크                                                       |
| ------------ | -------- | ---- | ---------------------------------------------------------- |
| 단말제어     | TCP      | 8801 | [SPEC](analysis/INTERFACE.spec.md#단말-제어-인터페이스)    |
| 스마트리부팅 | TCP      | 8801 | [SPEC](analysis/INTERFACE.spec.md#스마트리부팅-인터페이스) |
| STB재시작    | TCP      | 8801 | [SPEC](analysis/INTERFACE.spec.md#stb-리셋-인터페이스)     |

---

## 🛠️ 개발 참고 사항

### 공통 필드

모든 데이터 전송 인터페이스는 다음 공통 필드를 포함합니다:

- `hostId`: HOST ID (10자리)
- `macAddr`, `cmMac`: STB/CM MAC Address
- `stbIp`, `cmIp`: STB/CM IP 주소
- `stbModel`: STB 모델명
- `mwVer`, `localVer`, `cloudVer`: 버전 정보
- `loggingTime`, `sendingTime`: 수집/전송 시간

자세한 필드 정의는 [INTERFACE.spec.md](analysis/INTERFACE.spec.md)를 참조하세요.

### 플랫폼별 지원

일부 기능은 Android와 OCAP 플랫폼에서 지원 여부가 다릅니다:

**Android 지원 불가**: tvLock, zappingAd, morningAlarm, standbyMode  
**Android 검토 필요**: hdmiCec, hdcp  
**OCAP 특정 모델**: audioMode, hdmiCec (U300, HC100), hdr (U300)

### 빠른 링크

**인터페이스별 바로가기:**

데이터 전송 (6종):

- [주기전송](analysis/INTERFACE.spec.md#주기적-전송-인터페이스) | [일일전송](analysis/INTERFACE.spec.md#일일-전송-인터페이스) | [품질계측](analysis/INTERFACE.spec.md#품질계측전송-인터페이스)
- [자가진단](analysis/INTERFACE.spec.md#자가진단전송-인터페이스) | [망품질전환](analysis/INTERFACE.spec.md#망품질전환전송-인터페이스) | [전송요청](analysis/INTERFACE.spec.md#전송-요청에-의한-정보-전달-인터페이스)

단말 제어 (3종):

- [단말제어](analysis/INTERFACE.spec.md#단말-제어-인터페이스) | [스마트리부팅](analysis/INTERFACE.spec.md#스마트리부팅-인터페이스) | [STB재시작](analysis/INTERFACE.spec.md#stb-리셋-인터페이스)

---

## 🔧 문서 생성 시스템

### 워크플로우

```
엑셀 원본  →  parse_excel.py  →  JSON  →  generate_spec.py  →  SPEC 문서
  (v2.8)                        (15.8MB)                           (35.9KB)
                                                                      ↓
                                                              analysis/INTERFACE.spec.md
```

**검증 문서**는 별도 관리:

- `IMPLEMENTATION_VALIDATION.md`: 구현 검증
- `VALIDATION_REPORT.md`: 문서 검증

### 전체 프로세스 (엑셀 업데이트 시)

```bash
# docs 디렉토리에서 실행
python3 parse_excel.py && python3 generate_spec.py
```

**결과:**

- `interface_data.json` 생성 (15.8 MB)
- `analysis/INTERFACE.spec.md` 갱신 (29 KB)
