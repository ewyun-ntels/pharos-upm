package config_loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
)

// setupTestDir는 테스트용 임시 디렉터리를 생성합니다
func setupTestDir(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "config-loader-test")
	require.NoError(t, err)

	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	return tmpDir
}

// createTestConfigFile은 테스트용 설정 파일을 생성합니다
func createTestConfigFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
	return filePath
}

func TestCheckFileReadable(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr bool
	}{
		{
			name: "파일이 존재하고 읽을 수 있는 경우",
			setup: func(t *testing.T) string {
				tmpDir := setupTestDir(t)
				return createTestConfigFile(t, tmpDir, "config.yaml", "key: value")
			},
			wantErr: false,
		},
		{
			name: "파일이 존재하지 않는 경우",
			setup: func(t *testing.T) string {
				return "/nonexistent/path/config.yaml"
			},
			wantErr: true,
		},
		{
			name: "파일 권한이 없는 경우",
			setup: func(t *testing.T) string {
				tmpDir := setupTestDir(t)
				filePath := createTestConfigFile(t, tmpDir, "config.yaml", "key: value")
				// 파일 권한을 제거 (읽기 권한 없음)
				err := os.Chmod(filePath, 0000)
				require.NoError(t, err)
				return filePath
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup(t)
			err := checkFileReadable(filePath)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) []string
		wantErr     bool
		errContains string
	}{
		{
			name: "단일 설정 파일 로드 성공",
			setup: func(t *testing.T) []string {
				tmpDir := setupTestDir(t)
				configContent := `
serve:
  server_schema: "http"
database:
  driver: "postgresql"
  postgresql:
    host: "localhost"
    port: 5432
`
				configFile := createTestConfigFile(t, tmpDir, "config.yaml", configContent)
				return []string{configFile}
			},
			wantErr: false,
		},
		{
			name: "여러 설정 파일 로드 성공",
			setup: func(t *testing.T) []string {
				tmpDir := setupTestDir(t)

				// 기본 설정 파일
				baseConfig := `
serve:
  server_schema: "http"
database:
  driver: "postgresql"
  postgresql:
    host: "localhost"
    port: 5432
`
				// 추가 설정 파일 (오버라이드용)
				overrideConfig := `
serve:
  type: "https"
auth:
  global_secret: "test-secret"
`
				baseFile := createTestConfigFile(t, tmpDir, "base.yaml", baseConfig)
				overrideFile := createTestConfigFile(t, tmpDir, "override.yaml", overrideConfig)

				return []string{baseFile, overrideFile}
			},
			wantErr: false,
		},
		{
			name: "빈 설정 파일 배열",
			setup: func(t *testing.T) []string {
				return []string{}
			},
			wantErr:     true,
			errContains: "config file is empty",
		},
		{
			name: "첫 번째 파일이 존재하지 않는 경우",
			setup: func(t *testing.T) []string {
				return []string{"/nonexistent/config.yaml"}
			},
			wantErr: true,
		},
		{
			name: "두 번째 파일이 존재하지 않는 경우",
			setup: func(t *testing.T) []string {
				tmpDir := setupTestDir(t)
				configContent := `
serve:
  type: "http"
`
				configFile := createTestConfigFile(t, tmpDir, "config.yaml", configContent)
				return []string{configFile, "/nonexistent/second.yaml"}
			},
			wantErr: true,
		},
		{
			name: "잘못된 YAML 형식",
			setup: func(t *testing.T) []string {
				tmpDir := setupTestDir(t)
				invalidYaml := `
serve:
  type: "http"
  invalid yaml content: [
`
				configFile := createTestConfigFile(t, tmpDir, "invalid.yaml", invalidYaml)
				return []string{configFile}
			},
			wantErr: true,
		},
		{
			name: "두 번째 파일이 잘못된 형식인 경우",
			setup: func(t *testing.T) []string {
				tmpDir := setupTestDir(t)

				validConfig := `
serve:
  type: "http"
`
				invalidConfig := `
invalid yaml: [
`
				validFile := createTestConfigFile(t, tmpDir, "valid.yaml", validConfig)
				invalidFile := createTestConfigFile(t, tmpDir, "invalid.yaml", invalidConfig)

				return []string{validFile, invalidFile}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFiles := tt.setup(t)

			config, err := Load(configFiles)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				// 에러가 발생한 경우 빈 Config가 반환되어야 함
				assert.Equal(t, common.Config{}, config)
			} else {
				assert.NoError(t, err)
				// 성공한 경우 Config 구조체가 반환되어야 함
				assert.IsType(t, common.Config{}, config)
			}
		})
	}
}

