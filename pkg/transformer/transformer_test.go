package transformer

import (
	"context"
	"testing"

	. "github.com/cosmo-workspace/cosmo/pkg/kubeutil/test/gomega"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/template"
)

func TestApplyTransformers(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(cosmov1alpha1.AddToScheme(scheme))

	RegisterTestingT(t)

	type args struct {
		ctx          context.Context
		transformers []Transformer
		objects      []string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "OK",
			want: []string{`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    cosmo/ingress-patch-enable: "true"
    kubernetes.io/ingress.class: alb
  labels:
    cosmo-workspace.github.io/instance: cs1
    cosmo-workspace.github.io/template: code-server
    key: val
  name: cs1-test
  namespace: cosmo-user-tom
  ownerReferences:
  - apiVersion: cosmo-workspace.github.io/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Instance
    name: cs1
    uid: ""
spec:
  host: stg.example.com`},
			wantErr: false,
			args: args{
				transformers: AllTransformers(
					&cosmov1alpha1.Instance{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cs1",
							Namespace: "cosmo-user-tom",
						},
						Spec: cosmov1alpha1.InstanceSpec{
							Template: cosmov1alpha1.TemplateRef{
								Name: "code-server",
							},
							Override: cosmov1alpha1.OverrideSpec{
								PatchesJson6902: []cosmov1alpha1.Json6902{
									{
										Target: cosmov1alpha1.ObjectRef{
											ObjectReference: corev1.ObjectReference{
												APIVersion: "networking.k8s.io/v1",
												Kind:       "Ingress",
												Name:       "test",
											},
										},
										Patch: `
[
  {
    "op": "replace",
    "path": "/spec/host",
    "value": "stg.example.com"
  }
]`,
									},
								},
							},
						},
					},
					scheme,
					&cosmov1alpha1.Template{
						ObjectMeta: metav1.ObjectMeta{
							Name: "code-server",
						},
						Spec: cosmov1alpha1.TemplateSpec{},
					},
				),
				objects: []string{`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    cosmo/ingress-patch-enable: "true"
    kubernetes.io/ingress.class: alb
  labels:
    key: val
  name: test
spec:
  host: example.com`},
			},
		},
		{
			name:    "err on transform",
			wantErr: true,
			args: args{
				transformers: AllTransformers(
					&cosmov1alpha1.Instance{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "cs1",
							Namespace: "cosmo-user-tom",
						},
						Spec: cosmov1alpha1.InstanceSpec{
							Template: cosmov1alpha1.TemplateRef{
								Name: "code-server",
							},
							Override: cosmov1alpha1.OverrideSpec{
								PatchesJson6902: []cosmov1alpha1.Json6902{
									{
										Target: cosmov1alpha1.ObjectRef{
											ObjectReference: corev1.ObjectReference{
												APIVersion: "networking.k8s.io/v1",
												Kind:       "Ingress",
												Name:       "test",
											},
										},
										Patch: `invalid patch`,
									},
								},
							},
						},
					},
					scheme,
					&cosmov1alpha1.Template{
						ObjectMeta: metav1.ObjectMeta{
							Name: "code-server",
						},
						Spec: cosmov1alpha1.TemplateSpec{},
					},
				),
				objects: []string{`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    cosmo/ingress-patch-enable: "true"
    kubernetes.io/ingress.class: alb
  labels:
    key: val
  name: test
spec:
  host: example.com`},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects := make([]unstructured.Unstructured, len(tt.args.objects))
			for i, v := range tt.args.objects {
				_, o, err := template.StringToUnstructured(v)
				Expect(err).ShouldNot(HaveOccurred())
				objects[i] = *o
			}

			wants := make([]unstructured.Unstructured, len(tt.want))
			for i, v := range tt.want {
				_, o, err := template.StringToUnstructured(v)
				Expect(err).ShouldNot(HaveOccurred())
				wants[i] = *o
			}

			got, err := ApplyTransformers(tt.args.ctx, tt.args.transformers, objects)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyTransformers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i, g := range got {
				Expect(g).Should(BeEqualityDeepEqual(wants[i]))
			}
		})
	}
}

func TestName(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		tf := NewJSONPatchTransformer(nil, "")
		if got := Name(tf); got != "JSONPatchTransformer" {
			t.Errorf("Name() = %v, want JSONPatchTransformer", got)
		}
	})
}
