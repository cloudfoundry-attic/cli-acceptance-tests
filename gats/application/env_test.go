package application_test

import (
	"time"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	acceptanceTestHelpers "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	gatsHelpers "github.com/cloudfoundry/GATS/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Env", func() {
	const (
		assertionTimeout = 10 * time.Second
		appTimeout       = 1 * time.Minute
		copyAppTimeout   = appTimeout * 3
	)

	var (
		context *acceptanceTestHelpers.ConfiguredContext
		env     *acceptanceTestHelpers.Environment
	)

	config := acceptanceTestHelpers.LoadConfig()

	BeforeEach(func() {
		context = acceptanceTestHelpers.NewContext(config)
		env = acceptanceTestHelpers.NewEnvironment(context)

		env.Setup()
		AsUser(context.AdminUserContext(), 30*time.Second, func() {
			envVarGroups := Cf("ssevg", `{}`).Wait(assertionTimeout)
			Expect(envVarGroups).To(Exit(0))

			envVarGroups = Cf("srevg", `{}`).Wait(assertionTimeout)
			Expect(envVarGroups).To(Exit(0))
		})
	})

	AfterEach(func() {
		AsUser(context.AdminUserContext(), 30*time.Second, func() {
			envVarGroups := Cf("ssevg", `{}`).Wait(assertionTimeout)
			Expect(envVarGroups).To(Exit(0))

			envVarGroups = Cf("srevg", `{}`).Wait(assertionTimeout)
			Expect(envVarGroups).To(Exit(0))
			env.Teardown()
		})
	})

	It("returns ann applications running, staging, system provided and user defined environment variables", func() {
		AsUser(context.AdminUserContext(), 60*time.Second, func() {
			ssevgResult := Cf("ssevg", `{"name":"staging-val"}`).Wait(assertionTimeout)
			Expect(ssevgResult).To(Exit(0))
			srevgResult := Cf("srevg", `{"name":"running-val"}`).Wait(assertionTimeout)
			Expect(srevgResult).To(Exit(0))

			space := context.RegularUserContext().Space
			org := context.RegularUserContext().Org

			target := Cf("target", "-o", org, "-s", space).Wait(assertionTimeout)
			Expect(target.ExitCode()).To(Equal(0))

			appName := generator.RandomName()
			app := Cf("push", appName, "-p", gatsHelpers.NewAssets().ServiceBroker).Wait(appTimeout)
			Expect(app).To(Exit(0))

			setEnvResult := Cf("set-env", appName, "set-env-key", "set-env-val").Wait(assertionTimeout)
			Expect(setEnvResult).To(Exit(0))

			envResult := Cf("env", appName).Wait(assertionTimeout)
			Expect(envResult).To(Exit(0))

			output := envResult.Out.Contents()
			Expect(output).To(ContainSubstring("\"VCAP_APPLICATION\": {"))
			Expect(output).To(ContainSubstring("User-Provided:\nset-env-key: set-env-val"))
			Expect(output).To(ContainSubstring("Running Environment Variable Groups:\nname: running-val"))
			Expect(output).To(ContainSubstring("Staging Environment Variable Groups:\nname: staging-val"))
		})
	})
})
