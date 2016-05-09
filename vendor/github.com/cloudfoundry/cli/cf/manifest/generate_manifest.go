package manifest

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cli/cf/models"

	"gopkg.in/yaml.v2"

	"io"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

//go:generate counterfeiter . AppManifest

type AppManifest interface {
	BuildpackURL(string, string)
	DiskQuota(string, int64)
	Memory(string, int64)
	Service(string, string)
	StartCommand(string, string)
	EnvironmentVars(string, string, string)
	HealthCheckTimeout(string, int)
	Instances(string, int)
	Domain(string, string, string)
	GetContents() []models.Application
	Stack(string, string)
	AppPorts(string, []int)
	Save(f io.Writer) error
}

type ManifestApplication struct {
	Name       string                 `yaml:"name"`
	Instances  int                    `yaml:"instances,omitempty"`
	Memory     string                 `yaml:"memory,omitempty"`
	DiskQuota  string                 `yaml:"disk_quota,omitempty"`
	AppPorts   []int                  `yaml:"app-ports,omitempty"`
	Host       string                 `yaml:"host,omitempty"`
	Hosts      []string               `yaml:"hosts,omitempty"`
	Domain     string                 `yaml:"domain,omitempty"`
	Domains    []string               `yaml:"domains,omitempty"`
	NoHostname bool                   `yaml:"no-hostname,omitempty"`
	NoRoute    bool                   `yaml:"no-route,omitempty"`
	Buildpack  string                 `yaml:"buildpack,omitempty"`
	Command    string                 `yaml:"command,omitempty"`
	Env        map[string]interface{} `yaml:"env,omitempty"`
	Services   []string               `yaml:"services,omitempty"`
	Stack      string                 `yaml:"stack,omitempty"`
	Timeout    int                    `yaml:"timeout,omitempty"`
}

type ManifestApplications struct {
	Applications []ManifestApplication `yaml:"applications"`
}

type appManifest struct {
	contents []models.Application
}

func NewGenerator() AppManifest {
	return &appManifest{}
}

func (m *appManifest) Stack(appName string, stackName string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].Stack = &models.Stack{
		Name: stackName,
	}
}

func (m *appManifest) Memory(appName string, memory int64) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].Memory = memory
}

func (m *appManifest) DiskQuota(appName string, diskQuota int64) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].DiskQuota = diskQuota
}

func (m *appManifest) StartCommand(appName string, cmd string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].Command = cmd
}

func (m *appManifest) BuildpackURL(appName string, url string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].BuildpackURL = url
}

func (m *appManifest) HealthCheckTimeout(appName string, timeout int) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].HealthCheckTimeout = timeout
}

func (m *appManifest) Instances(appName string, instances int) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].InstanceCount = instances
}

func (m *appManifest) Service(appName string, name string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].Services = append(m.contents[i].Services, models.ServicePlanSummary{
		GUID: "",
		Name: name,
	})
}

func (m *appManifest) Domain(appName string, host string, domain string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].Routes = append(m.contents[i].Routes, models.RouteSummary{
		Host: host,
		Domain: models.DomainFields{
			Name: domain,
		},
	})
}

func (m *appManifest) EnvironmentVars(appName string, key, value string) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].EnvironmentVars[key] = value
}

func (m *appManifest) AppPorts(appName string, appPorts []int) {
	i := m.findOrCreateApplication(appName)
	m.contents[i].AppPorts = appPorts
}

func (m *appManifest) GetContents() []models.Application {
	return m.contents
}

func generateAppMap(app models.Application) (ManifestApplication, error) {
	if app.Stack == nil {
		return ManifestApplication{}, errors.New(T("required attribute 'stack' missing"))
	}

	if app.Memory == 0 {
		return ManifestApplication{}, errors.New(T("required attribute 'memory' missing"))
	}

	if app.DiskQuota == 0 {
		return ManifestApplication{}, errors.New(T("required attribute 'disk_quota' missing"))
	}

	if app.InstanceCount == 0 {
		return ManifestApplication{}, errors.New(T("required attribute 'instances' missing"))
	}

	var services []string
	for _, s := range app.Services {
		services = append(services, s.Name)
	}

	m := ManifestApplication{
		Name:      app.Name,
		Services:  services,
		Buildpack: app.BuildpackURL,
		Memory:    fmt.Sprintf("%dM", app.Memory),
		Command:   app.Command,
		Env:       app.EnvironmentVars,
		Timeout:   app.HealthCheckTimeout,
		Instances: app.InstanceCount,
		DiskQuota: fmt.Sprintf("%dM", app.DiskQuota),
		Stack:     app.Stack.Name,
		AppPorts:  app.AppPorts,
	}

	switch len(app.Routes) {
	case 0:
		m.NoRoute = true
	case 1:
		const noHostname = ""

		m.Domain = app.Routes[0].Domain.Name
		host := app.Routes[0].Host

		if host == noHostname {
			m.NoHostname = true
		} else {
			m.Host = host
		}
	default:
		hosts, domains := separateHostsAndDomains(app.Routes)

		switch len(hosts) {
		case 0:
			m.NoHostname = true
		case 1:
			m.Host = hosts[0]
		default:
			m.Hosts = hosts
		}

		switch len(domains) {
		case 1:
			m.Domain = domains[0]
		default:
			m.Domains = domains
		}
	}

	return m, nil
}

func (m *appManifest) Save(f io.Writer) error {
	apps := ManifestApplications{}

	for _, app := range m.contents {
		appMap, mapErr := generateAppMap(app)
		if mapErr != nil {
			return fmt.Errorf(T("Error saving manifest: {{.Error}}", map[string]interface{}{
				"Error": mapErr.Error(),
			}))
		}
		apps.Applications = append(apps.Applications, appMap)
	}

	contents, err := yaml.Marshal(apps)
	if err != nil {
		return err
	}

	_, err = f.Write(contents)
	if err != nil {
		return err
	}

	return nil
}

func (m *appManifest) findOrCreateApplication(name string) int {
	for i, app := range m.contents {
		if app.Name == name {
			return i
		}
	}
	m.addApplication(name)
	return len(m.contents) - 1
}

func (m *appManifest) addApplication(name string) {
	m.contents = append(m.contents, models.Application{
		ApplicationFields: models.ApplicationFields{
			Name:            name,
			EnvironmentVars: make(map[string]interface{}),
		},
	})
}

func separateHostsAndDomains(routes []models.RouteSummary) ([]string, []string) {
	var (
		hostSlice    []string
		domainSlice  []string
		hostPSlice   *[]string
		domainPSlice *[]string
		hosts        []string
		domains      []string
	)

	for i := 0; i < len(routes); i++ {
		hostSlice = append(hostSlice, routes[i].Host)
		domainSlice = append(domainSlice, routes[i].Domain.Name)
	}

	hostPSlice = removeDuplicatedValue(hostSlice)
	domainPSlice = removeDuplicatedValue(domainSlice)

	if hostPSlice != nil {
		hosts = *hostPSlice
	}
	if domainPSlice != nil {
		domains = *domainPSlice
	}

	return hosts, domains
}
