package goflow

import (
	"log/slog"

	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/third_party/goflow"
)

// workflow 에서 데이터 저장에 사용할 store로 차후 구현 필요

// DataStore is a simple key-value store.
type DataStore struct {
	config common.Config
}

// Set If an error is returned, gf.Run() raises an error
func (ds *DataStore) Set(_ string, value any) error {
	execution := goflow.Execution{}
	if ok, err := execution.Set(value); err != nil {
		slog.Error("execution Set error", "error", err)
		return nil
	} else if !ok {
		return nil
	}

	if err := execution.Upsert(ds.config); err != nil {
		slog.Error("execution Upsert error", "error", err)
		return nil
	}

	//switch v := value.(type) {
	switch value.(type) {
	default:
		//slog.Error("Not implemented", "type", v)
	}

	return nil
}

func (ds *DataStore) Get(_ string, _ any) (found bool, err error) {
	return false, nil
}

func (ds *DataStore) Delete(_ string) error {
	return nil
}

func (ds *DataStore) Close() error {
	return nil
}
