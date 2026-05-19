package control

import (
	"testing"

	"github.com/stretchr/testify/assert"
	core_command "ntels.com/pharos/core/pkg/command"
	"ntels.com/pharos/core/pkg/common"
)

func TestLoad_ReturnsNoError(t *testing.T) {
	config := common.Config{}

	err := Load(config)
	assert.NoError(t, err)
}

func TestLoad_CommandIsRegistered(t *testing.T) {
	// init()에서 catv-control-settopbox 커맨드가 등록되는지 검증
	cmds := core_command.GetCommands()
	assert.Contains(t, cmds, "catv-control")
}
