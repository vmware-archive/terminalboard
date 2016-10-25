package main_test

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Executable", func() {
	Describe("API", func() {
		var (
			port    int
			session *gexec.Session
		)

		BeforeEach(func() {
			port = 54321

			binPath, err := gexec.Build("github.com/pivotal-cf/terminalboard")
			Expect(err).NotTo(HaveOccurred())

			command := exec.Command(binPath)

			os.Setenv("PORT", fmt.Sprintf("%d", port))

			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session.Out).Should(Say("Listening on"))
			Eventually(session.Out).Should(Say("%d", port))
		})

		It("serves pipeline status", func() {
			apiURL := fmt.Sprintf("http://localhost:%d/api/pipeline_statuses", port)
			resp, err := http.Get(apiURL)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		AfterEach(func() {
			session = session.Kill()
			Eventually(session).Should(gexec.Exit())
		})
	})

	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
	})
})
