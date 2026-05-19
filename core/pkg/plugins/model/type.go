package model

import (
	"context"

	"github.com/centrifugal/centrifuge"
	"ntels.com/pharos/core/external/orm"
)

type PluginType string

func (pluginType PluginType) String() string {
	return string(pluginType)
}

const (
	PluginTypeDatasource PluginType = "datasource"
)

type Data interface {
	QueryData(context.Context, *QueryDataRequest) *QueryDataResponse

	RemoveDatasource(*RemoveDatasourceRequest) error
}

type Query struct {
	ID      string
	SQL     string
	Timeout int
}

type QueryDataResult struct {
	Frame orm.DatabaseResponse
	Error string
}

type QueryDataRequest struct {
	Datasource     Datasource
	DatabaseConfig orm.DatabaseConfig
	Queries        []Query
}

type QueryDataResponse struct {
	Results map[string]QueryDataResult
}

type RemoveDatasourceRequest struct {
	Datasource     Datasource
	DatabaseConfig orm.DatabaseConfig
}

type Stream interface {
	RunStream(*RunStreamRequest) *RunStreamResponse
	SubscribeStream(*SubscribeStreamRequest) *SubscribeStreamResponse
	UnsubscribeStream(*UnsubscribeStreamRequest) error
	PublishStream(*PublishStreamRequest) *PublishStreamResponse

	RemoveDatasource(*RemoveDatasourceRequest) error
}

type RunStreamRequest struct {
	Datasource Datasource

	Path string
	Data string

	Headers map[string]string
}

type RunStreamResponse struct {
	Data  string
	Error string
}

type SubscribeStreamRequest struct {
	Datasource Datasource
	Client     *centrifuge.Client

	Path string
	Data string

	Headers map[string]string
}

type SubscribeStreamResponse struct {
	Data  string
	Error string
}

type UnsubscribeStreamRequest struct {
	Datasource Datasource

	Path string
	Data string

	Headers map[string]string
}

type PublishStreamRequest struct {
	Datasource Datasource

	Path string
	Data string

	Headers map[string]string
}

type PublishStreamResponse struct {
	Data  string
	Error string
}
