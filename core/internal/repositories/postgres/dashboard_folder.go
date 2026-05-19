package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories"
)

var _ repositories.DashboardFolderRepository = (*dashboardFolderRepository)(nil)

type dashboardFolderRepository struct {
	repositories.BaseRepository
}

func newDashboardFolderRepository(config orm.DatabaseConfig) *dashboardFolderRepository {
	return &dashboardFolderRepository{
		BaseRepository: repositories.NewBaseRepository(config),
	}
}

type postgresDashboardFolder struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (pf *postgresDashboardFolder) toEntity() *repositories.DashboardFolderEntity {
	return &repositories.DashboardFolderEntity{
		ID:        pf.ID,
		Name:      pf.Name,
		CreatedAt: pf.CreatedAt,
		UpdatedAt: pf.UpdatedAt,
	}
}

func (repo *dashboardFolderRepository) Create(ctx context.Context, folder *repositories.DashboardFolderEntity) error {
	query := `INSERT INTO dashboard_folder (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`
	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.ExecContext(ctx, query, folder.ID, folder.Name, folder.CreatedAt, folder.UpdatedAt)
		return err
	})
}

func (repo *dashboardFolderRepository) GetByID(ctx context.Context, id string) (*repositories.DashboardFolderEntity, error) {
	query := `SELECT id, name, created_at, updated_at FROM dashboard_folder WHERE id = $1`
	var row postgresDashboardFolder
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.GetContext(ctx, &row, query, id)
	})
	if err != nil {
		return nil, err
	}
	return row.toEntity(), nil
}

func (repo *dashboardFolderRepository) ListAll(ctx context.Context) ([]*repositories.DashboardFolderEntity, error) {
	query := `SELECT id, name, created_at, updated_at FROM dashboard_folder ORDER BY name`
	var rows []postgresDashboardFolder
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.SelectContext(ctx, &rows, query)
	})
	if err != nil {
		return nil, err
	}
	result := make([]*repositories.DashboardFolderEntity, 0, len(rows))
	for _, r := range rows {
		result = append(result, r.toEntity())
	}
	return result, nil
}

func (repo *dashboardFolderRepository) Update(ctx context.Context, folder *repositories.DashboardFolderEntity) error {
	query := `UPDATE dashboard_folder SET name = $1, updated_at = $2 WHERE id = $3`
	return repo.Handler(func(db *sqlx.DB) error {
		res, err := db.ExecContext(ctx, query, folder.Name, folder.UpdatedAt, folder.ID)
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

func (repo *dashboardFolderRepository) DeleteByID(ctx context.Context, id string) error {
	query := `DELETE FROM dashboard_folder WHERE id = $1`
	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.ExecContext(ctx, query, id)
		return err
	})
}

func (repo *dashboardFolderRepository) MoveDashboard(ctx context.Context, dashboardID string, folderID *string) error {
	query := `UPDATE dashboard SET folder_id = $1 WHERE id = $2`
	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.ExecContext(ctx, query, folderID, dashboardID)
		return err
	})
}
