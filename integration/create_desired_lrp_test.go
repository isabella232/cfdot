package integration_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/models"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("create-desired-lrp", func() {
	var sess *gexec.Session

	itValidatesBBSFlags("create-desired-lrp")

	Context("when no spec is passed", func() {
		JustBeforeEach(func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "create-desired-lrp")

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess.Exited).Should(BeClosed())
		})

		It("exits with status code of 3", func() {
			Expect(sess.ExitCode()).To(Equal(3))
		})

		It("prints an error on stderr", func() {
			Expect(sess.Err).To(gbytes.Say(`missing spec`))
		})

		It("prints usage", func() {
			Expect(sess.Err).To(gbytes.Say("cfdot create-desired-lrp \\(spec\\|@file\\) .*"))
		})
	})

	Context("when bbs responds with 200 status code", func() {
		var (
			lrp  *models.DesiredLRP
			args []string
		)

		BeforeEach(func() {
			lrp = &models.DesiredLRP{
				ProcessGuid: "some-process-guid",
			}
		})

		JustBeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/desired_lrp/desire.r2"),
					ghttp.VerifyProtoRepresenting(&models.DesireLRPRequest{
						DesiredLrp: lrp,
					}),
					ghttp.RespondWithProto(200, &models.DesiredLRPLifecycleResponse{
						Error: nil,
					}),
				),
			)

			args = append([]string{"--bbsURL", bbsServer.URL(), "create-desired-lrp"}, args...)
			cfdotCmd := exec.Command(
				cfdotPath,
				args...,
			)
			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Exited).Should(BeClosed())
		})

		Context("as json", func() {
			BeforeEach(func() {
				spec, err := json.Marshal(lrp)
				Expect(err).NotTo(HaveOccurred())
				args = []string{string(spec)}
			})

			It("exits with status code of 0", func() {
				Expect(sess.ExitCode()).To(Equal(0))
			})
		})

		Context("as a file", func() {
			BeforeEach(func() {
				spec, err := json.Marshal(lrp)
				Expect(err).NotTo(HaveOccurred())
				f, err := ioutil.TempFile(os.TempDir(), "desired_lrp_spec")
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()
				_, err = f.Write(spec)
				Expect(err).NotTo(HaveOccurred())
				args = []string{"@" + f.Name()}
			})

			It("exits with status code 0", func() {
				Expect(sess.ExitCode()).To(Equal(0))
			})
		})

		Context("empty spec", func() {
			BeforeEach(func() {
				args = nil
			})

			It("exits with status code of 3", func() {
				Expect(sess.ExitCode()).To(Equal(3))
			})
		})

		Context("invalid spec", func() {
			BeforeEach(func() {
				args = []string{"foo"}
			})

			It("prints the error", func() {
				Expect(sess.Err).To(gbytes.Say("Invalid JSON:"))
			})

			It("exits with status code of 3", func() {
				Expect(sess.ExitCode()).To(Equal(3))
			})
		})

		Context("non-existing spec file", func() {
			BeforeEach(func() {
				args = []string{"/path/to/non/existing/file"}
			})

			It("exits with status code of 3", func() {
				Expect(sess.ExitCode()).To(Equal(3))
			})
		})
	})

	Context("when bbs responds with non-200 status code", func() {
		JustBeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/desired_lrp/desire.r2"),
					ghttp.RespondWithProto(500, &models.DesiredLRPLifecycleResponse{
						Error: &models.Error{
							Type:    models.Error_Deadlock,
							Message: "deadlock detected",
						},
					}),
				),
			)

			cfdotCmd := exec.Command(
				cfdotPath,
				"--bbsURL", bbsServer.URL(), "create-desired-lrp", "{}",
			)
			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Exited).Should(BeClosed())
		})

		It("exits with status code 4", func() {
			Expect(sess.ExitCode()).To(Equal(4))
		})

		It("prints the error", func() {
			Expect(sess.Err).To(gbytes.Say("deadlock"))
		})
	})
})
