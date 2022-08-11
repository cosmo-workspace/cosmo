package controllers

import (
	"context"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
)

var _ = Describe("Template controller", func() {
	const nsName string = "default"

	tmpl := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod-tmpl1",
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RawYaml: `apiVersion: v1
kind: Pod
metadata:
  name: alpine
spec:
  containers:
  - image: 'alpine:latest'
    name: alpine
    command:
    - echo
    - helloworld
---
apiVersion: v1
kind: Pod
metadata:
  name: alpine2
spec:
  containers:
  - image: 'alpine:latest'
    name: alpine
    command:
    - echo
    - helloworld
`,
		},
	}

	inst := cosmov1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-inst1",
			Namespace: nsName,
		},
		Spec: cosmov1alpha1.InstanceSpec{
			Template: cosmov1alpha1.TemplateRef{
				Name: tmpl.Name,
			},
			Override: cosmov1alpha1.OverrideSpec{},
		},
	}

	Context("when creating Template resource on new cluster", func() {
		It("should do nothing", func() {
			ctx := context.Background()

			By("creating template before instance")

			err := k8sClient.Create(ctx, &tmpl)
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(3 * time.Second)

			var createdTmpl cosmov1alpha1.Template
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: tmpl.Name}, &createdTmpl)
			}, time.Second*10).Should(Succeed())
			Ω(objectSnapshot(&createdTmpl)).To(MatchSnapShot())

			var pod corev1.Pod
			key := client.ObjectKey{
				Name:      instance.InstanceResourceName(inst.Name, "alpine"),
				Namespace: nsName,
			}
			err = k8sClient.Get(ctx, key, &pod)
			Expect(apierrs.IsNotFound(err)).Should(BeTrue())
		})
	})

	Context("when creating Instance resource", func() {
		It("should do instance reconcile and create child resources", func() {
			ctx := context.Background()

			err := k8sClient.Create(ctx, &inst)
			Expect(err).ShouldNot(HaveOccurred())

			By("fetching instance resource and checking if last applied resources added in instance status")

			var createdInst cosmov1alpha1.Instance
			Eventually(func() int {
				key := client.ObjectKey{
					Name:      inst.Name,
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &createdInst)
				Expect(err).ShouldNot(HaveOccurred())
				return createdInst.Status.LastAppliedObjectsCount
			}, time.Second*90).Should(BeEquivalentTo(2))
			Ω(instanceSnapshot(&createdInst)).To(MatchSnapShot())

			By("checking if child resources is as expected in template")
			var pod corev1.Pod
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(inst.Name, "alpine"),
					Namespace: nsName,
				}
				return k8sClient.Get(ctx, key, &pod)
			}, time.Second*10).Should(Succeed())
			Ω(objectSnapshot(&pod)).To(MatchSnapShot())
		})
	})

	Context("when updating Template resource", func() {
		It("should do instance reconcile and update child resources", func() {
			ctx := context.Background()

			var curInst cosmov1alpha1.Instance
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      inst.Name,
					Namespace: nsName,
				}
				return k8sClient.Get(ctx, key, &curInst)
			}, time.Second*10).Should(Succeed())

			// fetch current template
			var updatedTmpl cosmov1alpha1.Template
			Eventually(func() error {
				key := types.NamespacedName{
					Name: tmpl.Name,
				}
				return k8sClient.Get(ctx, key, &updatedTmpl)
			}, time.Second*10).Should(Succeed())
			Ω(objectSnapshot(&updatedTmpl)).To(MatchSnapShot())

			updatedTmpl.Spec.RawYaml = `apiVersion: v1
kind: Pod
metadata:
  name: alpine
spec:
  containers:
  - image: 'alpine:next'
    name: alpine
    command:
    - echo
    - helloworld
`

			// update template
			err := k8sClient.Update(ctx, &updatedTmpl)
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(3 * time.Second)

			var updatedInst cosmov1alpha1.Instance
			Eventually(func() int {
				key := client.ObjectKey{
					Name:      inst.Name,
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &updatedInst)
				Expect(err).ShouldNot(HaveOccurred())
				return updatedInst.Status.LastAppliedObjectsCount
			}, time.Second*90).Should(BeEquivalentTo(2))
			Ω(instanceSnapshot(&updatedInst)).To(MatchSnapShot())

			By("checking if pod updated")
			var pod corev1.Pod
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(inst.Name, "alpine"),
					Namespace: nsName,
				}
				return k8sClient.Get(ctx, key, &pod)
			}, time.Second*10).Should(Succeed())
			Ω(objectSnapshot(&pod)).To(MatchSnapShot())
		})
	})
})
