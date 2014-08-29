package feature_flags_test

import (
	CATS_helper "github.com/cloudfoundry/cf-acceptance-tests/helpers"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("CF feature flag commands", func() {
	const (
		assertionTimeout = 10.0
	)

	var (
		context *CATS_helper.ConfiguredContext
		env     *CATS_helper.Environment
		orgName string
	)

	config := CATS_helper.LoadConfig()

	BeforeEach(func() {
		context = CATS_helper.NewContext(config)
		env = CATS_helper.NewEnvironment(context)
		orgName = context.RegularUserContext().Org

		env.Setup()
	})

	AfterEach(func() {
		features := []string{"user_org_creation", "app_scaling", "private_domain_creation", "app_bits_upload", "route_creation"}

		AsUser(context.AdminUserContext(), func() {
			for _, feature := range features {
				featureFlags := Cf("enable-feature-flag", feature).Wait(assertionTimeout)
				Expect(featureFlags).To(Exit(0))
			}
		})

		env.Teardown()
	})

	Describe("feature flags", func() {
		It("can list feature flags and toggle them on and off", func() {
			AsUser(context.AdminUserContext(), func() {
				featureFlags := Cf("feature-flags").Wait(assertionTimeout)
				Expect(featureFlags).To(Exit(0))
				output := featureFlags.Out.Contents()
				Expect(output).To(ContainSubstring("Retrieving status of all flagged features as"))
				Expect(output).To(ContainSubstring("user_org_creation"))

				featureFlags = Cf("enable-feature-flag", "app_scaling").Wait(assertionTimeout)
				Expect(featureFlags).To(Exit(0))
				output = featureFlags.Out.Contents()
				Expect(output).To(ContainSubstring("Feature app_scaling Enabled."))

				featureFlags = Cf("feature-flag", "app_scaling").Wait(assertionTimeout)
				Expect(featureFlags).To(Exit(0))
				output = featureFlags.Out.Contents()
				Expect(output).To(ContainSubstring("Retrieving status of app_scaling as"))
				Expect(output).To(MatchRegexp("app_scaling\\s+enabled"))

				featureFlags = Cf("disable-feature-flag", "app_scaling").Wait(assertionTimeout)
				Expect(featureFlags).To(Exit(0))
				output = featureFlags.Out.Contents()
				Expect(output).To(ContainSubstring("Feature app_scaling Disabled."))
			})
		})
	})
})
