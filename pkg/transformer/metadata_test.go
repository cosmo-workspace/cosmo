package transformer

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/yaml"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/template"
)

func TestMetadataTransformer_Transform(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(cosmov1alpha1.AddToScheme(scheme))
	type fields struct {
		inst              *cosmov1alpha1.Instance
		disableNamePrefix bool
		scheme            *runtime.Scheme
	}
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "New label",
			fields: fields{
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cs1",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "code-server",
						},
						Override: cosmov1alpha1.OverrideSpec{},
						Vars:     map[string]string{"{{TEST}}": "OK"},
					},
				},
				disableNamePrefix: false,
				scheme:            scheme,
			},
			args: args{
				src: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    cosmo/ingress-patch-enable: "true"
    kubernetes.io/ingress.class: alb
  name: test
  namespace: cosmo-user-tom
spec:
  host: example.com
`,
			},
			want: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    cosmo/ingress-patch-enable: "true"
    kubernetes.io/ingress.class: alb
  labels:
    cosmo-workspace.github.io/instance: cs1
    cosmo-workspace.github.io/template: code-server
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
  host: example.com
`,
			wantErr: false,
		},
		{
			name: "Append label and Namespace",
			fields: fields{
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cs1",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "code-server",
						},
						Override: cosmov1alpha1.OverrideSpec{},
						Vars:     map[string]string{"{{TEST}}": "OK"},
					},
				},
				disableNamePrefix: false,
				scheme:            scheme,
			},
			args: args{
				src: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    cosmo/ingress-patch-enable: "true"
    kubernetes.io/ingress.class: alb
  labels:
    key: val
  name: test
spec:
  host: example.com
`,
			},
			want: `apiVersion: networking.k8s.io/v1
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
  host: example.com
`,
			wantErr: false,
		},
		{
			name: "Disable Name Prefix",
			fields: fields{
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "useraddon-eks-irsa",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "eks-irsa",
						},
					},
				},
				disableNamePrefix: true,
				scheme:            scheme,
			},
			args: args{
				src: `apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT_ID:role/IAM_ROLE_NAME
  name: default
`,
			},
			want: `apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT_ID:role/IAM_ROLE_NAME
  labels:
    cosmo-workspace.github.io/instance: useraddon-eks-irsa
    cosmo-workspace.github.io/template: eks-irsa
  name: default
  namespace: cosmo-user-tom
  ownerReferences:
  - apiVersion: cosmo-workspace.github.io/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Instance
    name: useraddon-eks-irsa
    uid: ""
`,
			wantErr: false,
		},
		{
			name: "Disable Name Prefix annotation invalid",
			fields: fields{
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "useraddon-eks-irsa",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "eks-irsa",
						},
					},
				},
				disableNamePrefix: false,
				scheme:            scheme,
			},
			args: args{
				src: `apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT_ID:role/IAM_ROLE_NAME
  name: default
`,
			},
			want: `apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT_ID:role/IAM_ROLE_NAME
  labels:
    cosmo-workspace.github.io/instance: useraddon-eks-irsa
    cosmo-workspace.github.io/template: eks-irsa
  name: useraddon-eks-irsa-default
  namespace: cosmo-user-tom
  ownerReferences:
  - apiVersion: cosmo-workspace.github.io/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Instance
    name: useraddon-eks-irsa
    uid: ""
`,
			wantErr: false,
		},
		{
			name: "Failed to set ownerref on resoruce due to already exist",
			fields: fields{
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "useraddon-pv",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "pv",
						},
					},
				},
				disableNamePrefix: false,
				scheme:            scheme,
			},
			args: args{
				src: `apiVersion: v1
kind: PersistentVolume
metadata:
  name: default
  ownerReferences:
    - apiVersion: cosmo-workspace.github.io/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Instance
      name: useraddon-other-instance
      uid: 5b286420-4444-4366-aa40-973b1092f840
      resourceVersion: "19512263"
`,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewMetadataTransformer(tt.fields.inst, tt.fields.scheme, tt.fields.disableNamePrefix)
			_, obj, err := template.StringToUnstructured(tt.args.src)
			if err != nil {
				t.Errorf("MetadataTransformer.Transform() template.StringToUnstructured error = %v", err)
				return
			}
			gotObj, err := tr.Transform(obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetadataTransformer.Transform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				var want unstructured.Unstructured
				err := yaml.Unmarshal([]byte(tt.want), &want)
				if err != nil {
					t.Errorf("yaml.Marshal err = %v", err)
				}
				if !equality.Semantic.DeepEqual(gotObj, &want) {
					t.Errorf("MetadataTransformer.Transform() = %v, want %v", *gotObj, want)
				}

			} else {
				if gotObj != nil {
					t.Errorf("MetadataTransformer.Transform() gotObj not nil %v", gotObj)
				}
			}
		})
	}
}
