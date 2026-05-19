package schedules

import (
	"fmt"
	"log/slog"
	net_http "net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/authhandler"
	"ntels.com/pharos/core/pkg/common"
	control_common "ntels.com/pharos/extensions/catv/business/pkg/control/common"
	"ntels.com/pharos/extensions/catv/business/pkg/control/common/tables"
	"ntels.com/pharos/extensions/catv/business/pkg/permissions"
	"ntels.com/pharos/shared/types/role"
)

type channelData struct {
	requestTime  orm.Datetime
	ScheduleName string
}

type Api struct {
	configPath string
	config     common.Config

	scheduleEntries sync.Map
	scheduleCron    *cron.Cron
	scheduleTable   *tables.StbControlScheduleTable

	done     chan struct{} // 종료 신호: Immediately/Once 송신 고루틴에 전달
	senderWg sync.WaitGroup
	wg       sync.WaitGroup
	channel  chan channelData
}

func (a *Api) Init(configPath string, config common.Config) {
	a.configPath = configPath
	a.config = config
}

func (a *Api) Use() bool {
	return true
}

func (a *Api) Load() error {
	a.done = make(chan struct{})
	a.channel = make(chan channelData)
	a.wg.Go(a.channelHandler)

	a.scheduleCron = cron.New(cron.WithSeconds(), cron.WithLocation(time.UTC))
	a.scheduleTable = tables.NewStbControlScheduleTable(a.config)

	stbControlScheduleRaws, err := a.scheduleTable.GetAll("update_time ASC")
	if err != nil {
		slog.Error("Failed to load scheduled control requests", "error", err)
		return fmt.Errorf("failed to load scheduled control requests: %w", err)
	}

	for _, raw := range stbControlScheduleRaws {
		if raw.ScheduleType == tables.ScheduleTypeImmediately {
			continue
		}

		if err := raw.Validate(); err != nil {
			slog.Info("Validation failed for scheduled control request, skipping schedule registration", "name", raw.Name, "error", err)
			continue
		}

		if err := a.registerSchedule(raw); err != nil {
			slog.Info("Failed to register schedule for scheduled control request", "name", raw.Name, "error", err)
			continue
		}
	}

	go a.scheduleCron.Start()

	return nil
}

func (a *Api) Unload() {
	slog.Info("Waiting for scheduled control executions to finish...")
	if a.scheduleCron != nil {
		<-a.scheduleCron.Stop().Done()
	}
	a.scheduleEntries.Range(func(key, _ any) bool {
		a.unregisterSchedule(key.(string))
		return true
	})
	close(a.done)
	a.senderWg.Wait()
	close(a.channel)
	a.wg.Wait()
	slog.Info("All scheduled control executions finished")
}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	routes.GET("/schedules", authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), permissions.Read), a.get)
	routes.POST("/schedules", authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), permissions.Create), a.post)
	routes.PUT("/schedules", authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), permissions.Update), a.put)
	routes.DELETE("/schedules/:name", authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), permissions.Delete), a.delete)

	routes.GET("/schedules/results", authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), permissions.Read), getSchedulesResultsListHandler(a.config))
	routes.GET("/schedules/results/:id", authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), permissions.Read), getSchedulesResultsHandler(a.config))
}

func (a *Api) GetRelativePath() string {
	return control_common.HttpRelativePath
}

