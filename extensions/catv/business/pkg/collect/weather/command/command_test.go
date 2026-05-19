package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand_Creation(t *testing.T) {
	cmd := Command()

	assert.NotNil(t, cmd)
	assert.Equal(t, "catv-collect-weather", cmd.Use)
	assert.Equal(t, "CATV Collect Weather Data Processor", cmd.Short)
	assert.True(t, cmd.SilenceUsage)
}

func TestCommand_Flags(t *testing.T) {
	cmd := Command()

	configFlag := cmd.PersistentFlags().Lookup("config")
	require.NotNil(t, configFlag)
	assert.Equal(t, "c", configFlag.Shorthand)
	assert.Equal(t, "configuration file path", configFlag.Usage)

	daemonFlag := cmd.PersistentFlags().Lookup("daemon")
	require.NotNil(t, daemonFlag)
	assert.Equal(t, "d", daemonFlag.Shorthand)
	assert.Equal(t, "run as daemon (background process)", daemonFlag.Usage)
	assert.Equal(t, "false", daemonFlag.DefValue)

	pidFlag := cmd.PersistentFlags().Lookup("pid-file")
	require.NotNil(t, pidFlag)
	assert.Equal(t, "/var/run/catv-weather.pid", pidFlag.DefValue)

	logFlag := cmd.PersistentFlags().Lookup("log-file")
	require.NotNil(t, logFlag)
	assert.Equal(t, "/var/log/catv-weather.log", logFlag.DefValue)
}

func TestCommand_DaemonFlagDefault(t *testing.T) {
	cmd := Command()

	isDaemon, err := cmd.PersistentFlags().GetBool("daemon")
	require.NoError(t, err)
	assert.False(t, isDaemon)
}

func TestCommand_ConfigFlagDefault(t *testing.T) {
	cmd := Command()

	configPaths, err := cmd.PersistentFlags().GetStringArray("config")
	require.NoError(t, err)
	require.Len(t, configPaths, 1)
	assert.Equal(t, "./config/config.toml", configPaths[0])
}

func TestCommand_HasPreRunE(t *testing.T) {
	cmd := Command()
	assert.NotNil(t, cmd.PreRunE)
}

func TestCommand_HasRunE(t *testing.T) {
	cmd := Command()
	assert.NotNil(t, cmd.RunE)
}
