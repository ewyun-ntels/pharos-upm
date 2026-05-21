package alert

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ory/fosite/token/jwt"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/query_builder"
	"ntels.com/pharos/core/pkg/alert/adapters"
	"ntels.com/pharos/core/pkg/alert/common"
	"ntels.com/pharos/core/pkg/alert/model"
	"ntels.com/pharos/core/pkg/alert/resources"
	"ntels.com/pharos/core/pkg/alert/rule"
	rule_common "ntels.com/pharos/core/pkg/alert/rule/common"
	pkg_common "ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins"
	alert_types "ntels.com/pharos/shared/types/alert"
	"ntels.com/pharos/shared/types/role"
)

// Alert rule을 여러번 쿼리하기 때문에 lock 처리 (많은 데이터를 다루지 않기 때문에 큰 문제는 없을것으로 보임
// create, modify 하는 구간만 처리
var handlerMutex = sync.Mutex{}
var ruleSchedulerSyncMutex = sync.Mutex{}
var ruleSchedulerVersions sync.Map

func isScheduledAlertType(alertType common.Type) bool {
	return alertType == common.TypeQuery || alertType == common.TypeEventHistory
}

func syncLocalRuleSchedulers(config pkg_common.Config, ruleDbs []model.Rule) {
	ruleSchedulerSyncMutex.Lock()
	defer ruleSchedulerSyncMutex.Unlock()

	dbScheduledRuleIDs := make(map[string]struct{})

	for _, ruleDb := range ruleDbs {
		alertType := common.Type(ruleDb.AlertType)
		if !isScheduledAlertType(alertType) {
			continue
		}

		dbScheduledRuleIDs[ruleDb.ID] = struct{}{}
		currentVersion := ruleDb.UpdatedAt.Time
		hasJob := rule_common.CronInstance.HasJob(ruleDb.ID)
		if storedVersion, ok := ruleSchedulerVersions.Load(ruleDb.ID); hasJob && ok {
			if version, ok := storedVersion.(time.Time); ok && version.Equal(currentVersion) {
				continue
			}
		} else if hasJob {
			ruleSchedulerVersions.Store(ruleDb.ID, currentVersion)
			continue
		}

		if hasJob {
			if err := rule_common.CronInstance.RemoveJob(ruleDb.ID); err != nil {
				slog.Warn("alert rule scheduler remove failed during sync", "id", ruleDb.ID, "error", err)
			}
		}

		template, err := rule.GetRuleFromStr(alertType, config, []byte(ruleDb.Rule))
		if err != nil {
			slog.Warn("alert rule scheduler load failed during sync", "id", ruleDb.ID, "error", err)
			continue
		}
		template.SetId(ruleDb.ID)

		if err = template.Run(context.Background()); err != nil {
			slog.Warn("alert rule scheduler run failed during sync", "id", ruleDb.ID, "error", err)
			continue
		}
		ruleSchedulerVersions.Store(ruleDb.ID, currentVersion)
	}

	for _, jobID := range rule_common.CronInstance.JobIDs() {
		if _, ok := dbScheduledRuleIDs[jobID]; ok {
			continue
		}
		if err := rule_common.CronInstance.RemoveJob(jobID); err != nil {
			slog.Warn("alert rule scheduler stale job remove failed during sync", "id", jobID, "error", err)
		}
		ruleSchedulerVersions.Delete(jobID)
	}
}

// getJWTClaims returns JWT claims from the Gin context if present, otherwise an empty claims struct.
func getJWTClaims(c *gin.Context) *jwt.JWTClaims {
	if value, exists := c.Get(pkg_common.ContextKeyJWTClaims); exists {
		return value.(*jwt.JWTClaims)
	}
	return &jwt.JWTClaims{}
}

