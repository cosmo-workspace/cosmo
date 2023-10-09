package webauthn_test

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	"github.com/go-webauthn/webauthn/webauthn"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	cosmowebauthn "github.com/cosmo-workspace/cosmo/pkg/auth/webauthn"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	testUtil "github.com/cosmo-workspace/cosmo/pkg/kosmo/test"
	//+kubebuilder:scaffold:imports
)

var cfg *rest.Config
var k8sClient kosmo.Client
var t testUtil.TestUtil
var testEnv *envtest.Environment

func init() {
	utilruntime.Must(cosmov1alpha1.AddToScheme(scheme.Scheme))
	//+kubebuilder:scaffold:scheme
}

func TestWebAuthn(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WebAuthn Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	k8sClient, err = kosmo.NewClientByRestConfig(cfg, scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	t = testUtil.NewTestUtil(k8sClient)
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("WebAuthn", func() {

	getCredSecret := func(ctx context.Context) (*corev1.Secret, error) {
		var sec corev1.Secret
		err := k8sClient.Get(ctx,
			types.NamespacedName{
				Name:      cosmowebauthn.CredentialSecretName,
				Namespace: cosmov1alpha1.UserNamespace("test-user")}, &sec)
		return &sec, err
	}
	deleteCredSecret := func(ctx context.Context) error {
		var sec corev1.Secret
		sec.SetName(cosmowebauthn.CredentialSecretName)
		sec.SetNamespace(cosmov1alpha1.UserNamespace("test-user"))
		err := k8sClient.Delete(ctx, &sec)
		return err
	}
	var _ = BeforeEach(func() {
		t.CreateCosmoUser("test-user", "test-display", nil, cosmov1alpha1.UserAuthTypePasswordSecert)
		t.CreateUserNameSpaceandDefaultPasswordIfAbsent("test-user")
	})

	var _ = AfterEach(func() {
		ctx := context.Background()
		t.DeleteCosmoUserAll()
		_, err := k8sClient.GetUser(ctx, "test-user")
		Expect(apierrs.IsNotFound(err)).To(BeTrue())
		deleteCredSecret(ctx)
	})

	var _ = It("should get new WebAuthn User", func() {
		ctx := context.Background()
		wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "test-user")
		Expect(err).NotTo(HaveOccurred())
		Expect(wu.WebAuthnID()).To(MatchSnapShot())
		Expect(wu.WebAuthnDisplayName()).To(MatchSnapShot())
		Expect(wu.WebAuthnName()).To(MatchSnapShot())
		Expect(wu.WebAuthnIcon()).To(MatchSnapShot())
		Expect(wu.WebAuthnCredentials()).To(MatchSnapShot())
	})

	var _ = It("should register new WebAuthn credential for new user", func() {
		ctx := context.Background()
		wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "test-user")
		Expect(err).NotTo(HaveOccurred())

		By("listing credentials")
		creds, err := wu.ListCredentials(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(creds).To(MatchSnapShot())

		By("fetching credential secret")
		_, err = getCredSecret(ctx)
		Expect(apierrs.IsNotFound(err)).To(BeTrue())

		cred := cosmowebauthn.Credential{
			DisplayName: "test-cred",
			Timestamp:   time.Date(2022, 4, 20, 21, 0, 0, 0, time.Local).Unix(),
			Cred: webauthn.Credential{
				ID:              []byte("AZK2rgkmjWkwLXkaKVCFdB7zvGelsgOU/dAN8XErN5E1f0NewA3MOEGfN1XfJhiLWZPs22CFOcfXvzB4LWsU0oY="),
				PublicKey:       []byte("pQECAyYgASFYIJvq3cxMy4dzWboxdWDs23t0LooTOsgaqCEobWypEfm4IlgguCfJg35XHVhGI2wh3++cbOSMNC2dqNcOL6U+bj+qJCk="),
				AttestationType: "none",
				Transport:       nil,
				Flags: webauthn.CredentialFlags{
					UserPresent:    true,
					UserVerified:   true,
					BackupEligible: false,
					BackupState:    false,
				},
				Authenticator: webauthn.Authenticator{
					AAGUID:       []byte("AAAAAAAAAAAAAAAAAAAAAA=="),
					SignCount:    0,
					CloneWarning: false,
					Attachment:   "",
				},
			},
		}

		By("registering credential")
		err = wu.RegisterCredential(ctx, &cred)
		Expect(err).NotTo(HaveOccurred())

		By("fetching credential secret again")
		sec, err := getCredSecret(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(ObjectSnapshot(sec)).To(MatchSnapShot())

		data := sec.Data[cosmowebauthn.CredentialListKey]
		Expect(string(data)).To(MatchSnapShot())

		By("list credentials again")
		creds, err = wu.ListCredentials(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(creds)).NotTo(BeZero())
		Expect(creds).To(MatchSnapShot())

		By("get user again")
		wu, err = cosmowebauthn.GetUser(ctx, k8sClient, "test-user")
		Expect(err).NotTo(HaveOccurred())
		Expect(wu.WebAuthnCredentials()).To(MatchSnapShot())
	})

	var _ = It("should be able to update and delete credentials", func() {
		ctx := context.Background()
		wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "test-user")
		Expect(err).NotTo(HaveOccurred())

		cred1 := cosmowebauthn.Credential{
			Base64URLEncodedId: "test-cred1",
			DisplayName:        "test-cred1",
			Timestamp:          time.Date(2022, 4, 20, 21, 0, 0, 0, time.Local).Unix(),
			Cred: webauthn.Credential{
				ID:              []byte("1ZK2rgkmjWkwLXkaKVCFdB7zvGelsgOU/dAN8XErN5E1f0NewA3MOEGfN1XfJhiLWZPs22CFOcfXvzB4LWsU0oY="),
				PublicKey:       []byte("1QECAyYgASFYIJvq3cxMy4dzWboxdWDs23t0LooTOsgaqCEobWypEfm4IlgguCfJg35XHVhGI2wh3++cbOSMNC2dqNcOL6U+bj+qJCk="),
				AttestationType: "none",
				Transport:       nil,
				Flags: webauthn.CredentialFlags{
					UserPresent:    true,
					UserVerified:   true,
					BackupEligible: false,
					BackupState:    false,
				},
				Authenticator: webauthn.Authenticator{
					AAGUID:       []byte("AAAAAAAAAAAAAAAAAAAAAA=="),
					SignCount:    0,
					CloneWarning: false,
					Attachment:   "",
				},
			},
		}
		cred2 := cosmowebauthn.Credential{
			Base64URLEncodedId: "test-cred2",
			DisplayName:        "test-cred2",
			Timestamp:          time.Date(2022, 4, 21, 21, 0, 0, 0, time.Local).Unix(),
			Cred: webauthn.Credential{
				ID:              []byte("2ZK2rgkmjWkwLXkaKVCFdB7zvGelsgOU/dAN8XErN5E1f0NewA3MOEGfN1XfJhiLWZPs22CFOcfXvzB4LWsU0oY="),
				PublicKey:       []byte("2QECAyYgASFYIJvq3cxMy4dzWboxdWDs23t0LooTOsgaqCEobWypEfm4IlgguCfJg35XHVhGI2wh3++cbOSMNC2dqNcOL6U+bj+qJCk="),
				AttestationType: "none",
				Transport:       nil,
				Flags: webauthn.CredentialFlags{
					UserPresent:    true,
					UserVerified:   true,
					BackupEligible: false,
					BackupState:    false,
				},
				Authenticator: webauthn.Authenticator{
					AAGUID:       []byte("AAAAAAAAAAAAAAAAAAAAAA=="),
					SignCount:    0,
					CloneWarning: false,
					Attachment:   "",
				},
			},
		}

		By("registering credentials")
		err = wu.RegisterCredential(ctx, &cred1)
		Expect(err).NotTo(HaveOccurred())
		err = wu.RegisterCredential(ctx, &cred2)
		Expect(err).NotTo(HaveOccurred())

		By("registering credentials again return already exists error")
		err = wu.RegisterCredential(ctx, &cred2)
		Expect(err).To(HaveOccurred())

		By("list credentials")
		creds, err := wu.ListCredentials(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(creds).To(MatchSnapShot())

		By("update cred1 name")
		err = wu.UpdateCredential(ctx, cred1.Base64URLEncodedId, pointer.String("new name"))
		Expect(err).NotTo(HaveOccurred())

		By("remove cred2")
		err = wu.RemoveCredential(ctx, cred2.Base64URLEncodedId)
		Expect(err).NotTo(HaveOccurred())

		By("remove cred2 again returns not found error")
		err = wu.RemoveCredential(ctx, cred2.Base64URLEncodedId)
		Expect(err).To(HaveOccurred())

		By("list credentials again")
		creds, err = wu.ListCredentials(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(creds).To(MatchSnapShot())
	})

	var _ = It("deals with invalid secret", func() {
		ctx := context.Background()
		wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "test-user")
		Expect(err).NotTo(HaveOccurred())

		cred1 := cosmowebauthn.Credential{
			Base64URLEncodedId: "test-cred1",
			DisplayName:        "test-cred1",
			Timestamp:          time.Date(2022, 4, 20, 21, 0, 0, 0, time.Local).Unix(),
			Cred: webauthn.Credential{
				ID:              []byte("1ZK2rgkmjWkwLXkaKVCFdB7zvGelsgOU/dAN8XErN5E1f0NewA3MOEGfN1XfJhiLWZPs22CFOcfXvzB4LWsU0oY="),
				PublicKey:       []byte("1QECAyYgASFYIJvq3cxMy4dzWboxdWDs23t0LooTOsgaqCEobWypEfm4IlgguCfJg35XHVhGI2wh3++cbOSMNC2dqNcOL6U+bj+qJCk="),
				AttestationType: "none",
				Transport:       nil,
				Flags: webauthn.CredentialFlags{
					UserPresent:    true,
					UserVerified:   true,
					BackupEligible: false,
					BackupState:    false,
				},
				Authenticator: webauthn.Authenticator{
					AAGUID:       []byte("AAAAAAAAAAAAAAAAAAAAAA=="),
					SignCount:    0,
					CloneWarning: false,
					Attachment:   "",
				},
			},
		}
		cred2 := cosmowebauthn.Credential{
			Base64URLEncodedId: "test-cred2",
			DisplayName:        "test-cred2",
			Timestamp:          time.Date(2022, 4, 21, 21, 0, 0, 0, time.Local).Unix(),
			Cred: webauthn.Credential{
				ID:              []byte("2ZK2rgkmjWkwLXkaKVCFdB7zvGelsgOU/dAN8XErN5E1f0NewA3MOEGfN1XfJhiLWZPs22CFOcfXvzB4LWsU0oY="),
				PublicKey:       []byte("2QECAyYgASFYIJvq3cxMy4dzWboxdWDs23t0LooTOsgaqCEobWypEfm4IlgguCfJg35XHVhGI2wh3++cbOSMNC2dqNcOL6U+bj+qJCk="),
				AttestationType: "none",
				Transport:       nil,
				Flags: webauthn.CredentialFlags{
					UserPresent:    true,
					UserVerified:   true,
					BackupEligible: false,
					BackupState:    false,
				},
				Authenticator: webauthn.Authenticator{
					AAGUID:       []byte("AAAAAAAAAAAAAAAAAAAAAA=="),
					SignCount:    0,
					CloneWarning: false,
					Attachment:   "",
				},
			},
		}

		By("registering credentials")
		err = wu.RegisterCredential(ctx, &cred1)
		Expect(err).NotTo(HaveOccurred())
		err = wu.RegisterCredential(ctx, &cred2)
		Expect(err).NotTo(HaveOccurred())

		By("update secret directly with invalid json data")
		sec, err := getCredSecret(ctx)
		Expect(err).NotTo(HaveOccurred())
		sec.Data[cosmowebauthn.CredentialListKey] = []byte("invalid data")

		err = k8sClient.Update(ctx, sec)
		Expect(err).NotTo(HaveOccurred())

		By("list credentials returns error")
		_, err = wu.ListCredentials(ctx)
		Expect(err).To(HaveOccurred())

		By("register credential returns error")
		err = wu.RegisterCredential(ctx, &cred2)
		Expect(err).To(HaveOccurred())

		By("remove credential returns error")
		err = wu.RemoveCredential(ctx, cred2.Base64URLEncodedId)
		Expect(err).To(HaveOccurred())
	})

	Describe("[UpdateCredential]", func() {
		registerCredential := func() {
			ctx := context.Background()
			wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "test-user")
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
			wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "test-user")
			Expect(err).NotTo(HaveOccurred())
			wu.RemoveCredential(ctx, "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU")
		}

		run_test := func(wantErr bool, base64urlEncodedCredId string, displayName *string) {

			By("registering credential")
			registerCredential()
			defer removeCredential()

			By("---------------test start----------------")
			ctx := context.Background()
			wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "test-user")
			Expect(err).NotTo(HaveOccurred())

			err = wu.UpdateCredential(ctx, base64urlEncodedCredId, displayName)
			if err == nil {
				Expect(err).NotTo(HaveOccurred())
				Expect(wantErr).To(BeFalse())

			} else {
				Expect(err).To(HaveOccurred())
				Ω(err.Error()).To(MatchSnapShot())
				Expect(wantErr).To(BeTrue())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry("update display name", false, "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU", pointer.String("new display name")),
			Entry("update display name to empty", false, "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU", pointer.String("")),
			Entry("update display name to empty", false, "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU", nil),
			Entry("no change", false, "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU", pointer.String("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36")),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("credential not found", true, "notfound", pointer.String("new display name")),
		)
	})

	Describe("[RemoveCredential]", func() {
		registerCredential := func() {
			ctx := context.Background()
			wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "test-user")
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
			wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "test-user")
			Expect(err).NotTo(HaveOccurred())
			wu.RemoveCredential(ctx, "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU")
		}

		run_test := func(wantErr bool, base64urlEncodedCredId string) {

			By("registering credential")
			registerCredential()
			defer removeCredential()

			By("---------------test start----------------")
			ctx := context.Background()
			wu, err := cosmowebauthn.GetUser(ctx, k8sClient, "test-user")
			Expect(err).NotTo(HaveOccurred())

			err = wu.RemoveCredential(ctx, base64urlEncodedCredId)
			if err == nil {
				Expect(err).NotTo(HaveOccurred())
				Expect(wantErr).To(BeFalse())

			} else {
				Expect(err).To(HaveOccurred())
				Ω(err.Error()).To(MatchSnapShot())
				Expect(wantErr).To(BeTrue())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry("OK", false, "D09Kc9k4zeoxF1Bq1o0ePtUpTnZDOMDMOwQGnXaiqTU"),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("no credential id", true, ""),
			Entry("credential not found", true, "notfound"),
		)
	})
})

