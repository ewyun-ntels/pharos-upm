package homeupm

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
		Datasource: config.HomeUPM.Datasource,
	}
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, resp)
	}
}
