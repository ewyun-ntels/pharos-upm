package sqlite

import (
	"time"

	_ "modernc.org/sqlite"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories"
)

func toDatabaseTime(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04:05.999")
}

func toNullableTime(t time.Time) *string {
	if !t.IsZero() {
		dbTime := toDatabaseTime(t)
		return &dbTime
	}
	return nil
}

func NewUserRepository(config orm.DatabaseConfig) repositories.UserRepository {
	return newUserRepository(config)
}

func NewDashboardRepository(config orm.DatabaseConfig) repositories.DashboardRepository {
	return newDashboardRepository(config)
}

func NewDataMigrationRepository(config orm.DatabaseConfig) repositories.DataMigrationRepository {
	return newDataMigrationRepository(config)
}

func NewRoleMetadataOverrideRepository(config orm.DatabaseConfig) repositories.RoleMetadataConfigRepository {
	return newRoleMetadataConfigRepository(config)
}

func NewUserMetadataConfigRepository(config orm.DatabaseConfig) repositories.UserMetadataConfigRepository {
	return newUserMetadataConfigRepository(config)
}

func NewDashboardFolderRepository(config orm.DatabaseConfig) repositories.DashboardFolderRepository {
	return newDashboardFolderRepository(config)
}

func NewDashboardHistoryRepository(config orm.DatabaseConfig) repositories.DashboardHistoryRepository {
	return newDashboardHistoryRepository(config)
}
