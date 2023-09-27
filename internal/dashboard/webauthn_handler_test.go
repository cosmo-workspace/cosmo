package dashboard

import (
	"context"
	"encoding/json"
	"net/http"

	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bufbuild/connect-go"
	"github.com/go-webauthn/webauthn/webauthn"

	cosmowebauthn "github.com/cosmo-workspace/cosmo/pkg/auth/webauthn"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1/dashboardv1alpha1connect"
)

var _ = Describe("Dashboard server [WebAuthn]", func() {
	var (
		userSession string
		client      dashboardv1alpha1connect.WebAuthnServiceClient
	)

	BeforeEach(func() {
		userSession = test_CreateLoginUserSession("normal-user", "user", nil, "password")
		client = dashboardv1alpha1connect.NewWebAuthnServiceClient(http.DefaultClient, "http://localhost:8888")
	})

	AfterEach(func() {
		clientMock.Clear()
		testUtil.DeleteCosmoUserAll()
	})
	//==================================================================================
	Describe("[ListCredentials]", func() {

		run_test := func(wantErr bool, req *dashv1alpha1.ListCredentialsRequest) {
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.ListCredentials(ctx, NewRequestWithSession(req, userSession))
			if err == nil {
				Expect(wantErr).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())
				Ω(res.Msg).To(MatchSnapShot())
			} else {
				Expect(wantErr).To(BeTrue())
				Expect(res).To(BeNil())
				Ω(err.Error()).To(MatchSnapShot())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry("OK", false, &dashv1alpha1.ListCredentialsRequest{UserName: "normal-user"}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("empty user name", true, &dashv1alpha1.ListCredentialsRequest{UserName: ""}),
			Entry("user not found", true, &dashv1alpha1.ListCredentialsRequest{UserName: "notfound"}),
		)
	})

	Describe("[Registration]", func() {

		run_test := func(wantErrBegin bool, reqBegin *dashv1alpha1.BeginRegistrationRequest, wantErrFin bool, reqFin *dashv1alpha1.FinishRegistrationRequest) {
			By("---------------test start----------------")
			ctx := context.Background()
			resBegin, err := client.BeginRegistration(ctx, connect.NewRequest(reqBegin))
			if err == nil {
				Expect(wantErrBegin).To(BeFalse())
				Expect(err).NotTo(HaveOccurred())
				Ω(webauthnSnapshot(resBegin.Msg.CredentialCreationOptions)).To(MatchSnapShot())

				By("---------------FinishRegistration----------------")
				resFin, err := client.FinishRegistration(ctx, connect.NewRequest(reqFin))
				if err == nil {
					Expect(wantErrFin).To(BeFalse())
					Expect(err).NotTo(HaveOccurred())
					Ω(resFin.Msg).To(MatchSnapShot())
				} else {
					Expect(wantErrFin).To(BeTrue())
					Expect(resFin).To(BeNil())
					Ω(err.Error()).To(MatchSnapShot())
				}
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(resBegin).To(BeNil())
				Expect(wantErrBegin).To(BeTrue())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in begin registration but error in finish registration:",
			run_test,
			Entry("cannot create mock for window.navigator.create() response",
				false, &dashv1alpha1.BeginRegistrationRequest{UserName: "normal-user"},
				true, &dashv1alpha1.FinishRegistrationRequest{UserName: "normal-user", CredentialCreationResponse: `{
					"id": "SMwLPQP1TEcS0qgAQe4HcaCln24GW7TkehoswjZr9ic",
					"rawId": "SMwLPQP1TEcS0qgAQe4HcaCln24GW7TkehoswjZr9ic",
					"type": "public-key",
					"response": {
						"clientDataJSON": "eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiOXljQVdNT3lTU0lEX2hJLWtINlV0RFJmSFo4SkJjZE5Tak5QY2dZakFScyIsIm9yaWdpbiI6Imh0dHBzOi8vZGFzaC1rM2QtY29kZS1zZXJ2ZXIuamxhbmRvd25lci5kZXYiLCJjcm9zc09yaWdpbiI6ZmFsc2UsIm90aGVyX2tleXNfY2FuX2JlX2FkZGVkX2hlcmUiOiJkbyBub3QgY29tcGFyZSBjbGllbnREYXRhSlNPTiBhZ2FpbnN0IGEgdGVtcGxhdGUuIFNlZSBodHRwczovL2dvby5nbC95YWJQZXgifQ",
						"attestationObject": "o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YVkBZ3clwhlksBwhsD4VrNEScgU34WN1s_gY-oKp89sh82y3RQAAAAAAAAAAAAAAAAAAAAAAAAAAACBIzAs9A_VMRxLSqABB7gdxoKWfbgZbtOR6GizCNmv2J6QBAwM5AQAgWQEAm49eL1TawnIGRRBAo7whklJO8oL0ePOST56sWmY4v8UcMNCcu_2jp9fcrrXaviVcb09TEn5EjrfclEO7s9idBSwoOUvepKavXmnE_6o5SxSNEnG0FwGV5ZAgpgTFwXzSE2OV41VVWaxwwN66TJUYMzojG5zws_FHTsA3TsyiIKxRp1ke2AZ1hyEWhpdwV-Dqs8w2DmFPvslbdu2rJtqSIRRbrmNIFGxWgEL9-CpnE5_7r2Oo2rS4cLB3G8Pq75dlLM5B8_eKdipRveuS08H8dEfQuJ4M1va0yrHZTf_tYkTsoPg3Iu-albVjrr-ZZux42O1JyC_B7YZ_R6EBLApUaSFDAQAB"
					}
				}`}),
		)

		DescribeTable("❌ fail in begin registration:",
			run_test,
			Entry("empty user name", true, &dashv1alpha1.BeginRegistrationRequest{UserName: ""}, true, nil),
			Entry("user not found", true, &dashv1alpha1.BeginRegistrationRequest{UserName: "notfound"}, true, nil),
		)
	})

	Describe("[Login]", func() {
		registerCredential := func() {
			ctx := context.Background()
			wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "normal-user")
			Expect(err).NotTo(HaveOccurred())

			cred := cosmowebauthn.Credential{
				Base64URLEncodedId: "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU",
				DisplayName:        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
				Timestamp:          1696436610,
				Cred: webauthn.Credential{
					ID:              []byte("D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU="),
					PublicKey:       []byte("pQECAyYgASFYIHYeITFpgzmctVCg/uRgvZWsXxej2aPHG+iiAidcreaiIlggyQy0xtTdTiqYqPlh8SQ0ViQH1vprBBKV9rFZZUhXHxA="),
					AttestationType: "none",
					Transport:       nil,
					Flags: webauthn.CredentialFlags{
						UserPresent:    true,
						UserVerified:   true,
						BackupEligible: false,
						BackupState:    false,
					},
					Authenticator: webauthn.Authenticator{
						AAGUID:       []byte("rc4AAjW8xgpkiwsl8fBVAw=="),
						SignCount:    0,
						CloneWarning: false,
						Attachment:   "",
					},
				},
			}
			wu.RegisterCredential(ctx, &cred)
		}
		removeCredential := func() {
			ctx := context.Background()
			wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "normal-user")
			Expect(err).NotTo(HaveOccurred())
			wu.RemoveCredential(ctx, "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU")
		}

		run_test := func(wantErrBegin bool, reqBegin *dashv1alpha1.BeginLoginRequest, wantErrFin bool, reqFin *dashv1alpha1.FinishLoginRequest) {

			By("registering credential")
			registerCredential()
			defer removeCredential()

			By("---------------test start----------------")
			ctx := context.Background()
			resBegin, err := client.BeginLogin(ctx, connect.NewRequest(reqBegin))
			if err == nil {
				Expect(err).NotTo(HaveOccurred())
				Ω(webauthnSnapshot(resBegin.Msg.CredentialRequestOptions)).To(MatchSnapShot())
				Expect(wantErrBegin).To(BeFalse())

				By("---------------FinishLogin----------------")
				resFin, err := client.FinishLogin(ctx, connect.NewRequest(reqFin))
				if err == nil {
					Expect(wantErrFin).To(BeFalse())
					Expect(err).NotTo(HaveOccurred())
					Ω(resFin.Msg).To(MatchSnapShot())
				} else {
					Expect(wantErrFin).To(BeTrue())
					Expect(resFin).To(BeNil())
					Ω(err.Error()).To(MatchSnapShot())
				}
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(resBegin).To(BeNil())
				Expect(wantErrBegin).To(BeTrue())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in begin login but error in finish login:",
			run_test,
			Entry("cannot create mock for window.navigator.create() response",
				false, &dashv1alpha1.BeginLoginRequest{UserName: "normal-user"},
				true, &dashv1alpha1.FinishLoginRequest{UserName: "normal-user", CredentialRequestResult: `{
					"id": "KxnR0-QIUysgrQAEOKgwwjMnhz45xqyMTxcHaM8DGIQ",
					"type": "public-key",
					"rawId": "KxnR0-QIUysgrQAEOKgwwjMnhz45xqyMTxcHaM8DGIQ",
					"response": {
						"clientDataJSON": "eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiS09rcEF0UWY0NVFvN28zSml0aWY4VlRJU25ncVEwXzlYVzBXNHVEQl9obyIsIm9yaWdpbiI6Imh0dHBzOi8vZGFzaC1rM2QtY29kZS1zZXJ2ZXIuamxhbmRvd25lci5kZXYiLCJjcm9zc09yaWdpbiI6ZmFsc2V9",
						"authenticatorData": "dyXCGWSwHCGwPhWs0RJyBTfhY3Wz-Bj6gqnz2yHzbLcFAAAAAQ",
						"signature": "hOOWKeP4EMDW5lM8uP4H_lMG-ZtvxaWycM-KlluSlRcypMyWKukBF3Onng8UDphhmm2lKO-rKDeE92r6cDN3bQ8U16uczAyzWp2oLGCS-tXDR1mQw6sObCEOw8fwxxmmmMokS0V-rcI7QuMTarfiebdBvx-8_imVeB-OD_bj5l9-1UaqdWxVYFLZEtJ1hVEOdKP1bez48qqgYCRngvXNx_NnTol3SJ29W03Pt7FWsh3h2QF1akRLXk9e_XvD5-YZAUOaJCE-4Yqwu_Z45q1e0pcNp_34hBRosFwRZcJOQ5EjRpK4uLyVuKsnBOYrh55kSg_smRisgf1k8ctuGKEJPw",
						"userHandle": "ODJmNGI0Yjc2ZDc4Zjg4YTU1OGU5NGU0YmFjYmJkOWRjYzVkYzdhY2JlMmQ5ZjI5ZDRlZTAwZjEwZWZmMWEyMg"
					}
				}`}),
		)

		DescribeTable("❌ fail in begin login:",
			run_test,
			Entry("empty user name", true, &dashv1alpha1.BeginLoginRequest{UserName: ""}, true, nil),
			Entry("user not found", true, &dashv1alpha1.BeginLoginRequest{UserName: "notfound"}, true, nil),
		)
	})

	Describe("[UpdateCredential]", func() {
		registerCredential := func() {
			ctx := context.Background()
			wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "normal-user")
			Expect(err).NotTo(HaveOccurred())

			cred := cosmowebauthn.Credential{
				Base64URLEncodedId: "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU",
				DisplayName:        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
				Timestamp:          1696436610,
				Cred: webauthn.Credential{
					ID:              []byte("D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU="),
					PublicKey:       []byte("pQECAyYgASFYIHYeITFpgzmctVCg/uRgvZWsXxej2aPHG+iiAidcreaiIlggyQy0xtTdTiqYqPlh8SQ0ViQH1vprBBKV9rFZZUhXHxA="),
					AttestationType: "none",
					Transport:       nil,
					Flags: webauthn.CredentialFlags{
						UserPresent:    true,
						UserVerified:   true,
						BackupEligible: false,
						BackupState:    false,
					},
					Authenticator: webauthn.Authenticator{
						AAGUID:       []byte("rc4AAjW8xgpkiwsl8fBVAw=="),
						SignCount:    0,
						CloneWarning: false,
						Attachment:   "",
					},
				},
			}
			wu.RegisterCredential(ctx, &cred)
		}
		removeCredential := func() {
			ctx := context.Background()
			wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "normal-user")
			Expect(err).NotTo(HaveOccurred())
			wu.RemoveCredential(ctx, "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU")
		}

		run_test := func(wantErr bool, req *dashv1alpha1.UpdateCredentialRequest) {

			By("registering credential")
			registerCredential()
			defer removeCredential()

			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.UpdateCredential(ctx, NewRequestWithSession(req, userSession))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
				Expect(err).NotTo(HaveOccurred())
				Expect(wantErr).To(BeFalse())

			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).To(BeNil())
				Expect(wantErr).To(BeTrue())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry("update display name", false, &dashv1alpha1.UpdateCredentialRequest{UserName: "normal-user", CredId: "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU", CredDisplayName: "new display name"}),
			Entry("update display name to empty", false, &dashv1alpha1.UpdateCredentialRequest{UserName: "normal-user", CredId: "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU", CredDisplayName: ""}),
			Entry("no change", false, &dashv1alpha1.UpdateCredentialRequest{UserName: "normal-user", CredId: "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU", CredDisplayName: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36"}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("credential not found", true, &dashv1alpha1.UpdateCredentialRequest{UserName: "normal-user", CredId: "notfound", CredDisplayName: "new display name"}),
			Entry("empty user name", true, &dashv1alpha1.UpdateCredentialRequest{UserName: "", CredId: "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU", CredDisplayName: "new display name"}),
			Entry("user not found", true, &dashv1alpha1.UpdateCredentialRequest{UserName: "notfound", CredId: "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU", CredDisplayName: "new display name"}),
		)
	})

	Describe("[RemoveCredential]", func() {
		registerCredential := func() {
			ctx := context.Background()
			wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "normal-user")
			Expect(err).NotTo(HaveOccurred())

			cred := cosmowebauthn.Credential{
				Base64URLEncodedId: "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU",
				DisplayName:        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
				Timestamp:          1696436610,
				Cred: webauthn.Credential{
					ID:              []byte("D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU="),
					PublicKey:       []byte("pQECAyYgASFYIHYeITFpgzmctVCg/uRgvZWsXxej2aPHG+iiAidcreaiIlggyQy0xtTdTiqYqPlh8SQ0ViQH1vprBBKV9rFZZUhXHxA="),
					AttestationType: "none",
					Transport:       nil,
					Flags: webauthn.CredentialFlags{
						UserPresent:    true,
						UserVerified:   true,
						BackupEligible: false,
						BackupState:    false,
					},
					Authenticator: webauthn.Authenticator{
						AAGUID:       []byte("rc4AAjW8xgpkiwsl8fBVAw=="),
						SignCount:    0,
						CloneWarning: false,
						Attachment:   "",
					},
				},
			}
			wu.RegisterCredential(ctx, &cred)
		}
		removeCredential := func() {
			ctx := context.Background()
			wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "normal-user")
			Expect(err).NotTo(HaveOccurred())
			wu.RemoveCredential(ctx, "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU")
		}

		run_test := func(wantErr bool, req *dashv1alpha1.DeleteCredentialRequest) {

			By("registering credential")
			registerCredential()
			defer removeCredential()

			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.DeleteCredential(ctx, NewRequestWithSession(req, userSession))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
				Expect(err).NotTo(HaveOccurred())
				Expect(wantErr).To(BeFalse())

			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).To(BeNil())
				Expect(wantErr).To(BeTrue())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry("OK", false, &dashv1alpha1.DeleteCredentialRequest{UserName: "normal-user", CredId: "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU"}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("no credential id", true, &dashv1alpha1.DeleteCredentialRequest{UserName: "normal-user", CredId: ""}),
			Entry("credential not found", true, &dashv1alpha1.DeleteCredentialRequest{UserName: "normal-user", CredId: "notfound"}),
			Entry("empty user name", true, &dashv1alpha1.DeleteCredentialRequest{UserName: "", CredId: "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU"}),
			Entry("user not found", true, &dashv1alpha1.DeleteCredentialRequest{UserName: "notfound", CredId: "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU"}),
		)
	})
})

func webauthnSnapshot(jsondata string) map[string]interface{} {
	var v map[string]interface{}
	err := json.Unmarshal([]byte(jsondata), &v)
	Expect(err).NotTo(HaveOccurred())

	if vv, ok := v["publicKey"]; ok {
		if vvi, ok := vv.(map[string]interface{}); ok {
			if _, ok := vvi["challenge"]; ok {
				vvi["challenge"] = "CHALLENGE"
			}
		}
	}

	return v
}
