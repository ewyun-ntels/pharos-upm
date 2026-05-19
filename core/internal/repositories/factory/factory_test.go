package factory

import (
	"strings"
	"testing"

	"ntels.com/pharos/core/external/orm"
)

func TestNewUserRepository(t *testing.T) {
	tests := []struct {
		name        string
		opts        RepositoryOptions
		expectedErr bool
		errContains string
	}{
		{
			name: "empty database config",
			opts: RepositoryOptions{
				DatabaseConfig: orm.DatabaseConfig{},
			},
			expectedErr: true,
			errContains: "database configuration is incomplete",
		},
		{
			name: "unsupported driver",
			opts: RepositoryOptions{
				DatabaseConfig: orm.DatabaseConfig{
					Driver: "unsupported",
				},
			},
			expectedErr: true,
			errContains: "unsupported database driver",
		},
		{
			name: "valid sqlite config",
			opts: RepositoryOptions{
				DatabaseConfig: orm.DatabaseConfig{
					Driver: orm.DriverSqlite,
					SQLite: orm.SQLiteConfig{
						Path: "/tmp/test.db",
					},
				},
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewUserRepository(tt.opts)

			if tt.expectedErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error message mismatch: got %q", err.Error())
				}
				if repo != nil {
					t.Errorf("expected nil repository but got %v", repo)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if repo == nil {
					t.Errorf("expected repository but got nil")
				}
			}
		})
	}
}
