package controllers

import (
	"context"
	"fmt"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	rbacv1 "k8s.io/api/rbac/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

var _ = Describe("ClusterTemplate controller", func() {
	const name string = "cluster-tmpl-test"
	const crName string = "pod-list-cr"

	tmpl := cosmov1alpha1.ClusterTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				cosmov1alpha1.TemplateAnnKeyDisableNamePrefix: "1",
			},
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RawYaml: fmt.Sprintf(`
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: %s
rules:
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
`, crName),
		},
	}

	Context("when creating ClusterTemplate resource on new cluster", func() {
		It("should do nothing", func() {
			ctx := context.Background()

			By("creating clustertemplate before clusterinstance")

			err := k8sClient.Create(ctx, &tmpl)
			Expect(err).ShouldNot(HaveOccurred())

			var createdTmpl cosmov1alpha1.ClusterTemplate
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: name}, &createdTmpl)
			}, time.Second*10).Should(Succeed())

			var cr rbacv1.ClusterRole
			key := client.ObjectKey{
				Name: crName,
			}
			err = k8sClient.Get(ctx, key, &cr)
			Expect(apierrs.IsNotFound(err)).Should(BeTrue())
		})
	})

	Context("when creating ClusterInstance resource", func() {
		It("should do clusterinstance reconcile and create child resources", func() {
			ctx := context.Background()

			inst := cosmov1alpha1.ClusterInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: name,
					},
					Override: cosmov1alpha1.OverrideSpec{},
				},
			}
			err := k8sClient.Create(ctx, &inst)
			Expect(err).ShouldNot(HaveOccurred())

			By("checking if child resources is as expected in template")

			// ClusterRole
			var cr rbacv1.ClusterRole
			Eventually(func() error {
				key := client.ObjectKey{
					Name: crName,
				}
				return k8sClient.Get(ctx, key, &cr)
			}, time.Second*10).Should(Succeed())
			Ω(ObjectSnapshot(&cr)).To(MatchSnapShot())

			By("fetching instance resource and checking if last applied resources added in instance status")

			var createdInst cosmov1alpha1.ClusterInstance
			Eventually(func() int {
				key := client.ObjectKey{
					Name: name,
				}
				err := k8sClient.Get(ctx, key, &createdInst)
				Expect(err).ShouldNot(HaveOccurred())

				return createdInst.Status.LastAppliedObjectsCount
			}, time.Second*60).Should(BeEquivalentTo(1))
		})
	})

	Context("when updating ClusterTemplate resource", func() {
		It("should do instance reconcile and update child resources", func() {
			ctx := context.Background()

			var curInst cosmov1alpha1.ClusterInstance
			Eventually(func() error {
				key := client.ObjectKey{
					Name: name,
				}
				return k8sClient.Get(ctx, key, &curInst)
			}, time.Second*10).Should(Succeed())

			// fetch current clustertemplate
			var tmpl cosmov1alpha1.ClusterTemplate
			Eventually(func() error {
				key := types.NamespacedName{
					Name: name,
				}
				err := k8sClient.Get(ctx, key, &tmpl)
				Expect(err).ShouldNot(HaveOccurred())

				tmpl.Spec.RawYaml = fmt.Sprintf(`
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: %s
rules:
  - apiGroups:
    - ""
    resources:
      - pods
    verbs:
      - get
      - list
      - watch
      - update # Add
`, crName)
				return k8sClient.Update(ctx, &tmpl)
			}, time.Second*60).Should(Succeed())

			By("checking if clusterrole updated")

			var cr rbacv1.ClusterRole
			Eventually(func() []string {
				key := client.ObjectKey{
					Name: crName,
				}
				err := k8sClient.Get(ctx, key, &cr)
				Expect(err).ShouldNot(HaveOccurred())

				return cr.Rules[0].Verbs
			}, time.Second*60).Should(Equal([]string{"get", "list", "watch", "update"}))
			Ω(ObjectSnapshot(&cr)).To(MatchSnapShot())

			// fetch current clusterinstance
			var updatedInst cosmov1alpha1.ClusterInstance
			Eventually(func() string {
				key := client.ObjectKey{
					Name: name,
				}
				err := k8sClient.Get(ctx, key, &updatedInst)
				Expect(err).ShouldNot(HaveOccurred())
				return updatedInst.Status.TemplateResourceVersion
			}, time.Second*60).ShouldNot(Equal(curInst.Status.TemplateResourceVersion))
			Ω(InstanceSnapshot(&updatedInst)).To(MatchSnapShot())
		})
	})
})
