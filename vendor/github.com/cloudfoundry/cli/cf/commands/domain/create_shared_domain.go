package domain

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type CreateSharedDomain struct {
	ui             terminal.UI
	config         coreconfig.Reader
	domainRepo     api.DomainRepository
	routingAPIRepo api.RoutingAPIRepository
}

func init() {
	commandregistry.Register(&CreateSharedDomain{})
}

func (cmd *CreateSharedDomain) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["router-group"] = &flags.StringFlag{Name: "router-group", Usage: T("Routes for this domain will be configured only on the specified router group")}
	return commandregistry.CommandMetadata{
		Name:        "create-shared-domain",
		Description: T("Create a domain that can be used by all orgs (admin-only)"),
		Usage: []string{
			T("CF_NAME create-shared-domain DOMAIN [--router-group ROUTER_GROUP]"),
		},
		Flags: fs,
	}
}

func (cmd *CreateSharedDomain) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires DOMAIN as an argument\n\n") + commandregistry.Commands.CommandUsage("create-shared-domain"))
	}

	requiredVersion, err := semver.Make("2.36.0")
	if err != nil {
		panic(err.Error())
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	if fc.String("router-group") != "" {
		reqs = append(reqs, []requirements.Requirement{
			requirementsFactory.NewMinAPIVersionRequirement("Option '--router-group'", requiredVersion),
			requirementsFactory.NewRoutingAPIRequirement(),
		}...)
	}

	return reqs
}

func (cmd *CreateSharedDomain) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	cmd.routingAPIRepo = deps.RepoLocator.GetRoutingAPIRepository()
	return cmd
}

func (cmd *CreateSharedDomain) Execute(c flags.FlagContext) {
	var routerGroup models.RouterGroup
	domainName := c.Args()[0]
	routerGroupName := c.String("router-group")

	if routerGroupName != "" {
		var routerGroupFound bool
		err := cmd.routingAPIRepo.ListRouterGroups(func(group models.RouterGroup) bool {
			if group.Name == routerGroupName {
				routerGroup = group
				routerGroupFound = true
				return false
			}

			return true
		})

		if err != nil {
			cmd.ui.Failed(err.Error())
		}
		if !routerGroupFound {
			cmd.ui.Failed(T("Router group {{.RouterGroup}} not found", map[string]interface{}{
				"RouterGroup": routerGroupName,
			}))
		}
	}

	cmd.ui.Say(T("Creating shared domain {{.DomainName}} as {{.Username}}...",
		map[string]interface{}{
			"DomainName": terminal.EntityNameColor(domainName),
			"Username":   terminal.EntityNameColor(cmd.config.Username())}))

	err := cmd.domainRepo.CreateSharedDomain(domainName, routerGroup.GUID)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
}
