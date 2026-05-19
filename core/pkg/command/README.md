# Command Package

## 개요

Command 패키지는 Pharos ETL 시스템의 CLI 명령어 등록 및 관리를 담당하는 경량 레지스트리 모듈입니다. [Cobra](https://github.com/spf13/cobra) 기반의 명령어를 중앙에서 관리하여 확장 모듈이 자신의 CLI 명령어를 메인 애플리케이션에 등록할 수 있도록 합니다.

### 주요 기능

- **명령어 레지스트리**: Extension에서 생성한 Cobra 명령어를 중앙 저장소에 등록
- **스레드 안전성**: Concurrent map을 사용한 안전한 명령어 관리
- **동적 로딩**: 런타임에 명령어 등록 및 조회 지원
- **확장 가능성**: 새로운 Extension 추가 시 명령어 자동 통합

## 아키텍처

### 구조도

```
┌─────────────────────────────────────────────────────────────┐
│                      Main Application                       │
│                     (cmd/pharos/main.go)                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ GetCommands()
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Command Registry                         │
│                (pkg/command/command.go)                     │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐  │
│  │         Thread-Safe Map[string]*cobra.Command       │  │
│  │  ┌─────────────┬─────────────┬─────────────────┐   │  │
│  │  │ "catv-etl"  │ "another"   │ "custom-cmd"    │   │  │
│  │  │   Command   │   Command   │   Command       │   │  │
│  │  └─────────────┴─────────────┴─────────────────┘   │  │
│  └─────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              ▲
                              │ Register(name, cmd)
                              │
┌─────────────────────────────────────────────────────────────┐
│                        Extensions                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │    CATV      │  │   Pharos     │  │   Custom     │     │
│  │  Extension   │  │  Extension   │  │  Extension   │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

### 데이터 흐름

1. **등록 단계 (init 함수)**
   ```
   Extension init() → command.Register() → Thread-Safe Map 저장
   ```

2. **조회 단계 (main 함수)**
   ```
   main() → command.GetCommands() → Map 반환 → rootCmd.AddCommand()
   ```

## API 참조

### Register

Extension의 명령어를 레지스트리에 등록합니다.

```go
func Register(name string, command *cobra.Command)
```

**파라미터:**
- `name` (string): 명령어 식별자 (로깅 및 디버깅용)
- `command` (*cobra.Command): 등록할 Cobra 명령어

**특징:**
- 스레드 안전
- 중복 이름 허용 (마지막 등록이 우선)
- nil 명령어 등록 가능 (main에서 필터링)

**예제:**
```go
func init() {
    myCmd := &cobra.Command{
        Use:   "mycommand",
        Short: "My custom command",
        Run: func(cmd *cobra.Command, args []string) {
            // 명령어 로직
        },
    }
    
    command.Register("my-extension-cmd", myCmd)
}
```

### GetCommands

등록된 모든 명령어를 조회합니다.

```go
func GetCommands() map[string]*cobra.Command
```

**반환값:**
- `map[string]*cobra.Command`: 명령어 이름과 Cobra 명령어 맵

**특징:**
- 스레드 안전
- 원본 맵 반환 (수정 주의)
- nil 명령어 포함 가능

**예제:**
```go
func main() {
    rootCmd := &cobra.Command{Use: "pharos"}
    
    for name, cmd := range command.GetCommands() {
        if cmd == nil {
            slog.Warn("Skipped nil command", "name", name)
            continue
        }
        
        slog.Info("Registering command", "name", name)
        rootCmd.AddCommand(cmd)
    }
    
    rootCmd.Execute()
}
```

## 사용 가이드

### Extension에서 명령어 등록

Extension의 `init()` 함수에서 명령어를 등록합니다.

```go
package myextension

import (
    "github.com/spf13/cobra"
    "ntels.com/pharos/core/pkg/command"
)

func init() {
    // 명령어 생성
    cmd := createMyCommand()
    
    // 레지스트리에 등록
    command.Register("my-extension", cmd)
}

func createMyCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "myext",
        Short: "My extension command",
        Long:  `Detailed description of my extension command`,
        Run: func(cmd *cobra.Command, args []string) {
            // 실행 로직
            fmt.Println("My extension is running!")
        },
    }
}
```

### 실제 사용 예시: CATV Extension

```go
// extensions/catv/extension.go
package catv

import (
    "ntels.com/pharos/core/pkg/command"
    "ntels.com/pharos/extensions/catv/pkg/etl"
)

func init() {
    // CATV ETL 명령어 등록
    command.Register("catv-etl", etl.Command())
}
```

실행:
```bash
$ pharos catv-etl --help
$ pharos catv-etl -c config.toml
```

### 여러 명령어 등록

하나의 Extension에서 여러 명령어를 등록할 수 있습니다.

```go
func init() {
    command.Register("ext-process", createProcessCommand())
    command.Register("ext-migrate", createMigrateCommand())
    command.Register("ext-status", createStatusCommand())
}
```

## 내부 구현

### Thread-Safe Map

Command 패키지는 `internal.Map[T]`을 사용하여 동시성을 보장합니다.

```go
var commands = internal.NewMap[*cobra.Command]()
```

주요 특징:
- Mutex 기반 동기화
- 제네릭 타입 지원
- Get/Set/GetAll 연산 제공

자세한 구현은 [internal/map.go](../../internal/map.go)를 참조하세요.

## 모범 사례

### 1. 명령어 이름 규칙

```go
// ✅ 권장: 확장 이름 포함
command.Register("catv-etl", cmd)
command.Register("pharos-sync", cmd)

