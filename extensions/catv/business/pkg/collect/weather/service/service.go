package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/robfig/cron/v3"
	"ntels.com/pharos/core/external/clickhouse_interface"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/extensions/catv/business/pkg/collect/weather/tables"
)

func NewService(config common.Config) (*Service, error) {
	ftpConnectionTimeout, err := time.ParseDuration(config.Catv.Collect.Weather.Timeout.Connection)
	if err != nil {
		slog.Error("Failed to parse FTP connection timeout", "error", err)
		return nil, fmt.Errorf("invalid FTP connection timeout: %w", err)
	}

	ftpShutTimeout, err := time.ParseDuration(config.Catv.Collect.Weather.Timeout.Shut)
	if err != nil {
		slog.Error("Failed to parse FTP shutdown timeout", "error", err)
		return nil, fmt.Errorf("invalid FTP shutdown timeout: %w", err)
	}

	return &Service{
		config: config,

		wg:                   new(sync.WaitGroup),
		ftpCron:              cron.New(cron.WithSeconds()),
		ftpConnectionTimeout: ftpConnectionTimeout,
		ftpShutTimeout:       ftpShutTimeout,
	}, nil
}

type Service struct {
	config common.Config

	wg                   *sync.WaitGroup
	ftpCron              *cron.Cron
	ftpConnectionTimeout time.Duration
	ftpShutTimeout       time.Duration
}

func (s *Service) Start() error {
	if _, err := s.ftpCron.AddFunc(s.config.Catv.Collect.Weather.CronSpec, s.cronHandler); err != nil {
		slog.Error("Failed to add FTP handler to cron", "error", err)
		return fmt.Errorf("failed to add FTP handler to cron: %w", err)
	}

	go s.ftpCron.Start()

	return nil
}

func (s *Service) Stop() error {
	if s.ftpCron != nil {
		s.ftpCron.Stop()
	}

	s.wg.Wait()

	return nil
}

func (s *Service) cronHandler() {
	s.etl(s.config.Catv.Collect.Weather.BaseDirectory, []string{FILE_NAME_3HOUR, FILE_NAME_LAND, FILE_NAME_SHKO, FILE_NAME_SHKO2}...)
}

func (s *Service) etl(directory string, files ...string) {
	datas, err := s.extract(directory, files...)
	if err != nil {
		slog.Error("Failed to extract data from FTP server", "error", err, "directory", directory, "files", files)
		return
	}

	for file, data := range datas {
		s.wg.Add(1)
		go func(file string, data string) {
			defer s.wg.Done()

			var insertTime = orm.Datetime{Time: time.Now().UTC()}
			var errors []string

			weatherTables, transformData, err := s.transform(insertTime, file, data)
			if err != nil {
				slog.Error("Failed to transform data", "error", err, "file", file)
				errors = append(errors, err.Error())
			} else if err := s.load(weatherTables); err != nil {
				slog.Error("Failed to load data into database", "error", err, "tables", weatherTables)
				errors = append(errors, err.Error())
			}

			if err := (tables.WeatherRawTable{InsertTime: insertTime, File: file, Raw: transformData, Errors: errors}).Insert(s.config); err != nil {
				slog.Error("Failed to insert raw weather data into database", "error", err, "file", file)
				return
			}
		}(file, data)
	}
}

func (s *Service) extract(directory string, files ...string) (map[string]string, error) {
	if s.config.Catv.Collect.Weather.Simulation {
		slog.Info("Extracting weather data in simulation mode")
		return s.getFilesForSimulation(directory, files...)
	}

	return s.getFilesForFTP(directory, files...)
}

func (s *Service) getFilesForFTP(directory string, files ...string) (map[string]string, error) {
	var ctx = context.Background()

	serverConn, err := ftp.Dial(
		s.config.Catv.Collect.Weather.Address,
		ftp.DialWithContext(ctx),
		ftp.DialWithTimeout(s.ftpConnectionTimeout),
		ftp.DialWithShutTimeout(s.ftpShutTimeout),
	)
	if err != nil {
		slog.Error("Failed to connect to FTP server", "error", err)
		return nil, err
	}
	defer func() {
		if err := serverConn.Quit(); err != nil {
			slog.Error("Failed to quit FTP server", "error", err)
		}
	}()

	if err = serverConn.Login(s.config.Catv.Collect.Weather.User, s.config.Catv.Collect.Weather.Password); err != nil {
		slog.Error("Failed to login to FTP server", "error", err)
		return nil, err
	}

	results := make(map[string]string)
	for _, file := range files {
		path := directory + string(filepath.Separator) + file

		data, err := s.getFile(serverConn, path)
		if err != nil {
			slog.Error("Failed to get file from FTP server", "error", err, "path", path)
			continue
		}

		results[file] = string(data)
	}
	return results, nil
}

