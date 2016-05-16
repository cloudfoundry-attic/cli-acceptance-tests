package application_test

import (
	"fmt"
	"time"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	acceptanceTestHelpers "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	gatsHelpers "github.com/cloudfoundry/cli-acceptance-tests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Env", func() {
	const (
		assertionTimeout = 10 * time.Second
		appTimeout       = 2 * time.Minute
		copyAppTimeout   = 3 * time.Minute
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
			envVarGroups := Cf("set-staging-environment-variable-group", `{}`).Wait(assertionTimeout)
			Expect(envVarGroups).To(Exit(0))

			envVarGroups = Cf("set-running-environment-variable-group", `{}`).Wait(assertionTimeout)
			Expect(envVarGroups).To(Exit(0))
		})
	})

	AfterEach(func() {
		AsUser(context.AdminUserContext(), 30*time.Second, func() {
			envVarGroups := Cf("set-staging-environment-variable-group", `{}`).Wait(assertionTimeout)
			Expect(envVarGroups).To(Exit(0))

			envVarGroups = Cf("set-running-environment-variable-group", `{}`).Wait(assertionTimeout)
			Expect(envVarGroups).To(Exit(0))
			env.Teardown()
		})
	})

	It("returns ann applications running, staging, system provided and user defined environment variables", func() {
		AsUser(context.AdminUserContext(), 60*time.Second, func() {
			stagingVal := fmt.Sprintf("staging-val-%d", time.Now().Nanosecond())
			runningVal := fmt.Sprintf("running-val-%d", time.Now().Nanosecond())
			setEnvVal := fmt.Sprintf("set-env-val-%d", time.Now().Nanosecond())

			ssevgResult := Cf("set-staging-environment-variable-group", fmt.Sprintf(`{"name":"%s"}`, stagingVal)).Wait(assertionTimeout)
			Expect(ssevgResult).To(Exit(0))
			srevgResult := Cf("set-running-environment-variable-group", fmt.Sprintf(`{"name":"%s"}`, runningVal)).Wait(assertionTimeout)
			Expect(srevgResult).To(Exit(0))

			space := context.RegularUserContext().Space
			org := context.RegularUserContext().Org

			target := Cf("target", "-o", org, "-s", space).Wait(assertionTimeout)
			Expect(target.ExitCode()).To(Equal(0))

			appName := generator.RandomName()
			app := Cf("push", appName, "-p", gatsHelpers.NewAssets().ServiceBroker).Wait(appTimeout)
			Expect(app).To(Exit(0))

			setEnvResult := Cf("set-env", appName, "set-env-key", setEnvVal).Wait(assertionTimeout)
			Expect(setEnvResult).To(Exit(0))

			envResult := Cf("env", appName).Wait(assertionTimeout)
			Expect(envResult).To(Exit(0))

			output := envResult.Out.Contents()
			Expect(output).To(ContainSubstring("\"VCAP_APPLICATION\": {"))
			Expect(output).To(ContainSubstring("User-Provided:\nset-env-key: %s", setEnvVal))
			Expect(output).To(ContainSubstring("Running Environment Variable Groups:\nname: %s", runningVal))
			Expect(output).To(ContainSubstring("Staging Environment Variable Groups:\nname: %s", stagingVal))
		})
	})
})
