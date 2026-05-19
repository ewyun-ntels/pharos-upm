package schedules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/pkg/common"
	control_common "ntels.com/pharos/extensions/catv/business/pkg/control/common"
)

func TestApi_GetRelativePath(t *testing.T) {
	a := &Api{}
	a.Init("", common.Config{})
	assert.Equal(t, control_common.HttpRelativePath, a.GetRelativePath())
}

func TestApi_Use(t *testing.T) {
	a := &Api{}
	assert.True(t, a.Use())
}
