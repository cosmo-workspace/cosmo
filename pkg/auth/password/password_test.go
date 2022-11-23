package password

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

var _ = Describe("password", func() {
	user1 := &cosmov1alpha1.User{Spec: cosmov1alpha1.UserSpec{}}
	user1.SetName("tom")
	user1.Spec.DisplayName = "tom the cat"

	Context("when reset password for existing user", func() {
		It("should create password secret", func() {
			ctx := context.Background()

			ns := corev1.Namespace{}
			ns.SetName(cosmov1alpha1.UserNamespace(user1.Name))

			err := k8sClient.Create(ctx, &ns)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(func() error {
				var ns corev1.Namespace
				key := client.ObjectKey{Name: cosmov1alpha1.UserNamespace(user1.Name)}
				err := k8sClient.Get(ctx, key, &ns)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())

			err = ResetPassword(ctx, k8sClient, user1.Name)
			Expect(err).ShouldNot(HaveOccurred())

			var secret corev1.Secret
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      cosmov1alpha1.UserPasswordSecretName,
					Namespace: cosmov1alpha1.UserNamespace(user1.Name),
				}
				err := k8sClient.Get(ctx, key, &secret)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())
		})
	})

	Context("when getting password from default password secret", func() {
		newPassword := "New Password"
		It("should return default password with correct password", func() {
			ctx := context.Background()

			// fiest get default password
			defaultPassword, err := GetDefaultPassword(ctx, k8sClient, user1.Name)
			Expect(err).ShouldNot(HaveOccurred())

			err = RegisterPassword(ctx, k8sClient, user1.Name, []byte(newPassword))
			Expect(err).ShouldNot(HaveOccurred())

			// failed to get default password
			_, err = GetDefaultPassword(ctx, k8sClient, user1.Name)
			Expect(err).Should(HaveOccurred())

			// not verified with invalid password
			verified, isDefault, err := VerifyPassword(ctx, k8sClient, user1.Name, []byte(*defaultPassword))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(isDefault).Should(BeFalse())
			Expect(verified).Should(BeFalse())

			// verified with new password
			verified, isDefault, err = VerifyPassword(ctx, k8sClient, user1.Name, []byte(newPassword))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(isDefault).Should(BeFalse())
			Expect(verified).Should(BeTrue())

			By("checking password is unreadable")
			var secret corev1.Secret
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      cosmov1alpha1.UserPasswordSecretName,
					Namespace: cosmov1alpha1.UserNamespace(user1.Name),
				}
				err := k8sClient.Get(ctx, key, &secret)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())

			p, ok := secret.Data[cosmov1alpha1.UserPasswordSecretDataKeyUserPasswordSecret]
			Expect(ok).Should(BeTrue())
			salt, ok := secret.Data[cosmov1alpha1.UserPasswordSecretDataKeyUserPasswordSalt]
			Expect(ok).Should(BeTrue())

			ex, _ := hash([]byte(newPassword), salt)
			Expect(bytesEqual(p, ex)).Should(BeTrue())

			ex, _ = hash([]byte(newPassword), nil)
			Expect(bytesEqual(p, ex)).Should(BeFalse())
		})
	})
})
