package controllers

import (
	"context"
	"errors"
	"time"

	"github.com/cosmo-workspace/cosmo/pkg/auth/password"
	. "github.com/cosmo-workspace/cosmo/pkg/kubeutil/test/gomega"
	"github.com/cosmo-workspace/cosmo/pkg/useraddon"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
)

var _ = Describe("User controller", func() {
	namespacedUserAddon := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "namespaced-addon",
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: wsv1alpha1.TemplateTypeUserAddon,
			},
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RawYaml: `apiVersion: batch/v1
kind: Job
metadata:
  labels:
    cosmo/instance: '{{INSTANCE}}'
    cosmo/template: code-server-test
  name: '{{INSTANCE}}-job'
  namespace: '{{NAMESPACE}}'
spec:
  template:
    metadata:
      labels:
        cosmo/instance: '{{INSTANCE}}'
        cosmo/template: '{{TEMPLATE}}'
    spec:
      containers:
      - name: eksctl
        image: weaveworks/eksctl:0.71.0
      restartPolicy: OnFailure
`,
		},
	}

	clusterUserAddon := cosmov1alpha1.ClusterTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-addon",
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: wsv1alpha1.TemplateTypeUserAddon,
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
  namespace: "{{USER_NAMESPACE}}"
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

	Context("when creating User resource", func() {
		It("should do create namespace, password and addons", func() {
			ctx := context.Background()

			By("creating template")

			err := k8sClient.Create(ctx, &namespacedUserAddon)
			Expect(err).ShouldNot(HaveOccurred())

			err = k8sClient.Create(ctx, &clusterUserAddon)
			Expect(err).ShouldNot(HaveOccurred())

			By("creating user")

			user := wsv1alpha1.User{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ua",
				},
				Spec: wsv1alpha1.UserSpec{
					DisplayName: "お名前",
					AuthType:    wsv1alpha1.UserAuthTypePasswordSecert,
					Addons: []wsv1alpha1.UserAddon{
						{
							Template: cosmov1alpha1.TemplateRef{
								Name: namespacedUserAddon.Name,
							},
							Vars: map[string]string{
								"KEY": "VAL",
							},
						},
						{
							Template: cosmov1alpha1.TemplateRef{
								Name:          clusterUserAddon.Name,
								ClusterScoped: true,
							},
						},
					},
				},
			}

			err = k8sClient.Create(ctx, &user)
			Expect(err).ShouldNot(HaveOccurred())

			var createdNs corev1.Namespace
			Eventually(func() error {
				key := client.ObjectKey{
					Name: wsv1alpha1.UserNamespace(user.Name),
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
			}, time.Second*30).Should(BeEquivalentTo(wsv1alpha1.UserNamespace(user.Name)))

			By("check namespace label")

			label := createdNs.GetLabels()
			Expect(label).ShouldNot(BeNil())

			userid, ok := label[wsv1alpha1.NamespaceLabelKeyUserID]
			Expect(ok).Should(BeTrue())
			Expect(userid).Should(BeEquivalentTo(user.Name))

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

})
