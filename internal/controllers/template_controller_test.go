package controllers

import (
	"context"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/kubeutil/test/gomega"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
)

var _ = Describe("Template controller", func() {
	const name string = "tmpl-test"
	const nsName string = "default"

	tmpl := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
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
`,
		},
	}

	expectedPodApply := func(ownerRef metav1.OwnerReference) *corev1apply.PodApplyConfiguration {
		return corev1apply.Pod(instance.InstanceResourceName(name, "alpine"), nsName).
			WithAPIVersion("v1").
			WithKind("Pod").
			WithLabels(map[string]string{
				cosmov1alpha1.LabelKeyInstance: name,
				cosmov1alpha1.LabelKeyTemplate: name,
			}).
			WithOwnerReferences(
				metav1apply.OwnerReference().
					WithAPIVersion(ownerRef.APIVersion).
					WithBlockOwnerDeletion(*ownerRef.BlockOwnerDeletion).
					WithController(*ownerRef.Controller).
					WithKind(ownerRef.Kind).
					WithName(ownerRef.Name).
					WithUID(ownerRef.UID),
			).
			WithSpec(corev1apply.PodSpec().
				WithContainers(corev1apply.Container().
					WithName("alpine").
					WithImage("alpine:latest").
					WithCommand("echo", "helloworld")))
	}

	Context("when creating Template resource on new cluster", func() {
		It("should do nothing", func() {
			ctx := context.Background()

			By("creating template before instance")

			err := k8sClient.Create(ctx, &tmpl)
			Expect(err).ShouldNot(HaveOccurred())

			var createdTmpl cosmov1alpha1.Template
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: name}, &createdTmpl)
			}, time.Second*10).Should(Succeed())

			var pod corev1.Pod
			key := client.ObjectKey{
				Name:      instance.InstanceResourceName(name, "alpine"),
				Namespace: nsName,
			}
			err = k8sClient.Get(ctx, key, &pod)
			Expect(apierrs.IsNotFound(err)).Should(BeTrue())
		})
	})

	Context("when creating Instance resource", func() {
		It("should do instance reconcile and create child resources", func() {
			ctx := context.Background()

			inst := cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: nsName,
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

			By("fetching instance resource and checking if last applied resources added in instance status")

			var createdInst cosmov1alpha1.Instance
			Eventually(func() int {
				key := client.ObjectKey{
					Name:      name,
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &createdInst)
				if err != nil {
					return 0
				}
				return createdInst.Status.LastAppliedObjectsCount
			}, time.Second*90).Should(BeEquivalentTo(1))

			By("checking if child resources is as expected in template")

			instOwnerRef := ownerRef(&inst, scheme.Scheme)

			// Pod
			var pod corev1.Pod
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(name, "alpine"),
					Namespace: nsName,
				}
				return k8sClient.Get(ctx, key, &pod)
			}, time.Second*10).Should(Succeed())

			podApplyCfg, err := corev1apply.ExtractPod(&pod, controllerFieldManager)
			Expect(err).ShouldNot(HaveOccurred())

			expectedPodApplyCfg := expectedPodApply(instOwnerRef)
			Expect(podApplyCfg).Should(BeEqualityDeepEqual(expectedPodApplyCfg))

			pod.SetGroupVersionKind(schema.FromAPIVersionAndKind(*podApplyCfg.APIVersion, *podApplyCfg.Kind))
			Expect(instance.ExistInLastApplyed(&createdInst, &pod)).Should(BeTrue())
		})
	})

	Context("when updating Template resource", func() {
		It("should do instance reconcile and update child resources", func() {
			ctx := context.Background()

			var inst cosmov1alpha1.Instance
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      name,
					Namespace: nsName,
				}
				return k8sClient.Get(ctx, key, &inst)
			}, time.Second*10).Should(Succeed())

			// fetch current template
			var tmpl cosmov1alpha1.Template
			Eventually(func() error {
				key := types.NamespacedName{
					Name: name,
				}
				return k8sClient.Get(ctx, key, &tmpl)
			}, time.Second*10).Should(Succeed())

			tmpl.Spec.RawYaml = `apiVersion: v1
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
			err := k8sClient.Update(ctx, &tmpl)
			Expect(err).ShouldNot(HaveOccurred())

			By("checking if child resources updated")

			instOwnerRef := ownerRef(&inst, scheme.Scheme)

			expectedPodApplyCfg := expectedPodApply(instOwnerRef)
			expectedPodApplyCfg.Spec.Containers[0].WithImage("alpine:next")

			By("checking if pod updated")

			var pod corev1.Pod
			Eventually(func() *corev1apply.PodApplyConfiguration {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(name, "alpine"),
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &pod)
				Expect(err).ShouldNot(HaveOccurred())

				podApplyCfg, err := corev1apply.ExtractPod(&pod, controllerFieldManager)
				Expect(err).ShouldNot(HaveOccurred())

				return podApplyCfg
			}, time.Second*10).Should(BeEqualityDeepEqual(expectedPodApplyCfg))
		})
	})
})
