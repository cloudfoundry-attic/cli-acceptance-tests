package spacequota

import (
	"github.com/cloudfoundry/cli/cf/api/spacequotas"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type DeleteSpaceQuota struct {
	ui             terminal.UI
	config         coreconfig.Reader
	spaceQuotaRepo spacequotas.SpaceQuotaRepository
}

func init() {
	commandregistry.Register(&DeleteSpaceQuota{})
}

func (cmd *DeleteSpaceQuota) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force delete (do not prompt for confirmation)")}

	return commandregistry.CommandMetadata{
		Name:        "delete-space-quota",
		Description: T("Delete a space quota definition and unassign the space quota from all spaces"),
		Usage: []string{
			T("CF_NAME delete-space-quota SPACE-QUOTA-NAME [-f]"),
		},
		Flags: fs,
	}
}

func (cmd *DeleteSpaceQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("delete-space-quota"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}

	return reqs
}

func (cmd *DeleteSpaceQuota) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceQuotaRepo = deps.RepoLocator.GetSpaceQuotaRepository()
	return cmd
}

func (cmd *DeleteSpaceQuota) Execute(c flags.FlagContext) {
	quotaName := c.Args()[0]

	if !c.Bool("f") {
		response := cmd.ui.ConfirmDelete("quota", quotaName)
		if !response {
			return
		}
	}

	cmd.ui.Say(T("Deleting space quota {{.QuotaName}} as {{.Username}}...", map[string]interface{}{
		"QuotaName": terminal.EntityNameColor(quotaName),
		"Username":  terminal.EntityNameColor(cmd.config.Username()),
	}))

	quota, apiErr := cmd.spaceQuotaRepo.FindByName(quotaName)
	switch (apiErr).(type) {
	case nil: // no error
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("Quota {{.QuotaName}} does not exist", map[string]interface{}{"QuotaName": quotaName}))
		return
	default:
		cmd.ui.Failed(apiErr.Error())
	}

	apiErr = cmd.spaceQuotaRepo.Delete(quota.GUID)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
	}

	cmd.ui.Ok()
}
