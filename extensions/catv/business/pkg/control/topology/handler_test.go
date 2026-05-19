package topology

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/pkg/common"
)

func newTestRouter(config common.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// 인증 미들웨어 없이 핸들러 직접 등록 (unit test)
	r.GET("/topology/sos", getSosHandler(config))
	r.GET("/topology/sos/:so_id/l3s", getL3sHandler(config))
	r.GET("/topology/sos/:so_id/l3s/:l3_id/cells", getCellsHandler(config))
	r.GET("/topology/sos/:so_id/l3s/:l3_id/cells/:cell_id/settopboxes", getSettopboxesHandler(config))
	return r
}

// RegisterRoutes — 등록된 경로 확인
func TestRegisterRoutes_RoutesRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	a := &Api{}
	a.Init("", common.Config{})
	a.RegisterRoutes(r)

	routes := r.Routes()
	routeMap := make(map[string]string)
	for _, route := range routes {
		routeMap[route.Method+":"+route.Path] = route.Path
	}

	assert.Contains(t, routeMap, "GET:/topology/sos")
	assert.Contains(t, routeMap, "GET:/topology/sos/:so_id/l3s")
	assert.Contains(t, routeMap, "GET:/topology/sos/:so_id/l3s/:l3_id/cells")
	assert.Contains(t, routeMap, "GET:/topology/sos/:so_id/l3s/:l3_id/cells/:cell_id/settopboxes")
}

// GET /topology/sos/:so_id/l3s — DB 없이 500 반환 확인 (핸들러 연결 확인)
func TestGetL3s_NoDB_Returns500(t *testing.T) {
	r := newTestRouter(common.Config{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/topology/sos/SO-01/l3s", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// GET /topology/sos/:so_id/l3s/:l3_id/cells — DB 없이 500 반환 확인
func TestGetCells_NoDB_Returns500(t *testing.T) {
	r := newTestRouter(common.Config{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/topology/sos/SO-01/l3s/L3-01/cells", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// GET /topology/sos/:so_id/l3s/:l3_id/cells/:cell_id/settopboxes — DB 없이 500 반환 확인
func TestGetSettopboxes_NoDB_Returns500(t *testing.T) {
	r := newTestRouter(common.Config{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/topology/sos/SO-01/l3s/L3-01/cells/CELL-01/settopboxes", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
