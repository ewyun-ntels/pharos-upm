# 설정 파일 레퍼런스

설정 파일은 **TOML** 형식이며, 환경변수로 오버라이드 가능합니다 (Viper 규칙: `.` → `_`, 대문자).

기본 설정 파일: `migrations/catv_quality/config/config.toml`  
참고 설정 파일: `core/config/master_config.toml`

---

## `[serve]`

```toml
[serve]
server_schema = "http"    # 또는 "https"
nats_schema   = "nats"    # 또는 "tls"
```

---

## `[servers.*]`

```toml
[servers.http]
port = 8080

[servers.https]
port = 8443
tls_min_version = "1.2"  # "1.2" 또는 "1.3"
tls_max_version = "1.3"

[servers.https.cert]
auto_generate = true           # true: 자동 생성, false: 파일 지정
dir = "./core/config/cert/tls"

[servers.nats]
port = 4222

[servers.nats.cert]
auto_generate = true
dir = "./core/config/cert/nats/tls"
insecure_skip_verify = false
```

---

## `[clients.nats]`

```toml
[clients.nats]
url      = "nats://127.0.0.1:4222"
key_file  = ""
cert_file = ""
ca_file   = ""
```

---

## `[client_tls]`

Outbound HTTP/gRPC 연결에 사용되는 클라이언트 TLS 설정.

```toml
[client_tls]
insecure_skip_verify = false  # 운영환경에서 반드시 false
min_version = "1.2"
max_version = "1.3"
```

---

## `[database]`

메타 DB (사용자, 대시보드, 권한 등).

```toml
[database]
driver = "sqlite"             # "sqlite" 또는 "postgresql"

[database.sqlite]
path = "./data/pharos.db"

# OR

[database.postgresql]
host     = "localhost"
port     = 5432
username = "pharos"
password = "password"
database = "pharos"
```

---

## `[auth]`

```toml
[auth]
global_secret         = "your-secret-key"       # HMAC 시크릿
access_token_lifespan  = "1h"                    # 액세스 토큰 유효기간
refresh_token_lifespan = "720h"                  # 리프레시 토큰 유효기간 (30일)
jwt_issuer             = "pharos"
cookie_mode            = false                   # true: 쿠키 기반 인증
csrf_protection        = false                   # cookie_mode=true 시 활성화 권장
secure_cookies         = false                   # HTTPS 환경에서 true
authenticate_hint      = 1                       # 1: 차단, 2: 전체 힌트

[auth.jwt_cert]
auto_generate = true
jwt_dir       = "./core/config/cert/jwt"

[[auth.clients]]
id     = "pharos-frontend"
secret = "client-secret"
```

---

## `[users]`

```toml
[users]
login_retry_limit          = 5     # 로그인 실패 허용 횟수
login_retry_reset_timeout  = 300   # 실패 카운터 리셋 시간 (초)
login_block_duration       = 1800  # 잠금 유지 시간 (초)
password_history_retention = 5     # 이전 비밀번호 재사용 금지 개수

[users.password_rule]
min_length    = 8
min_uppercase = 1
min_lowercase = 1
min_digits    = 1
min_special   = 1
password_ttl  = "90d"              # 비밀번호 만료 주기

[users.user_rule]
email_allowed = false
min_length    = 4
max_length    = 32
```

---

## `[master.statistics.database]`

Master 서버 통계 DB (ClickHouse 주로 사용).

```toml
[master.statistics.database]
driver = "clickhouse"

[master.statistics.database.clickhouse]
host               = "clickhouse-host"
port               = 9000              # TCP 포트
http_port          = 8123              # HTTP 포트
username           = "default"
password           = ""
database           = "catv"
max_open_connection = 10
max_lifetime       = 3600
```

---

## `[migration]`

```toml
[migration]
disable       = false
migration_dir = "./migrations/catv_quality/migration"
artifact_dir  = "./migrations/catv_quality/artifact"
# wait = "2s"   # 마이그레이션 시작 전 대기 (의존 서비스 준비 대기)

[[migration.env]]
name  = "CLICKHOUSE_HOST"
value = "clickhouse-host"

[[migration.env]]
name         = "CLICKHOUSE_PASSWORD"
valueFromEnv = "CH_PASSWORD"    # 호스트 환경변수에서 읽기
```

---

## `[nats]`