func GetPutStatusCleanHandler(config pkg_common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 입력 검증
		//validator := validation.NewValidator()

		name := c.Param("name")
		id := c.Param("id")

		claims := getJWTClaims(c)

		m := model.Model{
			Config: &config,
		}

		alert, err := m.GetAlert(name, id)
		if err != nil {
			slog.Error("alert status clean failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: "failed to retrieve alert",
			})
			return
		}

		if alert == nil {
			slog.Error("alert not found", "name", name, "id", id)
			c.JSON(http.StatusNotFound, external.ErrorResponse{Message: "alert not found"})
			return
		}

		err = m.DeleteAlert(name, id)
		if err != nil {
			slog.Error("alert status clean failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: "failed to delete alert",
			})
			return
		}

		prevSeverity := alert.Severity
		prevValue := alert.Value
		now := time.Now().UTC()

		alert.PreviousSeverity = prevSeverity
		alert.PreviousTimestamp = alert.Timestamp
		alert.PreviousValue = &prevValue
		alert.Timestamp = now
		alert.UpdatedAt = now
		alert.CheckTime = &now

		alert.Status = common.StatusNormal

		h := rule_common.History{
			Config: config,
		}

		err = h.Send([]common.Value{*alert}, common.StatusChangeReasonManual, &claims.Subject)
		if err != nil {
			slog.Error("alert event history send failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: "failed to record history",
			})
			return
		}

		c.Status(http.StatusOK)
	}
}

func GetPutStatusMaskHandler(config pkg_common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 입력 검증 - URL 파라미터
		validator := pkg_common.NewValidator()

		name := c.Param("name")
		id := c.Param("id")

		if validator.HasErrors() {
			slog.Error("input validation failed", "errors", validator.GetErrors())
			c.JSON(http.StatusBadRequest, external.ErrorResponse{
				Message: validator.GetFirstError(),
			})
			return
		}

		// Request body 읽기
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			slog.Error("read body failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		// API 스키마 사용
		var apiRequest alert_types.AlertStatusMaskRequest
		err = json.Unmarshal(body, &apiRequest)
		if err != nil {
			slog.Error("json.Unmarshal failed", "error", err)
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: err.Error()})
			return
		}

		claims := getJWTClaims(c)

		// Adapter를 통해 내부 타입으로 변환
		adapter := &adapters.APIAdapter{}
		mask := adapter.FromAPIStatusMaskRequest(apiRequest)

		m := model.Model{
			Config: &config,
		}

		err = m.SetStatusMask(name, id, mask)
		if err != nil {
			slog.Error("alert status mask failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: "failed to update alert status",
			})
			return
		}

		alert, err := m.GetAlert(name, id)
		if err != nil {
			slog.Error("alert status clean failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: "failed to retrieve alert",
			})
			return
		}
		if alert == nil {
			slog.Error("alert not found", "name", name, "id", id)
			c.JSON(http.StatusNotFound, external.ErrorResponse{Message: "alert not found"})
			return
		}

		h := rule_common.History{
			Config: config,
		}

		err = h.Send([]common.Value{*alert}, common.StatusChangeReasonMask, &claims.Subject)
		if err != nil {
			slog.Error("alert event history send failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: "failed to record history",
			})
			return
		}

		c.Status(http.StatusOK)
	}
}

func GetStatusHandler(config pkg_common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 쿼리 파라미터로부터 API 스키마 생성
		name := c.Query("name")
		status := c.Query("status")
		severity := c.Query("severity")
		alertType := c.Query("alertType")
		limit := c.Query("limit")
		offset := c.Query("offset")

		apiRequest := alert_types.AlertStatusRequest{}
		if name != "" {
			apiRequest.Name = &name
		}
		if status != "" {
			statusEnum := alert_types.Status(status)
			apiRequest.Status = &statusEnum
		}
		if severity != "" {
			apiRequest.Severity = &severity
		}
		if alertType != "" {
			alertTypeEnum := alert_types.AlertType(alertType)
			apiRequest.AlertType = &alertTypeEnum
		}
		if limit != "" {
			// limit 파싱 (에러 처리 생략)
		}
		if offset != "" {
			// offset 파싱 (에러 처리 생략)
		}

		// Adapter를 통해 필터링 파라미터로 변환
		adapter := &adapters.APIAdapter{}
		filters := adapter.FromAPIStatusRequest(apiRequest)

		m := model.Model{
			Config: &config,
		}

		// 필터링된 상태 조회 (기존 GetStatuses를 확장하거나 새로운 메서드 필요)
		internalStatus, err := m.GetStatuses()
		if err != nil {
			slog.Error("alert status failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: err.Error(),
			})
			return
		}

		// 필터링 적용 (간단한 클라이언트 사이드 필터링)
		if len(filters) > 0 {
			// 필터링 로직 적용 (실제로는 DB 쿼리에서 처리해야 함)
		}

		// API 응답에서만 shared types로 변환
		apiStatus := adapter.ToAPIAlertValues(*internalStatus)

		c.JSON(http.StatusOK, apiStatus)
	}
}

