package plugin

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/pluginconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/utils"
)

type Plugins struct {
	ui     terminal.UI
	config pluginconfig.PluginConfiguration
}

func init() {
	commandregistry.Register(&Plugins{})
}

func (cmd *Plugins) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["checksum"] = &flags.BoolFlag{Name: "checksum", Usage: T("Compute and show the sha1 value of the plugin binary file")}

	return commandregistry.CommandMetadata{
		Name:        "plugins",
		Description: T("List all available plugin commands"),
		Usage: []string{
			T("CF_NAME plugins"),
		},
		Flags: fs,
	}
}

func (cmd *Plugins) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
	}
	return reqs
}

func (cmd *Plugins) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.PluginConfig
	return cmd
}

func (cmd *Plugins) Execute(c flags.FlagContext) {
	var version string

	cmd.ui.Say(T("Listing Installed Plugins..."))

	plugins := cmd.config.Plugins()

	var table *terminal.UITable
	if c.Bool("checksum") {
		cmd.ui.Say(T("Computing sha1 for installed plugins, this may take a while ..."))
		table = cmd.ui.Table([]string{T("Plugin Name"), T("Version"), T("Command Name"), "sha1", T("Command Help")})
	} else {
		table = cmd.ui.Table([]string{T("Plugin Name"), T("Version"), T("Command Name"), T("Command Help")})
	}

	for pluginName, metadata := range plugins {
		if metadata.Version.Major == 0 && metadata.Version.Minor == 0 && metadata.Version.Build == 0 {
			version = "N/A"
		} else {
			version = fmt.Sprintf("%d.%d.%d", metadata.Version.Major, metadata.Version.Minor, metadata.Version.Build)
		}

		for _, command := range metadata.Commands {
			args := []string{pluginName, version}

			if command.Alias != "" {
				args = append(args, command.Name+", "+command.Alias)
			} else {
				args = append(args, command.Name)
			}

			if c.Bool("checksum") {
				checksum := utils.NewSha1Checksum(metadata.Location)
				sha1, err := checksum.ComputeFileSha1()
				if err != nil {
					args = append(args, "n/a")
				} else {
					args = append(args, fmt.Sprintf("%x", sha1))
				}
			}

			args = append(args, command.HelpText)
			table.Add(args...)
		}
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table.Print()
}
