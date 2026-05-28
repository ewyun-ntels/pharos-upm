package homev2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/common"
)

type ConfigResponse struct {
	Datasource string `json:"datasource"`
}

func GetConfigHandler(config common.Config) gin.HandlerFunc {
	resp := ConfigResponse{
		Datasource: config.HomeV2.Datasource,
	}
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, resp)
	}
}
