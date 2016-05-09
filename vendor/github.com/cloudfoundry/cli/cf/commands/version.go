package commands

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type Version struct {
	ui terminal.UI
}

func init() {
	commandregistry.Register(&Version{})
}

func (cmd *Version) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "version",
		Description: T("Print the version"),
		Usage: []string{
			"CF_NAME version",
			"\n\n   ",
			T("'{{.VersionShort}}' and '{{.VersionLong}}' are also accepted.", map[string]string{
				"VersionShort": "cf -v",
				"VersionLong":  "cf --version",
			}),
		},
	}
}

func (cmd *Version) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	return cmd
}

func (cmd *Version) Requirements(requirementsFactory requirements.Factory, context flags.FlagContext) []requirements.Requirement {
	reqs := []requirements.Requirement{}
	return reqs
}

func (cmd *Version) Execute(context flags.FlagContext) {
	cmd.ui.Say(fmt.Sprintf("%s version %s-%s", cf.Name, cf.Version, cf.BuiltOnDate))
}
