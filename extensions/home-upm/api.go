package homeupm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"ntels.com/pharos/core/pkg/authhandler"
	"ntels.com/pharos/core/pkg/common"
	sharedRole "ntels.com/pharos/shared/types/role"
)

const (
	defaultDatasource      = "prometheus-metric"
	defaultRestartWindow   = "1h"
	defaultWarningRestarts = 1
	defaultErrorRestarts   = 3

	kubernetesRequestTimeout = 15 * time.Second
)

var promDurationPattern = regexp.MustCompile(`^[1-9][0-9]*(ms|s|m|h|d|w|y)$`)

type homeUPMSection struct {
	Datasource       string `toml:"datasource"`
	RestartWindow    string `toml:"restart_window"`
	WarningRestarts  int    `toml:"warning_restarts"`
	ErrorRestarts    int    `toml:"error_restarts"`
}

type homeUPMTomlConfig struct {
	HomeUPM homeUPMSection `toml:"home_upm"`
}

type ConfigResponse struct {
	Datasource       string `json:"datasource"`
	RestartWindow    string `json:"restartWindow"`
	WarningRestarts  int    `json:"warningRestarts"`
	ErrorRestarts    int    `json:"errorRestarts"`
}

type Api struct {
	datasource       string
	restartWindow    string
	warningRestarts  int
	errorRestarts    int
}

func (a *Api) Init(configPath string, _ common.Config) {
	a.applyDefaults()

	if configPath == "" {
		slog.Warn("home-upm: config path is empty")
		return
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Warn("home-upm: failed to read config file", "path", configPath, "error", err)
		return
	}

	var cfg homeUPMTomlConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		slog.Warn("home-upm: failed to parse [home_upm] config", "error", err)
		return
	}

	a.applyConfig(cfg.HomeUPM)
	slog.Info(
		"home-upm: config loaded",
		"datasource", a.datasource,
		"restart_window", a.restartWindow,
		"warning_restarts", a.warningRestarts,
		"error_restarts", a.errorRestarts,
	)
}

func (a *Api) Use() bool { return true }

func (a *Api) Load() error { return nil }

func (a *Api) Unload() {}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	podDeleteAuth := authhandler.GetAuthenticationHandler(
		true,
		string(sharedRole.RoleSuperAdmin),
		string(sharedRole.RolePodDelete),
	)

	routes.GET("", a.getConfig)
	routes.DELETE("/pods/:namespace/:name", podDeleteAuth, a.deletePod)
}

func (a *Api) GetRelativePath() string {
	return "/home-upm"
}

func (a *Api) getConfig(c *gin.Context) {
	c.JSON(http.StatusOK, ConfigResponse{
		Datasource:       a.datasource,
		RestartWindow:    a.restartWindow,
		WarningRestarts:  a.warningRestarts,
		ErrorRestarts:    a.errorRestarts,
	})
}

func (a *Api) deletePod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace and pod name are required"})
		return
	}

	if err := deleteKubernetesPod(c.Request.Context(), namespace, name); err != nil {
		statusCode := http.StatusServiceUnavailable
		var statusErr apierrors.APIStatus
		if errors.As(err, &statusErr) {
			if code := statusErr.Status().Code; code > 0 {
				statusCode = int(code)
			}
		}
		slog.Error("home-upm: failed to delete pod", "namespace", namespace, "pod", name, "status", statusCode, "error", err)
		c.JSON(statusCode, gin.H{
			"error":   "failed to delete pod",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "pod delete requested",
		"namespace": namespace,
		"pod":       name,
	})
}

func (a *Api) applyDefaults() {
	a.datasource = defaultDatasource
	a.restartWindow = defaultRestartWindow
	a.warningRestarts = defaultWarningRestarts
	a.errorRestarts = defaultErrorRestarts
}

func deleteKubernetesPod(ctx context.Context, namespace, name string) error {
	ctx, cancel := context.WithTimeout(ctx, kubernetesRequestTimeout)
	defer cancel()

	config, source, err := buildKubernetesConfig()
	if err != nil {
		return err
	}
	config.Timeout = kubernetesRequestTimeout

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("create kubernetes client from %s: %w", source, err)
	}

	if err := clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("delete pod using %s: %w", source, err)
	}
	return nil
}

func buildKubernetesConfig() (*rest.Config, string, error) {
	kubeConfig, kubeConfigErr := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if kubeConfigErr == nil {
		return kubeConfig, "kubeconfig", nil
	}

	inClusterConfig, inClusterErr := rest.InClusterConfig()
	if inClusterErr == nil {
		return inClusterConfig, "in-cluster service account", nil
	}

	return nil, "", fmt.Errorf("load kubeconfig failed: %v; load in-cluster config failed: %w", kubeConfigErr, inClusterErr)
}

func (a *Api) applyConfig(cfg homeUPMSection) {
	if cfg.Datasource != "" {
		a.datasource = cfg.Datasource
	} else {
		slog.Warn("home-upm: [home_upm] datasource is not configured; using default", "datasource", a.datasource)
	}

	if isValidPromDuration(cfg.RestartWindow) {
		a.restartWindow = cfg.RestartWindow
	} else if cfg.RestartWindow != "" {
		slog.Warn("home-upm: invalid restart_window; using default", "restart_window", cfg.RestartWindow, "default", a.restartWindow)
	}

	if cfg.WarningRestarts > 0 {
		a.warningRestarts = cfg.WarningRestarts
	}
	if cfg.ErrorRestarts > 0 {
		a.errorRestarts = cfg.ErrorRestarts
	}
	if a.warningRestarts > a.errorRestarts {
		slog.Warn(
			"home-upm: warning_restarts is greater than error_restarts; clamping warning threshold",
			"warning_restarts", a.warningRestarts,
			"error_restarts", a.errorRestarts,
		)
		a.warningRestarts = a.errorRestarts
	}
}

func isValidPromDuration(value string) bool {
	return promDurationPattern.MatchString(value)
}
