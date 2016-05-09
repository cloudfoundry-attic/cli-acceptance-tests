package application

import (
	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type Scale struct {
	ui        terminal.UI
	config    coreconfig.Reader
	restarter ApplicationRestarter
	appReq    requirements.ApplicationRequirement
	appRepo   applications.ApplicationRepository
}

func init() {
	commandregistry.Register(&Scale{})
}

func (cmd *Scale) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["i"] = &flags.IntFlag{ShortName: "i", Usage: T("Number of instances")}
	fs["k"] = &flags.StringFlag{ShortName: "k", Usage: T("Disk limit (e.g. 256M, 1024M, 1G)")}
	fs["m"] = &flags.StringFlag{ShortName: "m", Usage: T("Memory limit (e.g. 256M, 1024M, 1G)")}
	fs["f"] = &flags.BoolFlag{ShortName: "f", Usage: T("Force restart of app without prompt")}

	return commandregistry.CommandMetadata{
		Name:        "scale",
		Description: T("Change or view the instance count, disk space limit, and memory limit for an app"),
		Usage: []string{
			T("CF_NAME scale APP_NAME [-i INSTANCES] [-k DISK] [-m MEMORY] [-f]"),
		},
		Flags: fs,
	}
}

func (cmd *Scale) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("scale"))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs
}

func (cmd *Scale) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()

	//get command from registry for dependency
	commandDep := commandregistry.Commands.FindCommand("restart")
	commandDep = commandDep.SetDependency(deps, false)
	cmd.restarter = commandDep.(ApplicationRestarter)

	return cmd
}

var bytesInAMegabyte int64 = 1024 * 1024

func (cmd *Scale) Execute(c flags.FlagContext) {
	currentApp := cmd.appReq.GetApplication()
	if !anyFlagsSet(c) {
		cmd.ui.Say(T("Showing current scale of app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
			map[string]interface{}{
				"AppName":     terminal.EntityNameColor(currentApp.Name),
				"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
			}))
		cmd.ui.Ok()
		cmd.ui.Say("")

		cmd.ui.Say("%s %s", terminal.HeaderColor(T("memory:")), formatters.ByteSize(currentApp.Memory*bytesInAMegabyte))
		cmd.ui.Say("%s %s", terminal.HeaderColor(T("disk:")), formatters.ByteSize(currentApp.DiskQuota*bytesInAMegabyte))
		cmd.ui.Say("%s %d", terminal.HeaderColor(T("instances:")), currentApp.InstanceCount)

		return
	}

	params := models.AppParams{}
	shouldRestart := false

	if c.String("m") != "" {
		memory, err := formatters.ToMegabytes(c.String("m"))
		if err != nil {
			cmd.ui.Failed(T("Invalid memory limit: {{.Memory}}\n{{.ErrorDescription}}",
				map[string]interface{}{
					"Memory":           c.String("m"),
					"ErrorDescription": err,
				}))
		}
		params.Memory = &memory
		shouldRestart = true
	}

	if c.String("k") != "" {
		diskQuota, err := formatters.ToMegabytes(c.String("k"))
		if err != nil {
			cmd.ui.Failed(T("Invalid disk quota: {{.DiskQuota}}\n{{.ErrorDescription}}",
				map[string]interface{}{
					"DiskQuota":        c.String("k"),
					"ErrorDescription": err,
				}))
		}
		params.DiskQuota = &diskQuota
		shouldRestart = true
	}

	if c.IsSet("i") {
		instances := c.Int("i")
		params.InstanceCount = &instances
	}

	if shouldRestart && !cmd.confirmRestart(c, currentApp.Name) {
		return
	}

	cmd.ui.Say(T("Scaling app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"AppName":     terminal.EntityNameColor(currentApp.Name),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	updatedApp, apiErr := cmd.appRepo.Update(currentApp.GUID, params)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()

	if shouldRestart {
		cmd.restarter.ApplicationRestart(updatedApp, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name)
	}
}

func (cmd *Scale) confirmRestart(context flags.FlagContext, appName string) bool {
	if context.Bool("f") {
		return true
	}

	result := cmd.ui.Confirm(T("This will cause the app to restart. Are you sure you want to scale {{.AppName}}?",
		map[string]interface{}{"AppName": terminal.EntityNameColor(appName)}))
	cmd.ui.Say("")
	return result
}

func anyFlagsSet(context flags.FlagContext) bool {
	return context.IsSet("m") || context.IsSet("k") || context.IsSet("i")
}
