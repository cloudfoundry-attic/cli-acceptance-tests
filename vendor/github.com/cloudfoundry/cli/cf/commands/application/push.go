package application

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/flags"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/api/stacks"
	"github.com/cloudfoundry/cli/cf/appfiles"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/words/generator"
)

type Push struct {
	ui            terminal.UI
	config        coreconfig.Reader
	manifestRepo  manifest.ManifestRepository
	appStarter    ApplicationStarter
	appStopper    ApplicationStopper
	serviceBinder service.ServiceBinder
	appRepo       applications.ApplicationRepository
	domainRepo    api.DomainRepository
	routeRepo     api.RouteRepository
	serviceRepo   api.ServiceRepository
	stackRepo     stacks.StackRepository
	authRepo      authentication.AuthenticationRepository
	wordGenerator generator.WordGenerator
	actor         actors.PushActor
	zipper        appfiles.Zipper
	appfiles      appfiles.AppFiles
}

func init() {
	commandregistry.Register(&Push{})
}

func (cmd *Push) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["b"] = &flags.StringFlag{ShortName: "b", Usage: T("Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'")}
	fs["c"] = &flags.StringFlag{ShortName: "c", Usage: T("Startup command, set to null to reset to default start command")}
	fs["d"] = &flags.StringFlag{ShortName: "d", Usage: T("Domain (e.g. example.com)")}
	fs["f"] = &flags.StringFlag{ShortName: "f", Usage: T("Path to manifest")}
	fs["i"] = &flags.IntFlag{ShortName: "i", Usage: T("Number of instances")}
	fs["k"] = &flags.StringFlag{ShortName: "k", Usage: T("Disk limit (e.g. 256M, 1024M, 1G)")}
	fs["m"] = &flags.StringFlag{ShortName: "m", Usage: T("Memory limit (e.g. 256M, 1024M, 1G)")}
	fs["hostname"] = &flags.StringFlag{Name: "hostname", ShortName: "n", Usage: T("Hostname (e.g. my-subdomain)")}
	fs["p"] = &flags.StringFlag{ShortName: "p", Usage: T("Path to app directory or to a zip file of the contents of the app directory")}
	fs["s"] = &flags.StringFlag{ShortName: "s", Usage: T("Stack to use (a stack is a pre-built file system, including an operating system, that can run apps)")}
	fs["t"] = &flags.StringFlag{ShortName: "t", Usage: T("Maximum time (in seconds) for CLI to wait for application start, other server side timeouts may apply")}
	fs["docker-image"] = &flags.StringFlag{Name: "docker-image", ShortName: "o", Usage: T("Docker-image to be used (e.g. user/docker-image-name)")}
	fs["health-check-type"] = &flags.StringFlag{Name: "health-check-type", ShortName: "u", Usage: T("Application health check type (e.g. 'port' or 'none')")}
	fs["no-hostname"] = &flags.BoolFlag{Name: "no-hostname", Usage: T("Map the root domain to this app")}
	fs["no-manifest"] = &flags.BoolFlag{Name: "no-manifest", Usage: T("Ignore manifest file")}
	fs["no-route"] = &flags.BoolFlag{Name: "no-route", Usage: T("Do not map a route to this app and remove routes from previous pushes of this app")}
	fs["no-start"] = &flags.BoolFlag{Name: "no-start", Usage: T("Do not start an app after pushing")}
	fs["random-route"] = &flags.BoolFlag{Name: "random-route", Usage: T("Create a random route for this app")}
	fs["route-path"] = &flags.StringFlag{Name: "route-path", Usage: T("Path for the route")}
	// Hidden:true to hide app-ports for release #117189491
	fs["app-ports"] = &flags.StringFlag{Name: "app-ports", Usage: T("Comma delimited list of ports the application may listen on"), Hidden: true}

	return commandregistry.CommandMetadata{
		Name:        "push",
		ShortName:   "p",
		Description: T("Push a new app or sync changes to an existing app"),
		Usage: []string{
			T("Push a single app (with or without a manifest)"),
			":\n   ",
			fmt.Sprintf("CF_NAME push %s ", T("APP_NAME")),
			fmt.Sprintf("[-b %s] ", T("BUILDPACK_NAME")),
			fmt.Sprintf("[-c %s] ", T("COMMAND")),
			fmt.Sprintf("[-d %s] ", T("DOMAIN")),
			fmt.Sprintf("[-f %s] ", T("MANIFEST_PATH")),
			fmt.Sprintf("[--docker-image %s]", T("DOCKER_IMAGE")),
			"\n   ",
			fmt.Sprintf("[-i %s] ", T("NUM_INSTANCES")),
			fmt.Sprintf("[-k %s] ", T("DISK")),
			fmt.Sprintf("[-m %s] ", T("MEMORY")),
			fmt.Sprintf("[--hostname %s] ", T("HOST")),
			fmt.Sprintf("[-p %s] ", T("PATH")),
			fmt.Sprintf("[-s %s] ", T("STACK")),
			fmt.Sprintf("[-t %s] ", T("TIMEOUT")),
			fmt.Sprintf("[-u %s] ", T("HEALTH_CHECK_TYPE")),
			fmt.Sprintf("[--route-path %s] ", T("ROUTE_PATH")),
			"\n   ",
			// Commented to hide app-ports for release #117189491
			// fmt.Sprintf("[--app-ports %s] ", T("APP_PORTS")),
			"[--no-hostname] [--no-manifest] [--no-route] [--no-start]\n",
			"\n   ",
			T("Push multiple apps with a manifest"),
			":\n   ",
			"CF_NAME push ",
			fmt.Sprintf("[-f %s] ", T("MANIFEST_PATH")),
			"\n",
		},
		Flags: fs,
	}
}

