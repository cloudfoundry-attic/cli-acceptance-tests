package integration

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Without Command", func() {
	Context("when cf is run without providing a command", func() {
		var (
			commandErr error
			session    *Session
		)

		BeforeEach(func() {
			command := exec.Command("cf")
			session, commandErr = Start(command, GinkgoWriter, GinkgoWriter)
		})

		It("displays the help text", func() {
			Expect(commandErr).NotTo(HaveOccurred())
			Eventually(session).Should(Exit(0))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("VERSION:"))
			Eventually(session).Should(Say("GETTING STARTED:"))
			Eventually(session).Should(Say("ENVIRONMENT VARIABLES:"))
			Eventually(session).Should(Say("GLOBAL OPTIONS:"))
		})
	})
})
