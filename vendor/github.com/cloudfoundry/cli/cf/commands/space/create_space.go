package space

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/spacequotas"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type CreateSpace struct {
	ui              terminal.UI
	config          coreconfig.Reader
	spaceRepo       spaces.SpaceRepository
	orgRepo         organizations.OrganizationRepository
	userRepo        api.UserRepository
	spaceRoleSetter user.SpaceRoleSetter
	spaceQuotaRepo  spacequotas.SpaceQuotaRepository
}

func init() {
	commandregistry.Register(&CreateSpace{})
}

func (cmd *CreateSpace) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["o"] = &flags.StringFlag{ShortName: "o", Usage: T("Organization")}
	fs["q"] = &flags.StringFlag{ShortName: "q", Usage: T("Quota to assign to the newly created space")}

	return commandregistry.CommandMetadata{
		Name:        "create-space",
		Description: T("Create a space"),
		Usage: []string{
			T("CF_NAME create-space SPACE [-o ORG] [-q SPACE-QUOTA]"),
		},
		Flags: fs,
	}
}

func (cmd *CreateSpace) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("create-space"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	if fc.String("o") == "" {
		reqs = append(reqs, requirementsFactory.NewTargetedOrgRequirement())
	}

	return reqs
}

func (cmd *CreateSpace) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.userRepo = deps.RepoLocator.GetUserRepository()
	cmd.spaceQuotaRepo = deps.RepoLocator.GetSpaceQuotaRepository()

	//get command from registry for dependency
	commandDep := commandregistry.Commands.FindCommand("set-space-role")
	commandDep = commandDep.SetDependency(deps, false)
	cmd.spaceRoleSetter = commandDep.(user.SpaceRoleSetter)

	return cmd
}

func (cmd *CreateSpace) Execute(c flags.FlagContext) {
	spaceName := c.Args()[0]
	orgName := c.String("o")
	spaceQuotaName := c.String("q")
	orgGUID := ""
	if orgName == "" {
		orgName = cmd.config.OrganizationFields().Name
		orgGUID = cmd.config.OrganizationFields().GUID
	}

	cmd.ui.Say(T("Creating space {{.SpaceName}} in org {{.OrgName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"SpaceName":   terminal.EntityNameColor(spaceName),
			"OrgName":     terminal.EntityNameColor(orgName),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	if orgGUID == "" {
		org, apiErr := cmd.orgRepo.FindByName(orgName)
		switch apiErr.(type) {
		case nil:
		case *errors.ModelNotFoundError:
			cmd.ui.Failed(T("Org {{.OrgName}} does not exist or is not accessible", map[string]interface{}{"OrgName": orgName}))
			return
		default:
			cmd.ui.Failed(T("Error finding org {{.OrgName}}\n{{.ErrorDescription}}",
				map[string]interface{}{
					"OrgName":          orgName,
					"ErrorDescription": apiErr.Error(),
				}))
			return
		}

		orgGUID = org.GUID
	}

	var spaceQuotaGUID string
	if spaceQuotaName != "" {
		spaceQuota, err := cmd.spaceQuotaRepo.FindByNameAndOrgGUID(spaceQuotaName, orgGUID)
		if err != nil {
			cmd.ui.Failed(err.Error())
		}
		spaceQuotaGUID = spaceQuota.GUID
	}

	space, err := cmd.spaceRepo.Create(spaceName, orgGUID, spaceQuotaGUID)
	if err != nil {
		if httpErr, ok := err.(errors.HTTPError); ok && httpErr.ErrorCode() == errors.SpaceNameTaken {
			cmd.ui.Ok()
			cmd.ui.Warn(T("Space {{.SpaceName}} already exists", map[string]interface{}{"SpaceName": spaceName}))
			return
		}
		cmd.ui.Failed(err.Error())
		return
	}
	cmd.ui.Ok()

	err = cmd.spaceRoleSetter.SetSpaceRole(space, models.RoleSpaceManager, cmd.config.UserGUID(), cmd.config.Username())
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	err = cmd.spaceRoleSetter.SetSpaceRole(space, models.RoleSpaceDeveloper, cmd.config.UserGUID(), cmd.config.Username())
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Say(T("\nTIP: Use '{{.CFTargetCommand}}' to target new space",
		map[string]interface{}{
			"CFTargetCommand": terminal.CommandColor(cf.Name + " target -o \"" + orgName + "\" -s \"" + space.Name + "\""),
		}))
}
