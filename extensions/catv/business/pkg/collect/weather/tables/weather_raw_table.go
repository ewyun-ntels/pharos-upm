package tables

import (
	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

type WeatherRawTable struct {
	InsertTime orm.Datetime `json:"insert_time" db:"insert_time"`
	File       string       `json:"file" db:"file"`
	Raw        string       `json:"raw" db:"raw"`
	Errors     []string     `json:"errors" db:"errors"`
}

func (w WeatherRawTable) TableName() string {
	return "dist_weather_raw"
}

func (w WeatherRawTable) Insert(config common.Config) error {
	var query = "INSERT INTO " + w.TableName() + " (insert_time, file, raw, errors) VALUES (:insert_time, :file, :raw, :errors)"

	handler := func(db *sqlx.DB) error {
		_, err := db.NamedExec(query, w)
		return err
	}

	return orm.StatisticsHandler(config.Catv.Database.Driver, &config.Catv.Database, handler)
}
