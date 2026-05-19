package types

import "fmt"

type Address struct {
	Ip   string `yaml:"ip,omitempty"`
	Port uint16 `yaml:"port,omitempty"`
}

func (a Address) String() string {
	return fmt.Sprintf("%s:%d", a.Ip, a.Port)
}
