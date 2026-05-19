package http_receiver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/internal/utility"
	"ntels.com/pharos/core/pkg/plugins/model"
)

type StreamClient struct {
	Servers *internal.Map[*http.Server]
}

func (streamClient *StreamClient) SubscribeStream(_ *model.SubscribeStreamRequest) *model.SubscribeStreamResponse {
	response := &model.SubscribeStreamResponse{}

	return response
}

func (streamClient *StreamClient) UnsubscribeStream(_ *model.UnsubscribeStreamRequest) error {
	return nil
}

func (streamClient *StreamClient) RunStream(request *model.RunStreamRequest) *model.RunStreamResponse {
	response := &model.RunStreamResponse{}

	if streamClient.Servers.Exist(request.Datasource.Name) {
		return response
	}

	uris := []struct {
		URI    string   `json:"uri"`
		Method []string `json:"method"`
	}{}

	if bytes, err := json.Marshal(request.Datasource.Data["uris"]); err != nil {
		response.Error = err.Error()
		return response
	} else if err := json.Unmarshal(bytes, &uris); err != nil {
		response.Error = err.Error()
		return response
	}

	engine := gin.Default()

	for _, uri := range uris {
		engine.Match(uri.Method, uri.URI, streamClient.handler(strings.Split(request.Headers["websocket-endpoints"], ","), request.Datasource.Name))
	}

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(int(request.Datasource.Data["port"].(float64))),
		Handler: engine,
	}

	streamClient.Servers.Set(request.Datasource.Name, server)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server ListenAndServe error", "error", err)
			err := streamClient.RemoveDatasource(&model.RemoveDatasourceRequest{Datasource: request.Datasource})
			if err != nil {
				slog.Error("RemoveDatasource error", "error", err)
				return
			}
		}
	}()

	return response
}

func (streamClient *StreamClient) PublishStream(_ *model.PublishStreamRequest) *model.PublishStreamResponse {
	response := &model.PublishStreamResponse{}

	return response
}

func (streamClient *StreamClient) RemoveDatasource(request *model.RemoveDatasourceRequest) error {
	streamClient.Servers.Remove(request.Datasource.Name, func(server *http.Server) {
		err := server.Shutdown(context.Background())
		if err != nil {
			slog.Error("server Shutdown error", "error", err)
			return
		}
	})

	return nil
}

func (streamClient *StreamClient) handler(websocketEndpoints []string, datasourceName string) func(c *gin.Context) {
	return func(c *gin.Context) {
		data := map[string]any{}
		data["fullPath"] = c.FullPath()
		data["method"] = c.Request.Method
		if body, err := io.ReadAll(c.Request.Body); err != nil {
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		} else {
			data["body"] = string(body)
		}

		bytes, err := json.Marshal(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		for _, websocketEndpoint := range websocketEndpoints {
			err := utility.CentrifugePublish(websocketEndpoint, "http-receiver:"+datasourceName, string(bytes))
			if err != nil {
				c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, nil)
	}
}
