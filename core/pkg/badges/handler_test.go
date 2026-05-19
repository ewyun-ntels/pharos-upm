package badges

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/query_builder"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins"
	"ntels.com/pharos/core/pkg/plugins/model"
)

// 테스트용 전역 config 변수
var config common.Config

// GetBadge는 테스트용 wrapper 함수입니다
func GetBadge(c *gin.Context) {
	handler := GetBadgeHandler(config)
	handler(c)
}

// PostBadge는 테스트용 wrapper 함수입니다
func PostBadge(c *gin.Context) {
	handler := GetPostBadgeHandler(config)
	handler(c)
}

// GetBadgeList는 테스트용 wrapper 함수입니다
func GetBadgeList(c *gin.Context) {
	handler := GetBadgeListHandler(config)
	handler(c)
}

// PutBadge는 테스트용 wrapper 함수입니다
func PutBadge(c *gin.Context) {
	handler := GetPutBadgeHandler(config)
	handler(c)
}

// DeleteBadge는 테스트용 wrapper 함수입니다
func DeleteBadge(c *gin.Context) {
	handler := GetDeleteBadgeHandler(config)
	handler(c)
}

// Load는 테스트용 wrapper 함수입니다
func Load(_ string, cfg common.Config, routes gin.IRoutes) error {
	config = cfg

	// initialize cache TTL from config
	var ttl time.Duration
	if cfg.Badges.CacheTTL == nil {
		ttl = time.Minute
	} else if *cfg.Badges.CacheTTL < 0 {
		ttl = time.Minute
	} else {
		ttl = *cfg.Badges.CacheTTL
	}
	badgeCache = newCache(ttl)

	RegisterRoutes(routes)
	return nil
}

// RegisterRoutes는 테스트용 함수입니다
func RegisterRoutes(routes gin.IRoutes) {
	routes.GET("/:name", GetBadge)
	routes.GET("/", GetBadgeList)
	routes.POST("/", PostBadge)
	routes.PUT("/:name", PutBadge)
	routes.DELETE("/:name", DeleteBadge)
}

