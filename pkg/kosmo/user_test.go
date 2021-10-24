package kosmo

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

var _ = Describe("Client", func() {
	user1 := &wsv1alpha1.User{
		ID:          "tom",
		DisplayName: "tom the cat",
	}
	user2 := &wsv1alpha1.User{
		ID:          "jry",
		DisplayName: "jry the mouse",
		Role:        "cosmo-admin",
	}

	Describe("CreateUser", func() {
		Context("when create new user", func() {
			It("should create user namespace and password secret", func() {
				ctx := clog.LogrIntoContext(context.Background(), testLogger)
				c := NewClient(k8sClient)
				Expect(c.Client).ShouldNot(BeNil())

				var currentns corev1.Namespace
				key := client.ObjectKey{
					Name: wsv1alpha1.UserNamespace(user1.ID),
				}
				err := c.Get(ctx, key, &currentns)
				Expect(apierrs.IsNotFound(err)).Should(BeTrue())

				// create cosmo-auth-proxy-rolebinding
				authProxyCRB := &rbacv1.ClusterRoleBinding{}
				authProxyCRB.SetName(wsv1alpha1.AuthProxyClusterRoleBindingName)
				authProxyCRB.RoleRef = rbacv1.RoleRef{
					APIGroup: rbacv1.GroupName,
					Kind:     "ClusterRole",
					Name:     wsv1alpha1.AuthProxyClusterRoleBindingName,
				}
				err = c.Create(ctx, authProxyCRB)
				Expect(err).ShouldNot(HaveOccurred())

				user1, err = c.CreateUser(ctx, user1)
				Expect(err).ShouldNot(HaveOccurred())

				var ns corev1.Namespace
				Eventually(func() error {
					key := client.ObjectKey{
						Name: wsv1alpha1.UserNamespace(user1.ID),
					}
					err := k8sClient.Get(ctx, key, &ns)
					if err != nil {
						return err
					}
					return nil
				}, time.Second*10).Should(Succeed())

				var secret corev1.Secret
				Eventually(func() error {
					key := client.ObjectKey{
						Name:      wsv1alpha1.UserPasswordSecretName,
						Namespace: wsv1alpha1.UserNamespace(user1.ID),
					}
					err := k8sClient.Get(ctx, key, &secret)
					if err != nil {
						return err
					}
					return nil
				}, time.Second*10).Should(Succeed())
			})
		})
	})

	Describe("GetDefaultPassword and VerifyPassword", func() {
		Context("when getting password from default password secret", func() {
			It("should return default password with correct password", func() {
				ctx := clog.LogrIntoContext(context.Background(), testLogger)
				c := NewClient(k8sClient)
				Expect(c.Client).ShouldNot(BeNil())

				defaultPassword, err := c.GetDefaultPassword(ctx, user1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				verified, isDefault, err := c.VerifyPassword(ctx, user1.ID, []byte(*defaultPassword))
				Expect(err).ShouldNot(HaveOccurred())
				Expect(isDefault).Should(BeTrue())
				Expect(verified).Should(BeTrue())
			})

			It("should not return default password with invalid password", func() {
				ctx := clog.LogrIntoContext(context.Background(), testLogger)
				c := NewClient(k8sClient)
				Expect(c.Client).ShouldNot(BeNil())

				verified, isDefault, err := c.VerifyPassword(ctx, user1.ID, []byte("invalid"))
				Expect(err).ShouldNot(HaveOccurred())
				Expect(isDefault).Should(BeTrue())
				Expect(verified).Should(BeFalse())
			})

			It("should not return default password if user not found", func() {
				ctx := clog.LogrIntoContext(context.Background(), testLogger)
				c := NewClient(k8sClient)
				Expect(c.Client).ShouldNot(BeNil())

				verified, isDefault, err := c.VerifyPassword(ctx, "notfound", []byte("invalid"))
				Expect(err).Should(HaveOccurred())
				Expect(isDefault).Should(BeFalse())
				Expect(verified).Should(BeFalse())
			})
		})
	})

	Describe("RegisterPassword and VerifyPassword", func() {
		Context("when getting password from default password secret", func() {
			newPassword := "New Password"
			It("should return default password with correct password", func() {
				ctx := clog.LogrIntoContext(context.Background(), testLogger)
				c := NewClient(k8sClient)
				Expect(c.Client).ShouldNot(BeNil())

				// fiest get default password
				defaultPassword, err := c.GetDefaultPassword(ctx, user1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				err = c.RegisterPassword(ctx, user1.ID, []byte(newPassword))
				Expect(err).ShouldNot(HaveOccurred())

				// failed to get default password
				_, err = c.GetDefaultPassword(ctx, user1.ID)
				Expect(err).Should(HaveOccurred())

				// not verified with invalid password
				verified, isDefault, err := c.VerifyPassword(ctx, user1.ID, []byte(*defaultPassword))
				Expect(err).ShouldNot(HaveOccurred())
				Expect(isDefault).Should(BeFalse())
				Expect(verified).Should(BeFalse())

				// verified with new password
				verified, isDefault, err = c.VerifyPassword(ctx, user1.ID, []byte(newPassword))
				Expect(err).ShouldNot(HaveOccurred())
				Expect(isDefault).Should(BeFalse())
				Expect(verified).Should(BeTrue())

				By("checking password is unreadable")
				var secret corev1.Secret
				Eventually(func() error {
					key := client.ObjectKey{
						Name:      wsv1alpha1.UserPasswordSecretName,
						Namespace: wsv1alpha1.UserNamespace(user1.ID),
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
				Expect(BytesEqual(p, ex)).Should(BeTrue())

				ex, _ = hash([]byte(newPassword), nil)
				Expect(BytesEqual(p, ex)).Should(BeFalse())
			})
		})
	})

	Describe("GetUser", func() {
		Context("when getting user", func() {
			It("should return user if found", func() {
				ctx := clog.LogrIntoContext(context.Background(), testLogger)
				c := NewClient(k8sClient)
				Expect(c.Client).ShouldNot(BeNil())

				user, err := c.GetUser(ctx, user1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				eq := equality.Semantic.DeepEqual(user, user1)
				Expect(eq).Should(BeTrue())
			})

			It("should not return user if not found", func() {
				ctx := clog.LogrIntoContext(context.Background(), testLogger)
				c := NewClient(k8sClient)
				Expect(c.Client).ShouldNot(BeNil())

				_, err := c.GetUser(ctx, "notfound")
				Expect(err).Should(HaveOccurred())
			})
		})
	})

	Describe("ListUser", func() {
		Context("when getting users", func() {
			It("should return all users in cluster", func() {
				ctx := clog.LogrIntoContext(context.Background(), testLogger)
				c := NewClient(k8sClient)
				Expect(c.Client).ShouldNot(BeNil())

				// create user2
				var err error
				user2, err = c.CreateUser(ctx, user2)
				Expect(err).ShouldNot(HaveOccurred())

				users, err := c.ListUsers(ctx)
				Expect(err).ShouldNot(HaveOccurred())

				expected := []wsv1alpha1.User{*user1, *user2}

				sort.Slice(users, func(i, j int) bool { return users[i].ID < users[j].ID })
				sort.Slice(expected, func(i, j int) bool { return expected[i].ID < expected[j].ID })

				eq := equality.Semantic.DeepEqual(users, expected)
				Expect(eq).Should(BeTrue())
			})
		})
	})

	Describe("UpdateUser", func() {
		Context("when updating user info", func() {
			It("should update user in cluster", func() {
				ctx := clog.LogrIntoContext(context.Background(), testLogger)
				c := NewClient(k8sClient)
				Expect(c.Client).ShouldNot(BeNil())

				// fetch user namespace
				user1ns, err := c.GetUserNamespace(ctx, user1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				updated := user1.DeepCopy()
				updated.DisplayName = "updated"
				updated.Role = wsv1alpha1.UserAdminRole

				_, err = c.UpdateUser(ctx, updated)
				Expect(err).ShouldNot(HaveOccurred())

				// fetch user namespace again
				updatedNs, err := c.GetUserNamespace(ctx, user1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				expectedNs := user1ns.DeepCopy()
				expectedNs.Annotations[wsv1alpha1.NamespaceAnnKeyUserName] = "updated"
				expectedNs.Annotations[wsv1alpha1.NamespaceAnnKeyUserRole] = wsv1alpha1.UserAdminRole.String()

				eq := LooseDeepEqual(updatedNs.DeepCopy(), expectedNs.DeepCopy(), WithPrintDiff())
				Expect(eq).Should(BeTrue())
			})
		})
	})

	Describe("DeleteUser", func() {
		Context("when deleting user", func() {
			It("should delete user namespace", func() {
				ctx := clog.LogrIntoContext(context.Background(), testLogger)
				c := NewClient(k8sClient)
				Expect(c.Client).ShouldNot(BeNil())

				// fetch user namespace
				_, err := c.GetUserNamespace(ctx, user1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				_, err = c.DeleteUser(ctx, user1.ID)
				Expect(err).ShouldNot(HaveOccurred())

				Eventually(func() error {
					// fetch user namespace again
					ns, err := c.GetUserNamespace(ctx, user1.ID)
					if apierrs.IsNotFound(err) {
						return nil
					} else if err != nil {
						return err
					}

					if ns.Status.Phase == corev1.NamespaceTerminating {
						return nil
					}
					return fmt.Errorf("still exist")
				}, time.Second*30).Should(Succeed())
			})
		})
	})

})

var testLogger = &TestLogger{Out: os.Stderr}

type TestLogger struct {
	Out io.Writer
}

func (l *TestLogger) Enabled() bool {
	return l.Out != nil
}
func (l *TestLogger) Info(msg string, keysAndValues ...interface{}) {
	fmt.Fprint(l.Out, "INFO", msg, keysAndValues)
}
func (l *TestLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	fmt.Fprint(l.Out, "ERROR", msg, keysAndValues)
}
func (l *TestLogger) V(level int) logr.Logger {
	return l
}
func (l *TestLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return l
}
func (l *TestLogger) WithName(name string) logr.Logger {
	return l
}
