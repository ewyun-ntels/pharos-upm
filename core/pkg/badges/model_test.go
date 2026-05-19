package badges

import (
	"encoding/json"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/query_builder"
	"ntels.com/pharos/core/pkg/common"
)

// setup sqlite DB for badges model tests
func newModelTestConfig(t *testing.T) common.Config {
	t.Helper()
	cfg := common.Config{}
	cfg.Database.Driver = orm.DriverSqlite
	cfg.Database.SQLite.Path = t.TempDir() + "/badges_model.db"
	// Create badges table explicitly for tests (production uses goose migrations)
	_ = orm.Handler(orm.DriverDefault, &cfg.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS badges (
			name TEXT PRIMARY KEY,
			datasource TEXT NOT NULL,
			query_json TEXT NOT NULL,
			column_name TEXT NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);
		`)
		return err
	})
	return cfg
}

func TestModel_UpsertGet_List_Delete(t *testing.T) {
	cfg := newModelTestConfig(t)
	m := &Model{Config: &cfg}

	// Upsert two badges
	spec1 := query_builder.QuerySpec{Query: "select 1 as v", Variables: map[string]any{"a": 1}}
	spec2 := query_builder.QuerySpec{Query: "select 2 as v", Variables: map[string]any{"b": 2}}

	require.NoError(t, m.Upsert("b1", "ds1", spec1, "v"))
	require.NoError(t, m.Upsert("b2", "ds2", spec2, "v"))

	// Get b1 and decode query json
	var decoded query_builder.QuerySpec
	row, err := m.Get("b1", &decoded)
	require.NoError(t, err)
	require.NotNil(t, row)
	require.Equal(t, "b1", row.Name)
	require.Equal(t, "ds1", row.Datasource)
	require.Equal(t, "v", row.ColumnName)
	require.Equal(t, spec1.Query, decoded.Query)
	// Variables is a map[string]any; ensure a key exists
	require.Equal(t, float64(1), decoded.Variables["a"]) // json numbers become float64

	// Get not-found
	r2, err := m.Get("nope", &decoded)
	require.NoError(t, err)
	require.Nil(t, r2)

	// List should return names sorted
	names, err := m.List()
	require.NoError(t, err)
	require.Equal(t, []string{"b1", "b2"}, names)

	// Delete b1
	require.NoError(t, m.Delete("b1"))
	names, err = m.List()
	require.NoError(t, err)
	require.Equal(t, []string{"b2"}, names)

	// Delete with empty name -> error
	require.Error(t, m.Delete(""))
}

func TestModel_Upsert_UpdatesExisting(t *testing.T) {
	cfg := newModelTestConfig(t)
	m := &Model{Config: &cfg}

	spec := query_builder.QuerySpec{Query: "sql1"}
	require.NoError(t, m.Upsert("badge", "ds1", spec, "v"))

	// update datasource and query
	spec2 := query_builder.QuerySpec{Query: "sql2"}
	require.NoError(t, m.Upsert("badge", "ds2", spec2, "v2"))

	// verify persisted
	// We can decode query json to ensure it is updated
	var dec query_builder.QuerySpec
	row, err := m.Get("badge", &dec)
	require.NoError(t, err)
	require.NotNil(t, row)
	require.Equal(t, "ds2", row.Datasource)
	require.Equal(t, "v2", row.ColumnName)
	require.Equal(t, "sql2", dec.Query)
}

func TestModel_Get_DstNil(t *testing.T) {
	cfg := newModelTestConfig(t)
	m := &Model{Config: &cfg}

	spec := query_builder.QuerySpec{Query: "sql"}
	require.NoError(t, m.Upsert("badge", "ds", spec, "v"))

	// Call Get with nil dst to ensure it doesn't panic and returns row
	row, err := m.Get("badge", nil)
	require.NoError(t, err)
	require.NotNil(t, row)

	// And check the raw persisted JSON is valid
	var js query_builder.QuerySpec
	require.NoError(t, json.Unmarshal([]byte(row.QueryJSON), &js))
	require.Equal(t, "sql", js.Query)
}
