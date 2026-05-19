package sqlite

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories"
)

var _ repositories.DashboardHistoryRepository = (*dashboardHistoryRepository)(nil)

type dashboardHistoryRepository struct {
	repositories.BaseRepository
}

func newDashboardHistoryRepository(config orm.DatabaseConfig) *dashboardHistoryRepository {
	return &dashboardHistoryRepository{
		BaseRepository: repositories.NewBaseRepository(config),
	}
}

type sqliteDashboardHistory struct {
	ID             string  `db:"id"`
	DashboardID    string  `db:"dashboard_id"`
	DashboardTitle string  `db:"dashboard_title"`
	Action         string  `db:"action"`
	ChangedBy      string  `db:"changed_by"`
	ChangedAt      string  `db:"changed_at"`
	ConfigJSON     *string `db:"config_json"`
}

func (r *sqliteDashboardHistory) toEntity() repositories.DashboardHistoryRecord {
	parse := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02 15:04:05.999", s)
		return t
	}
	return repositories.DashboardHistoryRecord{
		ID:             r.ID,
		DashboardID:    r.DashboardID,
		DashboardTitle: r.DashboardTitle,
		Action:         r.Action,
		ChangedBy:      r.ChangedBy,
		ChangedAt:      parse(r.ChangedAt),
		ConfigJSON:     r.ConfigJSON,
	}
}

func (repo *dashboardHistoryRepository) Record(ctx context.Context, record repositories.DashboardHistoryRecord) error {
	if record.ID == "" {
		record.ID = uuid.New().String()
	}
	if record.ChangedAt.IsZero() {
		record.ChangedAt = time.Now()
	}
	query := `INSERT INTO dashboard_history (id, dashboard_id, dashboard_title, action, changed_by, changed_at, config_json)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.ExecContext(ctx, query,
			record.ID,
			record.DashboardID,
			record.DashboardTitle,
			record.Action,
			record.ChangedBy,
			toDatabaseTime(record.ChangedAt),
			record.ConfigJSON,
		)
		return err
	})
}

func (repo *dashboardHistoryRepository) Query(ctx context.Context, limit, offset int, search string) ([]repositories.DashboardHistoryRecord, error) {
	var rows []sqliteDashboardHistory
	var err error
	if search != "" {
		like := "%" + search + "%"
		q := `SELECT id, dashboard_id, dashboard_title, action, changed_by, changed_at, config_json
			FROM dashboard_history
			WHERE dashboard_title LIKE ? OR action LIKE ? OR changed_by LIKE ?
			ORDER BY changed_at DESC LIMIT ? OFFSET ?`
		err = repo.Handler(func(db *sqlx.DB) error {
			return db.SelectContext(ctx, &rows, q, like, like, like, limit, offset)
		})
	} else {
		q := `SELECT id, dashboard_id, dashboard_title, action, changed_by, changed_at, config_json
			FROM dashboard_history ORDER BY changed_at DESC LIMIT ? OFFSET ?`
		err = repo.Handler(func(db *sqlx.DB) error {
			return db.SelectContext(ctx, &rows, q, limit, offset)
		})
	}
	if err != nil {
		return nil, err
	}
	result := make([]repositories.DashboardHistoryRecord, 0, len(rows))
	for _, r := range rows {
		result = append(result, r.toEntity())
	}
	return result, nil
}

func (repo *dashboardHistoryRepository) Count(ctx context.Context, search string) (int, error) {
	var count int
	var err error
	if search != "" {
		like := "%" + search + "%"
		q := `SELECT COUNT(*) FROM dashboard_history WHERE dashboard_title LIKE ? OR action LIKE ? OR changed_by LIKE ?`
		err = repo.Handler(func(db *sqlx.DB) error {
			return db.GetContext(ctx, &count, q, like, like, like)
		})
	} else {
		err = repo.Handler(func(db *sqlx.DB) error {
			return db.GetContext(ctx, &count, `SELECT COUNT(*) FROM dashboard_history`)
		})
	}
	return count, err
}

func (repo *dashboardHistoryRepository) Prune(ctx context.Context, maxKeep int) error {
	query := `DELETE FROM dashboard_history WHERE id NOT IN (
		SELECT id FROM dashboard_history ORDER BY changed_at DESC LIMIT ?
	)`
	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.ExecContext(ctx, query, maxKeep)
		return err
	})
}

func (repo *dashboardHistoryRepository) GetByID(ctx context.Context, historyID string) (*repositories.DashboardHistoryRecord, error) {
	var row sqliteDashboardHistory
	query := `SELECT id, dashboard_id, dashboard_title, action, changed_by, changed_at, config_json
		FROM dashboard_history WHERE id = ?`
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.GetContext(ctx, &row, query, historyID)
	})
	if err != nil {
		return nil, err
	}
	entity := row.toEntity()
	return &entity, nil
}

func (repo *dashboardHistoryRepository) QueryByDashboardIDs(ctx context.Context, ids []string, limit, offset int, search string) ([]repositories.DashboardHistoryRecord, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var q string
	var args []any
	var err error
	if search != "" {
		like := "%" + search + "%"
		q, args, err = sqlx.In(
			`SELECT id, dashboard_id, dashboard_title, action, changed_by, changed_at, config_json
			FROM dashboard_history
			WHERE dashboard_id IN (?) AND (dashboard_title LIKE ? OR action LIKE ? OR changed_by LIKE ?)
			ORDER BY changed_at DESC LIMIT ? OFFSET ?`,
			ids, like, like, like, limit, offset,
		)
	} else {
		q, args, err = sqlx.In(
			`SELECT id, dashboard_id, dashboard_title, action, changed_by, changed_at, config_json
			FROM dashboard_history WHERE dashboard_id IN (?) ORDER BY changed_at DESC LIMIT ? OFFSET ?`,
			ids, limit, offset,
		)
	}
	if err != nil {
		return nil, err
	}
	var rows []sqliteDashboardHistory
	err = repo.Handler(func(db *sqlx.DB) error {
		return db.SelectContext(ctx, &rows, q, args...)
	})
	if err != nil {
		return nil, err
	}
	result := make([]repositories.DashboardHistoryRecord, 0, len(rows))
	for _, r := range rows {
		result = append(result, r.toEntity())
	}
	return result, nil
}

func (repo *dashboardHistoryRepository) CountByDashboardIDs(ctx context.Context, ids []string, search string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var q string
	var args []any
	var err error
	if search != "" {
		like := "%" + search + "%"
		q, args, err = sqlx.In(
			`SELECT COUNT(*) FROM dashboard_history WHERE dashboard_id IN (?) AND (dashboard_title LIKE ? OR action LIKE ? OR changed_by LIKE ?)`,
			ids, like, like, like,
		)
	} else {
		q, args, err = sqlx.In(`SELECT COUNT(*) FROM dashboard_history WHERE dashboard_id IN (?)`, ids)
	}
	if err != nil {
		return 0, err
	}
	var count int
	err = repo.Handler(func(db *sqlx.DB) error {
		return db.GetContext(ctx, &count, q, args...)
	})
	return count, err
}

func (repo *dashboardHistoryRepository) DeleteByDashboardID(ctx context.Context, dashboardID string) error {
	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.ExecContext(ctx, `DELETE FROM dashboard_history WHERE dashboard_id = ?`, dashboardID)
		return err
	})
}