func TestCredentials_Default(t *testing.T) {
	type fields struct {
		Base64URLEncodedId string
		DisplayName        string
		Timestamp          int64
		ID                 string
	}
	type args struct {
		now time.Time
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   cosmowebauthn.Credential
	}{
		{
			name: "defaulting all",
			fields: fields{
				ID: "xxxx",
			},
			args: args{
				now: time.Date(2022, 4, 20, 9, 0, 0, 0, time.Local),
			},
			want: cosmowebauthn.Credential{
				Base64URLEncodedId: "eHh4eA",
				DisplayName:        "eHh4eA",
				Timestamp:          1650412800,
				Cred: webauthn.Credential{
					ID: []byte("xxxx"),
				},
			},
		},
		{
			name: "display name already defined",
			fields: fields{
				DisplayName: "defined",
				ID:          "xxxx",
			},
			args: args{
				now: time.Date(2022, 4, 20, 9, 0, 0, 0, time.Local),
			},
			want: cosmowebauthn.Credential{
				Base64URLEncodedId: "eHh4eA",
				DisplayName:        "defined",
				Timestamp:          1650412800,
				Cred: webauthn.Credential{
					ID: []byte("xxxx"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c := cosmowebauthn.Credential{
				Base64URLEncodedId: tt.fields.Base64URLEncodedId,
				DisplayName:        tt.fields.DisplayName,
				Timestamp:          tt.fields.Timestamp,
				Cred: webauthn.Credential{
					ID: []byte(tt.fields.ID),
				},
			}
			c.Default(tt.args.now)
			if !reflect.DeepEqual(c, tt.want) {
				t.Errorf("Credential.Default() = %v, want %v", c, tt.want)
			}
		})
	}
}
