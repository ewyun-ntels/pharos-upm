package command

import (
	"github.com/spf13/cobra"
	"ntels.com/pharos/core/internal"
)

var commands = internal.NewMap[*cobra.Command]()

func Register(name string, command *cobra.Command) {
	commands.Set(name, command)
}

func GetCommands() map[string]*cobra.Command {
	return commands.GetAll()
}
