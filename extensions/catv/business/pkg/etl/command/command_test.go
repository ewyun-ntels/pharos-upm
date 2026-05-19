package command

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommand_Creation은 Command 함수가 올바르게 Cobra 명령어를 생성하는지 검증합니다.
func TestCommand_Creation(t *testing.T) {
	cmd := Command()

	// 기본 속성 검증
	assert.NotNil(t, cmd)
	assert.Equal(t, "catv-etl", cmd.Use)
	assert.Equal(t, "CATV ETL Processor", cmd.Short)
	assert.True(t, cmd.SilenceUsage)
}

// TestCommand_Flags는 모든 필수 플래그가 올바르게 정의되어 있는지 검증합니다.
func TestCommand_Flags(t *testing.T) {
	cmd := Command()

	// config 플래그 검증
	configFlag := cmd.PersistentFlags().Lookup("config")
	require.NotNil(t, configFlag)
	assert.Equal(t, "c", configFlag.Shorthand)
	assert.Equal(t, "configuration file path", configFlag.Usage)

	// daemon 플래그 검증
	daemonFlag := cmd.PersistentFlags().Lookup("daemon")
	require.NotNil(t, daemonFlag)
	assert.Equal(t, "d", daemonFlag.Shorthand)
	assert.Equal(t, "run as daemon (background process)", daemonFlag.Usage)
	assert.Equal(t, "false", daemonFlag.DefValue)

	// pid-file 플래그 검증
	pidFlag := cmd.PersistentFlags().Lookup("pid-file")
	require.NotNil(t, pidFlag)
	assert.Equal(t, "PID file path (default: /var/run/catv-etl.pid)", pidFlag.Usage)

	// log-file 플래그 검증
	logFlag := cmd.PersistentFlags().Lookup("log-file")
	require.NotNil(t, logFlag)
	assert.Equal(t, "log file path for daemon mode (default: /var/log/catv-etl.log)", logFlag.Usage)
}

func TestCommand_FlagValues(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected map[string]any
	}{
		{
			name: "default values",
			args: []string{},
			expected: map[string]any{
				"daemon": false,
			},
		},
		{
			name: "daemon mode enabled",
			args: []string{"--daemon"},
			expected: map[string]any{
				"daemon": true,
			},
		},
		{
			name: "daemon short flag",
			args: []string{"-d"},
			expected: map[string]any{
				"daemon": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Command()
			cmd.SetArgs(tt.args)

			// Parse flags without running the command
			err := cmd.ParseFlags(tt.args)
			require.NoError(t, err)

			if expected, ok := tt.expected["daemon"].(bool); ok {
				daemon, err := cmd.Flags().GetBool("daemon")
				require.NoError(t, err)
				assert.Equal(t, expected, daemon)
			}
		})
	}
}

func TestCommand_Structure(t *testing.T) {
	cmd := Command()

	// PreRunE 함수가 설정되어 있는지 확인
	assert.NotNil(t, cmd.PreRunE)

	// RunE 함수가 설정되어 있는지 확인
	assert.NotNil(t, cmd.RunE)
}

func TestCommand_Type(t *testing.T) {
	cmd := Command()

	// cobra.Command 타입인지 확인
	assert.IsType(t, &cobra.Command{}, cmd)
}

func TestCommand_PersistentFlags(t *testing.T) {
	cmd := Command()

	// 모든 persistent 플래그가 존재하는지 확인
	requiredFlags := []string{"config", "daemon", "pid-file", "log-file"}
	for _, flagName := range requiredFlags {
		flag := cmd.PersistentFlags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}
}

// Benchmark tests
func BenchmarkCommand_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Command()
	}
}

func BenchmarkCommand_FlagParsing(b *testing.B) {
	args := []string{"--config", "test.toml", "--daemon"}

	for i := 0; i < b.N; i++ {
		cmd := Command()
		cmd.SetArgs(args)
		_ = cmd.ParseFlags(args)
	}
}
