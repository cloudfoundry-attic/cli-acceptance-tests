package service_access_test

import (
	"fmt"
	"time"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"code.cloudfoundry.org/cli-acceptance-tests/gats/service_access/servicebroker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var (
	assertionTimeout         = 1 * time.Minute
	extendedAssertionTimeout = 2 * time.Minute
)

var _ = Describe("CF service access commands", func() {
	// Create service broker, service, service plans, and service plan visibilities.

	var broker servicebroker.ServiceBroker
	config := helpers.LoadConfig()
	context := helpers.NewContext(config)
	env := helpers.NewEnvironment(context)
	orgName := context.RegularUserContext().Org
	BeforeEach(func() {
		env.Setup()
		broker = servicebroker.NewServiceBroker(generator.RandomName(), servicebroker.NewAssets().ServiceBroker, context)
		broker.Push()
		broker.Configure()

		AsUser(context.AdminUserContext(), context.ShortTimeout(), func() {
			broker.Create()
		})
	})

	AfterEach(func() {
		broker.Destroy()
		env.Teardown()
	})

	It("provides a reasonable workflow for seeing visibilities", func() {
		AsUser(context.AdminUserContext(), 100*time.Second, func() {
			access := Cf("service-access").Wait(assertionTimeout)
			Expect(access).To(Exit(0))
			output := access.Out.Contents()
			Expect(output).To(ContainSubstring(fmt.Sprintf("broker: %s", broker.Name)))
			Expect(output).To(ContainSubstring(broker.Service.Name))
			Expect(output).To(ContainSubstring(broker.SyncPlans[0].Name))
			Expect(output).To(ContainSubstring("none"))

			access = Cf("enable-service-access", broker.Service.Name, "-p", broker.SyncPlans[0].Name, "-o", orgName).Wait(extendedAssertionTimeout)
			Expect(access).To(Exit(0))
			Expect(access.Out.Contents()).To(ContainSubstring("OK"))

			access = Cf("service-access", "-o", orgName).Wait(assertionTimeout)
			Expect(access).To(Exit(0))
			output = access.Out.Contents()
			Expect(output).To(ContainSubstring(fmt.Sprintf("broker: %s", broker.Name)))
			Expect(output).To(ContainSubstring(broker.Service.Name))
			Expect(output).To(ContainSubstring(broker.SyncPlans[0].Name))
			Expect(output).To(ContainSubstring("limited"))
			Expect(output).To(ContainSubstring(orgName))

			access = Cf("enable-service-access", broker.Service.Name).Wait(extendedAssertionTimeout)
			Expect(access).To(Exit(0))
			Expect(access.Out.Contents()).To(ContainSubstring("OK"))

			access = Cf("service-access", "-e", broker.Service.Name).Wait(assertionTimeout)
			Expect(access).To(Exit(0))
			output = access.Out.Contents()
			Expect(output).To(ContainSubstring(fmt.Sprintf("broker: %s", broker.Name)))
			Expect(output).To(ContainSubstring(broker.Service.Name))
			Expect(output).To(ContainSubstring(broker.SyncPlans[0].Name))
			Expect(output).To(ContainSubstring("all"))

			access = Cf("disable-service-access", broker.Service.Name, "-p", broker.SyncPlans[0].Name).Wait(extendedAssertionTimeout)
			Expect(access).To(Exit(0))
			Expect(access.Out.Contents()).To(ContainSubstring("OK"))

			access = Cf("service-access", "-b", broker.Name).Wait(assertionTimeout)
			Expect(access).To(Exit(0))
			output = access.Out.Contents()
			Expect(output).To(ContainSubstring(fmt.Sprintf("broker: %s", broker.Name)))
			Expect(output).To(ContainSubstring(broker.Service.Name))
			Expect(output).To(ContainSubstring(broker.SyncPlans[0].Name))
			Expect(output).To(ContainSubstring("none"))
		})
	})
})
