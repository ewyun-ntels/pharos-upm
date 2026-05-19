package statistics

import "github.com/pressly/goose/v3"

// TableName Separate migration namespace for statistics DB
const TableName = "statistics_db_version"

var Migrations = map[string]goose.Migrations{}
