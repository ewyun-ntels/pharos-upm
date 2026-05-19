# CATV STB 제어 시스템 다이어그램

## 📋 개요

SKB CATV STB 제어 시스템의 전체 문서화 다이어그램 모음입니다.

## 📁 파일 구조

### 개별 파일 (원본)

1. **01-control-ui-wireframe.drawio** - STB 제어 화면 와이어프레임
2. **02.01-all-jobs-list-ui-wireframe.drawio** - Job 목록 화면 와이어프레임
3. **02.02-job-detail-result-ui-wireframe.drawio** - Job 상세 화면 와이어프레임
4. **03-control-sequence-diagram.drawio** - 제어 시퀀스 다이어그램
5. **04-state-transition-diagram.drawio** - 상태 전이 다이어그램
6. **05-system-architecture-diagram.drawio** - 시스템 아키텍처 다이어그램

### 통합 파일

- **catv-stb-control-all-diagrams.drawio** - 모든 다이어그램을 하나의 파일로 통합 (6개 페이지)

## 🔧 병합 스크립트

### merge_drawio_files.py

여러 개의 drawio 파일을 하나의 파일로 병합하는 Python 스크립트입니다.

#### 사용 방법

```bash
python3 merge_drawio_files.py
```

#### 기능

- ✅ 6개 drawio 파일을 하나로 병합
- ✅ 각 파일을 별도 페이지로 유지
- ✅ 페이지 이름에 번호 prefix 자동 추가
- ✅ diagram ID 충돌 방지 (파일명 기반 고유 ID)

#### 출력

```
catv-stb-control-all-diagrams.drawio
├─ [01] STB 제어 화면 Wireframe
├─ [02.01] Job 목록 화면 Wireframe
├─ [02.02] Job 상세 화면 Wireframe
├─ [03] 제어 시퀀스 다이어그램
├─ [04] 상태 전이 다이어그램
└─ [05] 시스템 아키텍처
```

## 📊 파일 정보

| 파일                                        | 크기     | 설명                                     |
| ------------------------------------------- | -------- | ---------------------------------------- |
| 01-control-ui-wireframe.drawio              | 35K      | POST /catv/stb/control UI                |
| 02.01-all-jobs-list-ui-wireframe.drawio     | 36K      | GET /catv/stb/control/results UI         |
| 02.02-job-detail-result-ui-wireframe.drawio | 61K      | GET /catv/stb/control/results/:job_id UI |
| 03-control-sequence-diagram.drawio          | 28K      | 시퀀스 다이어그램 (v3.0)                 |
| 04-state-transition-diagram.drawio          | 18K      | Task 상태 전이도                         |
| 05-system-architecture-diagram.drawio       | 24K      | 전체 시스템 구조                         |
| **catv-stb-control-all-diagrams.drawio**    | **189K** | **통합 파일 (6 페이지)**                 |

## 🎯 사용 가이드

### 통합 파일 열기

1. [draw.io](https://app.diagrams.net/) 또는 [diagrams.net](https://www.diagrams.net/) 접속
2. `catv-stb-control-all-diagrams.drawio` 열기
3. 하단 탭에서 페이지 전환

### 개별 파일 수정

1. 원본 개별 파일 수정 (01~05)
2. `python3 merge_drawio_files.py` 실행
3. 통합 파일 자동 재생성

## 📝 주요 내용

### 1️⃣ UI 와이어프레임 (01, 02.01, 02.02)

- **구조**: 상단 와이어프레임 + 하단 개발자 노트
- **원칙**: API 응답에서 직접 알 수 있는 데이터만 표시
- **스펙**: 실제 API 스펙과 100% 일치

### 2️⃣ 제어 시퀀스 다이어그램 (03)

- **버전**: v3.0 최적화 (메모리 캐시)
- **특징**: 실제 코드 기반 정확한 시퀀스
- **범위**: HTTP POST → K8s Job → Worker → STB

### 3️⃣ 상태 전이 다이어그램 (04)

- **상태**: pending → running → completed/failed
- **코드**: result_code (1/0/-1/"")
- **조건**: K8s 성공/실패, TCP 응답 패턴

### 4️⃣ 시스템 아키텍처 (05)

- **계층**: 4-Layer 구조
- **컴포넌트**: HTTP API, K8s, Worker, STB
- **최적화**: v3.0 (Connection Pool 제거, Rate Limit)

## 🔄 버전 정보

- **작성일**: 2026-02-13 (개별 파일)
- **통합일**: 2026-02-23 (병합 스크립트)
- **시스템 버전**: v3.0 (Worker Pool, Memory Cache)
- **API 버전**:
  - POST /catv/stb/control
  - GET /catv/stb/control/results (limit: 100~1000)
  - GET /catv/stb/control/results/:job_id (limit: 1000~10000)

## 🚀 주요 특징

### v3.0 최적화

- ✅ DB 중복 삽입 제거 (메모리 캐시)
- ✅ Connection Pool 제거 (직접 연결)
- ✅ Rate Limiter (300/초)
- ✅ 포트 고갈 대응
- ✅ 2단계 재시도 전략

### 문서 정확성

- ✅ 실제 코드 기반 작성
- ✅ API 스펙 100% 일치
- ✅ 운영 데이터 반영 (81,106대 실제 사례)

## 📚 참고 자료

### 관련 코드

- `pkg/control/http/handler.go` - HTTP API 핸들러
- `pkg/control/http/api.go` - API 라우팅
- `pkg/control/command/worker_pool_handler.go` - Worker Pool
- `pkg/control/common/stb_control_job_table.go` - DB 스키마

### 데이터베이스

- `dist_stb_information` - STB 정보 (TTL: 30일)
- `dist_stb_control_job` - Job/Task 결과 (TTL: 365일)

### API 문서

- 엔드포인트: `/catv/stb/*`
- 응답 형식: JSON
- 페이지네이션: limit/offset
- 필터: job_id, work_type, task_id, status, result, MAC

## ⚠️ 주의사항

1. **개별 파일 수정 시**: 병합 스크립트 재실행 필요
2. **API 스펙 변경 시**: 와이어프레임 개발자 노트 업데이트
3. **시스템 변경 시**: 시퀀스/아키텍처 다이어그램 업데이트
4. **통합 파일 직접 수정 시**: 개별 파일과 동기화 불가

## 💡 Tips

- 큰 화면에서 보기: zoom 조절 또는 전체 화면 모드
- 인쇄: 페이지별 인쇄 또는 PDF 내보내기
- 공유: PNG/SVG/PDF 형식으로 내보내기 가능
- 협업: draw.io 온라인 편집 기능 활용

---

**문의**: 시스템 관련 문의는 pharos 팀으로 연락하세요.