func (cmd *Push) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	var reqs []requirements.Requirement

	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd), "",
		func() bool {
			return len(fc.Args()) > 1
		},
	)

	reqs = append(reqs, usageReq)

	if fc.String("route-path") != "" {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--route-path'", cf.RoutePathMinimumAPIVersion))
	}

	if fc.String("app-ports") != "" {
		reqs = append(reqs, requirementsFactory.NewMinAPIVersionRequirement("Option '--app-ports'", cf.MultipleAppPortsMinimumAPIVersion))
	}

	reqs = append(reqs, []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
	}...)

	return reqs
}

func (cmd *Push) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.manifestRepo = deps.ManifestRepo

	//set appStarter
	appCommand := commandregistry.Commands.FindCommand("start")
	appCommand = appCommand.SetDependency(deps, false)
	cmd.appStarter = appCommand.(ApplicationStarter)

	//set appStopper
	appCommand = commandregistry.Commands.FindCommand("stop")
	appCommand = appCommand.SetDependency(deps, false)
	cmd.appStopper = appCommand.(ApplicationStopper)

	//set serviceBinder
	appCommand = commandregistry.Commands.FindCommand("bind-service")
	appCommand = appCommand.SetDependency(deps, false)
	cmd.serviceBinder = appCommand.(service.ServiceBinder)

	cmd.appRepo = deps.RepoLocator.GetApplicationRepository()
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	cmd.routeRepo = deps.RepoLocator.GetRouteRepository()
	cmd.serviceRepo = deps.RepoLocator.GetServiceRepository()
	cmd.stackRepo = deps.RepoLocator.GetStackRepository()
	cmd.authRepo = deps.RepoLocator.GetAuthenticationRepository()
	cmd.wordGenerator = deps.WordGenerator
	cmd.actor = deps.PushActor
	cmd.zipper = deps.AppZipper
	cmd.appfiles = deps.AppFiles

	return cmd
}

