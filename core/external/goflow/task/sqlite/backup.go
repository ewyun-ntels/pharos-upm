package sqlite

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	"maze.io/x/duration"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

type BackupTask struct {
	configPath string
	config     common.Config
	data       any
}

func (backupTask *BackupTask) GetName() string {
	return "sqlite-backup"
}

func (backupTask *BackupTask) SetConfig(configPath string, config common.Config) {
	backupTask.configPath = configPath
	backupTask.config = config
}

func (backupTask *BackupTask) SetData(data any) {
	backupTask.data = data
}

func (backupTask *BackupTask) GetDataFormat() any {
	return ``
}

func (backupTask *BackupTask) Run(_ context.Context) (any, error) {
	handler := func(db *sqlx.DB) error {
		backupFileName, err := backupTask.getBackupFileName()
		if err != nil {
			return err
		}

		// Use a parameterized statement to avoid injecting the path into SQL directly
		_, err = db.Exec(`VACUUM main INTO ?`, backupFileName)
		return err
	}
	if err := orm.Handler(orm.DriverDefault, &backupTask.config.Database, handler); err != nil {
		return nil, err
	}

	return nil, backupTask.ttl()
}

func (backupTask *BackupTask) ttl() error {
	d, err := duration.ParseDuration(backupTask.config.Database.SQLite.Backup.TTL)
	if err != nil {
		return err
	}
	ttl := time.Now().Add(-time.Second * time.Duration(d.Seconds()))

	walkFunc := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !info.ModTime().Before(ttl) {
			return nil
		}

		if err := os.Remove(path); err != nil {
			slog.Error("sqlite backup file remove error", "path", path, "error", err.Error())
			return nil // need to proceed to the next file
		}

		return nil
	}

	backupDirectory := backupTask.config.Database.SQLite.Backup.Directory
	if err := filepath.Walk(backupDirectory, walkFunc); err != nil {
		return err
	}

	return nil
}

func (backupTask *BackupTask) getBackupFileName() (string, error) {
	if err := backupTask.mkdirBackupDirectory(); err != nil {
		return "", err
	}

	backupDirectory := backupTask.config.Database.SQLite.Backup.Directory
	fileName := time.Now().Format("2006-01-02_15:04:05") + ".db"

	return backupDirectory + string(filepath.Separator) + fileName, nil
}

func (backupTask *BackupTask) mkdirBackupDirectory() error {
	backupDirectory := backupTask.config.Database.SQLite.Backup.Directory

	if _, err := os.Stat(backupDirectory); os.IsExist(err) {
		return nil
	}

	return os.MkdirAll(backupDirectory, 0o750)
}
