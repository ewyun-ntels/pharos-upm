package testutil

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// StartPGContainer starts a PostgreSQL container for testing purposes.
func StartPGContainer(ctx context.Context, dbname, user, password string) (*postgres.PostgresContainer, error) {
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbname),
		postgres.WithUsername(user),
		postgres.WithPassword(password),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, err
	}
	// wait until ready - testcontainers handles wait strategy
	return container, nil
}

// StartClickHouseContainer starts a ClickHouse container for testing purposes.
func StartClickHouseContainer(ctx context.Context, dbname, user, password string) (testcontainers.Container, error) {
	container, err := clickhouse.Run(ctx,
		"clickhouse/clickhouse-server:24-alpine",
		clickhouse.WithDatabase(dbname),
		clickhouse.WithUsername(user),
		clickhouse.WithPassword(password),
	)
	if err != nil {
		return nil, err
	}
	return container, nil
}
