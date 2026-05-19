package authhandler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/internal/repositories"
)

type loginHistoryResponse struct {
	Data  []repositories.LoginHistoryRecord `json:"data"`
	Total int                               `json:"total"`
}

func getLoginHistoryHandler(repo repositories.LoginHistoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Query("username")

		limit := 20
		if v := c.Query("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
				limit = n
			}
		}

		offset := 0
		if v := c.Query("offset"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 0 {
				offset = n
			}
		}

		total, err := repo.Count(c.Request.Context(), username)
		if err != nil {
			slog.Error("Failed to count login history", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		records, err := repo.Query(c.Request.Context(), username, limit, offset)
		if err != nil {
			slog.Error("Failed to query login history", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		c.JSON(http.StatusOK, loginHistoryResponse{Data: records, Total: total})
	}
}
