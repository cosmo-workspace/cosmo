package password

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

var _ = Describe("password", func() {
	user1 := &wsv1alpha1.User{Spec: wsv1alpha1.UserSpec{}}
	user1.SetName("tom")
	user1.Spec.DisplayName = "tom the cat"

	Context("when reset password for existing user", func() {
		It("should create password secret", func() {
			ctx := clog.LogrIntoContext(context.Background(), log.NullLogger{})

			ns := corev1.Namespace{}
			ns.SetName(wsv1alpha1.UserNamespace(user1.Name))

			err := k8sClient.Create(ctx, &ns)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(func() error {
				var ns corev1.Namespace
				key := client.ObjectKey{Name: wsv1alpha1.UserNamespace(user1.Name)}
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
					Name:      wsv1alpha1.UserPasswordSecretName,
					Namespace: wsv1alpha1.UserNamespace(user1.Name),
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
			ctx := clog.LogrIntoContext(context.Background(), log.NullLogger{})

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
					Name:      wsv1alpha1.UserPasswordSecretName,
					Namespace: wsv1alpha1.UserNamespace(user1.Name),
				}
				err := k8sClient.Get(ctx, key, &secret)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())

			p, ok := secret.Data[wsv1alpha1.UserPasswordSecretDataKeyUserPasswordSecret]
			Expect(ok).Should(BeTrue())
			salt, ok := secret.Data[wsv1alpha1.UserPasswordSecretDataKeyUserPasswordSalt]
			Expect(ok).Should(BeTrue())

			ex, _ := hash([]byte(newPassword), salt)
			Expect(bytesEqual(p, ex)).Should(BeTrue())

			ex, _ = hash([]byte(newPassword), nil)
			Expect(bytesEqual(p, ex)).Should(BeFalse())
		})
	})
})
