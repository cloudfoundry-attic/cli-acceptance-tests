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

var _ = Describe("CopySource", func() {
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
	})

	AfterEach(func() {
		env.Teardown()
	})

	It("can copy app bits between multiple apps", func() {
		AsUser(context.RegularUserContext(), 180*time.Second, func() {
			space := context.RegularUserContext().Space
			org := context.RegularUserContext().Org
			username := context.RegularUserContext().Username

			target := Cf("target", "-o", org, "-s", space).Wait(assertionTimeout)
			Expect(target.ExitCode()).To(Equal(0))

			appName1 := generator.RandomName()
			app1 := Cf("push", appName1, "-p", gatsHelpers.NewAssets().ServiceBroker).Wait(appTimeout)
			Expect(app1).To(Exit(0))

			appName2 := generator.RandomName()
			app2 := Cf("push", appName2, "-p", gatsHelpers.NewAssets().ServiceBroker).Wait(appTimeout)
			Expect(app2).To(Exit(0))

			copyBits := Cf("copy-source", appName1, appName2).Wait(copyAppTimeout)
			output := copyBits.Out.Contents()
			Expect(copyBits).To(Exit(0))
			Expect(output).To(ContainSubstring("Copying source from app %s to target app %s in org %s / space %s as %s...", appName1, appName2, org, space, username))
			Expect(output).To(ContainSubstring("Showing health and status for app %s in org %s / space %s as %s...", appName2, org, space, username))
		})
	})
})
