package schema

import "github.com/pressly/goose/v3"

func DisableMigrateLog() {
	goose.SetLogger(goose.NopLogger())
}
