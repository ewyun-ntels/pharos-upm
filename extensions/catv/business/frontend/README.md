# CATV Business Frontend

CATV STB(Set-Top Box) 제어 및 모니터링을 위한 프론트엔드 애플리케이션입니다.

---

## 📁 프로젝트 구조

```
src/
├── features/
│   └── control/                          # STB 제어 및 스케줄 관리
│       ├── page.tsx                      # 메인 페이지 (탭 레이아웃)
│       ├── index.ts
│       └── components/
│           ├── constants.ts              # 공통 상수 (WORK_TYPE_OPTIONS 등)
│           ├── schedule-list/            # 스케줄 목록 및 관리
│           ├── schedule-editor/          # 스케줄 등록·수정 에디터 (탭 내 페이지 형태)
│           ├── schedule-result/          # 스케줄 실행 결과 목록
│           └── schedule-result-detail/   # 결과 내 개별 STB 레코드 상세
├── providers/
│   └── catv-provider/                    # CATV 데이터 프로바이더
│       ├── types.ts                      # 타입 및 리소스 상수 정의
│       └── index.ts                      # DataProvider 구현
├── permissions.ts                        # 권한 상수
└── index.ts
```

---

## 🗂️ 도메인 계층 구조

```
Schedule                    스케줄 정의 (반복/한번/즉시)
  └── ScheduleResult        스케줄 1회 실행 결과 집계 (상태별 건수)
        └── ScheduleResultRecord   STB 1대 단위 제어 결과 레코드
```

- **Schedule** — 언제, 어떤 STB에, 어떤 명령을 실행할지 정의
- **ScheduleResult** — 실행 1회의 전체 진행 현황 (pending/running/succeeded/failed 집계)
- **ScheduleResultRecord** — 개별 STB의 제어 결과 (시작/종료 시간, 결과 코드, 결과 메시지 포함)

---

## 📦 컴포넌트 설계

### `schedule-list/`

스케줄 목록을 표시하고 등록·수정·삭제 진입점을 제공합니다.

- 생성 시간 내림차순 정렬 (최신순)
- `immediately` 타입, 또는 `once` 타입이고 실행 시간이 현재 이전인 경우 수정 불가 (버튼 미출력)
- 행 클릭 → `ScheduleDetailSheet`: 스케줄 정보 + 해당 스케줄의 결과 목록
- 권한에 따라 버튼 출력 여부 결정 (아래 권한 섹션 참고)

### `schedule-editor/`

스케줄을 생성·수정하는 에디터입니다. 메인 탭 내에서 별도 탭으로 열립니다 (다이얼로그 아님).

- **스케줄 정보**: 이름 / 실행 타입 (즉시·한번·반복)
  - 한번 실행: 인라인 Calendar + TimePicker로 날짜·시간 선택
  - 반복 실행: 타임존 선택 + Cron 표현식 입력
- **대상 선택**: SO → L3 → Cell 계층 선택, 선택 깊이에 따라 `area_type` 자동 결정
- **제어 명령**: work_type 선택
- **설정 요약**: 하단에 현재 설정 내용 실시간 표시
- 수정 모드에서 DB에 없는 L3/Cell 항목은 `(삭제됨)` 표시 후 선택 유지

### `schedule-result/`

스케줄 실행 결과 목록을 표시합니다.

- 서버 사이드 페이지네이션
- `running`/`pending` 결과가 있으면 5초 폴링
- 필터: 제어 명령(work_type), 스케줄 이름 검색, 기간 선택
  - 기간 선택: 시작/종료 날짜·시간 인라인 선택, 같은 날 종료 시간이 시작 이전이면 적용 비활성화
- 행 클릭 → 결과 ID와 work_type을 키로 하는 동적 탭 생성

### `schedule-result-detail/`

결과 내 개별 STB 레코드를 표시합니다. work_type에 따라 두 가지 렌더링 모드가 있습니다.

