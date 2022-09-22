package kubeutil

import (
	"context"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/go-cmp/cmp"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/cosmo-workspace/cosmo/pkg/template"
)

var _ = Describe("Client", func() {
	Describe("DryrunApply", func() {
		Context("when apply new object", func() {
			It("should create no object and dryrun succeed", func() {
				ctx := context.Background()

				_, err := GetUnstructured(ctx, k8sClient, schema.GroupVersionKind{Version: "v1", Kind: "Pod"}, "nginx", "default")
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

				_, err = Apply(ctx, k8sClient, applyObj, "test-controller", true, true)
				Expect(err).ShouldNot(HaveOccurred())

				time.Sleep(time.Second * 3)

				_, err = GetUnstructured(ctx, k8sClient, schema.GroupVersionKind{Version: "v1", Kind: "Pod"}, "nginx", "default")
				Expect(apierrs.IsNotFound(err)).Should(BeTrue())
			})
		})
	})

	Describe("Apply", func() {
		Context("when apply new object", func() {
			It("should create new object", func() {
				ctx := context.Background()

				_, err := GetUnstructured(ctx, k8sClient, schema.GroupVersionKind{Version: "v1", Kind: "Pod"}, "nginx", "default")
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

				_, err = Apply(ctx, k8sClient, applyObj, "test-controller", false, true)
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

				Expect(podApplyCfg).Should(Equal(expectedPodApplyCfg))
			})
		})

		Context("when apply existing object", func() {
			It("should update the object", func() {
				ctx := context.Background()

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

				_, err = Apply(ctx, k8sClient, applyObj, "test-controller", false, true)
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
				Expect(podApplyCfg).Should(Equal(expectedPodApplyCfg))

				By("checking resourceVersion is next one")

				befVer, err := strconv.Atoi(currentPod.ResourceVersion)
				Expect(err).ShouldNot(HaveOccurred())

				aftVer, err := strconv.Atoi(appliedPod.ResourceVersion)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(aftVer).Should(Equal(befVer + 1))

				By("checking other properties is not modified")

				// fix spec to applied values
				currentPodApplyCfg.Spec.Containers[0].Image = pointer.String("nginx:next")

				Expect(currentPodApplyCfg).Should(Equal(podApplyCfg))
			})
		})
	})

	Context("when changing invalid fields", func() {
		It("should not update the object", func() {
			ctx := context.Background()

			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-invalid-field",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "sh",
							Image: "busybox",
							Command: []string{
								"sleep", "inginity",
							},
						},
					},
				},
			}

			err := k8sClient.Create(ctx, &pod)
			Expect(err).ShouldNot(HaveOccurred())

			p := pod.DeepCopy()

			pod.Spec.Containers[0].Command = []string{"patch"}

			obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&pod)
			Expect(err).ShouldNot(HaveOccurred())

			var u unstructured.Unstructured
			u.Object = obj
			u.SetAPIVersion("v1")
			u.SetKind("Pod")

			_, err = Apply(ctx, k8sClient, &u, "test-controller", false, true)
			Expect(err).Should(HaveOccurred())

			var p2 corev1.Pod
			err = k8sClient.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, &p2)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(p.ResourceVersion).Should(BeEquivalentTo(p2.ResourceVersion))
		})
	})

	Describe("GetUnstructured", func() {
		It("should get pod as unstructured", func() {
			ctx := context.Background()

			svcYAML := `
apiVersion: v1
kind: Service
metadata:
  name: http
  namespace: default
spec:
  ports:
  - name: http
    port: 8080
    protocol: TCP
`
			var svc corev1.Service
			err := yaml.Unmarshal([]byte(svcYAML), &svc)
			Expect(err).ShouldNot(HaveOccurred())

			err = k8sClient.Create(ctx, &svc)
			Expect(err).ShouldNot(HaveOccurred())

			expectObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&svc)
			Expect(err).ShouldNot(HaveOccurred())

			expect := &unstructured.Unstructured{}
			expect.Object = expectObj
			expect.SetGroupVersionKind(ServiceGVK)

			got, err := GetUnstructured(ctx, k8sClient, ServiceGVK, svc.Name, svc.Namespace)
			Expect(err).ShouldNot(HaveOccurred())

			diff := cmp.Diff(got, expect)
			Expect(diff).Should(BeEmpty())
		})
	})
})
