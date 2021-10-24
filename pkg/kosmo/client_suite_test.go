package kosmo

import (
	"context"
	"os"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/template"
)

var _ = Describe("Client", func() {
	Describe("DryrunApply", func() {
		Context("when apply new object", func() {
			It("should create no object and dryrun succeed", func() {
				ctx := context.Background()

				c := NewClient(k8sClient)

				_, err := c.GetUnstructured(ctx, schema.GroupVersionKind{Version: "v1", Kind: "Pod"}, "nginx", "default")
				Expect(apierrs.IsNotFound(err)).Should(BeTrue())

				applyObjStr := `apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: default
spec:
  containers:
  - image: 'nginx:latest'
    name: nginx
`

				_, applyObj, err := template.StringToUnstructured(applyObjStr)
				Expect(err).ShouldNot(HaveOccurred())

				_, err = c.Apply(ctx, applyObj, "test-controller", true, true)
				Expect(err).ShouldNot(HaveOccurred())

				time.Sleep(time.Second * 3)

				_, err = c.GetUnstructured(ctx, schema.GroupVersionKind{Version: "v1", Kind: "Pod"}, "nginx", "default")
				Expect(apierrs.IsNotFound(err)).Should(BeTrue())
			})
		})
	})

	Describe("Apply", func() {
		Context("when apply new object", func() {
			It("should create new object", func() {
				ctx := context.Background()

				c, err := NewClientByRestConfig(cfg, scheme.Scheme)
				Expect(err).ShouldNot(HaveOccurred())

				_, err = c.GetUnstructured(ctx, schema.GroupVersionKind{Version: "v1", Kind: "Pod"}, "nginx", "default")
				Expect(apierrs.IsNotFound(err)).Should(BeTrue())

				applyObjStr := `apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: default
spec:
  containers:
  - image: 'nginx:latest'
    name: nginx
`
				expectedPodApplyCfg := corev1apply.Pod("nginx", "default").
					WithAPIVersion("v1").
					WithKind("Pod").
					WithSpec(corev1apply.PodSpec().
						WithContainers(corev1apply.Container().
							WithName("nginx").
							WithImage("nginx:latest")))

				_, applyObj, err := template.StringToUnstructured(applyObjStr)
				Expect(err).ShouldNot(HaveOccurred())

				_, err = c.Apply(ctx, applyObj, "test-controller", false, true)
				Expect(err).ShouldNot(HaveOccurred())

				var pod corev1.Pod
				Eventually(func() error {
					key := client.ObjectKey{
						Name:      "nginx",
						Namespace: "default",
					}
					err := k8sClient.Get(ctx, key, &pod)
					if err != nil {
						return err
					}
					return nil
				}, time.Second*10).Should(Succeed())

				podApplyCfg, err := corev1apply.ExtractPod(&pod, "test-controller")
				Expect(err).ShouldNot(HaveOccurred())

				eq := equality.Semantic.DeepEqual(podApplyCfg, expectedPodApplyCfg)
				Expect(eq).Should(BeTrue())

			})
		})

		Context("when apply existing object", func() {
			It("should update the object", func() {
				ctx := context.Background()
				c := NewClient(k8sClient)

				var currentPod corev1.Pod
				key := client.ObjectKey{
					Name:      "nginx",
					Namespace: "default",
				}
				err := k8sClient.Get(ctx, key, &currentPod)
				Expect(err).ShouldNot(HaveOccurred())

				currentPodApplyCfg, err := corev1apply.ExtractPod(&currentPod, "test-controller")
				Expect(err).ShouldNot(HaveOccurred())

				applyObjStr := `apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: default
spec:
  containers:
  - image: 'nginx:next'
    name: nginx
`

				expectedPodApplyCfg := corev1apply.Pod("nginx", "default").
					WithAPIVersion("v1").
					WithKind("Pod").
					WithSpec(corev1apply.PodSpec().
						WithContainers(corev1apply.Container().
							WithName("nginx").
							WithImage("nginx:next")))

				_, applyObj, err := template.StringToUnstructured(applyObjStr)
				Expect(err).ShouldNot(HaveOccurred())

				_, err = c.Apply(ctx, applyObj, "test-controller", false, true)
				Expect(err).ShouldNot(HaveOccurred())

				var appliedPod corev1.Pod
				Eventually(func() error {
					key := client.ObjectKey{
						Name:      "nginx",
						Namespace: "default",
					}
					err := k8sClient.Get(ctx, key, &appliedPod)
					if err != nil {
						return err
					}
					return nil
				}, time.Second*10).Should(Succeed())

				By("checking if applied properties are as expected")
				podApplyCfg, err := corev1apply.ExtractPod(&appliedPod, "test-controller")
				Expect(err).ShouldNot(HaveOccurred())

				clog.PrintObjectDiff(os.Stderr, podApplyCfg, expectedPodApplyCfg)
				eq := equality.Semantic.DeepEqual(podApplyCfg, expectedPodApplyCfg)
				Expect(eq).Should(BeTrue())

				By("checking resourceVersion is next one")
				clog.PrintObjectDiff(os.Stderr, currentPod, appliedPod)

				befVer, err := strconv.Atoi(currentPod.ResourceVersion)
				Expect(err).ShouldNot(HaveOccurred())

				aftVer, err := strconv.Atoi(appliedPod.ResourceVersion)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(befVer+1 == aftVer).Should(BeTrue())

				By("checking other properties is not modified")

				// fix spec to applied values
				currentPodApplyCfg.Spec.Containers[0].Image = pointer.String("nginx:next")

				eq = equality.Semantic.DeepEqual(currentPodApplyCfg, podApplyCfg)
				Expect(eq).Should(BeTrue())
			})
		})
	})
})

func dumpObject(obj interface{}) string {
	out, err := yaml.Marshal(obj)
	Expect(err).ShouldNot(HaveOccurred())
	return string(out)
}
