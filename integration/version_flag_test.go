package integration

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Version", func() {
	Context("when the -v option is provided without additional fields", func() {
		var (
			commandErr error
			session    *Session
		)

		BeforeEach(func() {
			command := exec.Command("cf", "-v")
			session, commandErr = Start(command, GinkgoWriter, GinkgoWriter)
		})

		It("displays the version", func() {
			Expect(commandErr).NotTo(HaveOccurred())
			Eventually(session).Should(Exit(0))
			Eventually(session).Should(Say("cf version"))
		})
	})
})
