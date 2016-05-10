package service_access_test

import (
	"testing"

	. "github.com/cloudfoundry/cf-acceptance-tests/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/cf-acceptance-tests/Godeps/_workspace/src/github.com/onsi/gomega"
)

func TestServiceAccess(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Service Access")
}
