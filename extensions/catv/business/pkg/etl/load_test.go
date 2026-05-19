package etl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	core_command "ntels.com/pharos/core/pkg/command"
)

func TestLoad_CommandIsRegistered(t *testing.T) {
	// init()에서 catv-etl 커맨드가 등록되는지 검증
	cmds := core_command.GetCommands()
	assert.Contains(t, cmds, "catv-etl")
	assert.Equal(t, "catv-etl", cmds["catv-etl"].Use)
}
