package rule

import (
	"errors"
	"log/slog"
	"sync"

	"ntels.com/pharos/core/pkg/notification/common"
)

type Manager struct {
	Rules map[string]Rule
	sync.RWMutex
}

var gManager Manager

func init() {
	gManager = Manager{
		Rules: make(map[string]Rule),
	}
}

func Register(rule Rule) error {
	gManager.Lock()
	defer gManager.Unlock()

	name := rule.GetName()
	slog.Debug("register new rule", "name", name)
	if _, ok := gManager.Rules[name]; ok {
		return errors.New("rule already registered(" + name + ")")
	}

	gManager.Rules[name] = rule

	return nil
}

func Unregister(name string) error {
	gManager.Lock()
	defer gManager.Unlock()

	delete(gManager.Rules, name)

	return nil
}

func Send(name string, value *common.AlertValue) (err error) {
	gManager.RLock()
	defer gManager.RUnlock()

	rule, ok := gManager.Rules[name]
	if !ok {
		return errors.New("rule not found")
	}

	err = rule.Send(value)
	if err != nil {
		return err
	}

	return
}
