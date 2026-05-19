package udp

import (
	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

type Raw struct {
	Timestamp        orm.Datetime `json:"timestamp" db:"timestamp"`
	TransmissionType string       `json:"transmission_type" db:"transmission_type"`
	RawData          string       `json:"raw_data" db:"raw_data"`
	Errors           []string     `json:"errors" db:"errors"`
}

func (r *Raw) insert(config common.Config) error {
	handler := func(db *sqlx.DB) error {
		query := `INSERT INTO dist_stb_transmission_raw
	(timestamp, transmission_type, raw_data, errors)
	VALUES
	(:timestamp, :transmission_type, :raw_data, :errors);`
		_, err := db.NamedExec(query, r)
		return err
	}

	return orm.StatisticsHandler(config.Catv.Database.Driver, &config.Catv.Database, handler)
}
