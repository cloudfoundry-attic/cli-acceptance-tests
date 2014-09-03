package environment_variable_groups_test

import (
	CATS_helper "github.com/cloudfoundry/cf-acceptance-tests/helpers"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("CF environment variable group commands", func() {
	const (
		assertionTimeout = 10.0
	)

	var (
		context *CATS_helper.ConfiguredContext
		env     *CATS_helper.Environment
	)

	config := CATS_helper.LoadConfig()

	BeforeEach(func() {
		context = CATS_helper.NewContext(config)
		env = CATS_helper.NewEnvironment(context)

		env.Setup()

		AsUser(context.AdminUserContext(), func() {
			envVarGroups := Cf("ssevg", `{}`).Wait(assertionTimeout)
			Expect(envVarGroups).To(Exit(0))
			output := envVarGroups.Out.Contents()
			Expect(output).To(ContainSubstring("Setting the contents of the staging environment variable group"))
			Expect(output).To(ContainSubstring("OK"))

			envVarGroups = Cf("srevg", `{}`).Wait(assertionTimeout)
			Expect(envVarGroups).To(Exit(0))
			output = envVarGroups.Out.Contents()
			Expect(output).To(ContainSubstring("Setting the contents of the running environment variable group"))
			Expect(output).To(ContainSubstring("OK"))
		})
	})

	AfterEach(func() {
		AsUser(context.AdminUserContext(), func() {
			envVarGroups := Cf("ssevg", `{}`).Wait(assertionTimeout)
			Expect(envVarGroups).To(Exit(0))
			output := envVarGroups.Out.Contents()
			Expect(output).To(ContainSubstring("Setting the contents of the staging environment variable group"))
			Expect(output).To(ContainSubstring("OK"))

			envVarGroups = Cf("srevg", `{}`).Wait(assertionTimeout)
			Expect(envVarGroups).To(Exit(0))
			output = envVarGroups.Out.Contents()
			Expect(output).To(ContainSubstring("Setting the contents of the running environment variable group"))
			Expect(output).To(ContainSubstring("OK"))
			env.Teardown()
		})
	})

	Describe("environment variable groups", func() {
		It("can list and set both running and staging environment variable groups", func() {
			AsUser(context.AdminUserContext(), func() {
				envVarGroups := Cf("ssevg", `{"foo":"bar"}`).Wait(assertionTimeout)
				Expect(envVarGroups).To(Exit(0))
				output := envVarGroups.Out.Contents()
				Expect(output).To(ContainSubstring("Setting the contents of the staging environment variable group"))
				Expect(output).To(ContainSubstring("OK"))

				envVarGroups = Cf("sevg").Wait(assertionTimeout)
				Expect(envVarGroups).To(Exit(0))
				output = envVarGroups.Out.Contents()
				Expect(output).To(ContainSubstring("foo"))
				Expect(output).To(ContainSubstring("bar"))

				envVarGroups = Cf("set-staging-environment-variable-group", `{"num": 123}`).Wait(assertionTimeout)
				Expect(envVarGroups).To(Exit(0))
				output = envVarGroups.Out.Contents()
				Expect(output).To(ContainSubstring("Setting the contents of the staging environment variable group"))
				Expect(output).To(ContainSubstring("OK"))

				envVarGroups = Cf("staging-environment-variable-group").Wait(assertionTimeout)
				Expect(envVarGroups).To(Exit(0))
				output = envVarGroups.Out.Contents()
				Expect(output).To(ContainSubstring("num"))
				Expect(output).To(ContainSubstring("123"))
				Expect(output).ToNot(ContainSubstring("foo"))
				Expect(output).ToNot(ContainSubstring("bar"))

				envVarGroups = Cf("srevg", `{"foo":"bar"}`).Wait(assertionTimeout)
				Expect(envVarGroups).To(Exit(0))
				output = envVarGroups.Out.Contents()
				Expect(output).To(ContainSubstring("Setting the contents of the running environment variable group"))
				Expect(output).To(ContainSubstring("OK"))

				envVarGroups = Cf("revg").Wait(assertionTimeout)
				Expect(envVarGroups).To(Exit(0))
				output = envVarGroups.Out.Contents()
				Expect(output).To(ContainSubstring("foo"))
				Expect(output).To(ContainSubstring("bar"))

				envVarGroups = Cf("set-running-environment-variable-group", `{"num": 123}`).Wait(assertionTimeout)
				Expect(envVarGroups).To(Exit(0))
				output = envVarGroups.Out.Contents()
				Expect(output).To(ContainSubstring("Setting the contents of the running environment variable group"))
				Expect(output).To(ContainSubstring("OK"))

				envVarGroups = Cf("running-environment-variable-group").Wait(assertionTimeout)
				Expect(envVarGroups).To(Exit(0))
				output = envVarGroups.Out.Contents()
				Expect(output).To(ContainSubstring("num"))
				Expect(output).To(ContainSubstring("123"))
				Expect(output).ToNot(ContainSubstring("foo"))
				Expect(output).ToNot(ContainSubstring("bar"))
			})
		})
	})
})
