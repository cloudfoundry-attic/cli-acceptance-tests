package security_groups_test

import (
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var context helpers.SuiteContext

func TestSecurityGroups(t *testing.T) {
	RegisterFailHandler(Fail)

	config := helpers.LoadConfig()
	context = helpers.NewContext(config)
	environment := helpers.NewEnvironment(context)

	BeforeSuite(func() {
		environment.Setup()
	})

	AfterSuite(func() {
		environment.Teardown()
	})

	RunSpecs(t, "Security Groups")
}
