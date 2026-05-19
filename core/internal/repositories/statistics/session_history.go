package statistics

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories"
)

var _ repositories.LoginHistoryRepository = (*loginHistoryRepository)(nil)

type loginHistoryRepository struct {
	config orm.DatabaseConfig
}

func NewLoginHistoryRepository(config orm.DatabaseConfig) repositories.LoginHistoryRepository {
	return &loginHistoryRepository{config: config}
}

func (r *loginHistoryRepository) tableName() string {
	if r.config.Driver == orm.DriverClickHouse {
		return "history_login FINAL"
	}
	return "history_login"
}

func (r *loginHistoryRepository) Query(ctx context.Context, username string, limit, offset int) ([]repositories.LoginHistoryRecord, error) {
	var records []repositories.LoginHistoryRecord

	err := orm.StatisticsHandler("", &r.config, func(db *sqlx.DB) error {
		var query string
		var args []any
		table := r.tableName()
		if username != "" {
			query = fmt.Sprintf(
				"SELECT session_id, user_id, COALESCE(client_ip, ''), COALESCE(user_agent, ''), started_at, ended_at, COALESCE(end_reason, '') FROM %s WHERE user_id LIKE ? ORDER BY started_at DESC LIMIT ? OFFSET ?",
				table,
			)
			args = []any{"%" + username + "%", limit, offset}
		} else {
			query = fmt.Sprintf(
				"SELECT session_id, user_id, COALESCE(client_ip, ''), COALESCE(user_agent, ''), started_at, ended_at, COALESCE(end_reason, '') FROM %s ORDER BY started_at DESC LIMIT ? OFFSET ?",
				table,
			)
			args = []any{limit, offset}
		}

		rows, err := db.QueryContext(ctx, db.Rebind(query), args...)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var rec repositories.LoginHistoryRecord
			var endedAt *time.Time
			if err := rows.Scan(&rec.SessionId, &rec.UserId, &rec.ClientIp, &rec.UserAgent, &rec.StartedAt, &endedAt, &rec.EndReason); err != nil {
				return err
			}
			rec.EndedAt = endedAt
			records = append(records, rec)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if records == nil {
		records = []repositories.LoginHistoryRecord{}
	}
	return records, nil
}

func (r *loginHistoryRepository) Count(ctx context.Context, username string) (int, error) {
	var total int

	err := orm.StatisticsHandler("", &r.config, func(db *sqlx.DB) error {
		var query string
		var args []any
		table := r.tableName()
		if username != "" {
			query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE user_id LIKE ?", table)
			args = []any{"%" + username + "%"}
		} else {
			query = fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		}
		return db.QueryRowContext(ctx, db.Rebind(query), args...).Scan(&total)
	})
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *loginHistoryRepository) UpsertSession(ctx context.Context, record repositories.LoginHistoryRecord) error {
	return orm.StatisticsHandler("", &r.config, func(db *sqlx.DB) error {
		var query string
		var args []any

		switch r.config.Driver {
		case orm.DriverClickHouse:
			query = `INSERT INTO history_login
				(session_id, user_id, client_ip, user_agent, started_at, ended_at, end_reason, version)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
			args = []any{
				record.SessionId, record.UserId, nilIfEmpty(record.ClientIp), nilIfEmpty(record.UserAgent),
				record.StartedAt, record.EndedAt, nilIfEmpty(record.EndReason),
				time.Now(),
			}
		default: // sqlite, postgresql
			query = `INSERT INTO history_login
				(session_id, user_id, client_ip, user_agent, started_at, ended_at, end_reason)
				VALUES (?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT(session_id) DO UPDATE SET
					ended_at   = CASE WHEN excluded.ended_at   IS NOT NULL THEN excluded.ended_at   ELSE history_login.ended_at   END,
					end_reason = CASE WHEN excluded.ended_at   IS NOT NULL THEN excluded.end_reason ELSE history_login.end_reason END`
			args = []any{
				record.SessionId, record.UserId, nilIfEmpty(record.ClientIp), nilIfEmpty(record.UserAgent),
				record.StartedAt, record.EndedAt, nilIfEmpty(record.EndReason),
			}
		}

		_, err := db.ExecContext(ctx, db.Rebind(query), args...)
		return err
	})
}

// nilIfEmpty returns nil for an empty string so the DB stores NULL instead of "".
func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *loginHistoryRepository) CloseOrphanedSessions(ctx context.Context, endedAt time.Time) (int64, error) {
	var affected int64

	err := orm.StatisticsHandler("", &r.config, func(db *sqlx.DB) error {
		var query string
		var args []any

		switch r.config.Driver {
		case orm.DriverClickHouse:
			// ClickHouse: UPDATE 불가 → ended_at이 NULL인 행을 SELECT해 closed 상태로 INSERT.
			// version = now()이므로 FINAL 조회 시 새 행(최신 version)이 선택된다.
			query = `INSERT INTO history_login
				(session_id, user_id, client_ip, started_at, ended_at, end_reason, version)
				SELECT session_id, user_id, client_ip, started_at, ?, 'server_restart', now()
				FROM history_login FINAL
				WHERE ended_at IS NULL`
			args = []any{endedAt}
		default: // sqlite, postgresql
			query = `UPDATE history_login SET ended_at = ?, end_reason = 'server_restart'
				WHERE ended_at IS NULL`
			args = []any{endedAt}
		}

		result, err := db.ExecContext(ctx, db.Rebind(query), args...)
		if err != nil {
			return err
		}
		affected, _ = result.RowsAffected()
		return nil
	})
	return affected, err
}

// noopLoginHistoryRepository is used when statistics DB is not configured.
type noopLoginHistoryRepository struct{}

func NoopLoginHistoryRepository() repositories.LoginHistoryRepository {
	return noopLoginHistoryRepository{}
}

func (n noopLoginHistoryRepository) Query(_ context.Context, _ string, _, _ int) ([]repositories.LoginHistoryRecord, error) {
	return []repositories.LoginHistoryRecord{}, nil
}

func (n noopLoginHistoryRepository) Count(_ context.Context, _ string) (int, error) {
	return 0, nil
}

func (n noopLoginHistoryRepository) UpsertSession(_ context.Context, _ repositories.LoginHistoryRecord) error {
	return nil
}

func (n noopLoginHistoryRepository) CloseOrphanedSessions(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}
