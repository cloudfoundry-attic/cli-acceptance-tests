package plugin_test

import (
	"runtime"
	"time"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"

	acceptanceTestHelpers "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var (
	context *acceptanceTestHelpers.ConfiguredContext
	env     *acceptanceTestHelpers.Environment
	config  acceptanceTestHelpers.Config
)

var _ = BeforeSuite(func() {
	config = acceptanceTestHelpers.LoadConfig()
	context = acceptanceTestHelpers.NewContext(config)
	env = acceptanceTestHelpers.NewEnvironment(context)

	env.Setup()

	var install *Session
	switch runtime.GOOS {
	case "darwin":
		install = Cf("install-plugin", "fixtures/plugin_api.osx").Wait(5 * time.Second)
	}
	Expect(install).To(Exit(0))
})

var _ = AfterSuite(func() {
	env.Teardown()

	uninstall := Cf("uninstall-plugin", "GatsPlugin").Wait(5 * time.Second)
	Expect(uninstall).To(Exit(0))
})

var _ = Describe("Plugin API", func() {
	const (
		apiTimeout       = 10 * time.Second
		operationTimeout = 20 * time.Second
	)

	Describe("GetCurrentOrg()", func() {
		It("gets the current targeted org", func() {
			apiResult := Cf("GetCurrentOrg").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult.Out.Contents()).To(ContainSubstring("CATS-ORG-"))
		})
	})

	Describe("GetCurrentSpace()", func() {
		It("gets the current targeted space", func() {
			AsUser(context.AdminUserContext(), 10*time.Second, func() {
				var cmd *Session

				org := generator.RandomName()
				space := generator.RandomName()

				cmd = Cf("create-org", org).Wait(operationTimeout)
				Expect(cmd).To(Exit(0))

				cmd = Cf("target", "-o", org).Wait(operationTimeout)
				Expect(cmd).To(Exit(0))

				cmd = Cf("create-space", space).Wait(operationTimeout)
				Expect(cmd).To(Exit(0))

				cmd = Cf("target", "-s", space).Wait(operationTimeout)
				Expect(cmd).To(Exit(0))

				apiResult := Cf("GetCurrentSpace").Wait(apiTimeout)
				Expect(apiResult).To(Exit(0))
				Expect(apiResult.Out.Contents()).To(ContainSubstring(space))

				cmd = Cf("delete-space", space, "-f").Wait(operationTimeout)
				Expect(cmd).To(Exit(0))

				cmd = Cf("delete-org", org, "-f").Wait(operationTimeout)
				Expect(cmd).To(Exit(0))
			})
		})
	})

	Describe("Username()", func() {
		It("gets the current Username", func() {
			apiResult := Cf("Username").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult.Out.Contents()).To(ContainSubstring("CATS-USER-"))
		})
	})

	Describe("UserGuid()", func() {
		It("gets the current UserGuid", func() {
			apiResult := Cf("UserGuid").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(len(apiResult.Out.Contents())).Should(BeNumerically(">", 40))
		})
	})

	Describe("UserEmail()", func() {
		It("gets the current UserEmail", func() {
			apiResult := Cf("UserEmail").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult.Out.Contents()).To(ContainSubstring("CATS-USER-"))
		})
	})

	Describe("IsLoggedIn()", func() {
		It("gets the current IsLoggedIn", func() {
			apiResult := Cf("IsLoggedIn").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult.Out.Contents()).To(ContainSubstring("true"))
		})
	})

	Describe("IsSSLDisabled()", func() {
		It("gets the current IsSSLDisabled", func() {
			apiResult := Cf("IsSSLDisabled").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult.Out.Contents()).ToNot(ContainSubstring("Error"))
		})
	})

	Describe("ApiEndpoint()", func() {
		It("gets the current ApiEndpoint", func() {
			apiResult := Cf("ApiEndpoint").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(len(apiResult.Out.Contents())).Should(BeNumerically(">", 25))
		})
	})

	Describe("ApiVersion()", func() {
		It("gets the current ApiVersion", func() {
			apiResult := Cf("ApiVersion").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(len(apiResult.Out.Contents())).Should(BeNumerically(">", 21))
		})
	})

	Describe("HasAPIEndpoint()", func() {
		It("gets HasAPIEndpoint", func() {
			apiResult := Cf("HasAPIEndpoint").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult.Out.Contents()).ToNot(ContainSubstring("Error"))
		})
	})

	Describe("HasOrganization()", func() {
		It("gets HasOrganization", func() {
			apiResult := Cf("HasOrganization").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult.Out.Contents()).To(ContainSubstring("true"))
		})
	})

	Describe("HasSpace()", func() {
		It("gets HasSpace", func() {
			apiResult := Cf("HasSpace").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult.Out.Contents()).To(ContainSubstring("true"))
		})
	})

	Describe("LoggregatorEndpoint()", func() {
		It("gets LoggregatorEndpoint", func() {
			apiResult := Cf("LoggregatorEndpoint").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult.Out.Contents()).To(ContainSubstring("wss://loggregator"))
		})
	})

	Describe("DopplerEndpoint()", func() {
		It("gets DopplerEndpoint", func() {
			apiResult := Cf("DopplerEndpoint").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult.Out.Contents()).To(ContainSubstring("wss://doppler"))
		})
	})

	Describe("AccessToken()", func() {
		It("gets AccessToken", func() {
			apiResult := Cf("AccessToken").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult.Out.Contents()).To(ContainSubstring("bearer"))
		})
	})

	Describe("GetOrg()", func() {
		It("gets the detail of a org", func() {
			org := generator.RandomName()

			AsUser(context.AdminUserContext(), 10*time.Second, func() {
				co := Cf("create-org", org).Wait(operationTimeout)
				Expect(co).To(Exit(0))

				apiResult := Cf("GetOrg", org).Wait(apiTimeout)
				Expect(apiResult).To(Exit(0))
				Expect(apiResult.Out.Contents()).To(ContainSubstring(org))

				do := Cf("delete-org", org, "-f").Wait(operationTimeout)
				Expect(do).To(Exit(0))
			})
		})
	})

	Describe("GetOrgs()", func() {
		It("gets a list of orgs", func() {
			apiResult := Cf("GetOrgs").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult.Out.Contents()).To(ContainSubstring("CATS-ORG-"))
		})
	})
})
