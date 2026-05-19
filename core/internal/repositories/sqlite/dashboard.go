package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/shared/types/dashboard"
)

var _ repositories.DashboardRepository = (*dashboardRepository)(nil)

// dashboardRepository implements the repositories.DashboardRepository interface
// using SQLite as the backend.
type dashboardRepository struct {
	repositories.BaseRepository
}

func newDashboardRepository(config orm.DatabaseConfig) *dashboardRepository {
	return &dashboardRepository{
		BaseRepository: repositories.NewBaseRepository(config),
	}
}

// SQLite 테이블 필드 구조
type sqliteDashboard struct {
	ID          string  `db:"id"`
	Config      string  `db:"config"`
	FolderID    *string `db:"folder_id"`
	Annotations string  `db:"annotations"`
}

func (sd *sqliteDashboard) set(dashboardEntity repositories.DashboardEntity) error {
	sd.ID = dashboardEntity.ID
	sd.FolderID = dashboardEntity.FolderID

	if bytes, err := dashboardEntity.Config.Marshal(); err != nil {
		return err
	} else {
		sd.Config = string(bytes)
	}

	if len(dashboardEntity.Annotations) == 0 {
		sd.Annotations = "[]"
	} else if bytes, err := json.Marshal(dashboardEntity.Annotations); err != nil {
		return err
	} else {
		sd.Annotations = string(bytes)
	}

	return nil
}

func (sd *sqliteDashboard) toDashboardEntity() (*repositories.DashboardEntity, error) {
	dashboardConfig, err := dashboard.UnmarshalDashboardConfig([]byte(sd.Config))
	if err != nil {
		return nil, err
	}

	var annotations []dashboard.Annotation
	if sd.Annotations != "" && sd.Annotations != "[]" {
		if err := json.Unmarshal([]byte(sd.Annotations), &annotations); err != nil {
			annotations = nil
		}
	}

	return &repositories.DashboardEntity{
		ID:          sd.ID,
		Config:      dashboardConfig,
		FolderID:    sd.FolderID,
		Annotations: annotations,
	}, nil
}

// Create inserts a new dashboard into the database
func (repo *dashboardRepository) Create(ctx context.Context, dashboard *repositories.DashboardEntity) error {
	query := `INSERT INTO dashboard (id, config, annotations) VALUES (?, ?, ?)`

	dbDashboard := sqliteDashboard{}
	if err := dbDashboard.set(*dashboard); err != nil {
		return err
	}

	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.ExecContext(ctx, query, dbDashboard.ID, dbDashboard.Config, dbDashboard.Annotations)
		return err
	})
}

// GetByID retrieves a dashboard by its ID
func (repo *dashboardRepository) GetByID(ctx context.Context, id string) (*repositories.DashboardEntity, error) {
	query := `SELECT id, config, folder_id, annotations FROM dashboard WHERE id = ?`

	var dbDashboard sqliteDashboard
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.GetContext(ctx, &dbDashboard, query, id)
	})
	if err != nil {
		return nil, err
	}

	return dbDashboard.toDashboardEntity()
}

// ListAll retrieves all dashboards from the database
func (repo *dashboardRepository) ListAll(ctx context.Context) ([]*repositories.DashboardEntity, error) {
	query := `SELECT id, config, folder_id, annotations FROM dashboard`

	var dbDashboards []sqliteDashboard
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.SelectContext(ctx, &dbDashboards, query)
	})
	if err != nil {
		return nil, err
	}

	dashboards := make([]*repositories.DashboardEntity, 0, len(dbDashboards))
	for _, dbDashboard := range dbDashboards {
		if dashboard, err := dbDashboard.toDashboardEntity(); err != nil {
			dashboards = append(dashboards, &repositories.DashboardEntity{ID: dbDashboard.ID, FolderID: dbDashboard.FolderID})
		} else {
			dashboards = append(dashboards, dashboard)
		}
	}

	return dashboards, nil
}

// Update updates an existing dashboard's config
func (repo *dashboardRepository) Update(ctx context.Context, dashboard *repositories.DashboardEntity) error {
	query := `UPDATE dashboard SET config = ? WHERE id = ?`

	dbDashboard := sqliteDashboard{}
	if err := dbDashboard.set(*dashboard); err != nil {
		return err
	}

	return repo.Handler(func(db *sqlx.DB) error {
		res, err := db.ExecContext(ctx, query, dbDashboard.Config, dbDashboard.ID)
		if err != nil {
			return err
		}

		n, _ := res.RowsAffected()
		if n == 0 {
			return sql.ErrNoRows
		}
		return nil
	})
}

// DeleteByID deletes a dashboard by its ID
func (repo *dashboardRepository) DeleteByID(ctx context.Context, id string) error {
	query := `DELETE FROM dashboard WHERE id = ?`

	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.ExecContext(ctx, query, id)
		return err
	})
}

// UpdateAnnotations replaces the annotations list for a dashboard
func (repo *dashboardRepository) UpdateAnnotations(ctx context.Context, id string, annotations []dashboard.Annotation) error {
	query := `UPDATE dashboard SET annotations = ? WHERE id = ?`

	var annotationsJSON string
	if len(annotations) == 0 {
		annotationsJSON = "[]"
	} else if bytes, err := json.Marshal(annotations); err != nil {
		return err
	} else {
		annotationsJSON = string(bytes)
	}

	return repo.Handler(func(db *sqlx.DB) error {
		res, err := db.ExecContext(ctx, query, annotationsJSON, id)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return sql.ErrNoRows
		}
		return nil
	})
}
