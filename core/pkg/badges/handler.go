package badges

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/internal/query_builder"
	"ntels.com/pharos/core/pkg/badges/adapters"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins"
	"ntels.com/pharos/shared/types/badge"
)

// indirection for easier testing
var dsQuery = func(ctx context.Context, req plugins.DsQueryRequest) *plugins.DsQueryResponse {
	return plugins.DatasourceQuery(ctx, req)
}

// cache instance for badge values; default to 1 minute to avoid nil in tests without Load()
var badgeCache = newCache(time.Minute)

// GetBadge executes the stored query for the given badge name and returns the configured column's value.

func GetBadgeHandler(config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		// cache hit?
		if v, ok := badgeCache.Get(name); ok {
			adapter := adapters.NewBadgeAdapter()
			internalResponse := adapters.BadgeValueResponse{Name: name, Value: v}
			apiResponse := adapter.ToAPIBadgeValueResponse(internalResponse, true)
			c.JSON(http.StatusOK, apiResponse)
			return
		}

		m := Model{Config: &config}

		var qs query_builder.QuerySpec
		row, err := m.Get(name, &qs)
		if err != nil {
			slog.Error("badge get failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}
		if row == nil {
			// Badge not found: return 204 No Content
			// Note: c.Status() alone doesn't write headers in test contexts
			// Using c.JSON() with nil body ensures proper status code propagation
			c.JSON(http.StatusNoContent, nil)
			return
		}

		// Render and execute the query
		queryResp, err := query_builder.GetQuery(qs.Query, qs.Variables)
		if err != nil {
			slog.Error("GetQuery failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		req := plugins.DsQueryRequest{Queries: []plugins.DsQuery{{ID: "q1", DatasourceName: row.Datasource, SQL: string(queryResp), Timeout: 30}}}
		resp := dsQuery(c, req)
		qres, ok := resp.Results["q1"]
		if !ok || qres.Error != "" {
			if !ok {
				slog.Error("DatasourceQuery returned no result for id", "id", "q1")
				c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: "datasource query missing result"})
				return
			}
			slog.Error("DatasourceQuery error", "error", qres.Error)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: qres.Error})
			return
		}

		// Extract value from first row by configured column
		var value any
		if len(qres.Frame.Data) > 0 {
			first := qres.Frame.Data[0]
			if v, ok := first[row.ColumnName]; ok {
				value = v
			}
		}

		// store to cache
		badgeCache.Set(name, value)

		// Convert to API response using adapter
		adapter := adapters.NewBadgeAdapter()
		internalResponse := adapters.BadgeValueResponse{Name: name, Value: value}
		apiResponse := adapter.ToAPIBadgeValueResponse(internalResponse, false)
		c.JSON(http.StatusOK, apiResponse)
	}
}

// PostBadge registers a new badge definition (or updates if exists).
func GetPostBadgeHandler(config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		m := Model{Config: &config}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			slog.Error("read body failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		var apiSpec badge.BadgeSpec
		if err := json.Unmarshal(body, &apiSpec); err != nil {
			slog.Error("json.Unmarshal failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		// Convert API spec to internal spec using adapter
		adapter := adapters.NewBadgeAdapter()
		adapterSpec := adapter.FromAPIBadgeSpec(apiSpec)

		spec := BadgeSpec{
			Name:       adapterSpec.Name,
			Datasource: adapterSpec.Datasource,
			Query:      adapterSpec.Query,
			Column:     adapterSpec.Column,
		}

		if spec.Name == "" {
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "name is empty"})
			return
		}
		if spec.Datasource == "" {
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "datasource is empty"})
			return
		}
		if spec.Query.Query == "" {
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "query is empty"})
			return
		}
		if spec.Column == "" {
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "column is empty"})
			return
		}

		if err := m.Upsert(spec.Name, spec.Datasource, spec.Query, spec.Column); err != nil {
			slog.Error("badge upsert failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}
		// invalidate cache for this badge
		badgeCache.Delete(spec.Name)
		c.Status(http.StatusOK)
	}
}

// PutBadge updates an existing badge with idempotent semantics.
// It requires the badge name in URL path and matches it with the body.
func GetPutBadgeHandler(config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		m := Model{Config: &config}

		pathName := c.Param("name")
		if pathName == "" {
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "name path parameter is required"})
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			slog.Error("read body failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		var apiSpec badge.BadgeSpec
		if err := json.Unmarshal(body, &apiSpec); err != nil {
			slog.Error("json.Unmarshal failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		// If body name empty, set from path; otherwise it must match
		if apiSpec.Name == "" {
			apiSpec.Name = pathName
		} else if apiSpec.Name != pathName {
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "name in path and body must match"})
			return
		}

		// Convert API spec to internal spec using adapter
		adapter := adapters.NewBadgeAdapter()
		adapterSpec := adapter.FromAPIBadgeSpec(apiSpec)

		spec := BadgeSpec{
			Name:       adapterSpec.Name,
			Datasource: adapterSpec.Datasource,
			Query:      adapterSpec.Query,
			Column:     adapterSpec.Column,
		}

		if spec.Datasource == "" {
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "datasource is empty"})
			return
		}
		if spec.Query.Query == "" {
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "query is empty"})
			return
		}
		if spec.Column == "" {
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "column is empty"})
			return
		}

		if err := m.Upsert(spec.Name, spec.Datasource, spec.Query, spec.Column); err != nil {
			slog.Error("badge upsert failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}
		// invalidate cache for this badge
		badgeCache.Delete(spec.Name)
		c.Status(http.StatusOK)
	}
}

// GetBadgeList returns all badge names.
func GetBadgeListHandler(config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		m := Model{Config: &config}
		names, err := m.List()
		if err != nil {
			slog.Error("badge list failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusOK, names)
	}
}

// DeleteBadge removes a badge by name.
func GetDeleteBadgeHandler(config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		m := Model{Config: &config}
		if err := m.Delete(name); err != nil {
			slog.Error("badge delete failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}
		// invalidate cache for this badge
		badgeCache.Delete(name)
		c.Status(http.StatusOK)
	}
}
