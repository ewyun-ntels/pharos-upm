package badges

import (
	"time"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/authhandler"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/shared/types/role"
)

type Api struct {
	configPath string
	config     common.Config
}

func (a *Api) Init(configPath string, config common.Config) {
	a.configPath = configPath
	a.config = config
}

func (a *Api) Use() bool {
	return true
}

func (a *Api) Load() error {
	// initialize cache TTL from config (pointer semantics)
	// nil => default 1m; 0 => disable; <0 => default 1m
	var ttl time.Duration
	if a.config.Badges.CacheTTL == nil {
		ttl = time.Minute
	} else if *a.config.Badges.CacheTTL < 0 {
		ttl = time.Minute
	} else {
		ttl = *a.config.Badges.CacheTTL
	}
	badgeCache = newCache(ttl)

	return nil
}

func (a *Api) Unload() {}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	routes.GET("",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin)),
		GetBadgeListHandler(a.config))
	routes.POST("",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin)),
		GetPostBadgeHandler(a.config))
	routes.PUT("/:name",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin)),
		GetPutBadgeHandler(a.config))
	routes.GET("/:name",
		GetBadgeHandler(a.config))
	routes.DELETE("/:name",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin)),
		GetDeleteBadgeHandler(a.config))
}

func (a *Api) GetRelativePath() string {
	return "/badges"
}