| 모드     | 조건                               | 특징                                                 |
| -------- | ---------------------------------- | ---------------------------------------------------- |
| 기본     | `work_type !== 'stb_request_info'` | 12개 기본 컬럼 (신원·시간·결과·메타)                 |
| STB 정보 | `work_type === 'stb_request_info'` | 기본 12개 + result_message JSON을 동적 컬럼으로 파싱 |

**컬럼 표시 설정 (`ColumnVisibilityPanel`)**

- `BASIC_COLUMN_IDS` (12개): 기본 컬럼 섹션 — 항목별 토글 가능
- 동적 컬럼 (STB 정보 모드): 별도 섹션, 전체 선택/해제 지원
- `DETAIL_DEFAULT_HIDDEN_COLUMN_IDS_BY_WORK_TYPE`: work_type별 초기 숨김 목록

---

## 📐 주요 타입 (`providers/catv-provider/types.ts`)

### `CATV_RESOURCES`

| 상수                      | 값                          | 용도                  |
| ------------------------- | --------------------------- | --------------------- |
| `SCHEDULES`               | `'schedules'`               | 스케줄 정의 CRUD      |
| `SCHEDULE_RESULTS`        | `'schedule-results'`        | 스케줄 실행 결과 목록 |
| `SCHEDULE_RESULT_RECORDS` | `'schedule-result-records'` | 개별 STB 제어 레코드  |

### `Schedule`

스케줄 정의. `schedule_type`은 `immediately` / `once` / `repeat`.

| 필드                   | 설명                                        |
| ---------------------- | ------------------------------------------- |
| `name`                 | 스케줄 고유 이름 (ID 역할)                  |
| `schedule_type`        | `immediately` / `once` / `repeat`           |
| `schedule_spec_once`   | 한번 실행 시각 (ISO 8601 UTC)               |
| `schedule_spec_repeat` | 반복 실행 Cron 표현식 (`TZ=Asia/Seoul ...`) |
| `area_type`            | 제어 대상 범위 (`so` / `l3` / `cell`)       |
| `work_type`            | 제어 명령 종류                              |

### `ScheduleResult`

실행 1회 집계. `count_by_status`로 pending/running/succeeded/failed 수 제공.

### `ScheduleResultRecord`

STB 1대 단위 결과. `result_code`: `1`(성공), `0`(기기 오류), `-1`(시스템 오류).

---

## 🔒 권한

| 권한 키                 | 기능                                  |
| ----------------------- | ------------------------------------- |
| `extension:catv:read`   | 페이지 접근, 모든 목록·상세 조회      |
| `extension:catv:create` | 스케줄 등록 버튼 출력 및 등록 탭 진입 |
| `extension:catv:update` | 스케줄 수정 버튼 출력 및 수정 탭 진입 |
| `extension:catv:delete` | 스케줄 삭제 버튼 출력 및 삭제 실행    |

권한 조합에 따른 버튼 출력 여부:

| Create | Update | Delete | 등록 버튼 | 수정 버튼¹ | 삭제 버튼 |
| :----: | :----: | :----: | :-------: | :--------: | :-------: |
|   ❌   |   ❌   |   ❌   |  미출력   |   미출력   |  미출력   |
|   ❌   |   ❌   |   ✅   |  미출력   |   미출력   |   출력    |
|   ❌   |   ✅   |   ❌   |  미출력   |    출력    |  미출력   |
|   ❌   |   ✅   |   ✅   |  미출력   |    출력    |   출력    |
|   ✅   |   ❌   |   ❌   |   출력    |   미출력   |  미출력   |
|   ✅   |   ❌   |   ✅   |   출력    |   미출력   |   출력    |
|   ✅   |   ✅   |   ❌   |   출력    |    출력    |  미출력   |
|   ✅   |   ✅   |   ✅   |   출력    |    출력    |   출력    |

¹ `immediately` 타입이거나 `once` 타입인데 실행 시간이 현재 이전이면 수정 버튼 미출력

> `role:super_admin` 계정은 모든 권한을 보유하므로 항상 모든 버튼이 출력됩니다.
