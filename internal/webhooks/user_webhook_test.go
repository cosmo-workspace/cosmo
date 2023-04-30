package webhooks

import (
	"context"
	"testing"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/kubeutil/test/gomega"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

var _ = Describe("User webhook", func() {
	normalUserAddon := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "normal-user-addon-test",
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: cosmov1alpha1.TemplateLabelEnumTypeUserAddon,
			},
		},
	}

	defaultUserAddon := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default-user-addon-test",
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: cosmov1alpha1.TemplateLabelEnumTypeUserAddon,
			},
			Annotations: map[string]string{
				cosmov1alpha1.UserAddonTemplateAnnKeyDefaultUserAddon: "true",
			},
		},
	}

	notUserAddon := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "notUserAddonTest",
		},
	}

	Context("when creating user with existing addon", func() {
		It("should pass", func() {
			ctx := context.Background()

			var err error
			err = k8sClient.Create(ctx, &normalUserAddon)
			Expect(err).ShouldNot(HaveOccurred())

			err = k8sClient.Create(ctx, &defaultUserAddon)
			Expect(err).ShouldNot(HaveOccurred())

			user := cosmov1alpha1.User{}
			user.SetName("testuser1")
			user.Spec = cosmov1alpha1.UserSpec{
				AuthType: cosmov1alpha1.UserAuthTypePasswordSecert,
				Addons: []cosmov1alpha1.UserAddon{
					{Template: cosmov1alpha1.UserAddonTemplateRef{Name: defaultUserAddon.GetName()}},
					{Template: cosmov1alpha1.UserAddonTemplateRef{Name: normalUserAddon.GetName()}},
				},
			}

			err = k8sClient.Create(ctx, &user)
			Expect(err).ShouldNot(HaveOccurred())

			var createdUser cosmov1alpha1.User
			Eventually(func() error {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: user.GetName()}, &createdUser)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())
		})
	})

	Context("when creating user with no default addon", func() {
		It("should pass with defaulting", func() {
			ctx := context.Background()

			user := cosmov1alpha1.User{
				TypeMeta: metav1.TypeMeta{
					Kind:       "User",
					APIVersion: cosmov1alpha1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "testuser2",
				},
				Spec: cosmov1alpha1.UserSpec{
					AuthType: cosmov1alpha1.UserAuthTypePasswordSecert,
				},
			}

			expectedUser := user.DeepCopy()
			expectedUser.Spec.Addons = []cosmov1alpha1.UserAddon{
				{Template: cosmov1alpha1.UserAddonTemplateRef{Name: defaultUserAddon.GetName()}},
			}

			err := k8sClient.Create(ctx, &user)
			Expect(err).ShouldNot(HaveOccurred())

			var createdUser cosmov1alpha1.User
			Eventually(func() error {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: user.GetName()}, &createdUser)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())

			expectedUser.ObjectMeta = createdUser.ObjectMeta
			Expect(&createdUser).Should(BeLooseDeepEqual(expectedUser))

		})
	})

	Context("when creating user with non-existing addon", func() {
		It("should deny", func() {
			ctx := context.Background()

			user := cosmov1alpha1.User{}
			user.SetName("testuser3")
			user.Spec = cosmov1alpha1.UserSpec{
				AuthType: cosmov1alpha1.UserAuthTypePasswordSecert,
				Addons: []cosmov1alpha1.UserAddon{
					{Template: cosmov1alpha1.UserAddonTemplateRef{Name: defaultUserAddon.GetName()}},
					{Template: cosmov1alpha1.UserAddonTemplateRef{Name: "notfound"}},
				},
			}
			err := k8sClient.Create(ctx, &user)
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("when creating user with addon template which is not labeled as addon", func() {
		It("should deny", func() {
			ctx := context.Background()

			user := cosmov1alpha1.User{}
			user.SetName("testuser4")
			user.Spec = cosmov1alpha1.UserSpec{
				AuthType: cosmov1alpha1.UserAuthTypePasswordSecert,
				Addons: []cosmov1alpha1.UserAddon{
					{Template: cosmov1alpha1.UserAddonTemplateRef{Name: defaultUserAddon.GetName()}},
					{Template: cosmov1alpha1.UserAddonTemplateRef{Name: notUserAddon.GetName()}},
				},
			}
			err := k8sClient.Create(ctx, &user)
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("when creating user with no auth type", func() {
		It("should pass with defaulting", func() {
			ctx := context.Background()

			user := cosmov1alpha1.User{
				TypeMeta: metav1.TypeMeta{
					Kind:       "User",
					APIVersion: cosmov1alpha1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "testuser5",
				},
				Spec: cosmov1alpha1.UserSpec{
					Addons: []cosmov1alpha1.UserAddon{
						{Template: cosmov1alpha1.UserAddonTemplateRef{Name: defaultUserAddon.GetName()}},
					},
				},
			}

			expectedUser := user.DeepCopy()
			expectedUser.Spec.AuthType = cosmov1alpha1.UserAuthTypePasswordSecert

			err := k8sClient.Create(ctx, &user)
			Expect(err).ShouldNot(HaveOccurred())

			var createdUser cosmov1alpha1.User
			Eventually(func() error {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: user.GetName()}, &createdUser)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())

			expectedUser.ObjectMeta = createdUser.ObjectMeta
			Expect(&createdUser).Should(BeLooseDeepEqual(expectedUser))
		})
	})

	Context("when creating user with innvalid auth type", func() {
		It("should deny", func() {
			ctx := context.Background()

			user := cosmov1alpha1.User{}
			user.SetName("testuser6")
			user.Spec = cosmov1alpha1.UserSpec{
				AuthType: "invalid",
				Addons: []cosmov1alpha1.UserAddon{
					{Template: cosmov1alpha1.UserAddonTemplateRef{Name: defaultUserAddon.GetName()}},
				},
			}
			err := k8sClient.Create(ctx, &user)
			Expect(err).Should(HaveOccurred())
		})
	})
})

func Test_validName(t *testing.T) {
	type args struct {
		v string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "only small alphanumeric",
			args: args{
				v: "hello",
			},
			want: true,
		},
		{
			name: "only small alphanumeric and -",
			args: args{
				v: "hello-world",
			},
			want: true,
		},
		{
			name: "endwith -",
			args: args{
				v: "hello-world-",
			},
			want: false,
		},
		{
			name: "startwith -",
			args: args{
				v: "-hello-world",
			},
			want: false,
		},
		{
			name: "contain .",
			args: args{
				v: "hello.world",
			},
			want: false,
		},
		{
			name: "capital",
			args: args{
				v: "helloWorld",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validName(tt.args.v); got != tt.want {
				t.Errorf("validName() = %v, want %v", got, tt.want)
			}
		})
	}
}