// setup sqlite for badges table
func newBadgesTestConfig(t *testing.T) common.Config {
	t.Helper()
	cfg := common.Config{}
	cfg.Database.Driver = orm.DriverSqlite
	cfg.Database.SQLite.Path = t.TempDir() + "/badges.db"
	// create badges table for tests (runtime code does not auto-create)
	_ = orm.Handler(orm.DriverDefault, &cfg.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS badges (
			name TEXT PRIMARY KEY,
			datasource TEXT NOT NULL,
			query_json TEXT NOT NULL,
			column_name TEXT NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);
		`)
		return err
	})
	return cfg
}

// stub dsQuery indirection for tests
func withStubDatasource(t *testing.T, frameData []map[string]any, errMsg string, fn func(callCount *int)) {
	old := dsQuery
	var calls int
	dsQuery = func(_ context.Context, req plugins.DsQueryRequest) *plugins.DsQueryResponse {
		calls++
		res := &plugins.DsQueryResponse{Results: map[string]model.QueryDataResult{}}
		res.Results[req.Queries[0].ID] = model.QueryDataResult{Frame: orm.DatabaseResponse{Data: frameData}, Error: errMsg}
		return res
	}
	t.Cleanup(func() { dsQuery = old })
	fn(&calls)
}

func TestPostAndGetBadgeFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newBadgesTestConfig(t)
	config = cfg

	// POST to create a badge
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	payload := BadgeSpec{
		Name:       "uptime",
		Datasource: "ds1",
		Query:      query_builder.QuerySpec{Query: "select 1 as v", Variables: map[string]any{}},
		Column:     "v",
	}
	b, _ := json.Marshal(payload)
	c.Request = httptest.NewRequest(http.MethodPost, "/badges", bytes.NewReader(b))

	PostBadge(c)
	require.Equal(t, http.StatusOK, rec.Code)

	// GET should execute stored query and return value from frame
	withStubDatasource(t, []map[string]any{{"v": 123}}, "", func(calls *int) {
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{gin.Param{Key: "name", Value: "uptime"}}
		GetBadge(c2)
		require.Equal(t, http.StatusOK, rec2.Code)
		var out BadgeValueResponse
		_ = json.Unmarshal(rec2.Body.Bytes(), &out)
		require.Equal(t, "uptime", out.Name)
		require.Equal(t, float64(123), out.Value)
		// second call should hit cache; dsQuery should not be called again
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Params = gin.Params{gin.Param{Key: "name", Value: "uptime"}}
		GetBadge(c3)
		require.Equal(t, http.StatusOK, rec3.Code)
		require.Equal(t, 1, *calls)
	})
}

func TestGetBadge_NoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newBadgesTestConfig(t)
	config = cfg

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{gin.Param{Key: "name", Value: "nope"}}
	GetBadge(c)
	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestGetBadge_ListAndErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newBadgesTestConfig(t)
	config = cfg

	// Seed two badges via PostBadge
	for _, nm := range []string{"b1", "b2"} {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		payload := BadgeSpec{Name: nm, Datasource: "ds1", Query: query_builder.QuerySpec{Query: "select 1 as v"}, Column: "v"}
		b, _ := json.Marshal(payload)
		c.Request = httptest.NewRequest(http.MethodPost, "/badges", bytes.NewReader(b))
		PostBadge(c)
		require.Equal(t, http.StatusOK, rec.Code)
	}

	// List badges
	recList := httptest.NewRecorder()
	cList, _ := gin.CreateTestContext(recList)
	GetBadgeList(cList)
	require.Equal(t, http.StatusOK, recList.Code)
	var names []string
	_ = json.Unmarshal(recList.Body.Bytes(), &names)
	require.Equal(t, []string{"b1", "b2"}, names)

	// GetBadge: query_builder error path by supplying an invalid template token
	// Save a badge with an invalid template that pongo2 cannot parse
	recBad := httptest.NewRecorder()
	cBad, _ := gin.CreateTestContext(recBad)
	bad := BadgeSpec{Name: "bad", Datasource: "ds1", Query: query_builder.QuerySpec{Query: "{{ invalid #"}, Column: "v"}
	bb, _ := json.Marshal(bad)
	cBad.Request = httptest.NewRequest(http.MethodPost, "/badges", bytes.NewReader(bb))
	PostBadge(cBad)
	require.Equal(t, http.StatusOK, recBad.Code)

	recGet := httptest.NewRecorder()
	cGet, _ := gin.CreateTestContext(recGet)
	cGet.Params = gin.Params{gin.Param{Key: "name", Value: "bad"}}
	GetBadge(cGet)
	require.Equal(t, http.StatusInternalServerError, recGet.Code)

	// GetBadge: Datasource missing result id path
	// First, store a normal badge
	recStore := httptest.NewRecorder()
	cStore, _ := gin.CreateTestContext(recStore)
	good := BadgeSpec{Name: "ds_missing", Datasource: "ds1", Query: query_builder.QuerySpec{Query: "select 1 as v"}, Column: "v"}
	gb, _ := json.Marshal(good)
	cStore.Request = httptest.NewRequest(http.MethodPost, "/badges", bytes.NewReader(gb))
	PostBadge(cStore)
	require.Equal(t, http.StatusOK, recStore.Code)

	old := dsQuery
	dsQuery = func(_ context.Context, req plugins.DsQueryRequest) *plugins.DsQueryResponse {
		// Return empty results map to simulate missing id
		return &plugins.DsQueryResponse{Results: map[string]model.QueryDataResult{}}
	}
	defer func() { dsQuery = old }()

	recMiss := httptest.NewRecorder()
	cMiss, _ := gin.CreateTestContext(recMiss)
	cMiss.Params = gin.Params{gin.Param{Key: "name", Value: "ds_missing"}}
	GetBadge(cMiss)
	require.Equal(t, http.StatusInternalServerError, recMiss.Code)

	// GetBadge: Datasource error string path
	dsQuery = func(_ context.Context, req plugins.DsQueryRequest) *plugins.DsQueryResponse {
		m := map[string]model.QueryDataResult{"q1": {Error: "boom"}}
		return &plugins.DsQueryResponse{Results: m}
	}
	defer func() { dsQuery = old }() // ensure restored

	recErr := httptest.NewRecorder()
	cErr, _ := gin.CreateTestContext(recErr)
	cErr.Params = gin.Params{gin.Param{Key: "name", Value: "ds_missing"}}
	GetBadge(cErr)
	require.Equal(t, http.StatusInternalServerError, recErr.Code)
}

func TestCacheInvalidatedOnPostAndDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newBadgesTestConfig(t)
	config = cfg

	// Create badge
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	payload := BadgeSpec{
		Name:       "cacheBadge",
		Datasource: "ds1",
		Query:      query_builder.QuerySpec{Query: "select 1 as v"},
		Column:     "v",
	}
	b, _ := json.Marshal(payload)
	c.Request = httptest.NewRequest(http.MethodPost, "/badges", bytes.NewReader(b))
	PostBadge(c)
	require.Equal(t, http.StatusOK, rec.Code)

	// First GET uses stub returning 100; second GET should be cached
	withStubDatasource(t, []map[string]any{{"v": 100}}, "", func(calls *int) {
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{gin.Param{Key: "name", Value: "cacheBadge"}}
		GetBadge(c1)
		require.Equal(t, http.StatusOK, rec1.Code)
		// cached call
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{gin.Param{Key: "name", Value: "cacheBadge"}}
		GetBadge(c2)
		require.Equal(t, http.StatusOK, rec2.Code)
		require.Equal(t, 1, *calls)
	})

	// Update badge (invalidate)
	recPut := httptest.NewRecorder()
	cPut, _ := gin.CreateTestContext(recPut)
	payload.Query = query_builder.QuerySpec{Query: "select 2 as v"}
	b2, _ := json.Marshal(payload)
	cPut.Request = httptest.NewRequest(http.MethodPut, "/badges/cacheBadge", bytes.NewReader(b2))
	cPut.Params = gin.Params{gin.Param{Key: "name", Value: "cacheBadge"}}
	PutBadge(cPut)
	require.Equal(t, http.StatusOK, recPut.Code)

	// Now stub returns 200; first GET after PUT should call dsQuery again
	withStubDatasource(t, []map[string]any{{"v": 200}}, "", func(calls *int) {
		rec3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(rec3)
		c3.Params = gin.Params{gin.Param{Key: "name", Value: "cacheBadge"}}
		GetBadge(c3)
		require.Equal(t, http.StatusOK, rec3.Code)
		require.Equal(t, 1, *calls)
	})

	// Delete (invalidate)
	recDel := httptest.NewRecorder()
	cDel, _ := gin.CreateTestContext(recDel)
	cDel.Params = gin.Params{gin.Param{Key: "name", Value: "cacheBadge"}}
	DeleteBadge(cDel)
	require.Equal(t, http.StatusOK, recDel.Code)
}

func TestPostBadge_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newBadgesTestConfig(t)
	config = cfg

	cases := []BadgeSpec{
		{},
		{Name: "n"},
		{Name: "n", Datasource: "d"},
		{Name: "n", Datasource: "d", Query: query_builder.QuerySpec{Query: ""}},
	}
	codes := []int{http.StatusBadRequest, http.StatusBadRequest, http.StatusBadRequest, http.StatusBadRequest}
	for i, in := range cases {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		b, _ := json.Marshal(in)
		c.Request = httptest.NewRequest(http.MethodPost, "/badges", bytes.NewReader(b))
		PostBadge(c)
		require.Equal(t, codes[i], rec.Code)
	}
}

func TestGetBadge_NoCacheWhenTTLZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newBadgesTestConfig(t)
	config = cfg

	// create badge
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	payload := BadgeSpec{
		Name:       "ncache",
		Datasource: "ds1",
		Query:      query_builder.QuerySpec{Query: "select 1 as v"},
		Column:     "v",
	}
	b, _ := json.Marshal(payload)
	c.Request = httptest.NewRequest(http.MethodPost, "/badges", bytes.NewReader(b))
	PostBadge(c)
	require.Equal(t, http.StatusOK, rec.Code)

	// Disable cache explicitly
	badgeCache = newCache(0)

	withStubDatasource(t, []map[string]any{{"v": 10}}, "", func(calls *int) {
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{gin.Param{Key: "name", Value: "ncache"}}
		GetBadge(c1)
		require.Equal(t, http.StatusOK, rec1.Code)

		// second call should NOT be cached; dsQuery called again
		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Params = gin.Params{gin.Param{Key: "name", Value: "ncache"}}
		GetBadge(c2)
		require.Equal(t, http.StatusOK, rec2.Code)
		require.Equal(t, 2, *calls)
	})
}

func TestGetBadge_ValueNullWhenColumnMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newBadgesTestConfig(t)
	config = cfg

	// create badge that expects column 'missing'
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	payload := BadgeSpec{
		Name:       "colnull",
		Datasource: "ds1",
		Query:      query_builder.QuerySpec{Query: "select 1 as v"},
		Column:     "missing",
	}
	b, _ := json.Marshal(payload)
	c.Request = httptest.NewRequest(http.MethodPost, "/badges", bytes.NewReader(b))
	PostBadge(c)
	require.Equal(t, http.StatusOK, rec.Code)

	withStubDatasource(t, []map[string]any{{"v": 10}}, "", func(calls *int) {
		rec1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(rec1)
		c1.Params = gin.Params{gin.Param{Key: "name", Value: "colnull"}}
		GetBadge(c1)
		require.Equal(t, http.StatusOK, rec1.Code)
		var out BadgeValueResponse
		_ = json.Unmarshal(rec1.Body.Bytes(), &out)
		require.Nil(t, out.Value)
	})
}

func TestPostBadge_InvalidJSON_Returns500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newBadgesTestConfig(t)
	config = cfg

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/badges", bytes.NewReader([]byte("{")))
	PostBadge(c)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDeleteBadge_EmptyName_Returns500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newBadgesTestConfig(t)
	config = cfg

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	// no param set -> name "" -> model.Delete returns error
	DeleteBadge(c)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestLoad_And_RegisterRoutes_Coverage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Test Load with nil CacheTTL (default 1m) and 0 (disabled) and negative
	{
		cfg := common.Config{}
		cfg.Badges.CacheTTL = nil
		r := gin.New().Group("/badges")
		require.NoError(t, Load("", cfg, r))
	}
	{
		cfg := common.Config{}
		zero := time.Duration(0)
		cfg.Badges.CacheTTL = &zero
		r := gin.New().Group("/badges2")
		require.NoError(t, Load("", cfg, r))
	}
	{
		cfg := common.Config{}
		neg := time.Duration(-1)
		cfg.Badges.CacheTTL = &neg
		r := gin.New().Group("/badges3")
		require.NoError(t, Load("", cfg, r))
	}

	// Cover RegisterRoutes
	r := gin.New()
	RegisterRoutes(r.Group("/badges"))
	// fire a simple request to ensure route exists (GET list requires auth but handler is invoked only via middleware; here we won't check code, just ensure no panic in route setup)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/badges/x", nil)
	r.ServeHTTP(rec, req)
}
