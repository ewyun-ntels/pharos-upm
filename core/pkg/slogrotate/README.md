##  SLOG 로그 파일 로테이션 기능 
slog를 사용하는 Application의 로그 파일 관리 기능을 제공

### 주요 기능
``` shell
1. 전역 로그 파일 로테이션 기능 제공
2. 개별 로그 파일 로테이션 기능 제공
3. 콘솔 로그 기능 제공
4. 서브 디렉토리 로그 파일 로테이션 기능 제공
5. gin web framework 로그 파일 로테이션 기능 제공
```

## 설정 정보
/config/config.toml
```shell
# Logger Rotation config
[logger]
# Enable console output
console = true
# Enable JSON format change
json = false
# Log path (e.g. "./logs")
logPath = "../../../../../logs"
# MB before rotation
maxSize = 10 
# Max old files to retain
maxBackups = 15
# Days to retain old files
maxAge = 15 
# Gzip compress rotated files
compress = false # true     
# Duration string for forced rotation (e.g. "24h", "2h30m")   
rotationInterval = "24h"
```

## 전역 로그 관리
/examples/history/chat/main.go

SetupGlobal(opts *slog.HandlerOptions, isJSON, isConsole bool)
```golang
import "ntels.com/pharos/pkg/slogrotate"

cfg, _ := config_loader.Load("/tarzan/config/config.toml")
rotate := slogrotate.New(&cfg, "main process name")

globalLogger, _ := rotate.SetupGlobal(&slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}, false, false)
go func() {
    for {
        slog.Info("Starting global_log !!!!")
        time.Sleep(50 * time.Millisecond)
    }
}()    
```

## 개별 / 서브 디렉토리 로그 관리
/examples/history/chat/main.go

NewSubdirLogger(subdir, logName string, opts *slog.HandlerOptions, isJSON bool)
```golang
import "ntels.com/pharos/pkg/slogrotate"

cfg, _ := config_loader.Load("/tarzan/config/config.toml")
rotate := slogrotate.New(&cfg, "main process name")

indiviLogger, _ := rotate.NewSubdirLogger("/Sub", "individuation_log", &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}, false)

go func() {
    for {
        indiviLogger.Info("Starting individuation_log !!!!")
        time.Sleep(50 * time.Millisecond)
    }
}()
```

## 콘솔 로그 관리
/examples/history/chat/main.go

NewConsoleLogger(opts *slog.HandlerOptions, isJSON bool)
```golang
import "ntels.com/pharos/pkg/slogrotate"

cfg, _ := config_loader.Load("/tarzan/config/config.toml")
rotate := slogrotate.New(&cfg, "main process name")

consoleLogger := rotate.NewConsoleLogger(&slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}, false)

go func() {
    for {
        consoleLogger.Info("Starting console_log !!!!")
        time.Sleep(1 * time.Second)
    }
}()
```

## GIN 로그 관리 
/examples/history/chat/main.go
```golang
import "ntels.com/pharos/pkg/slogrotate"

type SlogWriter struct {
	logger *slog.Logger
	level  slog.Level
}

func (w SlogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSuffix(string(p), "\n")
	w.logger.Log(context.Background(), w.level, msg)
	return len(p), nil
}

cfg, _ := config_loader.Load("/tarzan/config/config.toml")
rotate := slogrotate.New(&cfg, "main process name")

globalLogger, _ := rotate.SetupGlobal(&slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}, false, false)

// gin.Default() 이후에 아래 두 줄을 설정해야 gin 로그가 파일로만 나갑니다.
gin.DefaultWriter = SlogWriter{logger: globalLogger, level: slog.LevelInfo}
gin.DefaultErrorWriter = SlogWriter{logger: globalLogger, level: slog.LevelError}

r := gin.New()
r.Use(SlogMiddleware(globalLogger), SlogRecovery(globalLogger))

func SlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.RequestURI
		c.Next() 

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("http request",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}

func SlogRecovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic recovered",
					slog.Any("panic", rec),
					slog.String("path", c.Request.RequestURI),
				)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
```

##  SLOG 로그 파일 로테이션 기능 
slog를 사용하는 Application의 로그 파일 관리 기능을 제공

### 주요 기능
``` shell
1. 전역 로그 파일 로테이션 기능 제공
2. 개별 로그 파일 로테이션 기능 제공
3. 콘솔 로그 기능 제공
4. 서브 디렉토리 로그 파일 로테이션 기능 제공
5. gin web framework 로그 파일 로테이션 기능 제공
```

## 설정 정보
/config/config.toml
```shell
# Logger Rotation config
[logger]
# Enable console output
console = true
# Enable JSON format change
json = false
# Log path (e.g. "./logs")
logPath = "../../../../../logs"
# MB before rotation
maxSize = 10 
# Max old files to retain
maxBackups = 15
# Days to retain old files
maxAge = 15 
# Gzip compress rotated files
compress = false # true     
# Duration string for forced rotation (e.g. "24h", "2h30m")   
rotationInterval = "24h"
```

