package servicebroker

import (
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/runner"

	"io/ioutil"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var (
	DEFAULT_APP_TIMEOUT  = 30 * time.Second
	DEFAULT_TIMEOUT      = 45 * time.Second
	BROKER_START_TIMEOUT = 5 * time.Minute
	DEFAULT_MEMORY_LIMIT = "256M"
)

type Plan struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type ServiceBroker struct {
	Name    string
	Path    string
	context helpers.SuiteContext
	Service struct {
		Name            string `json:"name"`
		ID              string `json:"id"`
		DashboardClient struct {
			ID          string `json:"id"`
			Secret      string `json:"secret"`
			RedirectUri string `json:"redirect_uri"`
		}
	}
	SyncPlans  []Plan
	AsyncPlans []Plan
}

func NewServiceBroker(name string, path string, context helpers.SuiteContext) ServiceBroker {
	b := ServiceBroker{}
	b.Path = path
	b.Name = name
	b.Service.Name = generator.RandomName()
	b.Service.ID = generator.RandomName()
	b.SyncPlans = []Plan{
		{Name: generator.RandomName(), ID: generator.RandomName()},
		{Name: generator.RandomName(), ID: generator.RandomName()},
	}
	b.AsyncPlans = []Plan{
		{Name: generator.RandomName(), ID: generator.RandomName()},
		{Name: generator.RandomName(), ID: generator.RandomName()},
	}
	b.Service.DashboardClient.ID = generator.RandomName()
	b.Service.DashboardClient.Secret = generator.RandomName()
	b.Service.DashboardClient.RedirectUri = generator.RandomName()
	b.context = context
	return b
}

func (b ServiceBroker) Push() {
	config := helpers.LoadConfig()
	Expect(cf.Cf(
		"push", b.Name,
		"--no-start",
		"-b", config.RubyBuildpackName,
		"-m", DEFAULT_MEMORY_LIMIT,
		"-p", b.Path,
		"-d", config.AppsDomain,
	).Wait(BROKER_START_TIMEOUT)).To(Exit(0))
	SetBackend(b.Name)
	Expect(cf.Cf("start", b.Name).Wait(BROKER_START_TIMEOUT)).To(Exit(0))
}

func (b ServiceBroker) Configure() {
	Expect(runner.Curl(helpers.AppUri(b.Name, "/config"), "-d", b.ToJSON()).Wait(DEFAULT_TIMEOUT)).To(Exit(0))
}

func (b ServiceBroker) Create() {
	cf.AsUser(b.context.AdminUserContext(), DEFAULT_TIMEOUT, func() {
		Expect(cf.Cf("create-service-broker", b.Name, "username", "password", helpers.AppUri(b.Name, "")).Wait(DEFAULT_TIMEOUT)).To(Exit(0))
		Expect(cf.Cf("service-brokers").Wait(DEFAULT_TIMEOUT)).To(Say(b.Name))
	})
}

func (b ServiceBroker) Delete() {
	cf.AsUser(b.context.AdminUserContext(), DEFAULT_TIMEOUT, func() {
		Expect(cf.Cf("delete-service-broker", b.Name, "-f").Wait(DEFAULT_TIMEOUT)).To(Exit(0))

		brokers := cf.Cf("service-brokers").Wait(DEFAULT_TIMEOUT)
		Expect(brokers).To(Exit(0))
		Expect(brokers.Out.Contents()).ToNot(ContainSubstring(b.Name))
	})
}

func (b ServiceBroker) Destroy() {
	cf.AsUser(b.context.AdminUserContext(), DEFAULT_TIMEOUT, func() {
		Expect(cf.Cf("purge-service-offering", b.Service.Name, "-f").Wait(DEFAULT_TIMEOUT)).To(Exit(0))
	})
	b.Delete()
	Expect(cf.Cf("delete", b.Name, "-f", "-r").Wait(DEFAULT_TIMEOUT)).To(Exit(0))
}

func (b ServiceBroker) ToJSON() string {
	bytes, err := ioutil.ReadFile(NewAssets().ServiceBroker + "/cats.json")
	Expect(err).To(BeNil())

	replacer := strings.NewReplacer(
		"<fake-service>", b.Service.Name,
		"<fake-service-guid>", b.Service.ID,
		"<sso-test>", b.Service.DashboardClient.ID,
		"<sso-secret>", b.Service.DashboardClient.Secret,
		"<sso-redirect-uri>", b.Service.DashboardClient.RedirectUri,
		"<fake-plan>", b.SyncPlans[0].Name,
		"<fake-plan-guid>", b.SyncPlans[0].ID,
		"<fake-plan-2>", b.SyncPlans[1].Name,
		"<fake-plan-2-guid>", b.SyncPlans[1].ID,
		"<fake-async-plan>", b.AsyncPlans[0].Name,
		"<fake-async-plan-guid>", b.AsyncPlans[0].ID,
		"<fake-async-plan-2>", b.AsyncPlans[1].Name,
		"<fake-async-plan-2-guid>", b.AsyncPlans[1].ID,
	)

	return replacer.Replace(string(bytes))
}

func SetBackend(appName string) {
	config := helpers.LoadConfig()
	if config.Backend == "diego" {
		EnableDiego(appName)
	} else if config.Backend == "dea" {
		DisableDiego(appName)
	}
}

func EnableDiego(appName string) {
	guid := GetAppGuid(appName)
	Eventually(cf.Cf("curl", "/v2/apps/"+guid, "-X", "PUT", "-d", `{"diego": true}`), DEFAULT_TIMEOUT).Should(Exit(0))
}

func DisableDiego(appName string) {
	guid := GetAppGuid(appName)
	Eventually(cf.Cf("curl", "/v2/apps/"+guid, "-X", "PUT", "-d", `{"diego": false}`), DEFAULT_TIMEOUT).Should(Exit(0))
}

func GetAppGuid(appName string) string {
	cfApp := cf.Cf("app", appName, "--guid")
	Eventually(cfApp, DEFAULT_APP_TIMEOUT).Should(Exit(0))

	appGuid := strings.TrimSpace(string(cfApp.Out.Contents()))
	Expect(appGuid).NotTo(Equal(""))
	return appGuid
}

type Assets struct {
	ServiceBroker string
}

func NewAssets() Assets {
	return Assets{
		ServiceBroker: "../assets/service_broker",
	}
}
