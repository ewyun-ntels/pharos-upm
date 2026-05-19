package gnet

import (
	"fmt"
	"log/slog"

	"github.com/panjf2000/gnet/v2/pkg/logging"
)

// SlogGnetAdapter는 slog.Logger를 gnet의 logging.Logger 인터페이스로 변환하는 어댑터입니다.
// gnet은 다음 인터페이스를 요구합니다:
// - Debugf(format string, args ...any)
// - Infof(format string, args ...any)
// - Warnf(format string, args ...any)
// - Errorf(format string, args ...any)
// - Fatalf(format string, args ...any)
type SlogGnetAdapter struct {
	logger *slog.Logger
}

// NewSlogGnetAdapter는 slog.Logger를 래핑하여 gnet logging.Logger 인터페이스를 구현하는 어댑터를 생성합니다.
func NewSlogGnetAdapter(logger *slog.Logger) logging.Logger {
	return &SlogGnetAdapter{
		logger: logger,
	}
}

// Debugf는 DEBUG 레벨로 로그를 기록합니다.
func (a *SlogGnetAdapter) Debugf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.logger.Debug(msg, slog.String("source", "gnet"))
}

// Infof는 INFO 레벨로 로그를 기록합니다.
func (a *SlogGnetAdapter) Infof(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.logger.Info(msg, slog.String("source", "gnet"))
}

// Warnf는 WARN 레벨로 로그를 기록합니다.
func (a *SlogGnetAdapter) Warnf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.logger.Warn(msg, slog.String("source", "gnet"))
}

// Errorf는 ERROR 레벨로 로그를 기록합니다.
func (a *SlogGnetAdapter) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.logger.Error(msg, slog.String("source", "gnet"))
}

// Fatalf는 ERROR 레벨로 로그를 기록합니다.
// 주의: slog에는 Fatal 레벨이 없으므로 Error 레벨을 사용하고 panic을 발생시킵니다.
func (a *SlogGnetAdapter) Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	a.logger.Error(msg, slog.String("source", "gnet"), slog.String("level", "fatal"))
	panic(msg)
}
