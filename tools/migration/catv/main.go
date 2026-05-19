package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	ClickHouse struct {
		Debug                     bool     `yaml:"debug"`
		Timezone                  string   `yaml:"timezone"`
		Endpoints                 []string `yaml:"endpoints"`
		Username                  string   `yaml:"username"`
		Password                  string   `yaml:"password"`
		Database                  string   `yaml:"database"`
		DialTimeout               string   `yaml:"dial_timeout"`
		MaxOpenConnections        int      `yaml:"max_open_connections"`
		MaxIdleMaxOpenConnections int      `yaml:"max_idle_max_open_connections"`
		ConnectionMaxLifetime     string   `yaml:"connection_max_lifetime"`
		ReadTimeout               string   `yaml:"read_timeout"`
	} `yaml:"clickhouse"`

	Migration struct {
		MaxWorkers       int     `yaml:"max_workers"`
		FailureThreshold float64 `yaml:"failure_threshold"`
		ProgressInterval int     `yaml:"progress_interval"`
		SaveFailedRanges bool    `yaml:"save_failed_ranges"`
		Table            struct {
			Source      string `yaml:"source"`
			Destination string `yaml:"destination"`
		} `yaml:"table"`

		Where struct {
			Time struct {
				Field    string `yaml:"field"`
				Start    string `yaml:"start"`
				End      string `yaml:"end"`
				Interval string `yaml:"interval"`
			} `yaml:"time"`
		} `yaml:"where"`
	} `yaml:"migration"`
}

func (c *Config) Validate() error {
	if c.Migration.MaxWorkers < 0 {
		return fmt.Errorf("max_workers must be >= 0, got %d", c.Migration.MaxWorkers)
	}

	if c.Migration.FailureThreshold < 0 || c.Migration.FailureThreshold > 1 {
		return fmt.Errorf("failure_threshold must be between 0 and 1, got %f", c.Migration.FailureThreshold)
	}

	if c.Migration.ProgressInterval <= 0 {
		c.Migration.ProgressInterval = 100
	}

	if _, err := time.LoadLocation(c.ClickHouse.Timezone); err != nil {
		return fmt.Errorf("invalid timezone %q: %w", c.ClickHouse.Timezone, err)
	}

	if _, err := time.ParseDuration(c.ClickHouse.DialTimeout); err != nil {
		return fmt.Errorf("invalid dial_timeout: %w", err)
	}
	if _, err := time.ParseDuration(c.ClickHouse.ReadTimeout); err != nil {
		return fmt.Errorf("invalid read_timeout: %w", err)
	}
	if _, err := time.ParseDuration(c.ClickHouse.ConnectionMaxLifetime); err != nil {
		return fmt.Errorf("invalid connection_max_lifetime: %w", err)
	}
	if _, err := time.ParseDuration(c.Migration.Where.Time.Interval); err != nil {
		return fmt.Errorf("invalid interval: %w", err)
	}

	if err := validateIdentifier(c.Migration.Table.Source); err != nil {
		return fmt.Errorf("invalid source table: %w", err)
	}
	if err := validateIdentifier(c.Migration.Table.Destination); err != nil {
		return fmt.Errorf("invalid destination table: %w", err)
	}
	if err := validateIdentifier(c.Migration.Where.Time.Field); err != nil {
		return fmt.Errorf("invalid time field: %w", err)
	}

	return nil
}

func validateIdentifier(name string) error {
	if name == "" {
		return fmt.Errorf("identifier cannot be empty")
	}

	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '.') {
			return fmt.Errorf("invalid character %q in identifier %q", ch, name)
		}
	}
	return nil
}

func NewClickHouseClient(config Config) (*ClickHouseClient, error) {
	return &ClickHouseClient{
		config: config,
	}, nil
}

type ClickHouseClient struct {
	config Config

	ctx    context.Context
	cancel context.CancelFunc
	conn   clickhouse.Conn
}

func (c *ClickHouseClient) Open() error {
	c.ctx, c.cancel = context.WithCancel(context.Background())

	options, err := c.getOptions()
	if err != nil {
		c.cancel()
		return err
	}

	c.conn, err = clickhouse.Open(options)
	if err != nil {
		c.cancel()
		return err
	}

	return nil
}

func (c *ClickHouseClient) Ping() error {
	return c.conn.Ping(c.ctx)
}

func (c *ClickHouseClient) Close() error {
	c.cancel()

	return c.conn.Close()
}

