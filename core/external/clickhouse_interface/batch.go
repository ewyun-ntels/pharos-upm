package clickhouse_interface

import (
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
)

type ClickhouseBatch struct {
	batchData []byte
}

func (c *ClickhouseBatch) Get() (resp []byte) {
	resp = c.batchData
	c.batchData = nil
	return
}

func (c *ClickhouseBatch) Len() int {
	return len(c.batchData)
}

func (c *ClickhouseBatch) AddBatchJson(data any) error {
	marshal, err := json.Marshal(data)
	if err != nil {
		return err
	}
	c.batchData = append(c.batchData, marshal...)
	c.batchData = append(c.batchData, '\n')

	return nil
}

func (c *ClickhouseBatch) Insert(databaseConfig orm.DatabaseConfig, table string) error {
	if c.Len() == 0 {
		return nil
	}

	handler := func(db *sqlx.DB) error {
		query := fmt.Sprintf("INSERT INTO %s FORMAT JSONEachRow\n%s;", table, string(c.Get()))
		_, err := db.Exec(query)
		return err
	}

	return orm.Handler(databaseConfig.Driver, &databaseConfig, handler)
}