func (s *Service) getFilesForSimulation(_ string, files ...string) (map[string]string, error) {
	results := make(map[string]string)

	for _, file := range files {
		switch file {
		case FILE_NAME_3HOUR:
			data, err := tables.WeatherForecast3hTable{}.GetMockData(time.Now())
			if err != nil {
				slog.Error("Failed to get mock data for 3-hour forecast", "error", err)
				break
			}

			results[file] = data
		case FILE_NAME_LAND:
			data, err := tables.WeatherForecastDailyTable{}.GetMockData(time.Now())
			if err != nil {
				slog.Error("Failed to get mock data for daily forecast", "error", err)
				break
			}
			results[file] = data
		case FILE_NAME_SHKO, FILE_NAME_SHKO2:
			data, err := tables.WeatherObsTable{}.GetMockData(time.Now())
			if err != nil {
				slog.Error("Failed to get mock data for weather observation", "error", err)
				break
			}
			results[file] = data
		default:
			slog.Error("Unknown file name for simulation", "file", file)
		}
	}

	return results, nil
}

func (s *Service) transform(insertTime orm.Datetime, file string, data string) ([]tables.WeatherTable, string, error) {
	utf8DataBytes, err := tables.EUC_KRToUTF8(data)
	if err != nil {
		slog.Error("Failed to convert data encoding from EUC-KR to UTF-8", "error", err)
		return nil, data, fmt.Errorf("failed to convert data encoding from EUC-KR to UTF-8: %w", err)
	}
	utf8Data := string(utf8DataBytes)

	var table tables.WeatherTable

	switch file {
	case FILE_NAME_3HOUR:
		table = &tables.WeatherForecast3hTable{InsertTime: insertTime}
	case FILE_NAME_LAND:
		table = &tables.WeatherForecastDailyTable{InsertTime: insertTime}
	case FILE_NAME_SHKO, FILE_NAME_SHKO2:
		switch file {
		case FILE_NAME_SHKO:
			table = &tables.WeatherObsTable{InsertTime: insertTime, Source: OBS_SOURCE_SHKO}
		case FILE_NAME_SHKO2:
			table = &tables.WeatherObsTable{InsertTime: insertTime, Source: OBS_SOURCE_AWS}
		default:
			slog.Error("Unknown file name for weather observation", "file", file)
			return nil, utf8Data, fmt.Errorf("unknown file name for weather observation: %s", file)
		}
	default:
		slog.Error("Unknown file name", "file", file)
		return nil, utf8Data, fmt.Errorf("unknown file name: %s", file)
	}

	tables, err := table.Transform(insertTime, utf8Data)
	if err != nil {
		slog.Error("Failed to transform data into table format", "error", err)
		return nil, utf8Data, fmt.Errorf("failed to transform data into table format: %w", err)
	}
	return tables, utf8Data, nil
}

func (s *Service) load(tables []tables.WeatherTable) error {
	var batch = clickhouse_interface.ClickhouseBatch{}

	for _, t := range tables {
		data, err := t.ConvertBatchFormat()
		if err != nil {
			slog.Error("Failed to convert table data to ClickHouse format", "error", err, "table", t)
			continue
		}

		if err := batch.AddBatchJson(data); err != nil {
			slog.Error("Failed to add batch data to ClickHouse", "error", err, "data", data)
			continue
		}
	}

	if err := batch.Insert(s.config.Catv.Database, tables[0].TableName()); err != nil {
		slog.Error("Failed to insert batch data into ClickHouse", "error", err, "table", tables[0].TableName(), "batch_size", batch.Len())
		return err
	}

	return nil
}

func (s *Service) getFile(serverConn *ftp.ServerConn, name string) (string, error) {
	r, err := serverConn.Retr(name)
	if err != nil {
		return "", err
	}
	defer r.Close()

	buffer, err := io.ReadAll(r)
	return string(buffer), err
}