func (cmd *Push) Execute(c flags.FlagContext) {
	appsFromManifest := cmd.getAppParamsFromManifest(c)
	appFromContext := cmd.getAppParamsFromContext(c)
	appSet := cmd.createAppSetFromContextAndManifest(appFromContext, appsFromManifest)

	_, err := cmd.authRepo.RefreshAuthToken()
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	routeActor := actors.NewRouteActor(cmd.ui, cmd.routeRepo)

	for _, appParams := range appSet {
		if appParams.Name == nil {
			cmd.ui.Failed(T("Error: No name found for app"))
		}

		cmd.fetchStackGUID(&appParams)

		if c.IsSet("docker-image") {
			diego := true
			appParams.Diego = &diego
		}

		var app models.Application
		existingApp, err := cmd.appRepo.Read(*appParams.Name)
		switch err.(type) {
		case nil:
			cmd.ui.Say(T("Updating app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
				map[string]interface{}{
					"AppName":   terminal.EntityNameColor(existingApp.Name),
					"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
					"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
					"Username":  terminal.EntityNameColor(cmd.config.Username())}))

			if appParams.EnvironmentVars != nil {
				for key, val := range existingApp.EnvironmentVars {
					if _, ok := (*appParams.EnvironmentVars)[key]; !ok {
						(*appParams.EnvironmentVars)[key] = val
					}
				}
			}

			app, err = cmd.appRepo.Update(existingApp.GUID, appParams)
			if err != nil {
				cmd.ui.Failed(err.Error())
			}
		case *errors.ModelNotFoundError:
			spaceGUID := cmd.config.SpaceFields().GUID
			appParams.SpaceGUID = &spaceGUID

			cmd.ui.Say(T("Creating app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
				map[string]interface{}{
					"AppName":   terminal.EntityNameColor(*appParams.Name),
					"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
					"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
					"Username":  terminal.EntityNameColor(cmd.config.Username())}))

			app, err = cmd.appRepo.Create(appParams)
			if err != nil {
				cmd.ui.Failed(err.Error())
			}
		default:
			cmd.ui.Failed(err.Error())
		}

		cmd.ui.Ok()
		cmd.ui.Say("")

		cmd.updateRoutes(routeActor, app, appParams)

		if c.String("docker-image") == "" {
			err := cmd.actor.ProcessPath(*appParams.Path, cmd.processPathCallback(*appParams.Path, app))
			if err != nil {
				cmd.ui.Failed(
					T("Error processing app files: {{.Error}}",
						map[string]interface{}{
							"Error": err.Error(),
						}),
				)
				return
			}
		}

		if appParams.ServicesToBind != nil {
			cmd.bindAppToServices(*appParams.ServicesToBind, app)
		}

		cmd.restart(app, appParams, c)
	}
}

func (cmd *Push) processPathCallback(path string, app models.Application) func(string) {
	return func(appDir string) {
		localFiles, err := cmd.appfiles.AppFilesInDir(appDir)
		if err != nil {
			cmd.ui.Failed(
				T("Error processing app files in '{{.Path}}': {{.Error}}",
					map[string]interface{}{
						"Path":  path,
						"Error": err.Error(),
					}),
			)
		}

		if len(localFiles) == 0 {
			cmd.ui.Failed(
				T("No app files found in '{{.Path}}'",
					map[string]interface{}{
						"Path": path,
					}),
			)
		}

		cmd.ui.Say(T("Uploading {{.AppName}}...",
			map[string]interface{}{"AppName": terminal.EntityNameColor(app.Name)}))

		err = cmd.uploadApp(app.GUID, appDir, path, localFiles)
		if err != nil {
			cmd.ui.Failed(fmt.Sprintf(T("Error uploading application.\n{{.APIErr}}",
				map[string]interface{}{"APIErr": err.Error()})))
			return
		}
		cmd.ui.Ok()
	}
}

func (cmd *Push) updateRoutes(routeActor actors.RouteActor, app models.Application, appParams models.AppParams) {
	defaultRouteAcceptable := len(app.Routes) == 0
	routeDefined := appParams.Domains != nil || !appParams.IsHostEmpty() || appParams.NoHostname

	if appParams.NoRoute {
		if len(app.Routes) == 0 {
			cmd.ui.Say(T("App {{.AppName}} is a worker, skipping route creation",
				map[string]interface{}{"AppName": terminal.EntityNameColor(app.Name)}))
		} else {
			routeActor.UnbindAll(app)
		}
		return
	}

	if routeDefined || defaultRouteAcceptable {
		if appParams.Domains == nil {
			domain := cmd.findDomain(nil)
			appParams.UseRandomPort = isTcp(domain)
			cmd.processDomainsAndBindRoutes(appParams, routeActor, app, domain)
		} else {
			for _, d := range *(appParams.Domains) {
				domain := cmd.findDomain(&d)
				appParams.UseRandomPort = isTcp(domain)
				cmd.processDomainsAndBindRoutes(appParams, routeActor, app, domain)
			}
		}
	}
}

const TCP = "tcp"

func isTcp(domain models.DomainFields) bool {
	return domain.RouterGroupType == TCP
}

