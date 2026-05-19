package gin

import (
	"context"
	"log/slog"
	"strings"
)

type slogWriter struct {
	logger *slog.Logger
	level  slog.Level
}

func (w slogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSuffix(string(p), "\n")
	w.logger.Log(context.Background(), w.level, msg)
	return len(p), nil
}
