package rolehandler

import (
	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/internal/repositories/factory"
	"ntels.com/pharos/core/pkg/common"
	coreRole "ntels.com/pharos/core/pkg/role"
)

type Api struct {
	configPath      string
	config          common.Config
	metadataService *coreRole.MetadataService
}

func (a *Api) Init(configPath string, config common.Config) {
	a.configPath = configPath
	a.config = config
}

func (a *Api) Use() bool {
	return true
}

func (a *Api) Load() error {
	// Initialize metadata config repository and service
	configRepo, err := factory.NewRoleMetadataConfigRepository(factory.RepositoryOptions{
		DatabaseConfig: a.config.Database,
	})
	if err != nil {
		return err
	}

	a.metadataService = coreRole.NewMetadataService(configRepo)

	return nil
}

func (a *Api) Unload() {}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	router, _ := routes.(gin.IRouter)
	RegisterRoutes(router, a.metadataService)
}

func (a *Api) GetRelativePath() string {
	return ""
}
