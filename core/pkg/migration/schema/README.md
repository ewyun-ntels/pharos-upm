# Schema Migrations

## 개요

**스키마 마이그레이션**은 데이터베이스 테이블 구조를 정의하는 마이그레이션입니다.

## 특징

- **실행 시점**: 애플리케이션 시작 **전** (필수)
- **목적**: DDL (CREATE TABLE, ALTER TABLE, DROP TABLE)
- **테이블명**: `goose_db_version`
- **형식**: SQL 파일 또는 Go 함수
- **프레임워크**: [Goose](https://github.com/pressly/goose)

## 구조

```
schema/
├── agent/
│   └── base/
│       └── sqlite/
├── master/
│   ├── base/
│   │   ├── postgresql/
│   │   └── sqlite/
│   └── statistics/
│       ├── clickhouse/
│       ├── postgresql/
│       └── sqlite/
├── databases.go          # 실행 로직
└── migration_info.go     # Goose 래퍼
```

## 마이그레이션 파일 작성

### SQL 파일

```sql
-- 20250117000000_create_users.sql

-- +goose Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +goose Down
DROP TABLE users;
```

### Go 함수

```go
// 20250117000000_create_users.go
package postgresql

import (
    "database/sql"
    "github.com/pressly/goose/v3"
)

func init() {
    goose.AddMigrationContext(up, down)
}

func up(ctx context.Context, tx *sql.Tx) error {
    _, err := tx.Exec(`
        CREATE TABLE users (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL
        )
    `)
    return err
}

func down(ctx context.Context, tx *sql.Tx) error {
    _, err := tx.Exec("DROP TABLE users")
    return err
}
```

## 실행 방법

### 자동 실행 (권장)

```go
// main.go
import "ntels.com/pharos/core/pkg/migration"

if err := migration.Load(config); err != nil {
    log.Fatal(err)
}

// 내부적으로:
// 1. schema.Up(config)    - 스키마 생성 ✅
// 2. data.Run(config)     - 데이터 변환
```

### 수동 실행 (goose CLI)

```bash
# PostgreSQL
goose -dir pkg/migration/schema/master/base/postgresql \
      postgres "host=localhost user=pharos dbname=pharos" up

# SQLite
goose -dir pkg/migration/schema/master/base/sqlite \
      sqlite3 pharos.db up
```

## ClickHouse 클러스터 환경

### 문제: goose 버전 테이블의 클러스터 미지원으로 인한 동기화 불일치

goose v3의 ClickHouse dialect는 버전 관리 테이블(`goose_db_version`)을 다음과 같이 생성합니다:

```sql
-- goose 기본 동작 (단일 노드에만 생성됨)
CREATE TABLE IF NOT EXISTS <table> (...)
ENGINE = MergeTree()
ORDER BY (date)
```

반면 실제 마이그레이션 SQL은 모두 `ON CLUSTER` + `ReplicatedMergeTree`를 사용하기 때문에, 클러스터 환경에서 파드마다 연결되는 노드가 다르면 버전 테이블이 없는 노드에서 마이그레이션 상태를 읽지 못해 **동기화 불일치**가 발생합니다.

### 해결: 커스텀 Store로 ON CLUSTER 버전 테이블 생성

`database.Store` 인터페이스를 구현한 `clickhouseClusterStore`(`clickhouse_cluster_store.go`)가 `CreateVersionTable`을 오버라이드하여 아래 DDL로 테이블을 생성합니다:

```sql
CREATE TABLE IF NOT EXISTS <table> ON CLUSTER <cluster_name>
(
    version_id Int64,
    is_applied UInt8,
    date       Date     DEFAULT now(),
    tstamp     DateTime DEFAULT now()
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/goose/<table>', '{replica}')
ORDER BY (date)
```

#### 다중 샤드 환경에서 분산 테이블 없이 동기화되는 이유

이 프로젝트의 클러스터는 **멀티 샤드** 구성을 사용합니다:

```
clickhouse_cluster_replicated
  Shard 1: Replica1(01-01), Replica2(01-02)
  Shard 2: Replica1(02-01), Replica2(02-02)
```

ZooKeeper 복제는 **같은 ZK 경로를 공유하는 노드끼리만** 동기화됩니다.

| ZK 경로                              | 결과                                                                |
| ------------------------------------ | ------------------------------------------------------------------- |
| `/clickhouse/tables/{shard}/<table>` | 샤드별 독립 복제 그룹 → 샤드 간 데이터 불일치                       |
| `/clickhouse/tables/goose/<table>`   | 클러스터 전 노드가 단일 복제 그룹 → 분산 테이블 없이 전 노드 동기화 |

버전 테이블은 ZK 경로에서 `{shard}` 매크로를 **의도적으로 제거**합니다. 이로 인해 어느 노드(어느 샤드)에 마이그레이션 기록이 INSERT되더라도 ReplicatedMergeTree가 클러스터 전 노드로 자동 복제합니다.

#### 비동기 복제와 재기동 시 race condition

하지만 ReplicatedMergeTree의 INSERT는 **비동기**입니다. ClickHouse 공식 문서:

> "Recently inserted data appears on the other replicas with some latency."

즉, A노드에서 버전 기록 INSERT 후 반환 시점에 B노드는 아직 해당 데이터를 갖고 있지 않을 수 있습니다. 재기동 시 B노드에 연결되면 goose가 "마이그레이션 없음"으로 판단해 모든 마이그레이션을 재실행합니다.

| 단계                      | 상황                                                               |
| ------------------------- | ------------------------------------------------------------------ |
| 기동 1회 → A노드 연결     | 마이그레이션 실행, A에 버전 기록 INSERT                            |
| A→B 복제 진행 중 (비동기) | 수 ms~수 초 소요                                                   |
| 재기동 → B노드 연결       | **복제 완료 전이면** B에 버전 기록 없음 → 마이그레이션 재실행 시도 |

#### 해결: INSERT 후 SYSTEM SYNC REPLICA로 전 노드 동기 대기

`clickhouseClusterStore.Insert()`가 `inner.Insert()` 완료 후 추가로 아래를 실행합니다:

```sql
SYSTEM SYNC REPLICA <table> ON CLUSTER <cluster_name>
```

이 명령은 **클러스터 전 노드에서 실행**되어 각 노드가 자신의 복제 큐를 모두 처리할 때까지 기다린 후 반환합니다. INSERT가 완전히 반환된 시점에는 모든 노드가 해당 버전 기록을 갖고 있음이 보장됩니다.

### 접근 방법 비교

동기화 불일치 문제는 두 가지 독립적인 레이어로 나뉩니다.

```
레이어 1: 버전 테이블이 단일 노드에만 생성되는 문제
  └─→ 해결: Custom Store (CreateVersionTable 오버라이드) ← 필수 기반

레이어 2: INSERT 비동기 복제로 인한 재기동 시 race condition
  ├─→ 선택 A: SYSTEM SYNC REPLICA ON CLUSTER (현재 구현)
  └─→ 선택 B: insert_quorum

대안: Custom Store 방식 전체를 대체하는 아키텍처
  ├─→ ClickHouse 접속 노드 고정
  └─→ Kubernetes Job으로 마이그레이션 분리
```

---

#### 레이어 1: Custom Store로 버전 테이블 클러스터 생성 (필수)

`clickhouseClusterStore.CreateVersionTable()`이 `ON CLUSTER + ReplicatedMergeTree`로 버전 테이블을 전 노드에 생성합니다. 이는 레이어 2 선택과 무관하게 반드시 필요한 기반입니다.

---

#### 레이어 2: INSERT 동기화 방법 비교

버전 테이블이 전 노드에 존재해도, INSERT 후 복제가 완료되기 전에 다른 노드로 재연결되면 race condition이 발생합니다. 이를 방지하는 두 가지 방법을 비교합니다.

##### A. SYSTEM SYNC REPLICA ON CLUSTER (현재 구현)

INSERT 완료 후 별도로 `SYSTEM SYNC REPLICA ON CLUSTER`를 실행해 모든 노드의 복제 큐가 비워질 때까지 대기합니다. `receive_timeout`(10초)으로 장애 노드 대기 시간을 제한하며, SYNC 자체가 실패해도 이미 INSERT는 성공했으므로 경고 로그만 출력하고 진행합니다.

```sql
-- 1. 세션 타임아웃 설정 (SYSTEM SYNC REPLICA는 SETTINGS 절 미지원)
SET receive_timeout=10;
-- 2. 전 노드 복제 완료 대기 (실패 시 경고 로그 후 진행)
SYSTEM SYNC REPLICA ON CLUSTER <cluster> <db>.<table>;
```

**장점**

- INSERT 자체는 즉시 반환 → INSERT 성능에 영향 없음
- 버전 테이블 INSERT에만 적용 → 마이그레이션 SQL의 다른 INSERT에 영향 없음
- SYNC 실패가 non-fatal → 노드 일부 장애 시에도 마이그레이션 진행 가능
- `receive_timeout`(10초)으로 장애 노드 대기 시간 제어 (마이그레이션 스텝 수 × 10초로 최악 케이스 제한)

**단점**

- `SYSTEM SYNC REPLICA`는 `SETTINGS` 절 미지원 → `SET receive_timeout` 선행 필요
- `ON CLUSTER` 실행 시 원격 노드는 DB 컨텍스트 없음 → `db.table` 완전 지정 필요
- goose의 `database.Store` 인터페이스 변경 시 유지보수 필요
- SYNC가 경고로만 처리되므로 극히 짧은 재기동 window에서는 여전히 race condition 가능 (노드 장애 시 한정)

---

##### B. insert_quorum 세션 변수 사용

INSERT 전 `SET insert_quorum='all_replicas'`를 설정하여 INSERT 자체가 전 노드 복제 완료 후에 반환되도록 합니다.

```sql
SET insert_quorum='all_replicas';
SET insert_quorum_parallel=0;
INSERT INTO <table> ...;  -- 모든 replica 확인 후 반환
```

**장점**

- INSERT 완료 시점 = 전 노드 복제 완료 → 별도 SYNC 쿼리 불필요
- ClickHouse 네이티브 메커니즘 사용
- `SYSTEM SYNC REPLICA`의 문법 제약(`SETTINGS` 절, DB 컨텍스트 등) 없음

**단점**

- INSERT 자체가 블로킹 → 모든 노드 확인까지 INSERT 응답 지연
- 노드 일부 장애 시 INSERT 자체가 실패 (quorum 미충족)
- ClickHouse 버전에 따라 `all_replicas` 지원 여부 다름 (23.x 이상 권장, 24.x 이상 안정)

---

##### SYSTEM SYNC REPLICA vs insert_quorum 요약

| 항목                 | SYSTEM SYNC REPLICA (현재)                       | insert_quorum                  |
| -------------------- | ------------------------------------------------ | ------------------------------ |
| 동기화 시점          | INSERT 후 별도 쿼리                              | INSERT 자체가 블로킹           |
| INSERT 성능 영향     | 없음 (INSERT는 즉시 반환)                        | 있음 (모든 노드 확인까지 대기) |
| 적용 범위            | 버전 테이블 INSERT만                             | 세션 내 모든 INSERT            |
| 노드 장애 시         | 10초 후 경고 로그만 출력, 마이그레이션 계속 진행 | INSERT 자체 실패               |
| 구현 복잡도          | 중간 (SET 선행 + db.table 완전 지정 필요)        | 낮음                           |
| ClickHouse 버전 요건 | 없음                                             | 23.x 이상 권장                 |

---

#### 대안 아키텍처

Custom Store 방식 자체를 사용하지 않는 아키텍처 대안입니다.

##### C. ClickHouse 접속 노드 고정

DSN 호스트를 특정 노드 하나로 고정합니다.

**장점**

- 코드/로직 변경 없음, 설정만으로 해결
- 마이그레이션 흐름이 단순 명확

**단점**

- 고정 노드 장애 시 SPOF: 나머지 노드가 정상이어도 마이그레이션 실패로 앱 기동 불가
- 장애 시 대응 방향이 불명확
  - 마이그레이션 실패를 무시하고 진행 → 스키마 미보장 상태로 서비스 기동
  - 마이그레이션 실패 시 기동 중단 → 가용성 훼손
- 로드밸런서/서비스 디스커버리의 HA 이점 상실

---

##### D. Kubernetes Job으로 마이그레이션 분리

별도 Kubernetes Job(또는 Helm pre-install hook, ArgoCD Sync Wave)으로 마이그레이션을 실행합니다.

**장점**

- 앱 파드 재기동과 마이그레이션 실행이 완전히 분리 → race condition 원천 차단
- Job은 단일 파드로 1회 실행 보장 → 동시 마이그레이션 시도 없음
- 비동기 복제로도 충분 (재기동 시 마이그레이션을 실행하는 파드가 없으므로)

**단점**

- 배포 파이프라인 복잡도 증가 (Helm hook, ArgoCD Wave, Job 생명주기 관리 필요)
- Job 실패 시 배포 전체가 블로킹, 별도 운영 절차 필요
- 현재 아키텍처(앱 내 자동 실행) 전체 변경 필요

### 설정 방법

`statistics.database.migration.clickhouse.cluster_name` 키를 통해 활성화합니다.
`cluster_name`이 비어 있으면 기존 단일 노드 동작(`MergeTree`)을 그대로 유지합니다.

---

## 주의사항

⚠️ **스키마 마이그레이션은 앱 시작 전에 실행됩니다**

- 데이터베이스 연결이 필요함
- 실패 시 애플리케이션이 시작되지 않음
- 테이블이 없으면 data 마이그레이션이 실패함

⚠️ **제품별/드라이버별로 분리되어 있습니다**

- master/base - 마스터 서버 기본 테이블
- master/statistics - 통계 데이터 테이블
- agent/base - 에이전트 테이블

## 참고

- 상세 가이드: [MIGRATION_ARCHITECTURE.md](../../../../../docs/MIGRATION_ARCHITECTURE.md)
- Data Migrations: [../data/README.md](../data/README.md)
- API Migrations: [../api/README.md](../api/README.md)
