package log

import (
	"fmt"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

var std *logrus.Logger

func init() {
	std = logrus.StandardLogger()
	std.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339})
}

func GetLogger() *logrus.Logger {
	return std
}

func SetLogFile(path string, maxAge int) error {
	if std == nil {
		std = logrus.StandardLogger()
	}

	if maxAge <= 0 {
		maxAge = 30
	}

	if path == "" {
		std.SetOutput(os.Stdout)
	} else {
		rl, err := rotatelogs.New(
			fmt.Sprintf("%s.%%Y%%m%%d", path),
			rotatelogs.WithMaxAge(time.Duration(maxAge)*24*time.Hour),
		)
		if err != nil {
			return err
		}

		std.SetOutput(rl)
	}
	return nil
}

func Log(level logrus.Level, args ...interface{}) {
	std.Log(level, args...)
}

func Infof(format string, args ...interface{}) {
	std.Infof(format, args...)
}

func Errorf(format string, args ...interface{}) {
	std.Errorf(format, args...)
}
