package integration

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Verbose", func() {
	Context("when the -v option is provided with additional command", func() {
		var (
			commandErr error
			session    *Session
		)

		BeforeEach(func() {
			login := exec.Command("cf", "auth", "admin", "admin")
			loginSession, err := Start(login, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(loginSession).Should(Exit(0))

			command := exec.Command("cf", "-v", "orgs")
			session, commandErr = Start(command, GinkgoWriter, GinkgoWriter)
		})

		AfterEach(func() {
			logout := exec.Command("cf", "logout")
			logoutSession, err := Start(logout, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(logoutSession).Should(Exit(0))
		})

		It("displays verbose output", func() {
			Expect(commandErr).NotTo(HaveOccurred())
			Eventually(session).Should(Exit(0))
			Eventually(session).Should(Say("REQUEST:"))
			Eventually(session).Should(Say("GET /v2/organizations"))
			Eventually(session).Should(Say("RESPONSE:"))
		})
	})
})