// ❌ 비권장: 일반적인 이름
command.Register("process", cmd)  // 충돌 가능성
command.Register("cmd", cmd)      // 모호함
```

### 2. nil 체크

```go
func init() {
    cmd := createCommand()
    if cmd != nil {
        command.Register("my-cmd", cmd)
    }
}
```

### 3. 명령어 생성 함수 분리

```go
// ✅ 권장: 별도 함수로 분리
func init() {
    command.Register("my-ext", createCommand())
}

func createCommand() *cobra.Command {
    cmd := &cobra.Command{...}
    cmd.AddCommand(createSubCommand1())
    cmd.AddCommand(createSubCommand2())
    return cmd
}

// ❌ 비권장: init에 모든 로직
func init() {
    cmd := &cobra.Command{...}
    // 100줄의 코드...
    command.Register("my-ext", cmd)
}
```

### 4. 서브 명령어 구조화

```go
func createCommand() *cobra.Command {
    rootCmd := &cobra.Command{
        Use:   "myext",
        Short: "My extension commands",
    }
    
    // 서브 명령어 추가
    rootCmd.AddCommand(&cobra.Command{
        Use:   "start",
        Short: "Start the service",
        Run:   startHandler,
    })
    
    rootCmd.AddCommand(&cobra.Command{
        Use:   "stop",
        Short: "Stop the service",
        Run:   stopHandler,
    })
    
    return rootCmd
}
```

실행:
```bash
$ pharos myext start
$ pharos myext stop
```

## 디버깅

### 등록된 명령어 확인

Main 애플리케이션의 로그에서 확인:

```bash
$ pharos 2>&1 | grep "Register command"
INFO Register command name=catv-etl
INFO Register command name=pharos-sync
```

### nil 명령어 경고

```bash
$ pharos 2>&1 | grep "Skipped"
WARN Skipped registering nil command name=broken-extension
```

## 확장 시나리오

### 동적 명령어 생성

설정 파일을 기반으로 여러 명령어를 생성:

```go
func init() {
    configs := []struct {
        name string
        desc string
    }{
        {"worker1", "Worker 1 processor"},
        {"worker2", "Worker 2 processor"},
        {"worker3", "Worker 3 processor"},
    }
    
    for _, cfg := range configs {
        cmd := createWorkerCommand(cfg.name, cfg.desc)
        command.Register(cfg.name, cmd)
    }
}
```

### 조건부 명령어 등록

빌드 태그나 환경 변수에 따른 조건부 등록:

```go
func init() {
    if os.Getenv("ENABLE_DEBUG_COMMANDS") == "true" {
        command.Register("debug", createDebugCommand())
    }
    
    command.Register("main", createMainCommand())
}
```

## 관련 문서

- [Extension System](../server/README.md) - Server Extension 등록 시스템
- [Internal Map](../../internal/README.md) - Thread-safe map 구현
- [Cobra Documentation](https://github.com/spf13/cobra) - CLI 프레임워크
- [ETL Commands](../../../extensions/catv/pkg/etl/README.md) - ETL 명령어 예제

## 테스트

### 기본 테스트

```go
package command_test

import (
    "testing"
    
    "github.com/spf13/cobra"
    "ntels.com/pharos/core/pkg/command"
)

func TestRegisterAndGet(t *testing.T) {
    cmd := &cobra.Command{Use: "test"}
    
    command.Register("test-cmd", cmd)
    
    commands := command.GetCommands()
    if commands["test-cmd"] != cmd {
        t.Error("Command not registered correctly")
    }
}

func TestNilCommand(t *testing.T) {
    command.Register("nil-cmd", nil)
    
    commands := command.GetCommands()
    if commands["nil-cmd"] != nil {
        t.Error("Nil command should be stored")
    }
}
```

### 동시성 테스트

```go
func TestConcurrentRegister(t *testing.T) {
    var wg sync.WaitGroup
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            cmd := &cobra.Command{Use: fmt.Sprintf("cmd-%d", id)}
            command.Register(fmt.Sprintf("test-%d", id), cmd)
        }(i)
    }
    
    wg.Wait()
    
    commands := command.GetCommands()
    if len(commands) < 100 {
        t.Errorf("Expected at least 100 commands, got %d", len(commands))
    }
}
```

## FAQ

**Q: 같은 이름으로 두 번 등록하면?**
A: 마지막 등록이 우선됩니다. 경고는 발생하지 않으므로 고유한 이름 사용을 권장합니다.

**Q: 명령어를 동적으로 제거할 수 있나요?**
A: 현재는 지원하지 않습니다. 필요한 경우 `internal.Map`에 `Remove` 메서드를 추가할 수 있습니다.

**Q: Extension 로드 순서가 중요한가요?**
A: `init()` 함수의 실행 순서는 Go의 패키지 초기화 순서에 따르므로 보장되지 않습니다. 명령어 간 의존성이 없도록 설계하는 것이 좋습니다.

**Q: 명령어 이름과 Cobra Use 필드가 달라도 되나요?**
A: 네, 가능합니다. 레지스트리 이름은 로깅/디버깅용이고, 실제 CLI 명령어는 Cobra의 `Use` 필드로 결정됩니다.

## 라이선스

이 프로젝트는 회사 내부 사용을 위한 것입니다.
