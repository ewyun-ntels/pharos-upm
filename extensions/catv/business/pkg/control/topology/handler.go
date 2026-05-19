package topology

import (
	"log/slog"
	net_http "net/http"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/extensions/catv/business/pkg/control/common/tables"
)

func createListHandler[T any](
	queryFunc func(...string) ([]T, error),
	errorMsg string,
	pathParams []string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := make([]string, len(pathParams))
		for i, paramName := range pathParams {
			params[i] = c.Param(paramName)
		}

		results, err := queryFunc(params...)
		if err != nil {
			slog.Error(errorMsg, "error", err, "params", params)
			c.JSON(net_http.StatusInternalServerError, external.NewErrorResponse("failed to retrieve data"))
			return
		}

		c.JSON(net_http.StatusOK, map[string]any{
			"items": results,
			"count": len(results),
		})
	}
}

func getSosHandler(config common.Config) gin.HandlerFunc {
	return createListHandler(
		func(params ...string) ([]string, error) {
			return tables.NewStbInformationTable(config).GetSos()
		},
		"Failed to query SO list",
		[]string{})
}

func getL3sHandler(config common.Config) gin.HandlerFunc {
	return createListHandler(
		func(params ...string) ([]string, error) {
			return tables.NewStbInformationTable(config).GetL3s(params[0])
		},
		"Failed to query L3 list",
		[]string{"so_id"})
}

func getCellsHandler(config common.Config) gin.HandlerFunc {
	return createListHandler(
		func(params ...string) ([]string, error) {
			return tables.NewStbInformationTable(config).GetCells(params[0], params[1])
		},
		"Failed to query cell list",
		[]string{"so_id", "l3_id"})
}

func getSettopboxesHandler(config common.Config) gin.HandlerFunc {
	return createListHandler(
		func(params ...string) ([]string, error) {
			return tables.NewStbInformationTable(config).GetSettopboxes(params[0], params[1], params[2])
		},
		"Failed to query settopbox list",
		[]string{"so_id", "l3_id", "cell_id"})
}
