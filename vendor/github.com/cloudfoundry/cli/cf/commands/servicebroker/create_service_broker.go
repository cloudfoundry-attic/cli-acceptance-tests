package servicebroker

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type CreateServiceBroker struct {
	ui                terminal.UI
	config            coreconfig.Reader
	serviceBrokerRepo api.ServiceBrokerRepository
}

func init() {
	commandregistry.Register(&CreateServiceBroker{})
}

func (cmd *CreateServiceBroker) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["space-scoped"] = &flags.BoolFlag{Name: "space-scoped", Usage: T("Make the broker's service plans only visible within the targeted space")}

	return commandregistry.CommandMetadata{
		Name:        "create-service-broker",
		ShortName:   "csb",
		Description: T("Create a service broker"),
		Usage: []string{
			T("CF_NAME create-service-broker SERVICE_BROKER USERNAME PASSWORD URL [--space-scoped]"),
		},
		Flags: fs,
	}
}

func (cmd *CreateServiceBroker) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 4 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SERVICE_BROKER, USERNAME, PASSWORD, URL as arguments\n\n") + commandregistry.Commands.CommandUsage("create-service-broker"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	if fc.IsSet("space-scoped") {
		requiredVersion, err := semver.Make("2.47.0")
		if err != nil {
			panic(err.Error())
		}

		reqs = append(
			reqs,
			requirementsFactory.NewTargetedSpaceRequirement(),
			requirementsFactory.NewMinAPIVersionRequirement("--space-scoped", requiredVersion),
		)
	}

	return reqs
}

func (cmd *CreateServiceBroker) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.serviceBrokerRepo = deps.RepoLocator.GetServiceBrokerRepository()
	return cmd
}

func (cmd *CreateServiceBroker) Execute(c flags.FlagContext) {
	name := c.Args()[0]
	username := c.Args()[1]
	password := c.Args()[2]
	url := c.Args()[3]

	var err error
	if c.Bool("space-scoped") {
		cmd.ui.Say(T("Creating service broker {{.Name}} in org {{.Org}} / space {{.Space}} as {{.Username}}...",
			map[string]interface{}{
				"Name":     terminal.EntityNameColor(name),
				"Org":      terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"Space":    terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"Username": terminal.EntityNameColor(cmd.config.Username())}))
		err = cmd.serviceBrokerRepo.Create(name, url, username, password, cmd.config.SpaceFields().GUID)
	} else {
		cmd.ui.Say(T("Creating service broker {{.Name}} as {{.Username}}...",
			map[string]interface{}{
				"Name":     terminal.EntityNameColor(name),
				"Username": terminal.EntityNameColor(cmd.config.Username())}))
		err = cmd.serviceBrokerRepo.Create(name, url, username, password, "")
	}

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
