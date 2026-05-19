package notification

import (
	"github.com/flosch/pongo2/v6"
	notification_common "ntels.com/pharos/core/pkg/notification/common"
	"ntels.com/pharos/core/pkg/notification/rule"
)

// TODO tls, 일반 서버가 같이 돌아서 일단 flag 설정 차후 제거 필요
var isLoading = false

func init() {
	// 전역적으로 적용되기 때문에 init()에서 설정
	pongo2.SetAutoescape(false)
}

// SendNotification sends an alert notification for the given name and alert value. Returns an error if sending fails.
func SendNotification(name string, value *notification_common.AlertValue) (err error) {
	return rule.Send(name, value)
}