func TestLoadWithEnvironmentVariables(t *testing.T) {
	tmpDir := setupTestDir(t)

	configContent := `
serve:
  server_schema: "http"
database:
  driver: "postgresql"
  postgresql:
    host: "localhost"
    port: 5432
`
	configFile := createTestConfigFile(t, tmpDir, "config.yaml", configContent)

	// 환경변수 설정 (viper의 환경변수 매핑을 위해 _ 사용)
	t.Setenv("SERVE_SERVER_SCHEMA", "https")
	t.Setenv("DATABASE_POSTGRESQL_HOST", "remote-db")

	config, err := Load([]string{configFile})

	assert.NoError(t, err)
	assert.Equal(t, "https", config.Serve.ServerSchema)
	assert.Equal(t, "remote-db", config.Database.PostgreSQL.Host)
}

func TestLoadWithDotNotationEnvironmentVariables(t *testing.T) {
	tmpDir := setupTestDir(t)

	configContent := `
serve:
  type: "http"
auth:
  global_secret: "default-secret"
`
	configFile := createTestConfigFile(t, tmpDir, "config.yaml", configContent)

	// .으로 구분된 환경변수가 _로 변경되는지 테스트
	t.Setenv("AUTH_GLOBAL_SECRET", "env-secret")

	config, err := Load([]string{configFile})

	assert.NoError(t, err)
	// 환경변수로 오버라이드된 값이 적용되는지 확인
	assert.Equal(t, "env-secret", config.Auth.GlobalSecret)
}

// 벤치마크 테스트
func BenchmarkLoad(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "config-loader-bench")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	configContent := `
serve:
  type: "http"
  server_schema: "http"
database:
  driver: "postgresql"
  postgresql:
    host: "localhost"
    port: 5432
    database: "testdb"
auth:
  global_secret: "test-secret"
  access_token_lifespan: "24h"
`
	configFile := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Load([]string{configFile})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// 테이블 기반 테스트 - 다양한 시나리오
func TestLoadVariousScenarios(t *testing.T) {
	scenarios := []struct {
		name        string
		baseConfig  string
		extraConfig string
		envVars     map[string]string
		wantErr     bool
		validate    func(t *testing.T, config common.Config)
	}{
		{
			name: "기본 설정만 사용",
			baseConfig: `
serve:
  server_schema: "http"
`,
			wantErr: false,
			validate: func(t *testing.T, config common.Config) {
				assert.Equal(t, "http", config.Serve.ServerSchema)
			},
		},
		{
			name: "설정 파일 병합",
			baseConfig: `
serve:
  server_schema: "http"
database:
  driver: "postgresql"
  postgresql:
    host: "localhost"
`,
			extraConfig: `
serve:
database:
  postgresql:
    port: 5432
    database: "testdb"
`,
			wantErr: false,
			validate: func(t *testing.T, config common.Config) {
				assert.Equal(t, "http", config.Serve.ServerSchema)             // 유지됨
				assert.Equal(t, "localhost", config.Database.PostgreSQL.Host)  // 유지됨
				assert.Equal(t, 5432, config.Database.PostgreSQL.Port)         // 추가됨
				assert.Equal(t, "testdb", config.Database.PostgreSQL.Database) // 추가됨
			},
		},
		{
			name: "환경변수로 오버라이드",
			baseConfig: `
serve:
  server_schema: "http"
`,
			envVars: map[string]string{
				"SERVE_TYPE":          "https",
				"SERVE_SERVER_SCHEMA": "https",
			},
			wantErr: false,
			validate: func(t *testing.T, config common.Config) {
				assert.Equal(t, "https", config.Serve.ServerSchema)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			tmpDir := setupTestDir(t)

			// 환경변수 설정
			for key, value := range scenario.envVars {
				t.Setenv(key, value)
			}

			configFiles := []string{}

			// 기본 설정 파일 생성
			if scenario.baseConfig != "" {
				baseFile := createTestConfigFile(t, tmpDir, "base.yaml", scenario.baseConfig)
				configFiles = append(configFiles, baseFile)
			}

			// 추가 설정 파일 생성
			if scenario.extraConfig != "" {
				extraFile := createTestConfigFile(t, tmpDir, "extra.yaml", scenario.extraConfig)
				configFiles = append(configFiles, extraFile)
			}

			config, err := Load(configFiles)

			if scenario.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if scenario.validate != nil {
					scenario.validate(t, config)
				}
			}
		})
	}
}
