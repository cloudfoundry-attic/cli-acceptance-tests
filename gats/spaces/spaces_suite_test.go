package quotas_test

import (
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var context helpers.SuiteContext

var (
	assertionTimeout      = 10.0
	asyncAssertionTimeout = 15.0
)

func generateUniqueName() string {
	uuidBytes, err := uuid.NewV4()
	Expect(err).ToNot(HaveOccurred())
	return uuidBytes.String()
}

func TestSpaces(t *testing.T) {
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

	RunSpecs(t, "Spaces")
}
