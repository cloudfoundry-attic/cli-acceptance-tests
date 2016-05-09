package fakecommand

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type FakeCommand2 struct {
	Data string
	req  fakeReq2
	ui   terminal.UI
}

func init() {
	commandregistry.Register(FakeCommand2{Data: "FakeCommand2 data", req: fakeReq2{}})
}

func (cmd FakeCommand2) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "fake-command2",
		Description: "Description for fake-command2 with bad requirement",
		Usage: []string{
			"Usage of fake-command",
		},
	}
}

func (cmd FakeCommand2) Requirements(_ requirements.Factory, _ flags.FlagContext) []requirements.Requirement {
	reqs := []requirements.Requirement{cmd.req}
	return reqs
}

func (cmd FakeCommand2) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.req.ui = deps.UI
	cmd.ui = deps.UI
	cmd.ui.Say("SetDependency() called, pluginCall " + fmt.Sprintf("%t", pluginCall))

	return cmd
}

func (cmd FakeCommand2) Execute(c flags.FlagContext) {
	cmd.ui.Say("Command Executed")
}

type fakeReq2 struct {
	ui terminal.UI
}

func (f fakeReq2) Execute() error {
	f.ui.Say("Requirement executed and failed")
	return errors.New("Requirement executed and failed")
}
