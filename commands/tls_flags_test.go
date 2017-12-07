package commands_test

import (
	"os"

	"code.cloudfoundry.org/cfdot/commands"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/spf13/cobra"
)

var _ = FDescribe("TLSFlags", func() {
	var validTLSFlags map[string]string
	var dummyCmd *cobra.Command
	var err error
	var output *gbytes.Buffer

	BeforeEach(func() {
		dummyCmd = &cobra.Command{
			Use: "dummy",
			Run: func(cmd *cobra.Command, args []string) {},
		}
		commands.AddTLSFlags(dummyCmd)
		output = gbytes.NewBuffer()
		dummyCmd.SetOutput(output)

		validTLSFlags = map[string]string{
			"--clientCertFile": "fixtures/bbsClient.crt",
			"--clientKeyFile":  "fixtures/bbsClient.key",
			"--caCertFile":     "fixtures/bbsCACert.crt",
			"--skipCertVerify": "true",
		}
	})

	JustBeforeEach(func() {
		err = dummyCmd.PreRunE(dummyCmd, dummyCmd.Flags().Args())
	})

	FDescribe("BBSFlags", func() {
		BeforeEach(func() {
			mergeFlags(validTLSFlags, map[string]string{"--bbsURL": "https://example.com"})
		})

		Context("TLSFlags present", func() {
			BeforeEach(func() {
				commands.AddBBSFlags(dummyCmd)
			})

			Context("all flags are set right", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("should succeed", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Describe("skipCertVerify", func() {
				Context("when skipCertVerify is true", func() {
					Context("when the CA cert file is absent", func() {
						BeforeEach(func() {
							validTLSFlags["--skipCertVerify"] = "true"
							delete(validTLSFlags, "--caCertFile")
							parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
						})

						It("ignores the missing CA cert", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})
				})

				Context("when a SKIP_CERT_VERIFY environment variable is specified", func() {
					AfterEach(func() {
						os.Unsetenv("SKIP_CERT_VERIFY")
					})

					Context("when the SKIP_CERT_VERIFY is valid", func() {
						BeforeEach(func() {
							os.Setenv("SKIP_CERT_VERIFY", "true")
						})

						Context("when the flag is not present", func() {
							BeforeEach(func() {
								delete(validTLSFlags, "--skipCertVerify")
								delete(validTLSFlags, "caCertFile")
								parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
								Expect(parseFlagsErr).NotTo(HaveOccurred())
							})

							It("does not error", func() {
								Expect(err).NotTo(HaveOccurred())
							})
						})
					})

					Context("when the SKIP_CERT_VERIFY is not valid", func() {
						BeforeEach(func() {
							parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--skipCertVerify"))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
							os.Setenv("SKIP_CERT_VERIFY", "sponge")
						})

						It("returns an error", func() {
							Expect(err).To(MatchError("The value 'sponge' is not a valid value for SKIP_CERT_VERIFY. Please specify one of the following valid boolean values: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False"))
						})
					})

					Context("when the --skipCertVerify flag is also specified", func() {
						BeforeEach(func() {
							parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--skipCertVerify", "true"))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
							os.Setenv("SKIP_CERT_VERIFY", "false")
						})

						It("uses the value from the flag instead of the environment variable", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})
				})
			})

			Context("CA Cert file", func() {
				Context("when CA cert is not specified", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--caCertFile"))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("returns a validation error", func() {
						Expect(err).To(MatchError("--caCertFile must be specified if using HTTPS and --skipCertVerify is not set"))
					})

					It("exits with code 3", func() {
						cfdotError, ok := err.(commands.CFDotError)
						Expect(ok).To(BeTrue())
						Expect(cfdotError.ExitCode()).To(Equal(3))
					})
				})

				Context("when the CA cert file flag points to a nonexistent file", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--caCertFile", "notreal.cacrt"))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("returns a validation error", func() {
						certfile := validTLSFlags["--caCertFile"]
						Expect(err).To(MatchError(MatchRegexp("CA cert file '" + certfile + "' doesn't exist or is not readable: .*")))
					})

					It("exits with code 3", func() {
						cfdotError, ok := err.(commands.CFDotError)
						Expect(ok).To(BeTrue())
						Expect(cfdotError.ExitCode()).To(Equal(3))
					})
				})

				Context("when a CA_CERT_FILE environment variable is specified", func() {
					AfterEach(func() {
						os.Unsetenv("CA_CERT_FILE")
					})

					Context("when the CA_CERT_FILE is valid", func() {
						BeforeEach(func() {
							os.Setenv("CA_CERT_FILE", validTLSFlags["--caCertFile"])
							parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--caCertFile"))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
						})

						It("does not error", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})

					Context("when the CA_CERT_FILE points to a nonexistent file", func() {
						BeforeEach(func() {
							os.Setenv("CA_CERT_FILE", "sponge")
							parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--caCertFile"))

							Expect(parseFlagsErr).NotTo(HaveOccurred())
						})

						It("returns an error", func() {
							Expect(err).To(MatchError(MatchRegexp("^CA cert file 'sponge' doesn't exist or is not readable: .*")))
						})
					})

					Context("when the --CACertFile flag is also specified", func() {
						BeforeEach(func() {
							parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
							os.Setenv("CA_CERT_FILE", "not a key file")
						})

						It("uses the value from the flag instead of the environment variable", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})
				})
			})

			Describe("clientCertFile/clientKeyFile", func() {
				Context("when a cert file is specified, but a key file is not", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--clientKeyFile"))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("returns a validation error", func() {
						Expect(err).To(MatchError("--clientCertFile and --clientKeyFile must both be specified for TLS connections."))
					})

					It("exits with code 3", func() {
						cfdotError, ok := err.(commands.CFDotError)
						Expect(ok).To(BeTrue())
						Expect(cfdotError.ExitCode()).To(Equal(3))
					})
				})

				Context("when a key file is specified, but a cert file is not", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--clientCertFile"))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("returns a validation error", func() {
						Expect(err).To(MatchError("--clientCertFile and --clientKeyFile must both be specified for TLS connections."))
					})

					It("exits with code 3", func() {
						cfdotError, ok := err.(commands.CFDotError)
						Expect(ok).To(BeTrue())
						Expect(cfdotError.ExitCode()).To(Equal(3))
					})
				})

				Context("when the key file flag points to a nonexistent file", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--clientKeyFile", "sandwich.key"))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("returns a validation error", func() {
						keyfile := validTLSFlags["--clientKeyFile"]
						Expect(err).To(MatchError(MatchRegexp("^key file '" + keyfile + "' doesn't exist or is not readable: .*")))
					})

					It("exits with code 3", func() {
						cfdotError, ok := err.(commands.CFDotError)
						Expect(ok).To(BeTrue())
						Expect(cfdotError.ExitCode()).To(Equal(3))
					})
				})

				Context("when the cert file flag points to a nonexistent file", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--clientCertFile", "sandwich.crt"))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("returns a validation error", func() {
						certfile := validTLSFlags["--clientCertFile"]
						Expect(err).To(MatchError(MatchRegexp("^cert file '" + certfile + "' doesn't exist or is not readable: .*")))
					})

					It("exits with code 3", func() {
						cfdotError, ok := err.(commands.CFDotError)
						Expect(ok).To(BeTrue())
						Expect(cfdotError.ExitCode()).To(Equal(3))
					})
				})

				Context("when a CLIENT_CERT_FILE environment variable is specified", func() {
					AfterEach(func() {
						os.Unsetenv("CLIENT_CERT_FILE")
					})

					Context("when the CLIENT_CERT_FILE is valid", func() {
						BeforeEach(func() {
							os.Setenv("CLIENT_CERT_FILE", validTLSFlags["--clientCertFile"])
							parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--clientCertFile"))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
						})

						It("does not error", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})

					Context("when the CLIENT_CERT_FILE points to a nonexistent file", func() {
						BeforeEach(func() {
							os.Setenv("CLIENT_CERT_FILE", "sponge")
							parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--clientCertFile"))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
						})

						It("returns an error", func() {
							Expect(err).To(MatchError(MatchRegexp("^cert file 'sponge' doesn't exist or is not readable: .*")))
						})
					})

					Context("when the --clientCertFile flag is also specified", func() {
						BeforeEach(func() {
							parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
							os.Setenv("CLIENT_CERT_FILE", "not a cert file")
						})

						It("uses the value from the flag instead of the environment variable", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})
				})

				Context("when a CLIENT_KEY_FILE environment variable is specified", func() {
					AfterEach(func() {
						os.Unsetenv("CLIENT_KEY_FILE")
					})

					Context("when the CLIENT_KEY_FILE is valid", func() {
						BeforeEach(func() {
							os.Setenv("CLIENT_KEY_FILE", validTLSFlags["--clientKeyFile"])
							parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--clientKeyFile"))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
						})

						It("does not error", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})

					Context("when the CLIENT_KEY_FILE points to a nonexistent file", func() {
						BeforeEach(func() {
							os.Setenv("CLIENT_KEY_FILE", "sponge")
							parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--clientKeyFile"))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
						})

						It("returns an error", func() {
							Expect(err).To(MatchError(MatchRegexp("^key file 'sponge' doesn't exist or is not readable: .*")))
						})
					})

					Context("when the --clientKeyFile flag is also specified", func() {
						BeforeEach(func() {
							parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
							os.Setenv("CLIENT_KEY_FILE", "not a key file")
						})

						It("uses the value from the flag instead of the environment variable", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})
				})
			})

		})

		FContext("LegacyFlags", func() {
			BeforeEach(func() {
				mergeFlags(validTLSFlags, map[string]string{
					"--bbsCACertFile":     "fixtures/randomCACert.crt",
					"--bbsCertFile":       "fixtures/randomClient.crt",
					"--bbsKeyFile":        "fixtures/randomClient.key",
					"--bbsSkipCertVerify": "false",
				})
			})

			Context("both legacy flags and tls flags presnt", func() {
				BeforeEach(func() {
					commands.AddBBSFlags(dummyCmd)
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("prefers tls flags", func() {
					Expect(commands.BBSClientConfig.CACertFile).To(Equal("fixtures/bbsCACert.crt"))
					Expect(commands.BBSClientConfig.CertFile).To(Equal("fixtures/bbsClient.crt"))
					Expect(commands.BBSClientConfig.KeyFile).To(Equal("fixtures/bbsClient.key"))
					Expect(commands.BBSClientConfig.SkipCertVerify).To(BeTrue())
				})
			})

			Context("when the tls ca cert file is missing", func() {
				BeforeEach(func() {
					delete(validTLSFlags, "--caCertFile")
					commands.AddBBSFlags(dummyCmd)
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("picks the legacy flag for the missing tls flag", func() {
					Expect(commands.BBSClientConfig.CACertFile).To(Equal("fixtures/randomCACert.crt"))
				})

				It("picks tls flags", func() {
					Expect(commands.BBSClientConfig.CertFile).To(Equal("fixtures/bbsClient.crt"))
					Expect(commands.BBSClientConfig.KeyFile).To(Equal("fixtures/bbsClient.key"))
				})
			})

			Context("when the tls client cert file is missing", func() {
				BeforeEach(func() {
					delete(validTLSFlags, "--clientCertFile")
					commands.AddBBSFlags(dummyCmd)
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("picks the legacy flag for the missing tls flag", func() {
					Expect(commands.BBSClientConfig.CertFile).To(Equal("fixtures/randomClient.crt"))
				})

				It("picks tls flags", func() {
					Expect(commands.BBSClientConfig.CACertFile).To(Equal("fixtures/bbsCACert.crt"))
					Expect(commands.BBSClientConfig.KeyFile).To(Equal("fixtures/bbsClient.key"))
				})
			})

			Context("when the tls client key file is missing", func() {
				BeforeEach(func() {
					delete(validTLSFlags, "--clientKeyFile")
					commands.AddBBSFlags(dummyCmd)
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("picks the legacy flag for the missing tls flag", func() {
					Expect(commands.BBSClientConfig.KeyFile).To(Equal("fixtures/randomClient.key"))
				})

				It("picks tls flags", func() {
					Expect(commands.BBSClientConfig.CACertFile).To(Equal("fixtures/bbsCACert.crt"))
					Expect(commands.BBSClientConfig.CertFile).To(Equal("fixtures/bbsClient.crt"))
				})
			})

			Context("when the tls skip cert verify is missing", func() {
				BeforeEach(func() {
					delete(validTLSFlags, "--skipCertVerify")
					replaceFlagValue(validTLSFlags, "--bbsSkipCertVerify", "true")
					commands.AddBBSFlags(dummyCmd)
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("picks the legacy flag for skip cert verify", func() {
					Expect(commands.BBSClientConfig.SkipCertVerify).To(BeTrue())
				})
			})
		})
	})

	Describe("LocketFlags", func() {
		BeforeEach(func() {
			replaceFlagValue(validTLSFlags, "--locketAPILocation", "https://example.com")
		})

		Context("TLSFlags present", func() {
			BeforeEach(func() {
				commands.AddLocketFlags(dummyCmd)
			})

			Context("all flags are set right", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("should succeed", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Describe("skipCertVerify", func() {
				Context("when skipCertVerify is true", func() {
					Context("when the CA cert file is absent", func() {
						BeforeEach(func() {
							validTLSFlags["--skipCertVerify"] = "true"
							delete(validTLSFlags, "--caCertFile")
							parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
						})

						It("ignores the missing CA cert", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})
				})

				Context("when a SKIP_CERT_VERIFY environment variable is specified", func() {
					AfterEach(func() {
						os.Unsetenv("SKIP_CERT_VERIFY")
					})

					Context("when the SKIP_CERT_VERIFY is valid", func() {
						BeforeEach(func() {
							os.Setenv("SKIP_CERT_VERIFY", "true")
						})

						Context("when the flag is not present", func() {
							BeforeEach(func() {
								delete(validTLSFlags, "--skipCertVerify")
								delete(validTLSFlags, "caCertFile")
								parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
								Expect(parseFlagsErr).NotTo(HaveOccurred())
							})

							It("does not error", func() {
								Expect(err).NotTo(HaveOccurred())
							})
						})
					})

					Context("when the SKIP_CERT_VERIFY is not valid", func() {
						BeforeEach(func() {
							parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--skipCertVerify"))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
							os.Setenv("SKIP_CERT_VERIFY", "sponge")
						})

						It("returns an error", func() {
							Expect(err).To(MatchError("The value 'sponge' is not a valid value for SKIP_CERT_VERIFY. Please specify one of the following valid boolean values: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False"))
						})
					})

					Context("when the --skipCertVerify flag is also specified", func() {
						BeforeEach(func() {
							parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--skipCertVerify", "true"))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
							os.Setenv("SKIP_CERT_VERIFY", "false")
						})

						It("uses the value from the flag instead of the environment variable", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})
				})
			})

			Context("CA Cert file", func() {
				Context("when CA cert is not specified", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--caCertFile"))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("returns a validation error", func() {
						Expect(err).To(MatchError("--caCertFile must be specified if using HTTPS and --skipCertVerify is not set"))
					})

					It("exits with code 3", func() {
						cfdotError, ok := err.(commands.CFDotError)
						Expect(ok).To(BeTrue())
						Expect(cfdotError.ExitCode()).To(Equal(3))
					})
				})

				Context("when the CA cert file flag points to a nonexistent file", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--caCertFile", "notreal.cacrt"))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("returns a validation error", func() {
						certfile := validTLSFlags["--caCertFile"]
						Expect(err).To(MatchError(MatchRegexp("CA cert file '" + certfile + "' doesn't exist or is not readable: .*")))
					})

					It("exits with code 3", func() {
						cfdotError, ok := err.(commands.CFDotError)
						Expect(ok).To(BeTrue())
						Expect(cfdotError.ExitCode()).To(Equal(3))
					})
				})

				Context("when a CA_CERT_FILE environment variable is specified", func() {
					AfterEach(func() {
						os.Unsetenv("CA_CERT_FILE")
					})

					Context("when the CA_CERT_FILE is valid", func() {
						BeforeEach(func() {
							os.Setenv("CA_CERT_FILE", validTLSFlags["--caCertFile"])
							parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--caCertFile"))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
						})

						It("does not error", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})

					Context("when the CA_CERT_FILE points to a nonexistent file", func() {
						BeforeEach(func() {
							os.Setenv("CA_CERT_FILE", "sponge")
							parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--caCertFile"))

							Expect(parseFlagsErr).NotTo(HaveOccurred())
						})

						It("returns an error", func() {
							Expect(err).To(MatchError(MatchRegexp("^CA cert file 'sponge' doesn't exist or is not readable: .*")))
						})
					})

					Context("when the --CACertFile flag is also specified", func() {
						BeforeEach(func() {
							parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
							os.Setenv("CA_CERT_FILE", "not a key file")
						})

						It("uses the value from the flag instead of the environment variable", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})
				})
			})

			Describe("clientCertFile/clientKeyFile", func() {
				Context("when a cert file is specified, but a key file is not", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--clientKeyFile"))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("returns a validation error", func() {
						Expect(err).To(MatchError("--clientCertFile and --clientKeyFile must both be specified for TLS connections."))
					})

					It("exits with code 3", func() {
						cfdotError, ok := err.(commands.CFDotError)
						Expect(ok).To(BeTrue())
						Expect(cfdotError.ExitCode()).To(Equal(3))
					})
				})

				Context("when a key file is specified, but a cert file is not", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--clientCertFile"))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("returns a validation error", func() {
						Expect(err).To(MatchError("--clientCertFile and --clientKeyFile must both be specified for TLS connections."))
					})

					It("exits with code 3", func() {
						cfdotError, ok := err.(commands.CFDotError)
						Expect(ok).To(BeTrue())
						Expect(cfdotError.ExitCode()).To(Equal(3))
					})
				})

				Context("when the key file flag points to a nonexistent file", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--clientKeyFile", "sandwich.key"))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("returns a validation error", func() {
						keyfile := validTLSFlags["--clientKeyFile"]
						Expect(err).To(MatchError(MatchRegexp("^key file '" + keyfile + "' doesn't exist or is not readable: .*")))
					})

					It("exits with code 3", func() {
						cfdotError, ok := err.(commands.CFDotError)
						Expect(ok).To(BeTrue())
						Expect(cfdotError.ExitCode()).To(Equal(3))
					})
				})

				Context("when the cert file flag points to a nonexistent file", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--clientCertFile", "sandwich.crt"))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("returns a validation error", func() {
						certfile := validTLSFlags["--clientCertFile"]
						Expect(err).To(MatchError(MatchRegexp("^cert file '" + certfile + "' doesn't exist or is not readable: .*")))
					})

					It("exits with code 3", func() {
						cfdotError, ok := err.(commands.CFDotError)
						Expect(ok).To(BeTrue())
						Expect(cfdotError.ExitCode()).To(Equal(3))
					})
				})

				Context("when a CLIENT_CERT_FILE environment variable is specified", func() {
					AfterEach(func() {
						os.Unsetenv("CLIENT_CERT_FILE")
					})

					Context("when the CLIENT_CERT_FILE is valid", func() {
						BeforeEach(func() {
							os.Setenv("CLIENT_CERT_FILE", validTLSFlags["--clientCertFile"])
							parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--clientCertFile"))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
						})

						It("does not error", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})

					Context("when the CLIENT_CERT_FILE points to a nonexistent file", func() {
						BeforeEach(func() {
							os.Setenv("CLIENT_CERT_FILE", "sponge")
							parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--clientCertFile"))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
						})

						It("returns an error", func() {
							Expect(err).To(MatchError(MatchRegexp("^cert file 'sponge' doesn't exist or is not readable: .*")))
						})
					})

					Context("when the --clientCertFile flag is also specified", func() {
						BeforeEach(func() {
							parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
							os.Setenv("CLIENT_CERT_FILE", "not a cert file")
						})

						It("uses the value from the flag instead of the environment variable", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})
				})

				Context("when a CLIENT_KEY_FILE environment variable is specified", func() {
					AfterEach(func() {
						os.Unsetenv("CLIENT_KEY_FILE")
					})

					Context("when the CLIENT_KEY_FILE is valid", func() {
						BeforeEach(func() {
							os.Setenv("CLIENT_KEY_FILE", validTLSFlags["--clientKeyFile"])
							parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--clientKeyFile"))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
						})

						It("does not error", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})

					Context("when the CLIENT_KEY_FILE points to a nonexistent file", func() {
						BeforeEach(func() {
							os.Setenv("CLIENT_KEY_FILE", "sponge")
							parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--clientKeyFile"))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
						})

						It("returns an error", func() {
							Expect(err).To(MatchError(MatchRegexp("^key file 'sponge' doesn't exist or is not readable: .*")))
						})
					})

					Context("when the --clientKeyFile flag is also specified", func() {
						BeforeEach(func() {
							parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
							Expect(parseFlagsErr).NotTo(HaveOccurred())
							os.Setenv("CLIENT_KEY_FILE", "not a key file")
						})

						It("uses the value from the flag instead of the environment variable", func() {
							Expect(err).NotTo(HaveOccurred())
						})
					})
				})
			})
		})

		FContext("LegacyFlags", func() {
			BeforeEach(func() {
				mergeFlags(validTLSFlags, map[string]string{
					"--locketCACertFile":     "fixtures/randomCACert.crt",
					"--locketCertFile":       "fixtures/randomClient.crt",
					"--locketKeyFile":        "fixtures/randomClient.key",
					"--locketSkipCertVerify": "false",
				})
			})

			Context("both legacy flags and tls flags presnt", func() {
				BeforeEach(func() {
					commands.AddLocketFlags(dummyCmd)
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("prefers tls flags", func() {
					Expect(commands.LocketClientConfig.CACertFile).To(Equal("fixtures/bbsCACert.crt"))
					Expect(commands.LocketClientConfig.CertFile).To(Equal("fixtures/bbsClient.crt"))
					Expect(commands.LocketClientConfig.KeyFile).To(Equal("fixtures/bbsClient.key"))
					Expect(commands.LocketClientConfig.SkipCertVerify).To(BeTrue())
				})
			})

			Context("when the tls ca cert file is missing", func() {
				BeforeEach(func() {
					delete(validTLSFlags, "--caCertFile")
					commands.AddLocketFlags(dummyCmd)
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("picks the legacy flag for the missing tls flag", func() {
					Expect(commands.LocketClientConfig.CACertFile).To(Equal("fixtures/randomCACert.crt"))
				})

				It("picks tls flags", func() {
					Expect(commands.LocketClientConfig.CertFile).To(Equal("fixtures/bbsClient.crt"))
					Expect(commands.LocketClientConfig.KeyFile).To(Equal("fixtures/bbsClient.key"))
				})
			})

			Context("when the tls client cert file is missing", func() {
				BeforeEach(func() {
					delete(validTLSFlags, "--clientCertFile")
					commands.AddLocketFlags(dummyCmd)
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("picks the legacy flag for the missing tls flag", func() {
					Expect(commands.LocketClientConfig.CertFile).To(Equal("fixtures/randomClient.crt"))
				})

				It("picks tls flags", func() {
					Expect(commands.LocketClientConfig.CACertFile).To(Equal("fixtures/bbsCACert.crt"))
					Expect(commands.LocketClientConfig.KeyFile).To(Equal("fixtures/bbsClient.key"))
				})
			})

			Context("when the tls client key file is missing", func() {
				BeforeEach(func() {
					delete(validTLSFlags, "--clientKeyFile")
					commands.AddLocketFlags(dummyCmd)
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("picks the legacy flag for the missing tls flag", func() {
					Expect(commands.LocketClientConfig.KeyFile).To(Equal("fixtures/randomClient.key"))
				})

				It("picks tls flags", func() {
					Expect(commands.LocketClientConfig.CACertFile).To(Equal("fixtures/bbsCACert.crt"))
					Expect(commands.LocketClientConfig.CertFile).To(Equal("fixtures/bbsClient.crt"))
				})
			})

			Context("when the tls skip cert verify is missing", func() {
				BeforeEach(func() {
					delete(validTLSFlags, "--skipCertVerify")
					replaceFlagValue(validTLSFlags, "--locketSkipCertVerify", "true")
					commands.AddLocketFlags(dummyCmd)
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("picks the legacy flag for skip cert verify", func() {
					Expect(commands.LocketClientConfig.SkipCertVerify).To(BeTrue())
				})
			})
		})
	})
})