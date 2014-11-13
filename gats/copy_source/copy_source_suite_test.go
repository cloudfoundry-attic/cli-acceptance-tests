package copy_source_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCopySource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CopySource Suite")
}
