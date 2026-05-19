package userhandler

import (
	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/internal/repositories/factory"
	"ntels.com/pharos/core/internal/user"
	"ntels.com/pharos/core/pkg/common"
	coreRole "ntels.com/pharos/core/pkg/role"
	coreUser "ntels.com/pharos/core/pkg/user"
)

type Api struct {
	configPath          string
	config              common.Config
	roleMetadataService *coreRole.MetadataService
	userMetadataService *coreUser.MetadataService
	userStore           user.Store
}

func (a *Api) Init(configPath string, config common.Config) {
	a.configPath = configPath
	a.config = config
}

func (a *Api) Use() bool {
	return true
}

func (a *Api) Load() error {
	// Initialize role metadata service
	roleConfigRepo, err := factory.NewRoleMetadataConfigRepository(factory.RepositoryOptions{
		DatabaseConfig: a.config.Database,
	})
	if err != nil {
		return err
	}
	a.roleMetadataService = coreRole.NewMetadataService(roleConfigRepo)

	// Initialize user metadata service
	userConfigRepo, err := factory.NewUserMetadataConfigRepository(factory.RepositoryOptions{
		DatabaseConfig: a.config.Database,
	})
	if err != nil {
		return err
	}
	a.userMetadataService = coreUser.NewMetadataService(userConfigRepo)

	// Initialize user store
	userStore, err := user.NewStore(user.GetUserStoreConfig(a.config))
	if err != nil {
		return err
	}
	a.userStore = userStore

	return nil
}

func (a *Api) Unload() {}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	config = a.config

	router, _ := routes.(gin.IRouter)

	RegisterRoutes(router, a.userStore, a.roleMetadataService, a.userMetadataService)
}

func (a *Api) GetRelativePath() string {
	return ""
}
