package service_access_test

import (
	"testing"

	"github.com/cloudfoundry/cf-acceptance-tests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var context helpers.SuiteContext

func TestServiceAccess(t *testing.T) {
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

	RunSpecs(t, "Service Access")
}
