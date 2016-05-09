package fakecommand

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"
)

type FakeCommand1 struct {
	Data string
}

func init() {
	commandregistry.Register(FakeCommand1{Data: "FakeCommand1 data"})
}

func (cmd FakeCommand1) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: "Usage for BoolFlag"}
	fs["boolFlag"] = &flags.BoolFlag{Name: "BoolFlag", Usage: "Usage for BoolFlag"}
	fs["intFlag"] = &flags.IntFlag{Name: "intFlag", Usage: "Usage for intFlag"}

	return commandregistry.CommandMetadata{
		Name:        "fake-command",
		ShortName:   "fc1",
		Description: "Description for fake-command",
		Usage: []string{
			"CF_NAME Usage of fake-command",
		},
		Flags: fs,
	}
}

func (cmd FakeCommand1) Requirements(_ requirements.Factory, _ flags.FlagContext) []requirements.Requirement {
	return []requirements.Requirement{}
}

func (cmd FakeCommand1) SetDependency(deps commandregistry.Dependency, _ bool) commandregistry.Command {
	return cmd
}

func (cmd FakeCommand1) Execute(c flags.FlagContext) {
	fmt.Println("This is fake-command")
}
