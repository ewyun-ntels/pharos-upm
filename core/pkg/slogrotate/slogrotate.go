package slogrotate

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
	"ntels.com/pharos/core/pkg/common"
)

type InternalRotationConfig struct {
	Filename         string // Log file path (e.g. "./logs/app.log")
	MaxSize          int    // MB before rotation
	MaxBackups       int    // Max old files to retain
	MaxAge           int    // Days to retain old files
	Compress         bool   // Gzip compress rotated files
	RotationInterval string // Duration string for forced rotation (e.g. "24h", "2h30m")
}

func New(cfg *common.Config) *InternalRotationConfig {
	def := defaultRotationConfig()
	def.merge(cfg.Logger)

	return def
}

func (logger *InternalRotationConfig) String() string {
	return "slogrotate: " +
		"Filename: " + logger.Filename +
		", MaxSize: " + strconv.Itoa(logger.MaxSize) +
		", MaxBackups: " + strconv.Itoa(logger.MaxBackups) +
		", MaxAge: " + strconv.Itoa(logger.MaxAge) +
		", Compress: " + strconv.FormatBool(logger.Compress) +
		", RotationInterval: " + logger.RotationInterval
}

func defaultRotationConfig() *InternalRotationConfig {
	return &InternalRotationConfig{
		Filename:         "./logs/app.log",
		MaxSize:          100,
		MaxBackups:       7,
		MaxAge:           30,
		Compress:         false,
		RotationInterval: "24h",
	}
}

func (logger *InternalRotationConfig) copy() *InternalRotationConfig {
	return &InternalRotationConfig{
		Filename:         logger.Filename,
		MaxSize:          logger.MaxSize,
		MaxBackups:       logger.MaxBackups,
		MaxAge:           logger.MaxAge,
		Compress:         logger.Compress,
		RotationInterval: logger.RotationInterval,
	}
}

func (logger *InternalRotationConfig) merge(cfg common.RotationConfig) {

	logger.Filename = filepath.Join(cfg.LogPath, cfg.FileName)

	if cfg.MaxSize != 0 {
		logger.MaxSize = cfg.MaxSize
	}
	if cfg.MaxBackups != 0 {
		logger.MaxBackups = cfg.MaxBackups
	}
	if cfg.MaxAge != 0 {
		logger.MaxAge = cfg.MaxAge
	}
	if cfg.RotationInterval != "" {
		logger.RotationInterval = cfg.RotationInterval
	}
	if cfg.Compress {
		logger.Compress = cfg.Compress
	}
}

func ensureDirExists(path string) error {
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		return nil // current directory already exists
	}
	return os.MkdirAll(dir, 0o750)
}

// parseInterval parses the RotationInterval string to time.Duration.
func parseInterval(intervalStr string) (time.Duration, error) {
	d, err := time.ParseDuration(intervalStr)
	if err != nil {
		return 0, errors.New("invalid rotationInterval: " + intervalStr + ", must be like \"24h\" or \"2h30m\"")
	}
	return d, nil
}

func (logger *InternalRotationConfig) newFileSizeWriter() (io.Writer, error) {
	if err := ensureDirExists(logger.Filename); err != nil {
		return nil, err
	}
	return &lumberjack.Logger{
		Filename:   logger.Filename,
		MaxSize:    logger.MaxSize,
		MaxBackups: logger.MaxBackups,
		MaxAge:     logger.MaxAge,
		Compress:   logger.Compress,
	}, nil
}

func (logger *InternalRotationConfig) newFileSizeTimeWriter() (io.Writer, error) {
	ljWriter, err := logger.newFileSizeWriter()
	if err != nil {
		return nil, err
	}
	lj := ljWriter.(*lumberjack.Logger)

	interval, err := parseInterval(logger.RotationInterval)
	if err != nil {
		return nil, err
	}

	// Fire a rotation at the specified interval in a background goroutine.
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			if err := lj.Rotate(); err != nil {
				_, _ = os.Stderr.WriteString("slogrotate: failed to rotate log: " + err.Error() + "\n")
			}
		}
	}()
	return lj, nil
}

func (logger *InternalRotationConfig) SetupGlobal(opts *slog.HandlerOptions, isJSON, isConsole bool) (*slog.Logger, error) {

	if isConsole {
		global := logger.NewConsoleLogger(opts, isJSON)
		if global == nil {
			return nil, errors.New("failed to create global console logger")
		}
		slog.SetDefault(global)
		return global, nil
	}

	writer, err := logger.newFileSizeTimeWriter()
	if err != nil {
		return nil, err
	}

	global := logger.NewLogger(writer, opts, isJSON)
	if global == nil {
		return nil, errors.New("failed to create global file logger")
	}
	slog.SetDefault(global)

	return global, nil
}

func (logger *InternalRotationConfig) NewLogger(writer io.Writer, opts *slog.HandlerOptions, isJSON bool) *slog.Logger {
	var h slog.Handler

	if isJSON {
		h = slog.NewJSONHandler(writer, opts)
	} else {
		h = slog.NewTextHandler(writer, opts)
	}
	return slog.New(h)
}

func (logger *InternalRotationConfig) NewConsoleLogger(opts *slog.HandlerOptions, isJSON bool) *slog.Logger {
	return logger.NewLogger(os.Stdout, opts, isJSON)
}

func (logger *InternalRotationConfig) NewSubdirLogger(subdir, logName string, opts *slog.HandlerOptions, isJSON bool) (*slog.Logger, error) {
	if subdir == "" {
		return nil, errors.New("subdir name cannot be empty")
	}
	baseDir := filepath.Dir(logger.Filename)
	subPath := filepath.Join(baseDir, subdir)

	newLogger := logger.copy()
	newLogger.Filename = filepath.Join(subPath, logName+".log")

	writer, err := newLogger.newFileSizeTimeWriter()
	if err != nil {
		return nil, err
	}

	return newLogger.NewLogger(writer, opts, isJSON), nil
}
