package gin

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/internal/cert"
	"ntels.com/pharos/core/pkg/common"
	migrationAPI "ntels.com/pharos/core/pkg/migration/api"
	"ntels.com/pharos/core/pkg/server/http/api"
	"ntels.com/pharos/third_party/GoVisual"
)

type Server struct {
	mutex sync.RWMutex

	configPath string
	config     common.Config

	engine *gin.Engine
	server http.Server

	apis []api.Api
}

func (s *Server) Start(configPath string, config common.Config, logger *slog.Logger) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.configPath = configPath
	s.config = config

	// Store config globally for middleware access
	common.SetGlobalConfig(config)

	if _, exist := s.config.Servers[s.config.Serve.ServerSchema]; !exist {
		return external.ErrorInvalidServerSchema
	}

	s.engine = gin.New()

	gin.DefaultWriter = slogWriter{logger: logger, level: slog.LevelInfo}
	gin.DefaultErrorWriter = slogWriter{logger: logger, level: slog.LevelError}

	s.engine.Use(slogMiddleware(logger), slogRecovery(logger))
	// Security headers for UI and API responses
	s.engine.Use(securityHeaders())
	// CSRF protection for cookie-based authentication
	s.engine.Use(csrfMiddleware())

	s.engine.Any("/", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/ui") })
	s.engine.Any("/ui/*filename", Static)

	s.apis = api.GetApis(s.configPath, s.config)

	for index, apiGroup := range s.apis {
		if err := s.apis[index].Load(); err != nil {
			slog.Error("api load failed", "index", index, "path", s.apis[index].GetRelativePath(), "error", err)
			return err
		}

		s.apis[index].RegisterRoutes(s.engine.Group(apiGroup.GetRelativePath()))
	}

	go func() {
		if err := s.listenAndServe(); err != nil {
			panic(err)
		}
	}()

	go s.migration()

	return nil
}

func (s *Server) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for index := len(s.apis) - 1; index >= 0; index-- {
		s.apis[index].Unload()
	}

	if s.config.History.Use {
		GoVisual.Unload()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}
	<-ctx.Done()

	return nil
}

func (s *Server) listenAndServe() error {
	serverConfig := s.config.Servers[s.config.Serve.ServerSchema]

	if serverConfig.Port == 0 {
		serverConfig.Port = 8080
	}

	var handler http.Handler
	handler = s.engine
	if s.config.History.Use {
		ignorePaths := []string{"/ui/"}
		ignorePaths = append(ignorePaths, s.config.History.IgnorePaths...)

		handler = GoVisual.Wrap(
			s.engine,
			//GoVisual.WithClickHouseStorage(&s.config),
			GoVisual.WithPostgresStorage(s.config.Statistics.Database.GetDataSourceName(), 
					"history_requests"), // ewyun
			GoVisual.WithRequestBodyLogging(true),
			GoVisual.WithResponseBodyLogging(true),
			GoVisual.WithIgnorePaths(ignorePaths...),
		)
	}

	// https://gin-gonic.com/ko-kr/docs/examples/graceful-restart-or-stop/
	s.server = http.Server{
		Addr:    ":" + strconv.Itoa(serverConfig.Port),
		Handler: handler,
	}

	switch s.config.Serve.ServerSchema {
	case common.ServerTypeHttp:
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case common.ServerTypeHttps:
		if tlsConfig, err := cert.SetCertification(serverConfig); err != nil {
			return err
		} else {
			s.server.TLSConfig = tlsConfig
		}

		if err := s.server.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	default:
		return external.ErrorInvalidServerType
	}

	return nil
}

// Run API migrations in background after server starts listening
// so that HTTP endpoints are ready to accept migration requests.
// This uses github.com/loykin/apimigrate with settings from [migrataion] in config TOML.
func (s *Server) migration() {
	// Defer a small delay to maximize chance the listener is fully bound
	// (runEngine already started the listener in a goroutine, but add a tiny buffer).
	time.Sleep(200 * time.Millisecond)
	// start migration runner (non-blocking, logs errors internally)
	migrationAPI.Run(s.config)
}
