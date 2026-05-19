package schedules

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
	r.GET("/schedules/results", getSchedulesResultsListHandler(config))
	r.GET("/schedules/results/:id", getSchedulesResultsHandler(config))
	return r
}

// GET /schedules/results/:id — id 누락 시 404 (라우트 미매칭)
func TestGetSchedulesResults_MissingID_NotFound(t *testing.T) {
	r := newTestRouter(common.Config{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/schedules/results/", nil)
	r.ServeHTTP(w, req)
	// /schedules/results/ 는 /schedules/results/:id 에 매칭되지 않음
	assert.Equal(t, http.StatusMovedPermanently, w.Code)
}

// GET /schedules/results/:id — 잘못된 progress_status 값 → 400
func TestGetSchedulesResults_InvalidProgressStatus(t *testing.T) {
	r := newTestRouter(common.Config{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/schedules/results/some-id?progress_status=invalid", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// GET /schedules/results/:id — 잘못된 result_code → 400
func TestGetSchedulesResults_InvalidResultCode(t *testing.T) {
	r := newTestRouter(common.Config{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/schedules/results/some-id?result_code=99", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// GET /schedules/results/:id — 잘못된 MAC 주소 형식 → 400
func TestGetSchedulesResults_InvalidMacAddr(t *testing.T) {
	r := newTestRouter(common.Config{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/schedules/results/some-id?cm_mac_addr=*invalid", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
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

	assert.Contains(t, routeMap, "GET:/schedules")
	assert.Contains(t, routeMap, "POST:/schedules")
	assert.Contains(t, routeMap, "PUT:/schedules")
	assert.Contains(t, routeMap, "DELETE:/schedules/:name")
	assert.Contains(t, routeMap, "GET:/schedules/results")
	assert.Contains(t, routeMap, "GET:/schedules/results/:id")
}
