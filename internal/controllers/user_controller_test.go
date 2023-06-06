package controllers

import (
	"context"
	"errors"
	"fmt"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/kubeutil/test/gomega"
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/auth/password"
	"github.com/cosmo-workspace/cosmo/pkg/useraddon"
)

var _ = Describe("User controller", func() {
	namespacedUserAddon := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "namespaced-addon",
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: cosmov1alpha1.TemplateLabelEnumTypeUserAddon,
			},
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RawYaml: `apiVersion: batch/v1
kind: Job
metadata:
  labels:
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: code-server-test
  name: '{{INSTANCE}}-job'
  namespace: '{{NAMESPACE}}'
spec: 
  template:
    metadata:
      labels:
        cosmo-workspace.github.io/instance: '{{INSTANCE}}'
        cosmo-workspace.github.io/template: '{{TEMPLATE}}'
    spec:
      containers:
      - name: eksctl
        image: weaveworks/eksctl:{{IMAGE_TAG}}
      restartPolicy: OnFailure
`,
			RequiredVars: []cosmov1alpha1.RequiredVarSpec{{Var: "IMAGE_TAG"}},
		},
	}

	clusterUserAddon := cosmov1alpha1.ClusterTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-addon",
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: cosmov1alpha1.TemplateLabelEnumTypeUserAddon,
			},
			Annotations: map[string]string{
				cosmov1alpha1.TemplateAnnKeyDisableNamePrefix: "1",
			},
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RawYaml: `apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv0001
spec:
  capacity:
    storage: 1Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Delete
  storageClassName: slow
  hostPath:
    path: /data/pv0001
    type: DirectoryOrCreate
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pv-slow-claim
  namespace: "{{NAMESPACE}}"
spec:
  accessModes:
    - ReadWriteOnce
  volumeMode: Filesystem
  resources:
    requests:
      storage: 1Gi
  storageClassName: slow
`,
		},
	}

	emptyUserAddon := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "empty-addon",
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: cosmov1alpha1.TemplateLabelEnumTypeUserAddon,
			},
		},
	}

	Context("when creating User resource", func() {

		BeforeEach(func() {
			ctx := context.Background()

			By("creating template")
			addon := namespacedUserAddon.DeepCopy()
			err := k8sClient.Create(ctx, addon)
			Expect(err).ShouldNot(HaveOccurred())

			clusterAddon := clusterUserAddon.DeepCopy()
			err = k8sClient.Create(ctx, clusterAddon)
			Expect(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			By("delete template")
			addon := namespacedUserAddon.DeepCopy()
			err := k8sClient.Delete(ctx, addon)
			Expect(err).ShouldNot(HaveOccurred())
			clusterAddon := clusterUserAddon.DeepCopy()
			err = k8sClient.Delete(ctx, clusterAddon)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should do create namespace, password and addons", func() {

			By("creating user")

			user := cosmov1alpha1.User{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ua",
				},
				Spec: cosmov1alpha1.UserSpec{
					DisplayName: "お名前",
					AuthType:    cosmov1alpha1.UserAuthTypePasswordSecert,
					Addons: []cosmov1alpha1.UserAddon{
						{
							Template: cosmov1alpha1.UserAddonTemplateRef{
								Name: namespacedUserAddon.Name,
							},
							Vars: map[string]string{
								"IMAGE_TAG": "v0.71.0",
							},
						},
						{
							Template: cosmov1alpha1.UserAddonTemplateRef{
								Name:          clusterUserAddon.Name,
								ClusterScoped: true,
							},
						},
					},
				},
			}

			err := k8sClient.Create(ctx, &user)
			Expect(err).ShouldNot(HaveOccurred())

			var createdNs corev1.Namespace
			Eventually(func() error {
				key := client.ObjectKey{
					Name: cosmov1alpha1.UserNamespace(user.Name),
				}
				return k8sClient.Get(ctx, key, &createdNs)
			}, time.Second*10).Should(Succeed())

			Eventually(func() error {
				key := client.ObjectKey{
					Name: user.Name,
				}
				err := k8sClient.Get(ctx, key, &user)
				Expect(err).ShouldNot(HaveOccurred())

				if user.Status.Namespace.Name == "" {
					return fmt.Errorf("user namespace is empty")
				}
				if len(user.Status.Addons) != 2 {
					return fmt.Errorf("user addon count is not 2: %d", len(user.Status.Addons))
				}
				return nil
			}, time.Second*30).ShouldNot(HaveOccurred())
			Expect(UserSnapshot(&user)).Should(MatchSnapShot())

			By("check namespace label")

			label := createdNs.GetLabels()
			Expect(label).ShouldNot(BeNil())

			username, ok := label[cosmov1alpha1.NamespaceLabelKeyUserName]
			Expect(ok).Should(BeTrue())
			Expect(username).Should(BeEquivalentTo(user.Name))

			By("check namespace owner reference")

			ownerref := ownerRef(&user, scheme.Scheme)
			Expect(createdNs.OwnerReferences).Should(BeEqualityDeepEqual([]metav1.OwnerReference{ownerref}))

			By("check user's namespace reference")

			Expect(user.Status.Namespace.Name).Should(BeEquivalentTo(createdNs.GetName()))
			Expect(user.Status.Namespace.UID).Should(BeEquivalentTo(createdNs.GetUID()))
			Expect(user.Status.Namespace.ResourceVersion).Should(BeEquivalentTo(createdNs.GetResourceVersion()))

			By("check password secret is created")

			Eventually(func() error {
				_, err := password.GetDefaultPassword(ctx, k8sClient, user.Name)
				return err
			}, time.Second*10).Should(Succeed())

			By("check addon instance is created")

			Eventually(func() error {
				var addonInst cosmov1alpha1.Instance
				key := client.ObjectKey{
					Name:      useraddon.InstanceName(namespacedUserAddon.Name, ""),
					Namespace: createdNs.GetName(),
				}
				err := k8sClient.Get(ctx, key, &addonInst)
				if err != nil {
					return err
				}
				if addonInst.Spec.Template.Name != namespacedUserAddon.Name {
					return errors.New("invalid template name")
				}
				if equality.Semantic.DeepEqual(addonInst.Spec.Vars, user.Spec.Addons[0].Vars) {
					return errors.New("invalid template name")
				}
				return k8sClient.Get(ctx, client.ObjectKey{Name: user.Name}, &user)
			}, time.Second*10).Should(Succeed())

			Eventually(func() error {
				var clusterAddonInst cosmov1alpha1.ClusterInstance
				key := client.ObjectKey{
					Name: useraddon.InstanceName(clusterUserAddon.Name, user.GetName()),
				}
				err := k8sClient.Get(ctx, key, &clusterAddonInst)
				if err != nil {
					return err
				}
				if clusterAddonInst.Spec.Template.Name != clusterUserAddon.Name {
					return errors.New("invalid template name")
				}
				return nil
			}, time.Second*10).Should(Succeed())
		})

		It("should do create namespace and addons when authtype is ldap", func() {

			By("creating user")

			user := cosmov1alpha1.User{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ualdap",
				},
				Spec: cosmov1alpha1.UserSpec{
					DisplayName: "お名前",
					AuthType:    cosmov1alpha1.UserAuthTypeLDAP,
					Addons: []cosmov1alpha1.UserAddon{
						{
							Template: cosmov1alpha1.UserAddonTemplateRef{
								Name: namespacedUserAddon.Name,
							},
							Vars: map[string]string{
								"KEY": "VAL",
							},
						},
						{
							Template: cosmov1alpha1.UserAddonTemplateRef{
								Name:          clusterUserAddon.Name,
								ClusterScoped: true,
							},
						},
					},
				},
			}

			err := k8sClient.Create(ctx, &user)
			Expect(err).ShouldNot(HaveOccurred())

			var createdNs corev1.Namespace
			Eventually(func() error {
				key := client.ObjectKey{
					Name: cosmov1alpha1.UserNamespace(user.Name),
				}
				return k8sClient.Get(ctx, key, &createdNs)
			}, time.Second*10).Should(Succeed())

			Eventually(func() string {
				key := client.ObjectKey{
					Name: user.Name,
				}
				err := k8sClient.Get(ctx, key, &user)
				Expect(err).ShouldNot(HaveOccurred())

				return user.Status.Namespace.Name
			}, time.Second*30).Should(BeEquivalentTo(cosmov1alpha1.UserNamespace(user.Name)))

			By("check namespace label")

			label := createdNs.GetLabels()
			Expect(label).ShouldNot(BeNil())

			username, ok := label[cosmov1alpha1.NamespaceLabelKeyUserName]
			Expect(ok).Should(BeTrue())
			Expect(username).Should(BeEquivalentTo(user.Name))

			By("check namespace owner reference")

			ownerref := ownerRef(&user, scheme.Scheme)
			Expect(createdNs.OwnerReferences).Should(BeEqualityDeepEqual([]metav1.OwnerReference{ownerref}))

			By("check user's namespace reference")

			Expect(user.Status.Namespace.Name).Should(BeEquivalentTo(createdNs.GetName()))
			Expect(user.Status.Namespace.UID).Should(BeEquivalentTo(createdNs.GetUID()))
			Expect(user.Status.Namespace.ResourceVersion).Should(BeEquivalentTo(createdNs.GetResourceVersion()))

			By("check password secret is not created")

			Eventually(func() error {
				_, err := password.GetDefaultPassword(ctx, k8sClient, user.Name)
				return err
			}, time.Second*5).ShouldNot(Succeed())

			By("check addon instance is created")

			Eventually(func() error {
				var addonInst cosmov1alpha1.Instance
				key := client.ObjectKey{
					Name:      useraddon.InstanceName(namespacedUserAddon.Name, ""),
					Namespace: createdNs.GetName(),
				}
				err := k8sClient.Get(ctx, key, &addonInst)
				if err != nil {
					return err
				}
				if addonInst.Spec.Template.Name != namespacedUserAddon.Name {
					return errors.New("invalid template name")
				}
				if equality.Semantic.DeepEqual(addonInst.Spec.Vars, user.Spec.Addons[0].Vars) {
					return errors.New("invalid template name")
				}
				return nil
			}, time.Second*10).Should(Succeed())

			Eventually(func() error {
				var clusterAddonInst cosmov1alpha1.ClusterInstance
				key := client.ObjectKey{
					Name: useraddon.InstanceName(clusterUserAddon.Name, user.GetName()),
				}
				err := k8sClient.Get(ctx, key, &clusterAddonInst)
				if err != nil {
					return err
				}
				if clusterAddonInst.Spec.Template.Name != clusterUserAddon.Name {
					return errors.New("invalid template name")
				}
				return nil
			}, time.Second*10).Should(Succeed())
		})
	})

	Context("when updating user addon with invalid addon", func() {
		It("should try to create addon but status AddonFailed", func() {
			ctx := context.Background()

			By("creating invalid template")

			err := k8sClient.Create(ctx, &emptyUserAddon)
			Expect(err).ShouldNot(HaveOccurred())

			By("fetching and update user")
			var user cosmov1alpha1.User
			Eventually(func() error {
				err = k8sClient.Get(ctx, client.ObjectKey{Name: "ua"}, &user)
				Expect(err).ShouldNot(HaveOccurred())
				user.Spec.Addons = append(user.Spec.Addons, cosmov1alpha1.UserAddon{
					Template: cosmov1alpha1.UserAddonTemplateRef{
						Name: emptyUserAddon.Name,
					},
				})
				return k8sClient.Update(ctx, &user)
			}, time.Second*30).Should(Succeed())
			Expect(UserSnapshot(&user)).Should(MatchSnapShot())

			var updatedUser cosmov1alpha1.User
			Eventually(func() int {
				err = k8sClient.Get(ctx, client.ObjectKey{Name: "ua"}, &updatedUser)
				Expect(err).ShouldNot(HaveOccurred())
				return len(updatedUser.Status.Addons)
			}, time.Second*30).Should(Equal(3))
			Expect(UserSnapshot(&updatedUser)).Should(MatchSnapShot())

		})
	})

})