func (a *Api) get(c *gin.Context) {
	schedules, err := a.scheduleTable.GetAll("create_time DESC")
	if err != nil {
		slog.Error("Failed to get scheduled control requests", "error", err)
		c.JSON(net_http.StatusInternalServerError, external.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(net_http.StatusOK, schedules)
}

func (a *Api) post(c *gin.Context) {
	var request ScheduleRequest

	if err := request.serFromBody(c.Request.Body); err != nil {
		slog.Error("Failed to parse scheduled control request body", "error", err)
		c.JSON(net_http.StatusBadRequest, external.NewErrorResponse(err.Error()))
		return
	}

	if exists, err := a.scheduleTable.ExistsByName(request.Name); err != nil {
		slog.Error("Failed to check if scheduled control request exists", "name", request.Name, "error", err)
		c.JSON(net_http.StatusInternalServerError, external.NewErrorResponse(err.Error()))
		return
	} else if exists {
		slog.Error("Scheduled control request already exists", "name", request.Name)
		c.JSON(net_http.StatusBadRequest, external.NewErrorResponse("Scheduled control request already exists"))
		return
	}

	raw := tables.StbControlScheduleRaw{
		Name: request.Name,

		ScheduleType:       request.ScheduleType,
		ScheduleSpecOnce:   request.ScheduleSpecOnce,
		ScheduleSpecRepeat: request.ScheduleSpecRepeat,

		Sos:         request.Sos,
		L3s:         request.L3s,
		Cells:       request.Cells,
		Settopboxes: request.Settopboxes,

		WorkType:  request.WorkType,
		WorkValue: request.WorkValue,
		AreaType:  request.AreaType,
		AreaIDs:   request.AreaIDs,

		CreateTime: orm.Datetime{Time: time.Now().UTC()},
		IsDeleted:  0,
	}

	if err := raw.Validate(); err != nil {
		slog.Error("Validation failed for scheduled control request", "name", request.Name, "error", err)
		c.JSON(net_http.StatusBadRequest, external.NewErrorResponse(err.Error()))
		return
	}

	if err := a.scheduleTable.Create(raw); err != nil {
		slog.Error("Failed to create scheduled control request", "name", request.Name, "error", err)
		c.JSON(net_http.StatusInternalServerError, external.NewErrorResponse(err.Error()))
		return
	}

	if err := a.registerSchedule(raw); err != nil {
		slog.Error("Failed to register schedule for scheduled control request", "name", request.Name, "error", err)

		if deleteErr := a.scheduleTable.Delete(raw.Name); deleteErr != nil {
			slog.Error("Failed to rollback schedule creation after registration failure", "name", raw.Name, "error", deleteErr)
		}
		c.JSON(net_http.StatusBadRequest, external.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(net_http.StatusOK, nil)
}

func (a *Api) put(c *gin.Context) {
	var request ScheduleRequest

	if err := request.serFromBody(c.Request.Body); err != nil {
		slog.Error("Failed to parse scheduled control request body", "error", err)
		c.JSON(net_http.StatusBadRequest, external.NewErrorResponse(err.Error()))
		return
	}

	stbControlScheduleRaw, err := a.scheduleTable.GetByName(request.Name)
	if err != nil {
		slog.Error("Failed to get existing scheduled control request", "name", request.Name, "error", err)
		c.JSON(net_http.StatusInternalServerError, external.NewErrorResponse(err.Error()))
		return
	}
	stbControlScheduleRaw.ScheduleType = request.ScheduleType
	stbControlScheduleRaw.ScheduleSpecOnce = request.ScheduleSpecOnce
	stbControlScheduleRaw.ScheduleSpecRepeat = request.ScheduleSpecRepeat
	stbControlScheduleRaw.Sos = request.Sos
	stbControlScheduleRaw.L3s = request.L3s
	stbControlScheduleRaw.Cells = request.Cells
	stbControlScheduleRaw.Settopboxes = request.Settopboxes
	stbControlScheduleRaw.WorkType = request.WorkType
	stbControlScheduleRaw.WorkValue = request.WorkValue
	stbControlScheduleRaw.AreaType = request.AreaType
	stbControlScheduleRaw.AreaIDs = request.AreaIDs
	stbControlScheduleRaw.IsDeleted = 0

	if err := stbControlScheduleRaw.Validate(); err != nil {
		slog.Error("Validation failed for scheduled control request", "name", request.Name, "error", err)
		c.JSON(net_http.StatusBadRequest, external.NewErrorResponse(err.Error()))
		return
	}

	if err := a.scheduleTable.Update(stbControlScheduleRaw); err != nil {
		slog.Error("Failed to update scheduled control request", "name", stbControlScheduleRaw.Name, "error", err)
		c.JSON(net_http.StatusInternalServerError, external.NewErrorResponse(err.Error()))
		return
	}

	if err := a.reregisterSchedule(stbControlScheduleRaw); err != nil {
		slog.Error("Failed to register schedule for scheduled control request", "name", stbControlScheduleRaw.Name, "error", err)
		c.JSON(net_http.StatusBadRequest, external.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(net_http.StatusOK, nil)
}

func (a *Api) delete(c *gin.Context) {
	name := c.Param("name")

	a.unregisterSchedule(name)

	if err := a.scheduleTable.Delete(name); err != nil {
		slog.Error("Failed to delete scheduled control request", "name", name, "error", err)
		c.JSON(net_http.StatusInternalServerError, external.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(net_http.StatusOK, nil)
}

func (a *Api) registerSchedule(raw tables.StbControlScheduleRaw) error {
	a.unregisterSchedule(raw.Name)

	var data any
	var err error

	switch raw.ScheduleType {
	case tables.ScheduleTypeImmediately:
		a.senderWg.Go(func() {
			select {
			case a.channel <- channelData{requestTime: orm.Datetime{Time: time.Now().UTC()}, ScheduleName: raw.Name}:
			case <-a.done:
			}
		})
		return nil
	case tables.ScheduleTypeOnce:
		a.senderWg.Add(1)
		data = time.AfterFunc(time.Until(raw.ScheduleSpecOnce.Time.UTC()), func() {
			defer a.senderWg.Done()
			select {
			case a.channel <- channelData{requestTime: orm.Datetime{Time: time.Now().UTC()}, ScheduleName: raw.Name}:
				a.unregisterSchedule(raw.Name)
			case <-a.done:
			}
		})
	case tables.ScheduleTypeRepeat:
		data, err = a.scheduleCron.AddFunc(*raw.ScheduleSpecRepeat, func() {
			a.channel <- channelData{requestTime: orm.Datetime{Time: time.Now().UTC()}, ScheduleName: raw.Name}
		})
		if err != nil {
			return fmt.Errorf("failed to add repeat schedule: %w", err)
		}
	default:
		return fmt.Errorf("invalid schedule_type: %s", raw.ScheduleType)
	}

	a.scheduleEntries.Store(raw.Name, data)

	return nil
}

func (a *Api) reregisterSchedule(raw tables.StbControlScheduleRaw) error {
	return a.registerSchedule(raw)
}

func (a *Api) unregisterSchedule(name string) {
	// LoadAndDelete로 원자적 처리 — 동시 호출(콜백 내부 + Unload Range)에서 중복 실행 방지
	value, ok := a.scheduleEntries.LoadAndDelete(name)
	if ok {
		switch v := value.(type) {
		case *time.Timer:
			if v.Stop() {
				// 타이머가 발화 전에 정지됨 → 콜백이 실행되지 않으므로 senderWg 직접 정산
				a.senderWg.Done()
			}
			// Stop() == false: 이미 발화됨 → 콜백의 defer senderWg.Done()이 처리
		case cron.EntryID:
			a.scheduleCron.Remove(v)
		default:
			slog.Warn("Unexpected schedule data type", "name", name, "type", fmt.Sprintf("%T", value))
		}
	}
}

func (a *Api) channelHandler() {
	for data := range a.channel {
		scheduleRaw, err := a.scheduleTable.GetByName(data.ScheduleName)
		if err != nil {
			slog.Error("Failed to get scheduled control request for execution", "request_time", data.requestTime, "name", data.ScheduleName, "error", err)
			continue
		}

		scheduleRaw.Execute(a.configPath, a.config, data.requestTime)
	}
}