func (c *ClickHouseClient) getOptions() (*clickhouse.Options, error) {
	debug := c.config.ClickHouse.Debug

	endpoints := c.config.ClickHouse.Endpoints
	username := c.config.ClickHouse.Username
	password := c.config.ClickHouse.Password
	database := c.config.ClickHouse.Database

	dialTimeout, err := time.ParseDuration(c.config.ClickHouse.DialTimeout)
	if err != nil {
		return nil, err
	}

	maxOpenConns := c.config.ClickHouse.MaxOpenConnections
	maxIdleConns := c.config.ClickHouse.MaxIdleMaxOpenConnections

	connMaxLifetime, err := time.ParseDuration(c.config.ClickHouse.ConnectionMaxLifetime)
	if err != nil {
		return nil, err
	}

	readTimeout, err := time.ParseDuration(c.config.ClickHouse.ReadTimeout)
	if err != nil {
		return nil, err
	}

	return &clickhouse.Options{
		Protocol: clickhouse.Native,

		Addr: endpoints,
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},

		Debug: debug,
		Debugf: func(format string, v ...any) {
			slog.Debug("clickhouse debug", "log", fmt.Sprintf(format, v...))
		},

		DialTimeout:     dialTimeout,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,

		ReadTimeout: readTimeout,
	}, nil
}

func NewMigration(config Config) (*Migration, error) {
	clickHouseClient, err := NewClickHouseClient(config)
	if err != nil {
		return nil, err
	}

	return &Migration{
		config:           config,
		clickHouseClient: clickHouseClient,
	}, nil
}

type jobResult struct {
	StartTime time.Time
	EndTime   time.Time
	Success   bool
	Error     error
	Duration  time.Duration
}

type timeRange struct {
	Start time.Time
	End   time.Time
}

type Migration struct {
	config           Config
	clickHouseClient *ClickHouseClient
}

func (m *Migration) Run() error {
	if err := m.clickHouseClient.Open(); err != nil {
		slog.Error("clickhouse connect error", "error", err.Error())
		return err
	}
	if err := m.clickHouseClient.Ping(); err != nil {
		slog.Error("clickhouse ping error", "error", err.Error())
		return err
	}
	defer func() {
		if err := m.clickHouseClient.Close(); err != nil {
			slog.Error("clickhouse close error", "error", err.Error())
			return
		}
	}()

	timeRanges, err := m.getTimeRanges()
	if err != nil {
		slog.Error("time info parse error", "error", err.Error())
		return err
	}

	slog.Info("migration started", "total_jobs", len(timeRanges), "max_workers", m.config.Migration.MaxWorkers)

	maxWorkers := m.config.Migration.MaxWorkers
	if maxWorkers <= 0 {
		maxWorkers = 5
	}

	jobs := make(chan timeRange, len(timeRanges))
	results := make(chan jobResult, len(timeRanges))

	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go m.worker(i+1, jobs, results, &wg)
	}

	for _, tr := range timeRanges {
		jobs <- tr
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var successCount, failureCount int
	var failedJobs []jobResult
	progressInterval := m.config.Migration.ProgressInterval

	for result := range results {
		if result.Success {
			successCount++
			// Log only if the job took more than 10 seconds, because too many logs
			if result.Duration.Seconds() > 10 {
				slog.Info("job completed",
					"where_start_time", result.StartTime.Format(time.DateTime),
					"where_end_time", result.EndTime.Format(time.DateTime),
					"elapsed_time_seconds", result.Duration.Seconds())
			}
		} else {
			failureCount++
			failedJobs = append(failedJobs, result)
			slog.Error("job failed",
				"where_start_time", result.StartTime.Format(time.DateTime),
				"where_end_time", result.EndTime.Format(time.DateTime),
				"elapsed_time_seconds", result.Duration.Seconds(),
				"error", result.Error.Error())
		}

		totalProcessed := successCount + failureCount
		if totalProcessed%progressInterval == 0 {
			progressPercent := float64(totalProcessed) / float64(len(timeRanges)) * 100
			slog.Info("migration progress",
				"completed", totalProcessed,
				"total", len(timeRanges),
				"progress_percent", fmt.Sprintf("%.1f%%", progressPercent),
				"success", successCount,
				"failed", failureCount)
		}

		if m.config.Migration.FailureThreshold > 0 && totalProcessed >= 10 {
			failureRate := float64(failureCount) / float64(totalProcessed)
			if failureRate > m.config.Migration.FailureThreshold {
				slog.Error("failure threshold exceeded",
					"failure_rate", fmt.Sprintf("%.2f%%", failureRate*100),
					"threshold", fmt.Sprintf("%.2f%%", m.config.Migration.FailureThreshold*100),
					"failed_count", failureCount,
					"total_processed", totalProcessed)
			}
		}
	}

	slog.Info("migration summary",
		"total_jobs", len(timeRanges),
		"success", successCount,
		"failed", failureCount)

	if failureCount > 0 {
		slog.Warn("failed job details", "count", failureCount)
		for _, job := range failedJobs {
			slog.Warn("failed job",
				"start", job.StartTime.Format(time.DateTime),
				"end", job.EndTime.Format(time.DateTime),
				"error", job.Error.Error())
		}

		if m.config.Migration.SaveFailedRanges {
			if err := m.saveFailedRanges(failedJobs); err != nil {
				slog.Error("failed to save failed ranges", "error", err.Error())
			} else {
				slog.Info("failed ranges saved to file", "file", "failed_ranges.txt")
			}
		}

		return fmt.Errorf("%d out of %d jobs failed", failureCount, len(timeRanges))
	}

	return nil
}

