package migrations

import (
	"testing"

	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		config  common.Config
		wantErr bool
	}{
		{
			name: "empty config",
			config: common.Config{
				Serve: common.ServeConfig{},
				Catv: common.CatvConfig{
					Migration: struct {
						Database orm.DatabaseConfig "mapstructure:\"database\""
					}{
						Database: orm.DatabaseConfig{
							Driver: orm.DriverClickHouse,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "postgres driver",
			config: common.Config{
				Serve: common.ServeConfig{},
				Catv: common.CatvConfig{
					Migration: struct {
						Database orm.DatabaseConfig "mapstructure:\"database\""
					}{
						Database: orm.DatabaseConfig{
							Driver: orm.DriverPostgreSQL,
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Load(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadWithMasterClickHouse(t *testing.T) {
	config := common.Config{
		Serve: common.ServeConfig{},
		Catv: common.CatvConfig{
			Migration: struct {
				Database orm.DatabaseConfig "mapstructure:\"database\""
			}{
				Database: orm.DatabaseConfig{
					Driver: orm.DriverClickHouse,
					ClickHouse: orm.ClickHouseConfig{
						Host:     "localhost",
						Port:     9000,
						Database: "test",
					},
				},
			},
		},
	}

	err := Load(config)
	if err != nil {
		t.Errorf("Load() with master ClickHouse config failed: %v", err)
	}
}

func BenchmarkLoad(b *testing.B) {
	config := common.Config{
		Serve: common.ServeConfig{},
		Catv: common.CatvConfig{
			Migration: struct {
				Database orm.DatabaseConfig "mapstructure:\"database\""
			}{
				Database: orm.DatabaseConfig{
					Driver: orm.DriverClickHouse,
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Load(config)
	}
}
