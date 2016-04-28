package plugin_test

import (
	"runtime"
	"time"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"

	acceptanceTestHelpers "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	gatsHelpers "github.com/cloudfoundry/cli-acceptance-tests/helpers"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
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

	Expect(runtime.GOARCH).To(Equal("amd64"), "Plugin suite only runs under 64bit OS, please skip the plugin suite in 32bit OS (use flag -skipPackage='gats/plugin')")

	var install *Session
	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH == "amd64" {
			install = Cf("install-plugin", "-f", "fixtures/plugin_windows_amd64.exe").Wait(5 * time.Second)
		} else {
			install = Cf("install-plugin", "-f", "fixtures/plugin_windows_386.exe").Wait(5 * time.Second)
		}
	case "linux":
		if runtime.GOARCH == "amd64" {
			install = Cf("install-plugin", "-f", "fixtures/plugin_linux_amd64").Wait(5 * time.Second)
		} else {
			install = Cf("install-plugin", "-f", "fixtures/plugin_linux_386").Wait(5 * time.Second)
		}
	case "darwin":
		if runtime.GOARCH == "amd64" {
			install = Cf("install-plugin", "-f", "fixtures/plugin_darwin_amd64").Wait(5 * time.Second)
		} else {
			install = Cf("install-plugin", "-f", "fixtures/plugin_darwin_386").Wait(5 * time.Second)
		}
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
		apiTimeout       = 20 * time.Second
		appTimeout       = 5 * time.Minute
		assertionTimeout = 20 * time.Second
		operationTimeout = 20 * time.Second
	)

	Describe("CliCommand()", func() {
		It("calls the core cli command and output to terminal", func() {
			apiResult := Cf("CliCommand", "target").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult).Should(gbytes.Say("API endpoint"))
			Expect(apiResult).Should(gbytes.Say("API endpoint"))
		})
	})

	Describe("CliCommandWithoutTerminalOutput()", func() {
		It("calls the core cli command and without outputing to terminal", func() {
			apiResult := Cf("CliCommandWithoutTerminalOutput", "target").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult).Should(gbytes.Say("API endpoint"))
			Expect(apiResult).ShouldNot(gbytes.Say("API endpoint"))
		})
	})

	Describe("GetCurrentOrg()", func() {
		It("gets the current targeted org", func() {
			apiResult := Cf("GetCurrentOrg").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult.Out.Contents()).To(ContainSubstring("CATS-ORG-"))
		})
	})

	// Describe("GetCurrentSpace()", func() {
	// 	It("gets the current targeted space", func() {
	// 		AsUser(context.AdminUserContext(), 150*time.Second, func() {
	// 			var cmd *Session

	// 			org := generator.RandomName()
	// 			space := generator.RandomName()

	// 			cmd = Cf("create-org", org).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("target", "-o", org).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("create-space", space).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("target", "-s", space).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			apiResult := Cf("GetCurrentSpace").Wait(apiTimeout)
	// 			Expect(apiResult).To(Exit(0))
	// 			Expect(apiResult.Out.Contents()).To(ContainSubstring(space))

	// 			cmd = Cf("delete-space", space, "-f").Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("delete-org", org, "-f").Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))
	// 		})
	// 	})
	// })

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

	Describe("GetApp() and GetApps()", func() {
		It("gets app details and app list", func() {
			AsUser(context.RegularUserContext(), 250*time.Second, func() {
				space := context.RegularUserContext().Space
				org := context.RegularUserContext().Org

				target := Cf("target", "-o", org, "-s", space).Wait(assertionTimeout)
				Expect(target.ExitCode()).To(Equal(0))

				appName1 := generator.RandomName()
				app1 := Cf("push", appName1, "-p", gatsHelpers.NewAssets().ServiceBroker).Wait(appTimeout)
				Expect(app1).To(Exit(0))

				appName2 := generator.RandomName()
				app2 := Cf("push", appName2, "-p", gatsHelpers.NewAssets().ServiceBroker).Wait(appTimeout)
				Expect(app2).To(Exit(0))

				apiResult := Cf("GetApp", appName1).Wait(apiTimeout)
				Expect(apiResult).To(Exit(0))
				Expect(apiResult.Out.Contents()).To(ContainSubstring(appName1))

				apiResult = Cf("GetApps", appName1).Wait(apiTimeout)
				Expect(apiResult).To(Exit(0))
				Expect(apiResult.Out.Contents()).To(ContainSubstring(appName1))
				Expect(apiResult.Out.Contents()).To(ContainSubstring(appName2))

				app1 = Cf("delete", appName1, "-f").Wait(appTimeout)
				Expect(app1).To(Exit(0))

				app2 = Cf("delete", appName2, "-f").Wait(appTimeout)
				Expect(app2).To(Exit(0))
			})
		})
	})

	// Describe("GetOrg()", func() {
	// 	It("gets the detail of a org", func() {
	// 		org := generator.RandomName()

	// 		AsUser(context.AdminUserContext(), 50*time.Second, func() {
	// 			co := Cf("create-org", org).Wait(operationTimeout)
	// 			Expect(co).To(Exit(0))

	// 			apiResult := Cf("GetOrg", org).Wait(apiTimeout)
	// 			Expect(apiResult).To(Exit(0))
	// 			Expect(apiResult.Out.Contents()).To(ContainSubstring(org))

	// 			do := Cf("delete-org", org, "-f").Wait(operationTimeout)
	// 			Expect(do).To(Exit(0))
	// 		})
	// 	})
	// })

	Describe("GetOrgs()", func() {
		It("gets a list of orgs", func() {
			apiResult := Cf("GetOrgs").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult.Out.Contents()).To(ContainSubstring("CATS-ORG-"))
		})
	})

	// Describe("GetSpace()", func() {
	// 	It("gets the detail of a space", func() {
	// 		var cmd *Session

	// 		org := generator.RandomName()
	// 		space := generator.RandomName()

	// 		AsUser(context.AdminUserContext(), 120*time.Second, func() {
	// 			cmd = Cf("create-org", org).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("target", "-o", org).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("create-space", space).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			apiResult := Cf("GetSpace", space).Wait(apiTimeout)
	// 			Expect(apiResult).To(Exit(0))
	// 			Expect(apiResult.Out.Contents()).To(ContainSubstring(space))

	// 			cmd = Cf("delete-space", space, "-f").Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("delete-org", org, "-f").Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))
	// 		})
	// 	})
	// })

	// Describe("GetOrgUsers()", func() {
	// 	It("gets a list of users in the org", func() {
	// 		var cmd *Session

	// 		org := generator.RandomName()
	// 		user := generator.RandomName()

	// 		AsUser(context.AdminUserContext(), 120*time.Second, func() {
	// 			cmd = Cf("create-org", org).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("target", "-o", org).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("create-user", user, "password").Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("set-org-role", user, org, "OrgManager").Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			apiResult := Cf("GetOrgUsers", org, "-a").Wait(apiTimeout)
	// 			Expect(apiResult).To(Exit(0))
	// 			Expect(apiResult.Out.Contents()).To(ContainSubstring(user))

	// 			cmd = Cf("delete-org", org, "-f").Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))
	// 		})
	// 	})
	// })

	// Describe("GetSpaceUsers()", func() {
	// 	It("gets a list of users in the space", func() {
	// 		var cmd *Session

	// 		org := generator.RandomName()
	// 		space := generator.RandomName()
	// 		user := generator.RandomName()

	// 		AsUser(context.AdminUserContext(), 150*time.Second, func() {
	// 			cmd = Cf("create-org", org).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("create-space", space, "-o", org).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("target", "-o", org, "-s", space).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("create-user", user, "password").Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("set-org-role", user, org, "OrgManager").Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("set-space-role", user, org, space, "SpaceManager").Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			apiResult := Cf("GetSpaceUsers", org, space).Wait(apiTimeout)
	// 			Expect(apiResult).To(Exit(0))
	// 			Expect(apiResult.Out.Contents()).To(ContainSubstring(user))

	// 			cmd = Cf("delete-org", org, "-f").Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))
	// 		})
	// 	})
	// })

	Describe("GetSpaces()", func() {
		It("gets a list of spaces", func() {
			apiResult := Cf("GetSpaces").Wait(apiTimeout)
			Expect(apiResult).To(Exit(0))
			Expect(apiResult.Out.Contents()).To(ContainSubstring("CATS-SPACE-"))
		})
	})

	// Describe("GetServices()", func() {
	// 	It("gets a list of available services", func() {
	// 		var cmd *Session

	// 		service := generator.RandomName()
	// 		org := generator.RandomName()
	// 		space := generator.RandomName()

	// 		AsUser(context.AdminUserContext(), 120*time.Second, func() {
	// 			cmd = Cf("create-org", org).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("create-space", space, "-o", org).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("target", "-o", org, "-s", space).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("cups", service).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			apiResult := Cf("GetServices").Wait(apiTimeout)
	// 			Expect(apiResult).To(Exit(0))
	// 			Expect(apiResult.Out.Contents()).To(ContainSubstring(service))

	// 			do := Cf("delete-service", service, "-f").Wait(operationTimeout)
	// 			Expect(do).To(Exit(0))
	// 		})
	// 	})
	// })

	// Describe("GetService()", func() {
	// 	It("gets the details of a service", func() {
	// 		var cmd *Session

	// 		service := generator.RandomName()
	// 		org := generator.RandomName()
	// 		space := generator.RandomName()

	// 		AsUser(context.AdminUserContext(), 120*time.Second, func() {
	// 			cmd = Cf("create-org", org).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("create-space", space, "-o", org).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("target", "-o", org, "-s", space).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			cmd = Cf("cups", service).Wait(operationTimeout)
	// 			Expect(cmd).To(Exit(0))

	// 			apiResult := Cf("GetService", service).Wait(apiTimeout)
	// 			Expect(apiResult).To(Exit(0))
	// 			Expect(apiResult.Out.Contents()).To(ContainSubstring(service))

	// 			do := Cf("delete-service", service, "-f").Wait(operationTimeout)
	// 			Expect(do).To(Exit(0))
	// 		})
	// 	})
	// })

})
