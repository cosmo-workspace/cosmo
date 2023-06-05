package controllers

import (
	"context"
	"fmt"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
)

var _ = Describe("ClusterInstance controller", func() {
	const pvSufix string = "pv"
	const scSufix string = "sc"
	const varMountPath string = "MOUNT_PATH"
	const valMountPath string = "/tmp"

	tmpl := cosmov1alpha1.ClusterTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-storage-clustertmpl1",
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RawYaml: fmt.Sprintf(`
apiVersion: v1
kind: PersistentVolume
metadata:
  name: %s
spec:
  capacity:
    storage: 5Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Recycle
  storageClassName: {{INSTANCE}}-%s
  mountOptions:
    - hard
    - nfsvers=4.1
  nfs:
    server: nfs-server.example.com
    path: {{MOUNT_PATH}}
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: %s
provisioner: example.com/external-nfs
parameters:
  server: nfs-server.example.com
  path: {{MOUNT_PATH}}
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: %s
  namespace: default
spec:
  storageClassName: {{INSTANCE}}-%s
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
`, pvSufix, scSufix, scSufix, pvSufix, scSufix),
			RequiredVars: []cosmov1alpha1.RequiredVarSpec{
				{
					Var:     varMountPath,
					Default: "/mnt/pv",
				},
			},
		},
	}

	inst := cosmov1alpha1.ClusterInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pv-clusterinst1",
		},
		Spec: cosmov1alpha1.InstanceSpec{
			Template: cosmov1alpha1.TemplateRef{
				Name: tmpl.Name,
			},
			Override: cosmov1alpha1.OverrideSpec{},
			Vars: map[string]string{
				varMountPath: valMountPath,
			},
		},
	}

	Context("when creating ClusterTemplate resource on new cluster", func() {
		It("should do nothing", func() {
			ctx := context.Background()

			By("creating template before instance")

			err := k8sClient.Create(ctx, &tmpl)
			Expect(err).ShouldNot(HaveOccurred())

			var createdTmpl cosmov1alpha1.ClusterTemplate
			Eventually(func() error {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: tmpl.Name}, &createdTmpl)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())
		})
	})

	Context("when creating a Instance resource", func() {
		It("should do reconcile once and create child resources", func() {
			ctx := context.Background()

			err := k8sClient.Create(ctx, &inst)
			Expect(err).ShouldNot(HaveOccurred())

			By("fetching instance resource and checking if last applied resources added in instance status")

			var createdInst cosmov1alpha1.ClusterInstance
			Eventually(func() int {
				key := client.ObjectKey{
					Name: inst.Name,
				}
				err := k8sClient.Get(ctx, key, &createdInst)
				Expect(err).ShouldNot(HaveOccurred())

				return createdInst.Status.LastAppliedObjectsCount
			}, time.Second*10).Should(Equal(3))
			Ω(InstanceSnapshot(&createdInst)).To(MatchSnapShot())

			By("checking PersistentVolume is as expected in template")
			var pv corev1.PersistentVolume
			Eventually(func() error {
				key := client.ObjectKey{
					Name: instance.InstanceResourceName(inst.Name, pvSufix),
				}
				return k8sClient.Get(ctx, key, &pv)
			}, time.Second*10).Should(Succeed())
			Ω(ObjectSnapshot(&pv)).To(MatchSnapShot())

			// StorageClass
			By("checking StorageClass is as expected")

			var sc storagev1.StorageClass
			Eventually(func() error {
				key := client.ObjectKey{
					Name: instance.InstanceResourceName(inst.Name, scSufix),
				}
				return k8sClient.Get(ctx, key, &sc)
			}, time.Second*10).Should(Succeed())
			Ω(ObjectSnapshot(&sc)).To(MatchSnapShot())

			By("checking PVC is as expected")

			var pvc corev1.PersistentVolumeClaim
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(inst.Name, pvSufix),
					Namespace: "default",
				}
				return k8sClient.Get(ctx, key, &pvc)
			}, time.Second*10).Should(Succeed())
			Ω(ObjectSnapshot(&pvc)).To(MatchSnapShot())

		})
	})

	Context("when updating ClusterInstance resource", func() {
		It("should do reconcile again and update child resources", func() {
			ctx := context.Background()

			// fetch current instance
			var curInst cosmov1alpha1.ClusterInstance
			Eventually(func() error {
				key := types.NamespacedName{
					Name: inst.Name,
				}
				err := k8sClient.Get(ctx, key, &curInst)
				Expect(err).NotTo(HaveOccurred())

				// update instance override spec
				curInst.Spec.Override = cosmov1alpha1.OverrideSpec{
					PatchesJson6902: []cosmov1alpha1.Json6902{
						{
							Target: cosmov1alpha1.ObjectRef{
								ObjectReference: corev1.ObjectReference{
									APIVersion: "v1",
									Kind:       "PersistentVolume",
									Name:       pvSufix,
								},
							},
							Patch: `
[
  {
    "op": "replace",
    "path": "/spec/capacity/storage",
    "value": "10Gi"
  }
]
						`,
						},
					},
				}
				return k8sClient.Update(ctx, &curInst)
			}, time.Second*60).Should(Succeed())

			expectedQuantity, _ := resource.ParseQuantity("10Gi")

			By("checking if PersistentVolume updated")
			var pv corev1.PersistentVolume
			Eventually(func() resource.Quantity {
				key := client.ObjectKey{
					Name: instance.InstanceResourceName(inst.Name, pvSufix),
				}
				err := k8sClient.Get(ctx, key, &pv)
				Expect(err).ShouldNot(HaveOccurred())

				return *pv.Spec.Capacity.Storage()

			}, time.Second*30).Should(Equal(expectedQuantity))
			Ω(ObjectSnapshot(&pv)).To(MatchSnapShot())

			By("checking if StorageClass not updated")
			var sc storagev1.StorageClass
			Eventually(func() error {
				key := client.ObjectKey{
					Name: instance.InstanceResourceName(inst.Name, scSufix),
				}
				return k8sClient.Get(ctx, key, &sc)
			}, time.Second*10).Should(Succeed())
			Ω(ObjectSnapshot(&sc)).To(MatchSnapShot())
		})
	})

	Context("when creating pod without metadata.namespace", func() {
		It("success to create instance but child resources are not created", func() {
			ctx := context.Background()

			t := cosmov1alpha1.ClusterTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ns-level-with-no-ns-err",
				},
				Spec: cosmov1alpha1.TemplateSpec{
					RawYaml: `apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: nginx:alpine
`,
				},
			}
			err := k8sClient.Create(ctx, &t)
			Expect(err).ShouldNot(HaveOccurred())

			nsLevelInst := cosmov1alpha1.ClusterInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name: t.Name,
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: t.Name,
					},
				},
			}
			err = k8sClient.Create(ctx, &nsLevelInst)
			Expect(err).ShouldNot(HaveOccurred()) // pass here even though namespace is not found

			podName := instance.InstanceResourceName(t.Name, "nginx")

			time.Sleep(time.Second * 3)

			var createdInst cosmov1alpha1.ClusterInstance
			key := client.ObjectKey{
				Name: t.Name,
			}
			err = k8sClient.Get(ctx, key, &createdInst)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(createdInst.Status.LastAppliedObjectsCount).Should(BeZero()) // Pod is not created

			var pod corev1.Pod
			key = types.NamespacedName{Namespace: "default", Name: podName}
			err = k8sClient.Get(ctx, key, &pod)
			Expect(err).Should(HaveOccurred())
			Expect(apierrs.ReasonForError(err)).Should(Equal(metav1.StatusReasonNotFound))

			var pods corev1.PodList
			err = k8sClient.List(ctx, &pods)
			Expect(err).ShouldNot(HaveOccurred())

			for _, pod := range pods.Items {
				Expect(pod.Name).ShouldNot(Equal(podName))
			}
		})
	})

	Context("when removing namespaced resource in ClusterTemplate", func() {
		It("should remove unmanaged resources(GC)", func() {
			ctx := context.Background()

			var curInst cosmov1alpha1.ClusterInstance
			Eventually(func() error {
				key := client.ObjectKey{
					Name: inst.Name,
				}
				return k8sClient.Get(ctx, key, &curInst)
			}, time.Second*10).Should(Succeed())

			// fetch current clustertemplate
			var curTmpl cosmov1alpha1.ClusterTemplate
			Eventually(func() error {
				key := types.NamespacedName{
					Name: tmpl.Name,
				}
				err := k8sClient.Get(ctx, key, &curTmpl)
				Expect(err).ShouldNot(HaveOccurred())

				// remove pvc
				curTmpl.Spec.RawYaml = fmt.Sprintf(`
apiVersion: v1
kind: PersistentVolume
metadata:
  name: %s
spec:
  capacity:
    storage: 5Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Recycle
  storageClassName: {{INSTANCE}}-%s
  mountOptions:
    - hard
    - nfsvers=4.1
  nfs:
    server: nfs-server.example.com
    path: {{MOUNT_PATH}}
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: %s
provisioner: example.com/external-nfs
parameters:
  server: nfs-server.example.com
  path: {{MOUNT_PATH}}
`, pvSufix, scSufix, scSufix)
				return k8sClient.Update(ctx, &curTmpl)
			}, time.Second*60).Should(Succeed())

			By("checking if pvc is removed")

			var pvc corev1.PersistentVolumeClaim
			Eventually(func() error {
				key := client.ObjectKey{
					Name: instance.InstanceResourceName(inst.Name, pvSufix),
				}
				return k8sClient.Get(ctx, key, &pvc)
			}, time.Second*60).Should(HaveOccurred())

			// fetch current clusterinstance
			var updatedInst cosmov1alpha1.ClusterInstance
			Eventually(func() int {
				key := client.ObjectKey{
					Name: inst.Name,
				}
				err := k8sClient.Get(ctx, key, &updatedInst)
				Expect(err).ShouldNot(HaveOccurred())
				return updatedInst.Status.LastAppliedObjectsCount
			}, time.Second*60).Should(Equal(2))
			Ω(InstanceSnapshot(&updatedInst)).To(MatchSnapShot())
		})
	})
})
