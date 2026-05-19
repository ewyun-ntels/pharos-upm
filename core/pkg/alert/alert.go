package alert

import (
	"github.com/flosch/pongo2/v6"
)

func init() {
	// 전역적으로 적용되기 때문에 init()에서 설정
	pongo2.SetAutoescape(false)
}
