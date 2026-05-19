package workflow

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/external"
	core_goflow "ntels.com/pharos/core/external/goflow"
	"ntels.com/pharos/core/external/goflow/task"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/third_party/goflow"
)

func getTaskHandler(c *gin.Context) {
	name := c.Param("name")

	if task.Exist(name) {
		t := task.Get(name)

		c.JSON(http.StatusOK, map[string]any{"name": t.GetName(), "dataFormat": t.GetDataFormat()})
	} else {
		c.JSON(http.StatusOK, nil)
	}
}

func getsTaskHandler(c *gin.Context) {
	body := []map[string]any{}

	for _, t := range task.Gets() {
		body = append(body, map[string]any{"name": t.GetName(), "dataFormat": t.GetDataFormat()})
	}

	c.JSON(http.StatusOK, body)
}

func getJobHandler(configPath string, config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		job := core_goflow.Job{ConfigPath: configPath, Config: config}

		switch c.Request.Method {
		case http.MethodGet:
			c.JSON(job.GetHandler(c))
		case http.MethodPost:
			c.JSON(job.PostHandler(c))
		case http.MethodPut:
			c.JSON(job.PutHandler(c))
		case http.MethodDelete:
			c.JSON(job.DeleteHandler(c))
		default:
			c.JSON(http.StatusNotImplemented, nil)
		}
	}
}

func getExecutionHandler(config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		executions, err := (&goflow.Execution{}).GetsFromJob(config, c.Param("job"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		c.JSON(http.StatusOK, executions)
	}
}

func getExecutionsHandler(config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		executions, err := (&goflow.Execution{}).Gets(config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		c.JSON(http.StatusOK, executions)
	}
}