func (cmd *Push) processDomainsAndBindRoutes(
	appParams models.AppParams,
	routeActor actors.RouteActor,
	app models.Application,
	domain models.DomainFields,
) {
	if appParams.IsHostEmpty() {
		cmd.createAndBindRoute(
			nil,
			appParams.UseRandomRoute,
			appParams.UseRandomPort,
			routeActor,
			app,
			appParams.NoHostname,
			domain,
			appParams.RoutePath,
		)
	} else {
		for _, host := range *(appParams.Hosts) {
			cmd.createAndBindRoute(
				&host,
				appParams.UseRandomRoute,
				appParams.UseRandomPort,
				routeActor,
				app,
				appParams.NoHostname,
				domain,
				appParams.RoutePath,
			)
		}
	}
}

func (cmd *Push) createAndBindRoute(
	host *string,
	UseRandomRoute bool,
	UseRandomPort bool,
	routeActor actors.RouteActor,
	app models.Application,
	noHostName bool,
	domain models.DomainFields,
	routePath *string,
) {
	var hostname string
	if !noHostName {
		switch {
		case host != nil:
			hostname = *host
		case UseRandomPort:
			//do nothing
		case UseRandomRoute:
			hostname = hostNameForString(app.Name) + "-" + cmd.wordGenerator.Babble()
		default:
			hostname = hostNameForString(app.Name)
		}
	}

	var route models.Route
	if routePath != nil {
		route = routeActor.FindOrCreateRoute(hostname, domain, *routePath, UseRandomPort)
	} else {
		route = routeActor.FindOrCreateRoute(hostname, domain, "", UseRandomPort)
	}
	routeActor.BindRoute(app, route)
}

var forbiddenHostCharRegex = regexp.MustCompile("[^a-z0-9-]")
var whitespaceRegex = regexp.MustCompile(`[\s_]+`)

func hostNameForString(name string) string {
	name = strings.ToLower(name)
	name = whitespaceRegex.ReplaceAllString(name, "-")
	name = forbiddenHostCharRegex.ReplaceAllString(name, "")
	return name
}

func (cmd *Push) findDomain(domainName *string) models.DomainFields {
	domain, err := cmd.domainRepo.FirstOrDefault(cmd.config.OrganizationFields().GUID, domainName)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	return domain
}

