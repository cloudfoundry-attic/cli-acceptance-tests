package translations_test

import (
	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("i18n support", func() {
	It("returns the french translation for cf quota", func() {
		session := Cf("config", "--locale", "fr-FR")
		Eventually(session).Should(gexec.Exit(0))
		Eventually(Cf("help", "quota")).Should(Say("Afficher les informations de quota"))
	})
})
