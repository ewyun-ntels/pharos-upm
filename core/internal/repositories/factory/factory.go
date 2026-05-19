package factory

import (
	"errors"

	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/postgres"
	"ntels.com/pharos/core/internal/repositories/sqlite"
	"ntels.com/pharos/core/internal/repositories/statistics"
)

// RepositoryOptions holds configuration for repository creation
type RepositoryOptions struct {
	DatabaseConfig orm.DatabaseConfig
}

// StatisticsRepositoryOptions holds configuration for statistics DB repositories
type StatisticsRepositoryOptions struct {
	DatabaseConfig orm.DatabaseConfig
}

type constructorFunc[T any] func(orm.DatabaseConfig) T

type constructors[T any] struct {
	sqliteFunc   constructorFunc[T]
	postgresFunc constructorFunc[T]
}

func newRepository[T any](opts RepositoryOptions, constructor constructors[T]) (T, error) {
	var zero T
	if opts.DatabaseConfig.GetDriverName() == "" || opts.DatabaseConfig.GetDataSourceName() == "" {
		return zero, errors.New("database configuration is incomplete")
	}

	switch opts.DatabaseConfig.Driver {
	case orm.DriverSqlite:
		if constructor.sqliteFunc != nil {
			return constructor.sqliteFunc(opts.DatabaseConfig), nil
		}
	case orm.DriverPostgreSQL:
		if constructor.postgresFunc != nil {
			return constructor.postgresFunc(opts.DatabaseConfig), nil
		}
	}
	return zero, errors.New("unsupported database driver: " + opts.DatabaseConfig.Driver)
}

// NewUserRepository creates a new user repository based on the database driver
func NewUserRepository(opts RepositoryOptions) (repositories.UserRepository, error) {
	c := constructors[repositories.UserRepository]{
		sqliteFunc:   sqlite.NewUserRepository,
		postgresFunc: postgres.NewUserRepository,
	}
	return newRepository[repositories.UserRepository](opts, c)
}

// NewDashboardRepository creates a new dashboard repository based on the database driver
func NewDashboardRepository(opts RepositoryOptions) (repositories.DashboardRepository, error) {
	c := constructors[repositories.DashboardRepository]{
		sqliteFunc:   sqlite.NewDashboardRepository,
		postgresFunc: postgres.NewDashboardRepository,
	}
	return newRepository[repositories.DashboardRepository](opts, c)
}

// NewDataMigrationRepository creates a new data migration repository based on the database driver
func NewDataMigrationRepository(opts RepositoryOptions) (repositories.DataMigrationRepository, error) {
	c := constructors[repositories.DataMigrationRepository]{
		sqliteFunc:   sqlite.NewDataMigrationRepository,
		postgresFunc: postgres.NewDataMigrationRepository,
	}
	return newRepository[repositories.DataMigrationRepository](opts, c)
}

// NewRoleMetadataConfigRepository creates a new role metadata config repository based on the database driver
func NewRoleMetadataConfigRepository(opts RepositoryOptions) (repositories.RoleMetadataConfigRepository, error) {
	c := constructors[repositories.RoleMetadataConfigRepository]{
		sqliteFunc:   sqlite.NewRoleMetadataOverrideRepository,
		postgresFunc: postgres.NewRoleMetadataOverrideRepository,
	}
	return newRepository[repositories.RoleMetadataConfigRepository](opts, c)
}

// NewDashboardFolderRepository creates a new dashboard folder repository based on the database driver
func NewDashboardFolderRepository(opts RepositoryOptions) (repositories.DashboardFolderRepository, error) {
	c := constructors[repositories.DashboardFolderRepository]{
		sqliteFunc:   sqlite.NewDashboardFolderRepository,
		postgresFunc: postgres.NewDashboardFolderRepository,
	}
	return newRepository[repositories.DashboardFolderRepository](opts, c)
}

// NewDashboardHistoryRepository creates a new dashboard history repository based on the database driver
func NewDashboardHistoryRepository(opts RepositoryOptions) (repositories.DashboardHistoryRepository, error) {
	c := constructors[repositories.DashboardHistoryRepository]{
		sqliteFunc:   sqlite.NewDashboardHistoryRepository,
		postgresFunc: postgres.NewDashboardHistoryRepository,
	}
	return newRepository[repositories.DashboardHistoryRepository](opts, c)
}

// NewUserMetadataConfigRepository creates a new user metadata config repository based on the database driver
func NewUserMetadataConfigRepository(opts RepositoryOptions) (repositories.UserMetadataConfigRepository, error) {
	c := constructors[repositories.UserMetadataConfigRepository]{
		sqliteFunc:   sqlite.NewUserMetadataConfigRepository,
		postgresFunc: postgres.NewUserMetadataConfigRepository,
	}
	return newRepository[repositories.UserMetadataConfigRepository](opts, c)
}

// NewLoginHistoryRepository creates a login history repository backed by the statistics DB.
// Returns a noop repository if no statistics DB driver is configured.
func NewLoginHistoryRepository(opts StatisticsRepositoryOptions) repositories.LoginHistoryRepository {
	if opts.DatabaseConfig.Driver == "" {
		return statistics.NoopLoginHistoryRepository()
	}
	return statistics.NewLoginHistoryRepository(opts.DatabaseConfig)
}
