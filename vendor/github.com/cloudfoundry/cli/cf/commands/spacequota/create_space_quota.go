package spacequota

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/api/spacequotas"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type CreateSpaceQuota struct {
	ui        terminal.UI
	config    coreconfig.Reader
	quotaRepo spacequotas.SpaceQuotaRepository
	orgRepo   organizations.OrganizationRepository
}

func init() {
	commandregistry.Register(&CreateSpaceQuota{})
}

func (cmd *CreateSpaceQuota) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["allow-paid-service-plans"] = &flags.BoolFlag{Name: "allow-paid-service-plans", Usage: T("Can provision instances of paid service plans (Default: disallowed)")}
	fs["i"] = &flags.StringFlag{ShortName: "i", Usage: T("Maximum amount of memory an application instance can have (e.g. 1024M, 1G, 10G). -1 represents an unlimited amount. (Default: unlimited)")}
	fs["m"] = &flags.StringFlag{ShortName: "m", Usage: T("Total amount of memory a space can have (e.g. 1024M, 1G, 10G)")}
	fs["r"] = &flags.IntFlag{ShortName: "r", Usage: T("Total number of routes")}
	fs["s"] = &flags.IntFlag{ShortName: "s", Usage: T("Total number of service instances")}
	fs["a"] = &flags.IntFlag{ShortName: "a", Usage: T("Total number of application instances. -1 represents an unlimited amount. (Default: unlimited)")}

	return commandregistry.CommandMetadata{
		Name:        "create-space-quota",
		Description: T("Define a new space resource quota"),
		Usage: []string{
			"CF_NAME create-space-quota ",
			T("QUOTA"),
			" ",
			fmt.Sprintf("[-i %s] ", T("INSTANCE_MEMORY")),
			fmt.Sprintf("[-m %s] ", T("MEMORY")),
			fmt.Sprintf("[-r %s] ", T("ROUTES")),
			fmt.Sprintf("[-s %s] ", T("SERVICE_INSTANCES")),
			fmt.Sprintf("[-a %s] ", T("APP_INSTANCES")),
			"[--allow-paid-service-plans]",
		},
		Flags: fs,
	}
}

func (cmd *CreateSpaceQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("create-space-quota"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}

	if fc.IsSet("a") {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '-a'", cf.SpaceAppInstanceLimitMinimumAPIVersion))
	}

	return reqs
}

func (cmd *CreateSpaceQuota) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.quotaRepo = deps.RepoLocator.GetSpaceQuotaRepository()
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	return cmd
}

func (cmd *CreateSpaceQuota) Execute(context flags.FlagContext) {
	name := context.Args()[0]
	org := cmd.config.OrganizationFields()

	cmd.ui.Say(T("Creating space quota {{.QuotaName}} for org {{.OrgName}} as {{.Username}}...", map[string]interface{}{
		"QuotaName": terminal.EntityNameColor(name),
		"OrgName":   terminal.EntityNameColor(org.Name),
		"Username":  terminal.EntityNameColor(cmd.config.Username()),
	}))

	quota := models.SpaceQuota{
		Name:    name,
		OrgGUID: org.GUID,
	}

	memoryLimit := context.String("m")
	if memoryLimit != "" {
		parsedMemory, errr := formatters.ToMegabytes(memoryLimit)
		if errr != nil {
			cmd.ui.Failed(T("Invalid memory limit: {{.MemoryLimit}}\n{{.Err}}", map[string]interface{}{"MemoryLimit": memoryLimit, "Err": errr}))
		}

		quota.MemoryLimit = parsedMemory
	}

	instanceMemoryLimit := context.String("i")
	var parsedMemory int64
	var err error
	if instanceMemoryLimit == "-1" || instanceMemoryLimit == "" {
		parsedMemory = -1
	} else {
		parsedMemory, err = formatters.ToMegabytes(instanceMemoryLimit)
		if err != nil {
			cmd.ui.Failed(T("Invalid instance memory limit: {{.MemoryLimit}}\n{{.Err}}", map[string]interface{}{"MemoryLimit": instanceMemoryLimit, "Err": err}))
		}
	}

	quota.InstanceMemoryLimit = parsedMemory

	if context.IsSet("r") {
		quota.RoutesLimit = context.Int("r")
	}

	if context.IsSet("s") {
		quota.ServicesLimit = context.Int("s")
	}

	if context.IsSet("allow-paid-service-plans") {
		quota.NonBasicServicesAllowed = true
	}

	if context.IsSet("a") {
		quota.AppInstanceLimit = context.Int("a")
	} else {
		quota.AppInstanceLimit = resources.UnlimitedAppInstances
	}

	err = cmd.quotaRepo.Create(quota)

	httpErr, ok := err.(errors.HTTPError)
	if ok && httpErr.ErrorCode() == errors.QuotaDefinitionNameTaken {
		cmd.ui.Ok()
		cmd.ui.Warn(T("Space Quota Definition {{.QuotaName}} already exists", map[string]interface{}{"QuotaName": quota.Name}))
		return
	}

	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
