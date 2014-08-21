package service_access_test

import (
	"fmt"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	CATS_helper "github.com/cloudfoundry/cf-acceptance-tests/helpers"
	broker_helper "github.com/cloudfoundry/cf-acceptance-tests/services/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/GATS/helpers"
)

var assertionTimeout = 10.0

var _ = Describe("CF service access commands", func() {
	// Create service broker, service, service plans, and service plan visibilities.

	var broker broker_helper.ServiceBroker
	config := CATS_helper.LoadConfig()
	context := CATS_helper.NewContext(config)
	env := CATS_helper.NewEnvironment(context)
	orgName := context.RegularUserContext().Org
	BeforeEach(func() {
		env.Setup()
		broker = broker_helper.NewServiceBroker(generator.RandomName(), helpers.NewAssets().ServiceBroker, context)
		broker.Push()
		broker.Configure()
		broker.Create()
	})

	AfterEach(func() {
		broker.Destroy()
		env.Teardown()
	})

	It("provides a reasonable workflow for seeing visibilities", func() {
		AsUser(context.AdminUserContext(), func() {
			access := Cf("service-access").Wait(assertionTimeout)
			Expect(access).To(Exit(0))
			output := access.Out.Contents()
			Expect(output).To(ContainSubstring(fmt.Sprintf("broker: %s", broker.Name)))
			Expect(output).To(ContainSubstring(broker.Service.Name))
			Expect(output).To(ContainSubstring(broker.Plan.Name))
			Expect(output).To(ContainSubstring("none"))

			access = Cf("enable-service-access", broker.Service.Name, "-p", broker.Plan.Name, "-o", orgName).Wait(assertionTimeout)
			Expect(access).To(Exit(0))
			Expect(access.Out.Contents()).To(ContainSubstring("OK"))

			access = Cf("service-access", "-o", orgName).Wait(assertionTimeout)
			Expect(access).To(Exit(0))
			output = access.Out.Contents()
			Expect(output).To(ContainSubstring(fmt.Sprintf("broker: %s", broker.Name)))
			Expect(output).To(ContainSubstring(broker.Service.Name))
			Expect(output).To(ContainSubstring(broker.Plan.Name))
			Expect(output).To(ContainSubstring("limited"))
			Expect(output).To(ContainSubstring(orgName))

			access = Cf("enable-service-access", broker.Service.Name).Wait(assertionTimeout)
			Expect(access).To(Exit(0))
			Expect(access.Out.Contents()).To(ContainSubstring("OK"))

			access = Cf("service-access", "-e", broker.Service.Name).Wait(assertionTimeout)
			Expect(access).To(Exit(0))
			output = access.Out.Contents()
			Expect(output).To(ContainSubstring(fmt.Sprintf("broker: %s", broker.Name)))
			Expect(output).To(ContainSubstring(broker.Service.Name))
			Expect(output).To(ContainSubstring(broker.Plan.Name))
			Expect(output).To(ContainSubstring("all"))

			access = Cf("disable-service-access", broker.Service.Name, "-p", broker.Plan.Name).Wait(assertionTimeout)
			Expect(access).To(Exit(0))
			Expect(access.Out.Contents()).To(ContainSubstring("OK"))

			access = Cf("service-access", "-b", broker.Name).Wait(assertionTimeout)
			Expect(access).To(Exit(0))
			output = access.Out.Contents()
			Expect(output).To(ContainSubstring(fmt.Sprintf("broker: %s", broker.Name)))
			Expect(output).To(ContainSubstring(broker.Service.Name))
			Expect(output).To(ContainSubstring(broker.Plan.Name))
			Expect(output).To(ContainSubstring("none"))
		})
	})
})
