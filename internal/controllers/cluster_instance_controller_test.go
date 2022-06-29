package controllers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	. "github.com/cosmo-workspace/cosmo/pkg/kubeutil/test/gomega"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/equality"
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
	const tmplName string = "pv-test"
	const instName string = "cinst-test"
	const pvSufix string = "pv"
	const scSufix string = "sc"
	const varMountPath string = "MOUNT_PATH"
	const valMountPath string = "/tmp"

	tmpl := cosmov1alpha1.ClusterTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: tmplName,
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

	expectedPVApply := func(instName, mountPath string, ownerRef metav1.OwnerReference) *corev1apply.PersistentVolumeApplyConfiguration {
		return corev1apply.PersistentVolume(instance.InstanceResourceName(instName, pvSufix)).
			WithAPIVersion("v1").
			WithKind("PersistentVolume").
			WithLabels(map[string]string{
				cosmov1alpha1.LabelKeyInstance: instName,
				cosmov1alpha1.LabelKeyTemplate: tmplName,
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
				WithStorageClassName(instance.InstanceResourceName(instName, scSufix)).
				WithMountOptions("hard", "nfsvers=4.1").
				WithNFS(corev1apply.NFSVolumeSource().
					WithServer("nfs-server.example.com").
					WithPath(mountPath)))
	}

	expectedStorageClassApply := func(instName, mountPath string, ownerRef metav1.OwnerReference) *storagev1apply.StorageClassApplyConfiguration {
		return storagev1apply.StorageClass(instance.InstanceResourceName(instName, scSufix)).
			WithAPIVersion("storage.k8s.io/v1").
			WithKind("StorageClass").
			WithLabels(map[string]string{
				cosmov1alpha1.LabelKeyInstance: instName,
				cosmov1alpha1.LabelKeyTemplate: tmplName,
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
				err := k8sClient.Get(ctx, client.ObjectKey{Name: tmplName}, &createdTmpl)
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
					Name: instName,
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: tmplName,
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
			Eventually(func() error {
				key := client.ObjectKey{
					Name: instName,
				}
				err := k8sClient.Get(ctx, key, &createdInst)
				if err != nil {
					return err
				}
				if len(createdInst.Status.LastApplied) == 0 {
					return errors.New("child resources still not created")
				}
				return nil
			}, time.Second*10).Should(Succeed())

			By("checking if child resources is as expected in template")

			ownerRef := ownerRef(&inst, scheme.Scheme)

			// PersistentVolume
			By("checking PersistentVolume is as expected")

			var pv corev1.PersistentVolume
			Eventually(func() error {
				key := client.ObjectKey{
					Name: instance.InstanceResourceName(instName, pvSufix),
				}
				return k8sClient.Get(ctx, key, &pv)
			}, time.Second*10).Should(Succeed())

			pvApplyCfg, err := corev1apply.ExtractPersistentVolume(&pv, controllerFieldManager)
			Expect(err).ShouldNot(HaveOccurred())

			expectedPVApplyCfg := expectedPVApply(instName, valMountPath, ownerRef)
			Expect(pvApplyCfg).Should(BeEqualityDeepEqual(expectedPVApplyCfg))

			pv.SetGroupVersionKind(schema.FromAPIVersionAndKind(*pvApplyCfg.APIVersion, *pvApplyCfg.Kind))
			Expect(instance.ExistInLastApplyed(&createdInst, &pv)).Should(BeTrue())

			// StorageClass
			By("checking StorageClass is as expected")

			var sc storagev1.StorageClass
			Eventually(func() error {
				key := client.ObjectKey{
					Name: instance.InstanceResourceName(instName, scSufix),
				}
				return k8sClient.Get(ctx, key, &sc)
			}, time.Second*10).Should(Succeed())

			scApplyCfg, err := storagev1apply.ExtractStorageClass(&sc, controllerFieldManager)
			Expect(err).ShouldNot(HaveOccurred())

			expectedSCApplyCfg := expectedStorageClassApply(instName, valMountPath, ownerRef)

			Expect(scApplyCfg).Should(BeEqualityDeepEqual(expectedSCApplyCfg))

			sc.SetGroupVersionKind(schema.FromAPIVersionAndKind(*expectedSCApplyCfg.APIVersion, *expectedSCApplyCfg.Kind))
			Expect(instance.ExistInLastApplyed(&createdInst, &sc)).Should(BeTrue())

			By("checking creation time equal to update time")
			for _, v := range createdInst.Status.LastApplied {
				Expect(v.CreationTimestamp).Should(BeEquivalentTo(v.UpdateTimestamp))
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
					Name: instName,
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
			}).Should(Succeed())

			By("checking if child resources updated")

			ownerRef := ownerRef(&inst, scheme.Scheme)

			// expected PersistentVolume
			expectedPVApplyCfg := expectedPVApply(instName, valMountPath, ownerRef)
			expectedPVApplyCfg.Spec.WithCapacity(corev1.ResourceList{"storage": resource.MustParse("10Gi")})

			By("checking if PersistentVolume updated")

			var pv corev1.PersistentVolume
			Eventually(func() error {
				key := client.ObjectKey{
					Name: instance.InstanceResourceName(instName, pvSufix),
				}
				err := k8sClient.Get(ctx, key, &pv)
				if err != nil {
					return err
				}

				pvApplyCfg, err := corev1apply.ExtractPersistentVolume(&pv, controllerFieldManager)
				Expect(err).ShouldNot(HaveOccurred())

				eq := equality.Semantic.DeepEqual(pvApplyCfg, expectedPVApplyCfg)
				if !eq {
					return fmt.Errorf("not equal: %s", clog.Diff(pvApplyCfg, expectedPVApplyCfg))
				}
				return nil
			}, time.Second*10).Should(Succeed())

			// expected StorageClass
			expectedSCApplyCfg := expectedStorageClassApply(instName, valMountPath, ownerRef)

			By("checking if StorageClass updated")

			var sc storagev1.StorageClass
			Eventually(func() error {
				key := client.ObjectKey{
					Name: instance.InstanceResourceName(instName, scSufix),
				}
				err := k8sClient.Get(ctx, key, &sc)
				if err != nil {
					return err
				}

				scApplyCfg, err := storagev1apply.ExtractStorageClass(&sc, controllerFieldManager)
				Expect(err).ShouldNot(HaveOccurred())

				eq := equality.Semantic.DeepEqual(scApplyCfg, expectedSCApplyCfg)
				if !eq {
					return fmt.Errorf("not equal: %s", clog.Diff(scApplyCfg, expectedSCApplyCfg))
				}
				return nil
			}, time.Second*10).Should(Succeed())
		})
	})

	Context("when creating pod", func() {
		It("should not create namespaced-scope resource", func() {
			ctx := context.Background()

			nsLevelTmplName := "ns-level-ctmpl"
			nsLevelTmpl := cosmov1alpha1.ClusterTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: nsLevelTmplName,
				},
				Spec: cosmov1alpha1.TemplateSpec{
					RawYaml: `apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: {{USER_NAMESPACE}}
spec:
  containers:
  - name: nginx
    image: nginx:alpine
`,
				},
			}
			err := k8sClient.Create(ctx, &nsLevelTmpl)
			Expect(err).ShouldNot(HaveOccurred())

			nsLevelInstName := "ns-level-inst"
			nsLevelInst := cosmov1alpha1.ClusterInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name: nsLevelInstName,
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: nsLevelTmplName,
					},
					Vars: map[string]string{
						"USER_NAMESPACE": "default",
					},
				},
			}
			err = k8sClient.Create(ctx, &nsLevelInst)
			Expect(err).ShouldNot(HaveOccurred())

			podName := instance.InstanceResourceName(nsLevelInstName, "nginx")

			time.Sleep(time.Second * 3)

			var createdInst cosmov1alpha1.ClusterInstance
			key := client.ObjectKey{
				Name: nsLevelInstName,
			}
			err = k8sClient.Get(ctx, key, &createdInst)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(createdInst.Status.LastApplied)).Should(BeZero())

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
