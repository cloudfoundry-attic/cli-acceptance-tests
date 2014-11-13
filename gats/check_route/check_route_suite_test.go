package check_route_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCheckRoute(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CheckRoute Suite")
}
