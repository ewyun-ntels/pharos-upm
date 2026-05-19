package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats-server/v2/server"
)

// SlogNatsAdapter는 slog.Logger를 NATS 서버의 Logger 인터페이스로 변환하는 어댑터입니다.
// NATS 서버는 다음 인터페이스를 요구합니다:
// - Noticef(format string, v ...any)
// - Warnf(format string, v ...any)
// - Fatalf(format string, v ...any)
// - Errorf(format string, v ...any)
// - Debugf(format string, v ...any)
// - Tracef(format string, v ...any)
type SlogNatsAdapter struct {
	logger *slog.Logger
}

// NewSlogNatsAdapter는 slog.Logger를 래핑하여 NATS Logger 인터페이스를 구현하는 어댑터를 생성합니다.
func NewSlogNatsAdapter(logger *slog.Logger) server.Logger {
	return &SlogNatsAdapter{
		logger: logger,
	}
}

// Noticef는 INFO 레벨로 로그를 기록합니다.
func (a *SlogNatsAdapter) Noticef(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	a.logger.Info(msg, slog.String("source", "nats"))
}

// Warnf는 WARN 레벨로 로그를 기록합니다.
func (a *SlogNatsAdapter) Warnf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	a.logger.Warn(msg, slog.String("source", "nats"))
}

// Fatalf는 ERROR 레벨로 로그를 기록합니다.
// 주의: slog에는 Fatal 레벨이 없으므로 Error 레벨을 사용하고 panic을 발생시킵니다.
func (a *SlogNatsAdapter) Fatalf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	a.logger.Error(msg, slog.String("source", "nats"), slog.String("level", "fatal"))
	panic(msg)
}

// Errorf는 ERROR 레벨로 로그를 기록합니다.
func (a *SlogNatsAdapter) Errorf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	a.logger.Error(msg, slog.String("source", "nats"))
}

// Debugf는 DEBUG 레벨로 로그를 기록합니다.
func (a *SlogNatsAdapter) Debugf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	a.logger.Debug(msg, slog.String("source", "nats"))
}

// Tracef는 DEBUG 레벨로 로그를 기록합니다.
// 주의: slog에는 TRACE 레벨이 없으므로 DEBUG 레벨을 사용하고 trace 태그를 추가합니다.
func (a *SlogNatsAdapter) Tracef(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	a.logger.Debug(msg, slog.String("source", "nats"), slog.String("level", "trace"))
}

// SlogNatsAdapterWithContext는 context를 사용하는 향상된 어댑터입니다.
// Context를 통해 trace ID 등의 정보를 전달할 수 있습니다.
type SlogNatsAdapterWithContext struct {
	logger *slog.Logger
	ctx    context.Context
}

// NewSlogNatsAdapterWithContext는 context를 사용하는 어댑터를 생성합니다.
func NewSlogNatsAdapterWithContext(logger *slog.Logger, ctx context.Context) server.Logger {
	if ctx == nil {
		ctx = context.Background()
	}
	return &SlogNatsAdapterWithContext{
		logger: logger,
		ctx:    ctx,
	}
}

// Noticef는 INFO 레벨로 로그를 기록합니다.
func (a *SlogNatsAdapterWithContext) Noticef(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	a.logger.InfoContext(a.ctx, msg, slog.String("source", "nats"))
}

// Warnf는 WARN 레벨로 로그를 기록합니다.
func (a *SlogNatsAdapterWithContext) Warnf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	a.logger.WarnContext(a.ctx, msg, slog.String("source", "nats"))
}

// Fatalf는 ERROR 레벨로 로그를 기록하고 panic을 발생시킵니다.
func (a *SlogNatsAdapterWithContext) Fatalf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	a.logger.ErrorContext(a.ctx, msg, slog.String("source", "nats"), slog.String("level", "fatal"))
	panic(msg)
}

// Errorf는 ERROR 레벨로 로그를 기록합니다.
func (a *SlogNatsAdapterWithContext) Errorf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	a.logger.ErrorContext(a.ctx, msg, slog.String("source", "nats"))
}

// Debugf는 DEBUG 레벨로 로그를 기록합니다.
func (a *SlogNatsAdapterWithContext) Debugf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	a.logger.DebugContext(a.ctx, msg, slog.String("source", "nats"))
}

// Tracef는 DEBUG 레벨로 로그를 기록합니다.
func (a *SlogNatsAdapterWithContext) Tracef(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	a.logger.DebugContext(a.ctx, msg, slog.String("source", "nats"), slog.String("level", "trace"))
}
