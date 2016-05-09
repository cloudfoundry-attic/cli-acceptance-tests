package route

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type DeleteOrphanedRoutes struct {
	ui        terminal.UI
	routeRepo api.RouteRepository
	config    coreconfig.Reader
}

func init() {
	commandregistry.Register(&DeleteOrphanedRoutes{})
}

func (cmd *DeleteOrphanedRoutes) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force deletion without confirmation")}

	return commandregistry.CommandMetadata{
		Name:        "delete-orphaned-routes",
		Description: T("Delete all orphaned routes (i.e. those that are not mapped to an app)"),
		Usage: []string{
			T("CF_NAME delete-orphaned-routes [-f]"),
		},
		Flags: fs,
	}
}

func (cmd *DeleteOrphanedRoutes) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *DeleteOrphanedRoutes) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	return cmd
}

func (cmd *DeleteOrphanedRoutes) Execute(c flags.FlagContext) {
	force := c.Bool("f")
	if !force {
		response := cmd.ui.Confirm(T("Really delete orphaned routes?{{.Prompt}}",
			map[string]interface{}{"Prompt": terminal.PromptColor(">")}))

		if !response {
			return
		}
	}

	cmd.ui.Say(T("Getting routes as {{.Username}} ...\n",
		map[string]interface{}{"Username": terminal.EntityNameColor(cmd.config.Username())}))

	apiErr := cmd.routeRepo.ListRoutes(func(route models.Route) bool {

		if len(route.Apps) == 0 {
			cmd.ui.Say(T("Deleting route {{.Route}}...",
				map[string]interface{}{"Route": terminal.EntityNameColor(route.URL())}))
			apiErr := cmd.routeRepo.Delete(route.GUID)
			if apiErr != nil {
				cmd.ui.Failed(apiErr.Error())
				return false
			}
		}
		return true
	})

	if apiErr != nil {
		cmd.ui.Failed(T("Failed fetching routes.\n{{.Err}}", map[string]interface{}{"Err": apiErr.Error()}))
		return
	}
	cmd.ui.Ok()
}
