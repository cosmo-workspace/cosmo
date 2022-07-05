package controllers

import (
	"context"
	"fmt"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/kubeutil/test/gomega"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	storagev1apply "k8s.io/client-go/applyconfigurations/storage/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
)

var _ = Describe("ClusterInstance controller", func() {
	const name string = "clusterinst-test"
	const pvSufix string = "pv"
	const scSufix string = "sc"
	const varMountPath string = "MOUNT_PATH"
	const valMountPath string = "/tmp"

	tmpl := cosmov1alpha1.ClusterTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
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
`, pvSufix, scSufix, scSufix),
			RequiredVars: []cosmov1alpha1.RequiredVarSpec{
				{
					Var:     varMountPath,
					Default: "/mnt/pv",
				},
			},
		},
	}

	expectedPVApply := func(mountPath string, ownerRef metav1.OwnerReference) *corev1apply.PersistentVolumeApplyConfiguration {
		return corev1apply.PersistentVolume(instance.InstanceResourceName(name, pvSufix)).
			WithAPIVersion("v1").
			WithKind("PersistentVolume").
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
			WithSpec(corev1apply.PersistentVolumeSpec().
				WithCapacity(corev1.ResourceList{"storage": resource.MustParse("5Gi")}).
				WithVolumeMode(corev1.PersistentVolumeMode("Filesystem")).
				WithAccessModes(corev1.ReadWriteOnce).
				WithPersistentVolumeReclaimPolicy(corev1.PersistentVolumeReclaimRecycle).
				WithStorageClassName(instance.InstanceResourceName(name, scSufix)).
				WithMountOptions("hard", "nfsvers=4.1").
				WithNFS(corev1apply.NFSVolumeSource().
					WithServer("nfs-server.example.com").
					WithPath(mountPath)))
	}

	expectedStorageClassApply := func(mountPath string, ownerRef metav1.OwnerReference) *storagev1apply.StorageClassApplyConfiguration {
		return storagev1apply.StorageClass(instance.InstanceResourceName(name, scSufix)).
			WithAPIVersion("storage.k8s.io/v1").
			WithKind("StorageClass").
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
			WithProvisioner("example.com/external-nfs").
			WithParameters(map[string]string{
				"server": "nfs-server.example.com",
				"path":   mountPath,
			})
	}

	Context("when creating ClusterTemplate resource on new cluster", func() {
		It("should do nothing", func() {
			ctx := context.Background()

			By("creating template before instance")

			err := k8sClient.Create(ctx, &tmpl)
			Expect(err).ShouldNot(HaveOccurred())

			var createdTmpl cosmov1alpha1.ClusterTemplate
			Eventually(func() error {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: name}, &createdTmpl)
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

			inst := cosmov1alpha1.ClusterInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: name,
					},
					Override: cosmov1alpha1.OverrideSpec{},
					Vars: map[string]string{
						varMountPath: valMountPath,
					},
				},
			}
			err := k8sClient.Create(ctx, &inst)
			Expect(err).ShouldNot(HaveOccurred())

			By("fetching instance resource and checking if last applied resources added in instance status")

			var createdInst cosmov1alpha1.ClusterInstance
			Eventually(func() int {
				key := client.ObjectKey{
					Name: name,
				}
				err := k8sClient.Get(ctx, key, &createdInst)
				Expect(err).ShouldNot(HaveOccurred())

				return createdInst.Status.LastAppliedObjectsCount
			}, time.Second*60).Should(BeEquivalentTo(2))

			By("checking if child resources is as expected in template")

			ownerRef := ownerRef(&inst, scheme.Scheme)

			// PersistentVolume
			By("checking PersistentVolume is as expected")

			var pv corev1.PersistentVolume
			Eventually(func() error {
				key := client.ObjectKey{
					Name: instance.InstanceResourceName(name, pvSufix),
				}
				return k8sClient.Get(ctx, key, &pv)
			}, time.Second*10).Should(Succeed())

			pvApplyCfg, err := corev1apply.ExtractPersistentVolume(&pv, controllerFieldManager)
			Expect(err).ShouldNot(HaveOccurred())

			expectedPVApplyCfg := expectedPVApply(valMountPath, ownerRef)
			Expect(pvApplyCfg).Should(BeEqualityDeepEqual(expectedPVApplyCfg))

			pv.SetGroupVersionKind(schema.FromAPIVersionAndKind(*pvApplyCfg.APIVersion, *pvApplyCfg.Kind))
			Expect(instance.ExistInLastApplyed(&createdInst, &pv)).Should(BeTrue())

			// StorageClass
			By("checking StorageClass is as expected")

			var sc storagev1.StorageClass
			Eventually(func() error {
				key := client.ObjectKey{
					Name: instance.InstanceResourceName(name, scSufix),
				}
				return k8sClient.Get(ctx, key, &sc)
			}, time.Second*10).Should(Succeed())

			scApplyCfg, err := storagev1apply.ExtractStorageClass(&sc, controllerFieldManager)
			Expect(err).ShouldNot(HaveOccurred())

			expectedSCApplyCfg := expectedStorageClassApply(valMountPath, ownerRef)

			Expect(scApplyCfg).Should(BeEqualityDeepEqual(expectedSCApplyCfg))

			sc.SetGroupVersionKind(schema.FromAPIVersionAndKind(*expectedSCApplyCfg.APIVersion, *expectedSCApplyCfg.Kind))
			Expect(instance.ExistInLastApplyed(&createdInst, &sc)).Should(BeTrue())

			By("checking creation time equal to update time")
			for _, v := range createdInst.Status.LastApplied {
				Expect(*v.CreationTimestamp).Should(BeEquivalentTo(*v.UpdateTimestamp))
			}
		})
	})

	Context("when updating ClusterInstance resource", func() {
		It("should do reconcile again and update child resources", func() {
			ctx := context.Background()

			// fetch current instance
			var inst cosmov1alpha1.ClusterInstance
			Eventually(func() error {
				key := types.NamespacedName{
					Name: name,
				}
				return k8sClient.Get(ctx, key, &inst)
			}, time.Second*10).Should(Succeed())

			// update instance override spec
			inst.Spec.Override = cosmov1alpha1.OverrideSpec{
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

			Eventually(func() error {
				return k8sClient.Update(ctx, &inst)
			}, time.Second*60).Should(Succeed())

			By("checking if child resources updated")

			ownerRef := ownerRef(&inst, scheme.Scheme)

			// expected PersistentVolume
			expectedPVApplyCfg := expectedPVApply(valMountPath, ownerRef)
			expectedPVApplyCfg.Spec.WithCapacity(corev1.ResourceList{"storage": resource.MustParse("10Gi")})

			By("checking if PersistentVolume updated")

			var pv corev1.PersistentVolume
			Eventually(func() *corev1apply.PersistentVolumeApplyConfiguration {
				key := client.ObjectKey{
					Name: instance.InstanceResourceName(name, pvSufix),
				}
				err := k8sClient.Get(ctx, key, &pv)
				Expect(err).ShouldNot(HaveOccurred())

				pvApplyCfg, err := corev1apply.ExtractPersistentVolume(&pv, controllerFieldManager)
				Expect(err).ShouldNot(HaveOccurred())

				return pvApplyCfg
			}, time.Second*10).Should(BeEqualityDeepEqual(expectedPVApplyCfg))

			// expected StorageClass
			expectedSCApplyCfg := expectedStorageClassApply(valMountPath, ownerRef)

			By("checking if StorageClass updated")

			var sc storagev1.StorageClass
			Eventually(func() *storagev1apply.StorageClassApplyConfiguration {
				key := client.ObjectKey{
					Name: instance.InstanceResourceName(name, scSufix),
				}
				err := k8sClient.Get(ctx, key, &sc)
				Expect(err).ShouldNot(HaveOccurred())

				scApplyCfg, err := storagev1apply.ExtractStorageClass(&sc, controllerFieldManager)
				Expect(err).ShouldNot(HaveOccurred())

				return scApplyCfg
			}, time.Second*10).Should(BeEqualityDeepEqual(expectedSCApplyCfg))
		})
	})

	Context("when creating pod", func() {
		It("create namespaced-scope resource even though namespace is not found", func() {
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
})