func (m *Migration) worker(_ int, jobs <-chan timeRange, results chan<- jobResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for tr := range jobs {
		startTime := time.Now()
		err := m.job(tr.Start, tr.End)
		duration := time.Since(startTime)

		results <- jobResult{
			StartTime: tr.Start,
			EndTime:   tr.End,
			Success:   err == nil,
			Error:     err,
			Duration:  duration,
		}
	}
}

func (m *Migration) getTimeRanges() ([]timeRange, error) {
	sourceLocation, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return nil, err
	}

	destinationLocation, err := time.LoadLocation(m.config.ClickHouse.Timezone)
	if err != nil {
		return nil, err
	}

	start, err := time.ParseInLocation(time.DateTime, m.config.Migration.Where.Time.Start, sourceLocation)
	if err != nil {
		return nil, err
	}
	end, err := time.ParseInLocation(time.DateTime, m.config.Migration.Where.Time.End, sourceLocation)
	if err != nil {
		return nil, err
	}

	start = start.In(destinationLocation)
	end = end.In(destinationLocation)

	interval, err := time.ParseDuration(m.config.Migration.Where.Time.Interval)
	if err != nil {
		return nil, err
	}

	var timeRanges []timeRange

	if end.Sub(start) < interval {
		timeRanges = append(timeRanges, timeRange{Start: start, End: end})
		return timeRanges, nil
	}

	for current := start; current.Before(end) || current.Equal(end); {
		intervalEnd := current.Add(interval).Add(-1 * time.Second)

		if intervalEnd.After(end) || intervalEnd.Equal(end) {
			timeRanges = append(timeRanges, timeRange{Start: current, End: end})
			break
		}

		timeRanges = append(timeRanges, timeRange{Start: current, End: intervalEnd})
		current = intervalEnd.Add(1 * time.Second)
	}

	return timeRanges, nil
}

func (m *Migration) saveFailedRanges(failedJobs []jobResult) error {
	file, err := os.Create("failed_ranges.txt")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	fmt.Fprintf(file, "# Failed time ranges - %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "# Format: START_TIME | END_TIME | ERROR\n\n")

	for _, job := range failedJobs {
		fmt.Fprintf(file, "%s | %s | %s\n",
			job.StartTime.Format(time.DateTime),
			job.EndTime.Format(time.DateTime),
			job.Error.Error())
	}

	return nil
}

func (m *Migration) job(startTime, endTime time.Time) error {
	// Note: Identifiers are already validated in Config.Validate()
	// This prevents SQL injection through table/column names
	var query = fmt.Sprintf(`
INSERT INTO %s
SELECT * FROM %s
WHERE %s >= toDateTime('%s') AND %s <= toDateTime('%s');
`, m.config.Migration.Table.Destination, m.config.Migration.Table.Source, m.config.Migration.Where.Time.Field, startTime.Format(time.DateTime), m.config.Migration.Where.Time.Field, endTime.Format(time.DateTime))

	if err := m.clickHouseClient.conn.Exec(m.clickHouseClient.ctx, query); err != nil {
		slog.Error("clickhouse exec error", "error", err.Error())
		return err
	}

	return nil
}

func loadConfig(filePath string) (Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	config := Config{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	return config, nil
}

func main() {
	var processStartTime = time.Now()
	defer func() {
		slog.Info("process time info",
			"elapsed_time_seconds", time.Since(processStartTime).Seconds(),
		)
	}()

	configFile := flag.String("config", "", "Path to configuration file (YAML format)")
	flag.Parse()

	if *configFile == "" {
		slog.Info("Usage: migrator -config <config-file.yml>")
		return
	}

	config, err := loadConfig(*configFile)
	if err != nil {
		slog.Error("Failed to load configuration", "error", err.Error())
		return
	}

	if err := config.Validate(); err != nil {
		slog.Error("Configuration validation failed", "error", err.Error())
		return
	}

	slog.Info("Configuration loaded successfully", "config_file", *configFile)

	migration, err := NewMigration(config)
	if err != nil {
		slog.Error("migration create error", "error", err.Error())
		return
	}
	if err := migration.Run(); err != nil {
		slog.Error("migration run error", "error", err.Error())
		return
	}
}
