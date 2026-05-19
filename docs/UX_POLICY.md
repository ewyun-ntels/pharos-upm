# UX 공통 정책
필요한 정책, 이슈가 있을 경우 작성
<br>

## UX Case


#### 💠 표기 방식
| 번호 | 정책 | 세부 설명 |
|------|------|-----------|
| 1 | 날짜 표기 | - 한국 : yyyy. MM. dd 오전 hh:mm (24시간제, 구분자 ".") <br> - 미국 : MM/dd/yyyy, hh:mm (12시간제, 구분자 "/") <br> - 영국 : dd/MM/yyyy hh:mm (24시간제, 구분자 "/") <br> - 일본 : yyyy/MM/dd hh:mm (24시간제, 구분자 "/") <br> - 독일 : dd.MM.yyyy hh:mm (24시간제, 구분자 ".") |
| 2 | Title 대문자 표기 | New dashboard, Alert name, Dashboard ... |
| 3 | Header Title 표기 | - List 화면 : Title 노출, Breadcrumbs 미노출 <br> - Dashboard 화면 : Title 미노출, Breadcrumbs 노출 |
| 4 | Column Button 명칭 | - Table Column에 버튼이 여러개일 경우 : Column명 **Actions** <br> - 단일 버튼일 경우 : 해당 버튼명을 Column명과 일치 (ex. Delete, Notification) |
| 5 | New 와 Add 의 차이 | - New : 신규, 새로운 페이지 생성일 경우 New <br> - Add : 기존 또는 무엇인가 추가하려고 했을 때 Add |

#### 💠 컴포넌트 

| 번호 | 정책 | 세부 설명 |
|------|------|-----------|
| 1 | 필수값 입력 후 [Save] 버튼 활성화 | ex. New user modal |
| 2 | Input <br>(Status : Error) | - Input 입력 시 즉시 유효성검사 <br> - Input Ring 색상 **Red** 변경 <br> - Description과 error msg 노출될 경우 <br> 🔸 case1. Input 아래 error - Description 노출 <br> 🔸 case2. Description 내용을 icon button으로 tooltip 노출 |
| 3 | Main Primary Button [▫️Add button] | - Primary + Add (최상위 레벨 버튼의 경우) <br> - 액션과 구체적인 엔티티로 작성 <br> ex. + Add rule, + New User, + Add Panel |
| 4 | [▫️icon button] tooltip | 아이콘 버튼만 있는 경우 tooltip 노출
| 5 | Form size | - 편집(Edit, New) <br> panel, 신규, 수정 페이지 사용되는 form height=36px <br>- Filter <br> dashboard 필터 또는 list 필터 화면 form height=32px
| 6 | 권한 관리 | - CRUD API 기반으로 Component 활성화 여부 <br> - Sidebar Menu도 CRUD 기반으로 활성화 여부 표현(CRUD중 한개만 있어도 표현) <br> - View 권한은 주로 "목록/데이터 조회"에 한정해서 사용 <br> - Create/Update/Delete 동작은 각각의 권한만으로도 관리 <br> - 조회 화면은 View 권한으로도 제어 가능(버튼, 액션은 개별 권한으로 필터링하여 적용함) <br> - Dashboard 화면의 경우 별도 권한으로 관리함 (Setting > Permission) |
| 7 | List view <br> [▫️Dele] 버튼 | - [▫️전체 삭제] 버튼은 최상단에 별도 표기 <br> - 부분 삭제는 table 또는 컨텐츠 주변에 [icon button] 표현 <br> - Table에서 checkbox 선택 시 [Delate] 버튼 노출, 기본 형태는 미노출 상태 | 
| 8 | List view <br> 표현 방식 | - Table 첫번째 컬럼의 명칭 클릭 시 Detail Sheet  액션 <br> - 수정은 Dialog 표현 <br> - 왼쪽 명칭 링크표시  (예외의 경우 별도 표시 필요.) |
| 9 | Table scroll | Table scroll 상시 노출
| 10 | Breadcrumb | UI 위치만 잡고 옵션에 따라 show/hide 할 수 있도록 반영
| 11 | 목록형 테이블(Icon button) | border가 없는 아이콘 형태 사용
| 12 | Select Hover&Focus style <br>(Select, React-Select, Datetime-range등) | - Hover : border-foreground 적용 <br> - Focus : border-foreground 적용 <br> - 모든 select 컴포넌트 및 option에 cursor-pointer (Hand) 적용 - custom.css 에 적용함 |

#### 💠 Font 

| 번호 | 정책 | 세부 설명 |
|------|------|-----------|
| 1 | h1 (tailwind : text-2xl) <br> 로그인 화면 사용중 | H1 Title 24 Text Label |
| 2 | h2 (tailwind : text-xl) <br> 페이지 타이틀 사용중 | H2 Title 20 Text Label |
| 3 | h3 (tailwind : text-lg) <br> 모달, 카드 타이틀 사용중 | H3 Title 18 Text Label |
| 4 | h4 (tailwind : text-base) | H4 Title 16 Text Label |
| 5 | text-[13px] | - Tailwind에 존재하지 않는 13px 사이즈 추가 |


---
#### 🔔 이슈
1. Table row UX 구체적인 정의 필요
    - Side sheet 노출되었을 경우 또는 row hover 했을 경우
    - 특정 row 클릭했을 때 UI 표현
    - side sheet와 트리거 항목이 포함된 화면(dashboard 또는 list 페이지 등..)과 연결되어 있지 않음.