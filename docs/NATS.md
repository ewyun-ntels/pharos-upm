# NATS 통신 시스템

Pharos의 NATS 기반 Master-Agent 통신 설정 및 사용 가이드

## 🚀 빠른 시작 (5분 설정)

### 1. 빌드
```bash
cd /path/to/pharos/core
go build -o bin/pharos cmd/pharos/main.go
```

### 2. Master 설정
```toml
# config/master_config.toml
[serve]
type = "master"
nats_schema = "nats"  # 개발용 (TLS 비활성)

[servers.nats]
port = 4222

[nats]
jetstream = true
jetstream_max_memory = 268435456  # 256MB
store_dir = "./nats/storage"
```

### 3. Agent 설정
```toml
# config/agent_config.toml
[serve]
type = "agent"

[clients.nats]
url = "nats://MASTER_IP:4222"  # Master IP로 변경

[agent]
host = "http://:8081"
master_host = "http://MASTER_IP:8080"
```

### 4. 실행 및 테스트
```bash
# 터미널 1: Master
./bin/pharos serve -c config/master_config.toml

# 터미널 2: Agent  
./bin/pharos serve -c config/agent_config.toml

# 로그 확인: "NATS server started", "Agent connected" 메시지 확인
```

## 🔒 운영 환경 설정 (TLS 보안)

### Master (운영용)
```toml
[serve]
type = "master"
nats_schema = "tls"  # TLS 활성화

[servers.nats]
port = 4222

[servers.nats.cert]
auto_generate = true
dir = "./core/config/cert/nats/tls"
insecure_skip_verify = false

[servers.nats.cert.auto_generate_info]
common_name = 'pharos-master'
organization = ['your-company.com']
subject_alternate_name_ips = ["127.0.0.1", "YOUR_MASTER_IP"]

[nats]
jetstream = true
jetstream_max_memory = 1073741824  # 1GB
store_dir = "./nats/storage"

# 기능 활성화/비활성화
[nats.subjects]
custom_metrics = true
system_info = true
log_collection = true
file_transfer = false      # 보안상 비활성화
command_execution = false  # 보안상 비활성화
```

### Agent (운영용)
```toml
[serve]
type = "agent"

[clients.nats]
url = "tls://MASTER_IP:4222"  # TLS 사용
ca_file = "./config/cert/nats/ca.crt"

[agent]
host = "http://:8081"
master_host = "https://MASTER_IP:8443"
```

## � 주요 기능

### 기본 Subject들
- `agent.ping.{name}` - 연결 상태 확인
- `agent.collect.information.health.{name}` - 헬스체크 수집
- `agent.collect.information.service_database.{name}` - DB 상태 수집

### 확장 Subject들 (선택적)
- `agent.collect.custom.metrics.{name}` - 커스텀 메트릭 (`custom_metrics = true`)
- `agent.collect.system.info.{name}` - 시스템 정보 (`system_info = true`)
- `agent.collect.logs.{name}` - 로그 수집 (`log_collection = true`)
- `agent.file.transfer.{name}` - 파일 전송 (`file_transfer = true`)

## 🛠️ 트러블슈팅

### 연결 안 됨
```bash
# 포트 확인
netstat -tlnp | grep 4222

# Agent 설정 확인  
grep "url.*nats" config/agent_config.toml
```

### TLS 오류
```bash
# 인증서 재생성
rm -rf core/config/cert/nats/tls/*
./bin/pharos serve -c config/master_config.toml
```

### 상태 확인
```bash
# 프로세스 확인
ps aux | grep pharos

# 로그 모니터링
tail -f pharos.log | grep -E "(NATS|Agent)"
```

## 🔧 커스텀 Subject 추가

새로운 Subject를 추가하려면 `core/pkg/server/nats/subjects/` 폴더에 파일 추가:

```go
// agent_custom_feature.go
package subjects

func GetAgentCustomFeatureSubject(config common.Config) core_nats.Subject {
    return core_nats.Subject{
        Key:            core_nats.AgentCustomFeature,
        ServeType:      common.ServeTypeAgent,
        RequestSubject: "agent.custom.feature.{{.name}}",
        RequestHandler: func(msg *nats.Msg) {
            // 요청 처리 로직
            msg.Respond([]byte(`{"status": "success"}`))
        },
    }
}
```

## 🔍 보안 체크리스트

- [ ] 운영환경: `nats_schema = "tls"` 사용
- [ ] `insecure_skip_verify = false` 설정
- [ ] 방화벽: 4222 포트만 허용
- [ ] 위험 기능 비활성화: `file_transfer = false`, `command_execution = false`
- [ ] 정기 인증서 갱신 자동화

---

**📞 문의:** [GitHub Issues](../../issues) | **개선:** [Pull Requests](../../pulls)