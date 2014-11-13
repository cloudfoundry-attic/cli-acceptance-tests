package check_route_test

import (
	"fmt"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("CheckRoute", func() {

	var (
		context  *helpers.ConfiguredContext
		hostName string
	)

	config := helpers.LoadConfig()

	BeforeEach(func() {
		context = helpers.NewContext(config)

		uuidBytes, err := uuid.NewV4()
		Expect(err).ToNot(HaveOccurred())
		hostName = uuidBytes.String()
	})

	AfterEach(func() {
		Cf("delete-route", config.AppsDomain, "-n", hostName)
	})

	It("can check if a route exists", func() {
		fmt.Println("domain:", config.AppsDomain)
		AsUser(context.AdminUserContext(), func() {

			createRoute := Cf("create-route", config.PersistentAppSpace, config.AppsDomain, "-n", hostName)
			Expect(createRoute.ExitCode()).To(Equal(0))
		})
	})
})
