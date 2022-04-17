package controllers

import (
	"context"
	"errors"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
)

var _ = Describe("Template controller", func() {
	const tmplName string = "alpine"
	const instName string = "tmpl-test-inst"
	const nsName string = "default"
	var instOwnerRef metav1.OwnerReference

	tmpl := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: tmplName,
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

	expectedPodApply := func(instName, namespace string, ownerRef metav1.OwnerReference) *corev1apply.PodApplyConfiguration {
		return corev1apply.Pod(instance.InstanceResourceName(instName, "alpine"), namespace).
			WithAPIVersion("v1").
			WithKind("Pod").
			WithLabels(map[string]string{
				cosmov1alpha1.LabelKeyInstance: instName,
				cosmov1alpha1.LabelKeyTemplate: "alpine",
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
				err := k8sClient.Get(ctx, client.ObjectKey{Name: tmplName}, &createdTmpl)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())
		})
	})

	Context("when creating Instance resource", func() {
		It("should do instance reconcile and create child resources", func() {
			ctx := context.Background()

			inst := cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      instName,
					Namespace: nsName,
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: tmplName,
					},
					Override: cosmov1alpha1.OverrideSpec{},
				},
			}
			err := k8sClient.Create(ctx, &inst)
			Expect(err).ShouldNot(HaveOccurred())

			By("fetching instance resource and checking if last applied resources added in instance status")

			var createdInst cosmov1alpha1.Instance
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instName,
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &createdInst)
				if err != nil {
					return err
				}
				if len(createdInst.Status.LastApplied) == 0 {
					return errors.New("child resources still not created")
				}
				return nil
			}, time.Second*30).Should(Succeed())

			By("checking if child resources is as expected in template")

			instOwnerRef = ownerRef(&inst, scheme.Scheme)

			// Pod
			var pod corev1.Pod
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(instName, "alpine"),
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &pod)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())

			podApplyCfg, err := corev1apply.ExtractPod(&pod, InstControllerFieldManager)
			Expect(err).ShouldNot(HaveOccurred())

			expectedPodApplyCfg := expectedPodApply(instName, nsName, instOwnerRef)

			clog.PrintObjectDiff(os.Stderr, podApplyCfg, expectedPodApplyCfg)
			eq := equality.Semantic.DeepEqual(podApplyCfg, expectedPodApplyCfg)
			Expect(eq).Should(BeTrue())

			pod.SetGroupVersionKind(schema.FromAPIVersionAndKind("v1", "Pod"))
			Expect(instance.ExistInLastApplyed(createdInst, &pod)).Should(BeTrue())
		})
	})

	Context("when updating Template resource", func() {
		It("should do instance reconcile and update child resources", func() {
			ctx := context.Background()

			// fetch current template
			var tmpl cosmov1alpha1.Template
			Eventually(func() error {
				key := types.NamespacedName{
					Name: tmplName,
				}
				err := k8sClient.Get(ctx, key, &tmpl)
				if err != nil {
					return err
				}
				return nil
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

			expectedPodApplyCfg := expectedPodApply(instName, nsName, instOwnerRef)
			expectedPodApplyCfg.Spec.Containers[0].WithImage("alpine:next")

			By("checking if pod updated")

			var pod corev1.Pod
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(instName, "alpine"),
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &pod)
				if err != nil {
					return err
				}

				podApplyCfg, err := corev1apply.ExtractPod(&pod, InstControllerFieldManager)
				Expect(err).ShouldNot(HaveOccurred())

				clog.PrintObjectDiff(os.Stderr, podApplyCfg, expectedPodApplyCfg)
				eq := equality.Semantic.DeepEqual(podApplyCfg, expectedPodApplyCfg)
				if !eq {
					return errors.New("not equal")
				}
				return nil
			}, time.Second*10).Should(Succeed())
		})
	})
})
