package ui_config

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

type HomeDashboard struct {
	DashboardID string `json:"dashboard_id" db:"dashboard_id"`

	config common.Config
}

func (h *HomeDashboard) GetHandler(_ *gin.Context) (int, any) {
	handler := func(db *sqlx.DB) error {
		return db.Get(h, `SELECT dashboard_id FROM home_dashboard LIMIT 1;`)
	}
	if err := orm.Handler(orm.DriverDefault, &h.config.Database, handler); errors.Is(err, sql.ErrNoRows) {
		return http.StatusOK, nil
	} else if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, h
}

func (h *HomeDashboard) PutHandler(c *gin.Context) (int, any) {
	if err := c.ShouldBindJSON(h); err != nil {
		return http.StatusBadRequest, external.ErrorResponse{Message: err.Error()}
	}

	handler := func(db *sqlx.DB) error {
		result, err := db.NamedExec(`UPDATE home_dashboard SET dashboard_id = :dashboard_id;`, h)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			_, err = db.NamedExec(`INSERT INTO home_dashboard(dashboard_id) VALUES (:dashboard_id);`, h)
			return err
		}

		return nil
	}
	if err := orm.Handler(orm.DriverDefault, &h.config.Database, handler); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

func (h *HomeDashboard) DeleteHandler(_ *gin.Context) (int, any) {
	handler := func(db *sqlx.DB) error {
		_, err := db.Exec(`DELETE FROM home_dashboard;`)
		return err
	}
	if err := orm.Handler(orm.DriverDefault, &h.config.Database, handler); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

func getHomeHandler(config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := HomeDashboard{config: config}

		switch c.Request.Method {
		case http.MethodGet:
			c.JSON(h.GetHandler(c))
		case http.MethodPut:
			c.JSON(h.PutHandler(c))
		case http.MethodDelete:
			c.JSON(h.DeleteHandler(c))
		default:
			c.JSON(http.StatusNotImplemented, nil)
		}
	}
}
