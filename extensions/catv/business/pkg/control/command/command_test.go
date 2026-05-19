package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	control_common "ntels.com/pharos/extensions/catv/business/pkg/control/common"
)

func TestCommand_UseAndShort(t *testing.T) {
	cmd := Command()
	require.NotNil(t, cmd)
	assert.Equal(t, control_common.CommandUse, cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestCommand_FlagsDefined(t *testing.T) {
	cmd := Command()
	require.NotNil(t, cmd)

	assert.NotNil(t, cmd.Flag("config"))
	assert.NotNil(t, cmd.Flag("daemon"))
	assert.NotNil(t, cmd.Flag("pid-file"))
	assert.NotNil(t, cmd.Flag("log-file"))
	assert.NotNil(t, cmd.Flag("k8s-job-name"))
}

func TestCommand_Defaults(t *testing.T) {
	cmd := Command()

	configFlag := cmd.Flag("config")
	require.NotNil(t, configFlag)
	// StringArray flags show default as "[value]"
	assert.Equal(t, "[./config/config.toml]", configFlag.DefValue)

	daemonFlag := cmd.Flag("daemon")
	require.NotNil(t, daemonFlag)
	assert.Equal(t, "false", daemonFlag.DefValue)

	kjFlag := cmd.Flag("k8s-job-name")
	require.NotNil(t, kjFlag)
	assert.Equal(t, "", kjFlag.DefValue)
}

func TestCommand_SilenceUsage(t *testing.T) {
	cmd := Command()
	assert.True(t, cmd.SilenceUsage)
}