## 전역 로그 관리
/examples/history/chat/main.go

SetupGlobal(opts *slog.HandlerOptions, isJSON, isConsole bool)
```golang
import "ntels.com/pharos/pkg/slogrotate"

cfg, _ := config_loader.Load("/tarzan/config/config.toml")
rotate := slogrotate.New(&cfg, "main process name")

globalLogger, _ := rotate.SetupGlobal(&slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}, false, false)
go func() {
    for {
        slog.Info("Starting global_log !!!!")
        time.Sleep(50 * time.Millisecond)
    }
}()    
```

## 개별 / 서브 디렉토리 로그 관리
/examples/history/chat/main.go

NewSubdirLogger(subdir, logName string, opts *slog.HandlerOptions, isJSON bool)
```golang
import "ntels.com/pharos/pkg/slogrotate"

cfg, _ := config_loader.Load("/tarzan/config/config.toml")
rotate := slogrotate.New(&cfg, "main process name")

indiviLogger, _ := rotate.NewSubdirLogger("/Sub", "individuation_log", &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}, false)

go func() {
    for {
        indiviLogger.Info("Starting individuation_log !!!!")
        time.Sleep(50 * time.Millisecond)
    }
}()
```

## 콘솔 로그 관리
/examples/history/chat/main.go

NewConsoleLogger(opts *slog.HandlerOptions, isJSON bool)
```golang
import "ntels.com/pharos/pkg/slogrotate"

cfg, _ := config_loader.Load("/tarzan/config/config.toml")
rotate := slogrotate.New(&cfg, "main process name")

consoleLogger := rotate.NewConsoleLogger(&slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}, false)

go func() {
    for {
        consoleLogger.Info("Starting console_log !!!!")
        time.Sleep(1 * time.Second)
    }
}()
```

## GIN 로그 관리 
/examples/history/chat/main.go
```golang
import "ntels.com/pharos/pkg/slogrotate"

type SlogWriter struct {
	logger *slog.Logger
	level  slog.Level
}

func (w SlogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSuffix(string(p), "\n")
	w.logger.Log(context.Background(), w.level, msg)
	return len(p), nil
}

cfg, _ := config_loader.Load("/tarzan/config/config.toml")
rotate := slogrotate.New(&cfg, "main process name")

globalLogger, _ := rotate.SetupGlobal(&slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}, false, false)

// gin.Default() 이후에 아래 두 줄을 설정해야 gin 로그가 파일로만 나갑니다.
gin.DefaultWriter = SlogWriter{logger: globalLogger, level: slog.LevelInfo}
gin.DefaultErrorWriter = SlogWriter{logger: globalLogger, level: slog.LevelError}

r := gin.New()
r.Use(SlogMiddleware(globalLogger), SlogRecovery(globalLogger))

func SlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.RequestURI
		c.Next() 

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("http request",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}

func SlogRecovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic recovered",
					slog.Any("panic", rec),
					slog.String("path", c.Request.RequestURI),
				)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
```

## console = false 인데 콘솔로 출력이 나오는 이유와 해결 방법
- 원인: 일부 컴포넌트가 slog를 거치지 않고 직접 콘솔로 출력하고 있었습니다.
  - HashiCorp 플러그인 로더(hclog)가 기본으로 os.Stdout에 출력
  - 플러그인 서브프로세스의 stdout/stderr가 부모 프로세스 콘솔로 전파
- 개선: console=false일 때 다음과 같이 동작합니다.
  - 플러그인 로더의 hclog Output을 io.Discard로 설정하여 콘솔 출력 억제
  - go-plugin 실행 커맨드(exec.Cmd)의 Stdout/Stderr를 io.Discard로 설정하여 플러그인 프로세스 출력 억제
  - Gin 로그는 gin.DefaultWriter, gin.DefaultErrorWriter를 slog 기반 writer로 대체

설정 요약:
```toml
[logger]
console = false
json = true # 또는 false
logPath = "./logs"
```

주의 사항:
- 애플리케이션 시작 초기에 생성되는 일시적인 로그는 기본 stdout으로 나갈 수 있으나, 서버 초기화가 끝나면 모두 파일로만 기록됩니다.
- 별도의 커맨드(예: simulator)는 자체적으로 os.Stdout을 사용할 수 있으므로, 해당 커맨드에서도 slogrotate를 사용하도록 구성해야 합니다.

## 유닛 테스트 (README)
- 테스트 파일: pkg/slogrotate/slogrotate_test.go
- 목적: console=false 설정 시 stdout으로 아무 것도 출력되지 않고, 파일에만 로그가 기록되는지 검증
- 방법 요약:
  - os.Stdout을 파이프로 임시 교체하여 캡처
  - slogrotate.SetupGlobal(console=false) 후 slog.Info 호출
  - 캡처된 stdout이 비어 있는지 확인
  - 로그 파일(tarzan-test.log)에 메시지가 기록되었는지 확인
