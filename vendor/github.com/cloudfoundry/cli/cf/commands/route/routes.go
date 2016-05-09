package route

import (
	"fmt"
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type ListRoutes struct {
	ui         terminal.UI
	routeRepo  api.RouteRepository
	domainRepo api.DomainRepository
	config     coreconfig.Reader
}

func init() {
	commandregistry.Register(&ListRoutes{})
}

func (cmd *ListRoutes) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["orglevel"] = &flags.BoolFlag{Name: "orglevel", Usage: T("List all the routes for all spaces of current organization")}

	return commandregistry.CommandMetadata{
		Name:        "routes",
		ShortName:   "r",
		Description: T("List all routes in the current space or the current organization"),
		Usage: []string{
			"CF_NAME routes [--orglevel]",
		},
		Flags: fs,
	}
}

func (cmd *ListRoutes) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}

	return reqs
}

func (cmd *ListRoutes) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	return cmd
}

func (cmd *ListRoutes) Execute(c flags.FlagContext) {
	orglevel := c.Bool("orglevel")

	if orglevel {
		cmd.ui.Say(T("Getting routes for org {{.OrgName}} as {{.Username}} ...\n",
			map[string]interface{}{
				"Username": terminal.EntityNameColor(cmd.config.Username()),
				"OrgName":  terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			}))
	} else {
		cmd.ui.Say(T("Getting routes for org {{.OrgName}} / space {{.SpaceName}} as {{.Username}} ...\n",
			map[string]interface{}{
				"Username":  terminal.EntityNameColor(cmd.config.Username()),
				"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			}))
	}

	table := cmd.ui.Table([]string{T("space"), T("host"), T("domain"), T("port"), T("path"), T("type"), T("apps"), T("service")})

	d := make(map[string]models.DomainFields)
	cmd.domainRepo.ListDomainsForOrg(cmd.config.OrganizationFields().GUID, func(domain models.DomainFields) bool {
		d[domain.GUID] = domain
		return true
	})

	var routesFound bool
	cb := func(route models.Route) bool {
		routesFound = true
		appNames := []string{}
		for _, app := range route.Apps {
			appNames = append(appNames, app.Name)
		}

		var port string
		if route.Port != 0 {
			port = fmt.Sprintf("%d", route.Port)
		}

		domain := d[route.Domain.GUID]

		table.Add(
			route.Space.Name,
			route.Host,
			route.Domain.Name,
			port,
			route.Path,
			domain.RouterGroupType,
			strings.Join(appNames, ","),
			route.ServiceInstance.Name,
		)
		return true
	}

	var err error
	if orglevel {
		err = cmd.routeRepo.ListAllRoutes(cb)
	} else {
		err = cmd.routeRepo.ListRoutes(cb)
	}

	table.Print()
	if err != nil {
		cmd.ui.Failed(T("Failed fetching routes.\n{{.Err}}", map[string]interface{}{"Err": err.Error()}))
	}

	if !routesFound {
		cmd.ui.Say(T("No routes found"))
	}
}