```toml
[nats]
max_payload            = 1048576  # 1MB
jetstream              = true
jetstream_max_memory   = 1073741824  # 1GB
jetstream_max_store    = 10737418240 # 10GB
store_dir              = "./data/nats"

[nats.cluster]
name   = "pharos-cluster"
listen = "0.0.0.0:6222"

[nats.cluster.routes]
# 클러스터 내 다른 NATS 서버 주소
# routes = ["nats://node2:6222", "nats://node3:6222"]
```

---

## `[pool]`

```toml
[pool.workers]
worker_count    = 10   # 워커 고루틴 수
job_queue_size  = 100  # 작업 큐 크기
batch_size      = 50   # 배치 처리 크기
enable_metrics  = true

[pool.connection]
max_size     = 100    # 최대 커넥션 수
shard_count  = 16     # 커넥션 풀 샤드 수
max_idle_time = "10m" # 유휴 커넥션 제거 주기
```

---

## `[catv]`

CATV Extension 전용 설정.

```toml
[catv.database]
driver = "clickhouse"
# [catv.database.clickhouse] 하위에 ClickHouse 접속 정보 기재

# UDP 수집 포트
[catv.collect.transmission.periodic]
port = 9001

[catv.collect.transmission.daily]
port = 9002

[catv.collect.transmission.diagnostic]
port = 9003

[catv.collect.transmission.quality_measurement]
port = 9004

[catv.collect.transmission.network_quality_transition]
port = 9005

# 기상 데이터 수집
[catv.collect.weather]
simulation     = false
address        = "ftp.kma.go.kr"
user           = "anonymous"
password       = ""
cron_spec      = "0 30 6 * * *"    # 매일 06:30:00
base_directory = "/pub/forecast/gridded"

[catv.collect.weather.timeout]
connection = "30s"
shut       = "5s"

# STB 원격 제어
[catv.control]
simulation              = false    # true: 실제 STB 연결 없이 시뮬레이션
port                    = 0        # 사용하지 않음 (K8s Job 내부에서만 사용)
max_connections_per_second = 100   # Rate Limiter

[catv.control.batch]
max_count      = 1000
max_size       = 4096
flush_interval = "1s"

[catv.control.timeout]
connection = "5s"
send       = "10s"

[catv.control.retry.total]
count = 3
delay = "1s"

[catv.control.retry.port_exhaustion]
count = 5
delay = "500ms"

[catv.control.k8s_job]
count                    = 5         # 병렬 Job 파드 수
namespace                = "catv"
active_deadline_seconds  = 3600
backoff_limit            = 0         # 실패 시 재시도 없음
ttl_seconds_after_finished = 86400   # Job 완료 후 1일 후 자동 삭제
image                    = "ntels.harbor.core/pharos/pharos-catv-dev:latest"
service_account_name     = "catv-job-sa"
allowed_env_vars         = ["CONFIG_PATH", "LOG_LEVEL"]
```

---

## `[metrics]`

```toml
[metrics]
use = true   # false: /metrics 엔드포인트 비활성화
```

---

## `[etl]`

```toml
[etl.elasticsearch]
host     = "https://elasticsearch:9200"
username = "elastic"
password = "password"
```

---

## `[logger]`

```toml
[logger]
level             = "info"     # debug / info / warn / error
json              = true       # JSON 로그 포맷
console           = true       # 콘솔 출력
file_name         = "pharos.log"
log_path          = "./log"
max_size          = 100        # MB
max_backups       = 5
max_age           = 30         # days
compress          = true       # gzip 압축
rotation_interval = "24h"      # 강제 로테이션 주기
```

---

## `[daemon]`

```toml
[daemon]
pid_file       = "./data/pharos.pid"
log_file       = "./log/pharos.log"
restart_policy = "on-failure"  # "no" / "on-failure" / "always"
max_restarts   = 5             # 0 = 무제한
restart_delay  = 5             # 초
```

---

## 환경변수 오버라이드 예시

Viper는 설정 키의 `.`을 `_`로, 소문자를 대문자로 변환한 환경변수를 우선 적용합니다.

```bash
# [catv.database.clickhouse] host 오버라이드
export CATV_DATABASE_CLICKHOUSE_HOST="clickhouse.prod.internal"

# [auth] global_secret 오버라이드
export AUTH_GLOBAL_SECRET="prod-secret-key"

# [master.statistics.database.clickhouse] password 오버라이드
export MASTER_STATISTICS_DATABASE_CLICKHOUSE_PASSWORD="secret"
```