func (cmd *Push) bindAppToServices(services []string, app models.Application) {
	for _, serviceName := range services {
		serviceInstance, err := cmd.serviceRepo.FindInstanceByName(serviceName)

		if err != nil {
			cmd.ui.Failed(T("Could not find service {{.ServiceName}} to bind to {{.AppName}}",
				map[string]interface{}{"ServiceName": serviceName, "AppName": app.Name}))
			return
		}

		cmd.ui.Say(T("Binding service {{.ServiceName}} to app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
			map[string]interface{}{
				"ServiceName": terminal.EntityNameColor(serviceInstance.Name),
				"AppName":     terminal.EntityNameColor(app.Name),
				"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"Username":    terminal.EntityNameColor(cmd.config.Username())}))

		err = cmd.serviceBinder.BindApplication(app, serviceInstance, nil)

		switch httpErr := err.(type) {
		case errors.HTTPError:
			if httpErr.ErrorCode() == errors.ServiceBindingAppServiceTaken {
				err = nil
			}
		}

		if err != nil {
			cmd.ui.Failed(T("Could not bind to service {{.ServiceName}}\nError: {{.Err}}",
				map[string]interface{}{"ServiceName": serviceName, "Err": err.Error()}))
		}

		cmd.ui.Ok()
	}
}

func (cmd *Push) fetchStackGUID(appParams *models.AppParams) {
	if appParams.StackName == nil {
		return
	}

	stackName := *appParams.StackName
	cmd.ui.Say(T("Using stack {{.StackName}}...",
		map[string]interface{}{"StackName": terminal.EntityNameColor(stackName)}))

	stack, err := cmd.stackRepo.FindByName(stackName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
	appParams.StackGUID = &stack.GUID
}

func (cmd *Push) restart(app models.Application, params models.AppParams, c flags.FlagContext) {
	if app.State != T("stopped") {
		cmd.ui.Say("")
		app, _ = cmd.appStopper.ApplicationStop(app, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name)
	}

	cmd.ui.Say("")

	if c.Bool("no-start") {
		return
	}

	if params.HealthCheckTimeout != nil {
		cmd.appStarter.SetStartTimeoutInSeconds(*params.HealthCheckTimeout)
	}

	cmd.appStarter.ApplicationStart(app, cmd.config.OrganizationFields().Name, cmd.config.SpaceFields().Name)
}

func (cmd *Push) getAppParamsFromManifest(c flags.FlagContext) []models.AppParams {
	if c.Bool("no-manifest") {
		return []models.AppParams{}
	}

	var path string
	if c.String("f") != "" {
		path = c.String("f")
	} else {
		var err error
		path, err = os.Getwd()
		if err != nil {
			cmd.ui.Failed(T("Could not determine the current working directory!"), err)
		}
	}

	m, err := cmd.manifestRepo.ReadManifest(path)

	if err != nil {
		if m.Path == "" && c.String("f") == "" {
			return []models.AppParams{}
		}
		cmd.ui.Failed(T("Error reading manifest file:\n{{.Err}}", map[string]interface{}{"Err": err.Error()}))
	}

	apps, err := m.Applications()
	if err != nil {
		cmd.ui.Failed("Error reading manifest file:\n%s", err)
	}

	cmd.ui.Say(T("Using manifest file {{.Path}}\n",
		map[string]interface{}{"Path": terminal.EntityNameColor(m.Path)}))
	return apps
}

func (cmd *Push) createAppSetFromContextAndManifest(contextApp models.AppParams, manifestApps []models.AppParams) []models.AppParams {
	var err error
	var apps []models.AppParams

	switch len(manifestApps) {
	case 0:
		if contextApp.Name == nil {
			cmd.ui.Failed(
				T("Manifest file is not found in the current directory, please provide either an app name or manifest") +
					"\n\n" +
					commandregistry.Commands.CommandUsage("push"),
			)
		} else {
			err = addApp(&apps, contextApp)
		}
	case 1:
		manifestApps[0].Merge(&contextApp)
		err = addApp(&apps, manifestApps[0])
	default:
		selectedAppName := contextApp.Name
		contextApp.Name = nil

		if !contextApp.IsEmpty() {
			cmd.ui.Failed("%s", T("Incorrect Usage. Command line flags (except -f) cannot be applied when pushing multiple apps from a manifest file."))
		}

		if selectedAppName != nil {
			var foundApp bool
			for _, appParams := range manifestApps {
				if appParams.Name != nil && *appParams.Name == *selectedAppName {
					foundApp = true
					addApp(&apps, appParams)
				}
			}

			if !foundApp {
				err = errors.New(T("Could not find app named '{{.AppName}}' in manifest", map[string]interface{}{"AppName": *selectedAppName}))
			}
		} else {
			for _, manifestApp := range manifestApps {
				addApp(&apps, manifestApp)
			}
		}
	}

	if err != nil {
		cmd.ui.Failed(T("Error: {{.Err}}", map[string]interface{}{"Err": err.Error()}))
	}

	return apps
}

func addApp(apps *[]models.AppParams, app models.AppParams) error {
	if app.Name == nil {
		return errors.New(T("App name is a required field"))
	}

	if app.Path == nil {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		app.Path = &cwd
	}

	*apps = append(*apps, app)

	return nil
}

func (cmd *Push) getAppParamsFromContext(c flags.FlagContext) models.AppParams {
	appParams := models.AppParams{
		NoRoute:        c.Bool("no-route"),
		UseRandomRoute: c.Bool("random-route"),
		NoHostname:     c.Bool("no-hostname"),
	}

	if len(c.Args()) > 0 {
		appParams.Name = &c.Args()[0]
	}

	if c.String("n") != "" {
		appParams.Hosts = &[]string{c.String("n")}
	}

	if c.String("route-path") != "" {
		routePath := c.String("route-path")
		appParams.RoutePath = &routePath
	}

	if c.String("app-ports") != "" {
		appPortStrings := strings.Split(c.String("app-ports"), ",")
		appPorts := make([]int, len(appPortStrings))

		for i, s := range appPortStrings {
			p, err := strconv.Atoi(s)
			if err != nil {
				cmd.ui.Failed(T("Invalid app port: {{.AppPort}}\nApp port must be a number", map[string]interface{}{
					"AppPort": s,
				}))
			}
			appPorts[i] = p
		}

		appParams.AppPorts = &appPorts
	}

	if c.String("b") != "" {
		buildpack := c.String("b")
		if buildpack == "null" || buildpack == "default" {
			buildpack = ""
		}
		appParams.BuildpackURL = &buildpack
	}

	if c.String("c") != "" {
		command := c.String("c")
		if command == "null" || command == "default" {
			command = ""
		}
		appParams.Command = &command
	}

	if c.String("d") != "" {
		appParams.Domains = &[]string{c.String("d")}
	}

	if c.IsSet("i") {
		instances := c.Int("i")
		if instances < 1 {
			cmd.ui.Failed(T("Invalid instance count: {{.InstancesCount}}\nInstance count must be a positive integer",
				map[string]interface{}{"InstancesCount": instances}))
		}
		appParams.InstanceCount = &instances
	}

	if c.String("k") != "" {
		diskQuota, err := formatters.ToMegabytes(c.String("k"))
		if err != nil {
			cmd.ui.Failed(T("Invalid disk quota: {{.DiskQuota}}\n{{.Err}}",
				map[string]interface{}{"DiskQuota": c.String("k"), "Err": err.Error()}))
		}
		appParams.DiskQuota = &diskQuota
	}

	if c.String("m") != "" {
		memory, err := formatters.ToMegabytes(c.String("m"))
		if err != nil {
			cmd.ui.Failed(T("Invalid memory limit: {{.MemLimit}}\n{{.Err}}",
				map[string]interface{}{"MemLimit": c.String("m"), "Err": err.Error()}))
		}
		appParams.Memory = &memory
	}

	if c.String("docker-image") != "" {
		dockerImage := c.String("docker-image")
		appParams.DockerImage = &dockerImage
	}

	if c.String("p") != "" {
		path := c.String("p")
		appParams.Path = &path
	}

	if c.String("s") != "" {
		stackName := c.String("s")
		appParams.StackName = &stackName
	}

	if c.String("t") != "" {
		timeout, err := strconv.Atoi(c.String("t"))
		if err != nil {
			cmd.ui.Failed("Error: %s", fmt.Errorf(T("Invalid timeout param: {{.Timeout}}\n{{.Err}}",
				map[string]interface{}{"Timeout": c.String("t"), "Err": err.Error()})))
		}

		appParams.HealthCheckTimeout = &timeout
	}

	if healthCheckType := c.String("u"); healthCheckType != "" {
		if healthCheckType != "port" && healthCheckType != "none" {
			cmd.ui.Failed("Error: %s", fmt.Errorf(T("Invalid health-check-type param: {{.healthCheckType}}",
				map[string]interface{}{"healthCheckType": healthCheckType})))
		}

		appParams.HealthCheckType = &healthCheckType
	}

	return appParams
}

func (cmd *Push) uploadApp(appGUID, appDir, appDirOrZipFile string, localFiles []models.AppFileFields) error {
	uploadDir, err := ioutil.TempDir("", "apps")
	if err != nil {
		return err
	}
	defer os.RemoveAll(uploadDir)

	remoteFiles, hasFileToUpload, err := cmd.actor.GatherFiles(localFiles, appDir, uploadDir)
	if err != nil {
		return err
	}

	zipFile, err := ioutil.TempFile("", "uploads")
	if err != nil {
		return err
	}
	defer func() {
		zipFile.Close()
		os.Remove(zipFile.Name())
	}()

	if hasFileToUpload {
		err = cmd.zipper.Zip(uploadDir, zipFile)
		if err != nil {
			if emptyDirErr, ok := err.(*errors.EmptyDirError); ok {
				return emptyDirErr
			}
			return fmt.Errorf("%s: %s", T("Error zipping application"), err.Error())
		}

		zipFileSize, err := cmd.zipper.GetZipSize(zipFile)
		if err != nil {
			return err
		}

		zipFileCount := cmd.appfiles.CountFiles(uploadDir)
		if zipFileCount > 0 {
			cmd.ui.Say(T("Uploading app files from: {{.Path}}", map[string]interface{}{"Path": appDir}))
			cmd.ui.Say(T("Uploading {{.ZipFileBytes}}, {{.FileCount}} files",
				map[string]interface{}{
					"ZipFileBytes": formatters.ByteSize(zipFileSize),
					"FileCount":    zipFileCount}))
		}
	}

	return cmd.actor.UploadApp(appGUID, zipFile, remoteFiles)
}
