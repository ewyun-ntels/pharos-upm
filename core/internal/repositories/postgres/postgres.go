package postgres

import (
	_ "github.com/lib/pq"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories"
)

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