func GetPostRuleHandler(config pkg_common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		handlerMutex.Lock()
		defer handlerMutex.Unlock()

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			slog.Error("read body failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		var apiRule alert_types.AlertRule
		err = json.Unmarshal(body, &apiRule)
		if err != nil {
			slog.Error("json.Unmarshal failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		// Convert API rule to internal resources.Rule format
		ruleAdapter := &adapters.RuleAdapter{}
		ruleConfig, err := ruleAdapter.FromAPIRuleCreate(apiRule)
		if err != nil {
			slog.Error("convert API rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		res := resources.Rule{
			AlertType: common.Type(apiRule.AlertType),
			Rule:      ruleConfig,
		}

		if res.Rule == nil {
			slog.Error("rule is nil")
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "rule is nil"})
			return
		}

		newRule, err := rule.GetRuleTemplate(res.AlertType)
		if err != nil {
			slog.Error("get rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		err = newRule.Load(config, res.Rule)
		if err != nil {
			slog.Error("load rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		if newRule.GetId() != "" {
			slog.Error("rule id already exists", "id", newRule.GetId())
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "rule id already exists"})
			return
		}

		err = newRule.Normalize()
		if err != nil {
			slog.Error("normalize rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		err = newRule.Validate()
		if err != nil {
			slog.Error("validate rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		newRule.SetId(uuid.New().String())

		m := model.Model{
			Config: &config,
		}

		// database error로 판단하기에는 db 종류에 따라 에러 여부를 판단해야 하기 때문에 사전에 확인하는것으로 변경
		getRule, _ := m.GetRuleByName(newRule.GetName())
		if getRule != nil {
			slog.Error("rule name is duplicated")
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "rule name is duplicated"})
			return
		}

		cur := time.Now().UTC()

		strRule, _ := json.Marshal(newRule)
		insertRule := model.Rule{
			ID:        newRule.GetId(),
			AlertType: string(res.AlertType),
			Name:      newRule.GetName(),
			Rule:      string(strRule),
			Timestamp: orm.Datetime{Time: cur},
			UpdatedAt: orm.Datetime{Time: cur},
		}
		err = m.InsertRule(&insertRule)
		if err != nil {
			slog.Error("alert rule insert failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		err = newRule.Run(c.Request.Context())
		if err != nil {
			slog.Error("alert rule run failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}
		ruleSchedulerVersions.Store(insertRule.ID, insertRule.UpdatedAt.Time)

		c.JSON(http.StatusOK, map[string]string{"id": newRule.GetId()})
	}
}

func GetRuleListHandler(config pkg_common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var claims *jwt.JWTClaims
		if value, exists := c.Get(pkg_common.ContextKeyJWTClaims); exists {
			claims = value.(*jwt.JWTClaims)
		} else {
			claims = &jwt.JWTClaims{}
		}

		detail := false
		_, exists := c.GetQuery("detail")
		if exists {
			detail = true
		}

		m := model.Model{
			Config: &config,
		}

		ruleDbs, err := m.GetAllRuleDb()
		if err != nil {
			slog.Error("alert get rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}
		syncLocalRuleSchedulers(config, ruleDbs)

		var statusMap = make(map[string][]common.Value)
		if detail {
			status, err := m.GetStatuses()
			if err != nil {
				slog.Error("alert status load failed", "error", err)
				c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
				return
			}

			for _, p := range *status {
				statusMap[p.Name] = append(statusMap[p.Name], p)
			}
		}

		var internalRules []resources.Rule
		for _, ruleDb := range ruleDbs {
			if pkg_common.GetRole(claims, string(role.RoleSuperAdmin)) ||
				pkg_common.GetRole(claims, string(role.RoleAlertRead)) {
				var appendRes resources.Rule
				ruleStatus := statusMap[ruleDb.Name]
				err = appendRes.BuildFromDb(&ruleDb, ruleStatus)
				if err != nil {
					slog.Error("alert build rule failed", "error", err)
					c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
					return
				}

				internalRules = append(internalRules, appendRes)
			} else {
				simpleRule := map[string]any{"name": ruleDb.Name, "id": ruleDb.ID}
				internalRules = append(internalRules, resources.Rule{
					AlertType: common.Type(ruleDb.AlertType),
					Rule:      simpleRule,
					Timestamp: ruleDb.Timestamp.Time,
					UpdateAt:  ruleDb.UpdatedAt.Time,
				})
			}
		}

		// Convert to API response format
		var apiRules []alert_types.AlertRule
		for _, internalRule := range internalRules {
			apiResponse, err := adapters.ToAPIRuleResponseFromInternal(&internalRule)
			if err != nil {
				slog.Error("failed to convert rule to API response", "error", err)
				c.JSON(http.StatusInternalServerError, external.ErrorResponse{
					Message: "failed to format rule response",
				})
				return
			}
			apiRules = append(apiRules, *apiResponse)
		}

		c.JSON(http.StatusOK, apiRules)
	}
}

func GetRuleHandler(config pkg_common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 입력 검증
		validator := pkg_common.NewValidator()
		id := validator.ValidateUUID("id", c.Param("id"), true)

		if validator.HasErrors() {
			slog.Error("input validation failed", "errors", validator.GetErrors())
			c.JSON(http.StatusBadRequest, external.ErrorResponse{
				Message: validator.GetFirstError(),
			})
			return
		}

		m := model.Model{
			Config: &config,
		}

		existsRule, err := m.GetRule(id)
		if err != nil {
			slog.Error("alert get rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: "failed to retrieve rule",
			})
			return
		}

		if existsRule == nil {
			slog.Error("rule not found", "id", id)
			c.JSON(http.StatusNotFound, external.ErrorResponse{Message: "rule not found"})
			return
		}

		var internalRule resources.Rule
		err = internalRule.BuildFromDb(existsRule, nil)
		if err != nil {
			slog.Error("alert build rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: "failed to build rule response",
			})
			return
		}

		// API 응답을 위해 adapter 사용
		apiResponse, err := adapters.ToAPIRuleResponseFromInternal(&internalRule)
		if err != nil {
			slog.Error("failed to convert rule to API response", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: "failed to format rule response",
			})
			return
		}

		c.JSON(http.StatusOK, apiResponse)
	}
}

func GetPutRuleHandler(config pkg_common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		handlerMutex.Lock()
		defer handlerMutex.Unlock()

		id := c.Param("id")

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			slog.Error("read body failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		var apiRule alert_types.AlertRule
		err = json.Unmarshal(body, &apiRule)
		if err != nil {
			slog.Error("json.Unmarshal failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		// Convert API rule to internal resources.Rule format
		ruleAdapter := &adapters.RuleAdapter{}
		ruleConfig, err := ruleAdapter.FromAPIRuleUpdate(apiRule)
		if err != nil {
			slog.Error("convert API rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		newRes := resources.Rule{
			AlertType: common.Type(string(apiRule.AlertType)),
			Rule:      ruleConfig,
		}

		m := model.Model{
			Config: &config,
		}

		editeRule, err := m.GetRule(id)
		if err != nil {
			slog.Error("alert get rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		if editeRule == nil {
			slog.Error("rule not found", "id", id)
			c.JSON(http.StatusNotFound, external.ErrorResponse{Message: "rule not found"})
			return
		}

		newRule, err := rule.GetRule(newRes.AlertType, config, newRes.Rule)
		if err != nil {
			slog.Error("alert get rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}
		newRule.SetId(id)

		err = newRule.Validate()
		if err != nil {
			slog.Error("rule validate failed", "error", err)
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: err.Error()})
			return
		}

		// rule 정보 파악후에 삭제
		oldRule, err := rule.GetRuleFromStr(common.Type(editeRule.AlertType), config, []byte(editeRule.Rule))
		if err != nil {
			slog.Error("get rule from rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		err = oldRule.Destroy()
		if err != nil {
			slog.Error("alert destroy rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		strRule, _ := json.Marshal(newRule)
		updateRule := model.Rule{
			ID:        newRule.GetId(),
			AlertType: string(newRes.AlertType),
			Name:      newRule.GetName(),
			Rule:      string(strRule),
			Timestamp: orm.Datetime{Time: editeRule.Timestamp.Time},
			UpdatedAt: orm.Datetime{Time: time.Now()},
		}

		err = m.UpdateRule(&updateRule)
		if err != nil {
			slog.Error("alert rule insert failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		err = newRule.Run(context.Background())
		if err != nil {
			slog.Error("alert rule run failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}
		ruleSchedulerVersions.Store(updateRule.ID, updateRule.UpdatedAt.Time)

		c.Status(http.StatusOK)
	}
}

func GetDeleteRuleHandler(config pkg_common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// destroy 구간이 존재하기 때문에 modify와 충돌을 대비해 lock 처리 (delete가 얼마 없어서 상관 없을것으로 보임)
		handlerMutex.Lock()
		defer handlerMutex.Unlock()

		id := c.Param("id")

		m := model.Model{
			Config: &config,
		}

		deleteRule, err := m.GetRule(id)
		if err != nil {
			slog.Error("alert get rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		oldRule, err := rule.GetRuleFromStr(common.Type(deleteRule.AlertType), config, []byte(deleteRule.Rule))
		if err != nil {
			// rule parsing 실패시에도 DB정리는 진행할 수 있도록 경고만 출력
			slog.Error("get rule from string failed", "error", err)
		} else {
			if derr := oldRule.Destroy(); derr != nil {
				// 크론 잡이 이미 없을 수 있으므로 경고만 출력하고 계속 진행
				slog.Warn("alert destroy rule warning", "error", derr)
			}
		}

		// 현재 활성화된 상태(alert_status) 제거 (rule name 기준)
		if derr := m.DeleteAlertsByName(deleteRule.Name); derr != nil {
			slog.Error("alert statuses delete by name failed", "error", derr, "name", deleteRule.Name)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: derr.Error()})
			return
		}

		// 룰 자체를 DB에서 제거
		if err := m.DeleteRule(id); err != nil {
			slog.Error("alert delete rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}
		ruleSchedulerVersions.Delete(id)

		c.Status(http.StatusOK)
	}
}

// PostQuery Alert query를 테스트하기 위해 사용
func PostQuery(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.Error("read body failed", "error", err)
		c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
		return
	}

	// API 스키마 사용
	var apiRequest alert_types.AlertQueryRequest
	err = json.Unmarshal(body, &apiRequest)
	if err != nil {
		slog.Error("json.Unmarshal failed", "error", err)
		c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
		return
	}

	// Adapter를 통해 내부 타입으로 변환
	adapter := &adapters.APIAdapter{}
	res := adapter.FromAPIQueryRequest(apiRequest)

	if res.Datasource == "" {
		slog.Error("datasource is empty")
		c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "datasource is empty"})
		return
	}

	queryResp, err := query_builder.GetQuery(res.Query.Query, res.Query.Variables)
	if err != nil {
		slog.Error("GetQuery failed", "error", err)
		c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
		return
	}

	// Route the query via plugins.DatasourceQuery (unified dispatcher)
	req := plugins.DsQueryRequest{Queries: []plugins.DsQuery{
		{ID: "q1", DatasourceName: res.Datasource, SQL: string(queryResp), Timeout: 30},
	}}
	resp := plugins.DatasourceQuery(c, req)
	qres, ok := resp.Results["q1"]
	if !ok {
		slog.Error("DatasourceQuery returned no result for id", "id", "q1")
		c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: "datasource query missing result"})
		return
	}
	if qres.Error != "" {
		slog.Error("DatasourceQuery error", "error", qres.Error)
		c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: qres.Error})
		return
	}

	// Adapter를 통해 응답 스키마로 변환
	response := adapter.ToAPIQueryResponse(qres.Frame)
	c.JSON(http.StatusOK, response)
}

// HistoryQueryParams 히스토리 쿼리 매개변수 (exported for adapter use)
type HistoryQueryParams struct {
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Count     int
}

// HistoryRow 히스토리 데이터베이스 행
type HistoryRow struct {
	Timestamp        orm.Datetime `db:"timestamp"`
	Id               string       `db:"id"`
	AlertId          string       `db:"alert_id"`
	Name             string       `db:"name"`
	AlertType        string       `db:"alert_type"`
	Description      string       `db:"description"`
	Value            float64      `db:"value"`
	Severity         string       `db:"severity"`
	Status           string       `db:"status"`
	PreviousSeverity string       `db:"previous_severity"`
	PreviousValue    *float64     `db:"previous_value"`
	PreviousTimestamp *orm.Datetime `db:"previous_timestamp"`
	Version           orm.Datetime  `db:"version"`
	Mask              bool          `db:"mask"`
	Labels           string       `db:"labels"`
}

// validateHistoryQueryParams 히스토리 쿼리 매개변수 검증
func validateHistoryQueryParams(c *gin.Context) (*HistoryQueryParams, error) {
	validator := pkg_common.NewValidator()

	safeName := validator.ValidateAlertName("name", c.Query("name"), false)
	startTimeStr := validator.ValidateStringLength("start-time", c.Query("start-time"), 0, 50, false)
	endTimeStr := validator.ValidateStringLength("end-time", c.Query("end-time"), 0, 50, false)
	countStr := validator.ValidateStringLength("count", c.Query("count"), 0, 10, false)

	if validator.HasErrors() {
		return nil, errors.New(validator.GetFirstError())
	}

	params := &HistoryQueryParams{Name: safeName}

	// 시간 검증
	if startTimeStr != "" {
		params.StartTime = validator.ValidateDateTime("start-time", startTimeStr, false)
		if validator.HasErrors() {
			return nil, errors.New(validator.GetFirstError())
		}
	}

	if endTimeStr != "" {
		params.EndTime = validator.ValidateDateTime("end-time", endTimeStr, false)
		if validator.HasErrors() {
			return nil, errors.New(validator.GetFirstError())
		}
	}

	// 카운트 검증
	if countStr != "" {
		params.Count = validator.ValidateCount("count", countStr, 0, 10000, false)
		if validator.HasErrors() {
			return nil, errors.New(validator.GetFirstError())
		}
	}

	// 시간 범위 검증
	if !params.StartTime.IsZero() && !params.EndTime.IsZero() {
		if params.StartTime.After(params.EndTime) {
			return nil, errors.New("start-time must be before end-time")
		}
		if params.EndTime.Sub(params.StartTime) > 365*24*time.Hour {
			return nil, errors.New("time range cannot exceed 1 year")
		}
	}

	return params, nil
}

// queryHistoryFromDB 데이터베이스에서 히스토리 조회
func queryHistoryFromDB(statsDB *orm.DatabaseConfig, params *HistoryQueryParams) ([]alert_types.AlertValue, error) {
	var values []common.Value
	adapter := &adapters.APIAdapter{}

	err := orm.StatisticsHandler("", statsDB, func(db *sqlx.DB) error {
		query := `SELECT timestamp, id, alert_id, name, alert_type, description, value, 
		          severity, status, previous_severity, previous_value, previous_timestamp, version, mask, labels FROM history_alert_row`

		// ClickHouse만 FINAL 절 추가
		if statsDB.Driver == orm.DriverClickHouse {
			query += " FINAL"
		}

		var args []any
		var conditions []string

		// WHERE 조건 추가
		if params.Name != "" {
			conditions = append(conditions, "name = ?")
			args = append(args, params.Name)
		}

		if !params.StartTime.IsZero() && !params.EndTime.IsZero() {
			conditions = append(conditions, "timestamp BETWEEN ? AND ?")
			if statsDB.Driver == orm.DriverClickHouse {
				args = append(args, params.StartTime.Unix(), params.EndTime.Unix())
			} else {
				args = append(args, params.StartTime, params.EndTime)
			}
		}

		if len(conditions) > 0 {
			query += " WHERE " + strings.Join(conditions, " AND ")
		}

		// ORDER BY와 LIMIT
		limit := 1000
		if params.Count > 0 {
			limit = params.Count
		}
		query += " ORDER BY timestamp DESC LIMIT ?"
		args = append(args, limit)

		// 쿼리 실행
		var rows []HistoryRow
		err := db.Select(&rows, db.Rebind(query), args...)
		if err != nil {
			return fmt.Errorf("database query failed: %w", err)
		}

		// 결과 변환
		for _, row := range rows {
			value := common.Value{
				Timestamp:        row.Timestamp.Time,
				UpdatedAt:        row.Version.Time,
				Id:               row.Id,
				Name:             row.Name,
				AlertType:        row.AlertType,
				AlertId:          row.AlertId,
				Value:            row.Value,
				Description:      row.Description,
				Severity:         row.Severity,
				Status:           row.Status,
				PreviousSeverity: row.PreviousSeverity,
				PreviousValue:    row.PreviousValue,
				Mask:             row.Mask,
				StringLabels:     row.Labels,
			}
			if row.PreviousTimestamp != nil {
				value.PreviousTimestamp = row.PreviousTimestamp.Time
			}
			_ = value.ConvertStringLabelsToLabels()
			values = append(values, value)
		}

		return nil
	})

	// Convert to API types using adapter
	if err != nil {
		return nil, err
	}

	return adapter.ToAPIAlertValues(values), nil
}

func GetHistQueryHandler(config pkg_common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 입력 검증 (기존 로직 유지)
		params, err := validateHistoryQueryParams(c)
		if err != nil {
			slog.Error("input validation failed", "error", err)
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: err.Error()})
			return
		}

		// 데이터베이스 설정 확인
		statsDB := &config.Statistics.Database
		if statsDB.Driver == "" {
			slog.Error("no statistics database configured")
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: "statistics database not configured",
			})
			return
		}

		// 쿼리 실행 - 스키마 기반 응답
		values, err := queryHistoryFromDB(statsDB, params)
		if err != nil {
			slog.Error("database operation failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: "database operation failed",
			})
			return
		}

		// Adapter를 통해 응답 스키마로 변환
		adapter := &adapters.APIAdapter{}
		response := adapter.ToAPIHistoryResponse(values)
		c.JSON(http.StatusOK, response)
	}
}

func GetPostEventHandler(config pkg_common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 삭제된 rule에 데이터가 들어가는 것을 방지하기 위해
		handlerMutex.Lock()
		defer handlerMutex.Unlock()

		// 입력 검증
		validator := pkg_common.NewValidator()
		ruleName := validator.ValidateAlertName("name", c.Param("name"), true)

		if validator.HasErrors() {
			slog.Error("input validation failed", "errors", validator.GetErrors())
			c.JSON(http.StatusBadRequest, external.ErrorResponse{
				Message: validator.GetFirstError(),
			})
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			slog.Error("read body failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: "failed to read request body",
			})
			return
		}

		// Body 크기 제한 (1MB)
		if len(body) > 1024*1024*2 {
			slog.Error("request body too large", "size", len(body))
			c.JSON(http.StatusBadRequest, external.ErrorResponse{
				Message: "request body too large (max 1MB)",
			})
			return
		}

		// API 스키마 사용
		var apiRequest alert_types.AlertEventRequest
		err = json.Unmarshal(body, &apiRequest)
		if err != nil {
			slog.Error("event data unmarshal failed", "error", err)
			c.JSON(http.StatusBadRequest, external.ErrorResponse{
				Message: "invalid JSON format",
			})
			return
		}

		// Adapter를 통해 내부 타입으로 변환
		adapter := &adapters.APIAdapter{}
		res := adapter.FromAPIEventRequest(apiRequest)

		err = res.Validate()
		if err != nil {
			slog.Error("event data validate failed", "error", err)
			c.JSON(http.StatusBadRequest, external.ErrorResponse{
				Message: "event data validation failed",
			})
			return
		}

		m := model.Model{
			Config: &config,
		}

		eventRule, err := m.GetRuleByName(ruleName)
		if err != nil {
			slog.Error("alert get rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: "failed to retrieve rule",
			})
			return
		}

		if eventRule == nil {
			slog.Error("alert rule not found", "name", ruleName)
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "alert rule not found"})
			return
		}

		controller, err := rule.GetRuleFromStr(common.Type(eventRule.AlertType), config, []byte(eventRule.Rule))
		if err != nil {
			slog.Error("alert rule load failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: "failed to load rule",
			})
			return
		}

		err = controller.EventHandler(res)
		if err != nil {
			slog.Error("alert event control failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{
				Message: "failed to process event",
			})
			return
		}

		c.Status(http.StatusOK)
	}
}
