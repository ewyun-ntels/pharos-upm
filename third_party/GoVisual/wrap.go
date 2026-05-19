package GoVisual

import (
	"context"
	"net/http"
	"sync"

	"golang.org/x/exp/slog"
	middleware2 "ntels.com/pharos/third_party/GoVisual/internal/middleware"
	store2 "ntels.com/pharos/third_party/GoVisual/internal/store"
)

var (
	// Global signal handler to ensure we only have one
	shutdownFuncs []func(context.Context) error
	shutdownMutex sync.Mutex
)

// addShutdownFunc adds a shutdown function to be called on signal
func addShutdownFunc(fn func(context.Context) error) {
	if fn == nil {
		slog.Warn("Warning: Attempted to register nil shutdown function, ignoring")
		return
	}
	shutdownMutex.Lock()
	defer shutdownMutex.Unlock()
	shutdownFuncs = append(shutdownFuncs, fn)
}

func Unload() {
	ctx := context.Background()
	shutdownMutex.Lock()
	funcs := make([]func(context.Context) error, len(shutdownFuncs))
	copy(funcs, shutdownFuncs)
	shutdownMutex.Unlock()

	// Execute all shutdown functions
	for _, fn := range funcs {
		if err := fn(ctx); err != nil {
			slog.Error("Error during shutdown: %v", err)
		}
	}
}

// Wrap wraps an http.Handler with request visualization middleware
func Wrap(handler http.Handler, opts ...Option) http.Handler {
	// Apply options to default config
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	// Create store based on configuration
	var requestStore store2.Store
	var err error

	storeConfig := &store2.StorageConfig{
		Type:             config.StorageType,
		Capacity:         config.MaxRequests,
		ConnectionString: config.ConnectionString,
		TableName:        config.TableName,
		TTL:              config.RedisTTL,
		ExistingDB:       config.ExistingDB,
		ClickHouseConf:   config.ClickHouseConf,
	}

	requestStore, err = store2.NewStore(storeConfig)
	if err != nil {
		slog.Error("Failed to create configured storage backend: %v. Falling back to in-memory storage.", err)
		requestStore = store2.NewInMemoryStore(config.MaxRequests)
	}

	// Add store cleanup to shutdown functions
	addShutdownFunc(func(ctx context.Context) error {
		if err := requestStore.Close(); err != nil {
			slog.Error("Error closing storage: %v", err)
			return err
		}
		return nil
	})

	// Create middleware wrapper
	wrapped := middleware2.Wrap(handler, requestStore, config.LogRequestBody, config.LogResponseBody, config)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Otherwise, serve the application
		wrapped.ServeHTTP(w, r)
	})
}
