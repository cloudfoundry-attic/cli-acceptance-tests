package check_route_test

import (
	"fmt"
	"time"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CheckRoute", func() {
	const (
		assertionTimeout = 10.0
	)

	var (
		context  *helpers.ConfiguredContext
		hostName string

		env *helpers.Environment
	)

	config := helpers.LoadConfig()

	BeforeEach(func() {
		hostName = generator.RandomName()

		context = helpers.NewContext(config)
		env = helpers.NewEnvironment(context)

		env.Setup()
	})

	AfterEach(func() {
		Cf("delete-route", config.AppsDomain, "-n", hostName)

		env.Teardown()
	})

	It("can check if a route exists", func() {
		AsUser(context.AdminUserContext(), 60*time.Second, func() {
			space := context.RegularUserContext().Space

			target := Cf("target", "-o", context.RegularUserContext().Org, "-s", space).Wait(assertionTimeout)
			Expect(target.ExitCode()).To(Equal(0))

			createRoute := Cf("create-route", space, config.AppsDomain, "-n", hostName).Wait(assertionTimeout)
			Expect(createRoute.ExitCode()).To(Equal(0))

			checkRoute := Cf("check-route", hostName, config.AppsDomain).Wait(assertionTimeout)
			Expect(checkRoute.Out.Contents()).To(ContainSubstring(fmt.Sprintf("Route %s.%s does exist", hostName, config.AppsDomain)))

			deleteRoute := Cf("delete-route", config.AppsDomain, "-n", hostName, "-f").Wait(assertionTimeout + 30)
			Expect(deleteRoute.ExitCode()).To(Equal(0))

			checkRoute = Cf("check-route", hostName, config.AppsDomain).Wait(assertionTimeout)
			Expect(checkRoute.Out.Contents()).To(ContainSubstring(fmt.Sprintf("Route %s.%s does not exist", hostName, config.AppsDomain)))
		})
	})
})
