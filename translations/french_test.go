package translations_test

import (
	"github.com/XenoPhex/jibber_jabber"

	//	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("i18n support and language detection", func() {
	BeforeEach(func() {
		userLocale, err := jibber_jabber.DetectIETF()
		Expect(err).NotTo(HaveOccurred())
		Expect(userLocale).To(Equal("fr-FR"), "This test can only be run when the system's language is set to french")
	})

	It("does nothing yet", func() {

	})
})
