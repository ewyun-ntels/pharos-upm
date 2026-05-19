package schema

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pressly/goose/v3/database"
)

// clickhouseClusterStore wraps the standard ClickHouse store and overrides
// CreateVersionTable to use ON CLUSTER + ReplicatedMergeTree so that the
// version table is properly replicated across all nodes in a ClickHouse cluster.
type clickhouseClusterStore struct {
	inner        database.Store
	tableName    string
	databaseName string
	clusterName  string
}

var _ database.Store = (*clickhouseClusterStore)(nil)

func newClickhouseClusterStore(tableName, databaseName, clusterName string) (database.Store, error) {
	inner, err := database.NewStore(database.DialectClickHouse, tableName)
	if err != nil {
		return nil, err
	}
	return &clickhouseClusterStore{
		inner:        inner,
		tableName:    tableName,
		databaseName: databaseName,
		clusterName:  clusterName,
	}, nil
}

func (s *clickhouseClusterStore) Tablename() string {
	return s.tableName
}

// CreateVersionTable creates the goose version table on the entire cluster using
// ReplicatedMergeTree so every replica shares the same migration state.
//
// The ZooKeeper path intentionally omits the {shard} macro. In a multi-shard cluster,
// using {shard} would create an independent replication group per shard, causing pods
// connected to different shards to see different (out-of-sync) migration histories.
// By using a shard-agnostic path, all nodes in the cluster share a single replication
// group, so any write on any node is propagated to every other node without needing a
// Distributed table.
func (s *clickhouseClusterStore) CreateVersionTable(ctx context.Context, db database.DBTxConn) error {
	q := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s ON CLUSTER %s
(
    version_id Int64,
    is_applied UInt8,
    date       Date     DEFAULT now(),
    tstamp     DateTime DEFAULT now()
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/goose/%s', '{replica}')
ORDER BY (date)`,
		s.tableName, s.clusterName, s.tableName,
	)
	if _, err := db.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("failed to create version table %q on cluster %q: %w", s.tableName, s.clusterName, err)
	}
	return nil
}

// Insert writes the version record and then blocks until every replica in the
// cluster has received and applied the replication entry. This prevents the
// race condition where the process restarts and connects to a different
// ClickHouse node that has not yet received the version record written in the
// previous run, causing goose to re-run all migrations.
//
// ReplicatedMergeTree INSERTs are asynchronous by default — the INSERT returns
// as soon as the receiving node confirms the write and registers the entry in
// ZooKeeper. Other nodes fetch and apply the entry in the background. Without
// this sync step, there is a window (typically sub-second, but not guaranteed)
// where other nodes do not yet see the new version record.
//
// SYSTEM SYNC REPLICA ON CLUSTER makes every node in the cluster drain its own
// replication queue before the statement returns to the client, eliminating
// that window at the cost of a small additional latency per migration step.
func (s *clickhouseClusterStore) Insert(ctx context.Context, db database.DBTxConn, req database.InsertRequest) error {
	if err := s.inner.Insert(ctx, db, req); err != nil {
		return err
	}

	// Wait for all cluster replicas to apply the entry we just wrote.
	// The table name must be fully qualified (database.table) so that remote
	// nodes in the cluster can resolve it — without a database prefix they
	// default to the 'default' database and fail with UNKNOWN_TABLE.
	qualifiedTable := s.tableName
	if s.databaseName != "" {
		qualifiedTable = s.databaseName + "." + s.tableName
	}
	// receive_timeout must be set as a session variable before SYSTEM SYNC REPLICA
	// because the command does not support an inline SETTINGS clause.
	// 10 seconds is sufficient for a healthy cluster where replication is typically
	// sub-second. This prevents a single unhealthy node from blocking startup for
	// more than 10 seconds per migration step.
	if _, err := db.ExecContext(ctx, "SET receive_timeout=10"); err != nil {
		slog.Warn("failed to set receive_timeout before syncing replicas, proceeding anyway",
			"table", s.tableName, "error", err)
		return nil
	}
	syncQ := fmt.Sprintf(
		"SYSTEM SYNC REPLICA ON CLUSTER %s %s", s.clusterName, qualifiedTable,
	)
	// SYSTEM SYNC REPLICA failure is non-fatal: the INSERT already succeeded on
	// the local node and ReplicatedMergeTree will propagate it asynchronously.
	// A sync failure typically means a replica is temporarily unavailable; the
	// worst case is the brief race window we were trying to close, which is the
	// same as running without sync at all. Logging a warning is sufficient.
	if _, err := db.ExecContext(ctx, syncQ); err != nil {
		slog.Warn("failed to sync replicas after inserting version; replication will complete asynchronously",
			"version", req.Version, "table", s.tableName, "error", err)
	}
	return nil
}

func (s *clickhouseClusterStore) Delete(ctx context.Context, db database.DBTxConn, version int64) error {
	return s.inner.Delete(ctx, db, version)
}

func (s *clickhouseClusterStore) GetMigration(ctx context.Context, db database.DBTxConn, version int64) (*database.GetMigrationResult, error) {
	return s.inner.GetMigration(ctx, db, version)
}

func (s *clickhouseClusterStore) GetLatestVersion(ctx context.Context, db database.DBTxConn) (int64, error) {
	return s.inner.GetLatestVersion(ctx, db)
}

func (s *clickhouseClusterStore) ListMigrations(ctx context.Context, db database.DBTxConn) ([]*database.ListMigrationsResult, error) {
	return s.inner.ListMigrations(ctx, db)
}
