package controllers

import (
	"context"
	"errors"
	"fmt"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/kubeutil/test/gomega"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	rbacv1apply "k8s.io/client-go/applyconfigurations/rbac/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
)

var _ = Describe("ClusterTemplate controller", func() {
	const tmplName string = "pod-list-role"
	const instName string = "clustertmpl-role"
	const crName string = "pod-list-cr"

	tmpl := cosmov1alpha1.ClusterTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: tmplName,
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

	expectedClusterRoleApply := func(instName string, ownerRef metav1.OwnerReference) *rbacv1apply.ClusterRoleApplyConfiguration {
		return rbacv1apply.ClusterRole(crName).
			WithAPIVersion("rbac.authorization.k8s.io/v1").
			WithKind("ClusterRole").
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
			WithRules(
				rbacv1apply.PolicyRule().
					WithResources("pods").
					WithVerbs("get", "list", "watch").
					WithAPIGroups(""))
	}

	Context("when creating ClusterTemplate resource on new cluster", func() {
		It("should do nothing", func() {
			ctx := context.Background()

			By("creating clustertemplate before clusterinstance")

			err := k8sClient.Create(ctx, &tmpl)
			Expect(err).ShouldNot(HaveOccurred())

			var createdTmpl cosmov1alpha1.ClusterTemplate
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: tmplName}, &createdTmpl)
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
					Name: instName,
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
			}, time.Second*30).Should(Succeed())

			By("checking if child resources is as expected in template")

			instOwnerRef := ownerRef(&inst, scheme.Scheme)

			// ClusterRole
			var cr rbacv1.ClusterRole
			Eventually(func() error {
				key := client.ObjectKey{
					Name: crName,
				}
				err := k8sClient.Get(ctx, key, &cr)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())

			crApplyCfg, err := rbacv1apply.ExtractClusterRole(&cr, controllerFieldManager)
			Expect(err).ShouldNot(HaveOccurred())

			expectedCRApplyCfg := expectedClusterRoleApply(instName, instOwnerRef)
			Expect(crApplyCfg).Should(BeEqualityDeepEqual(expectedCRApplyCfg))

			cr.SetGroupVersionKind(schema.FromAPIVersionAndKind(*crApplyCfg.APIVersion, *crApplyCfg.Kind))
			Expect(instance.ExistInLastApplyed(&createdInst, &cr)).Should(BeTrue())
		})
	})

	Context("when updating ClusterTemplate resource", func() {
		It("should do instance reconcile and update child resources", func() {
			ctx := context.Background()

			// fetch current clusterinstance
			var inst cosmov1alpha1.ClusterInstance
			Eventually(func() error {
				key := client.ObjectKey{
					Name: instName,
				}
				return k8sClient.Get(ctx, key, &inst)
			}, time.Second*30).Should(Succeed())

			// fetch current clustertemplate
			var tmpl cosmov1alpha1.ClusterTemplate
			Eventually(func() error {
				key := types.NamespacedName{
					Name: tmplName,
				}
				return k8sClient.Get(ctx, key, &tmpl)
			}, time.Second*10).Should(Succeed())

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
  - update
`, crName)

			// update template
			err := k8sClient.Update(ctx, &tmpl)
			Expect(err).ShouldNot(HaveOccurred())

			By("checking if child resources updated")

			instOwnerRef := ownerRef(&inst, scheme.Scheme)

			expectedClusterRoleApply := expectedClusterRoleApply(instName, instOwnerRef)
			expectedClusterRoleApply.Rules[0].Verbs = append(expectedClusterRoleApply.Rules[0].Verbs, "update")

			By("checking if clusterrole updated")

			var cr rbacv1.ClusterRole
			Eventually(func() error {
				key := client.ObjectKey{
					Name: crName,
				}
				err := k8sClient.Get(ctx, key, &cr)
				if err != nil {
					return err
				}

				crApplyCfg, err := rbacv1apply.ExtractClusterRole(&cr, controllerFieldManager)
				Expect(err).ShouldNot(HaveOccurred())

				eq := equality.Semantic.DeepEqual(crApplyCfg, expectedClusterRoleApply)
				if !eq {
					return fmt.Errorf("not equal: %s", clog.Diff(crApplyCfg, expectedClusterRoleApply))
				}
				return nil
			}, time.Second*10).Should(Succeed())
		})
	})
})
