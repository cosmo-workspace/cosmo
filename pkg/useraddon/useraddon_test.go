package useraddon

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/pointer"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/template"
)

func TestEmptyTemplateObject(t *testing.T) {
	type args struct {
		addon cosmov1alpha1.UserAddon
	}
	tests := []struct {
		name string
		args args
		want cosmov1alpha1.TemplateObject
	}{
		{
			name: "namespaced",
			args: args{
				addon: cosmov1alpha1.UserAddon{
					Template: cosmov1alpha1.UserAddonTemplateRef{
						Name: "tmpl",
					},
				},
			},
			want: &cosmov1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tmpl",
				},
			},
		},
		{
			name: "cluster",
			args: args{
				addon: cosmov1alpha1.UserAddon{
					Template: cosmov1alpha1.UserAddonTemplateRef{
						Name:          "ctmpl",
						ClusterScoped: true,
					},
				},
			},
			want: &cosmov1alpha1.ClusterTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ctmpl",
				},
			},
		},
		{
			name: "empty",
			args: args{
				addon: cosmov1alpha1.UserAddon{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EmptyTemplateObject(tt.args.addon); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EmptyTemplateObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmptyInstanceObject(t *testing.T) {
	type args struct {
		addon    cosmov1alpha1.UserAddon
		username string
	}
	tests := []struct {
		name string
		args args
		want cosmov1alpha1.InstanceObject
	}{
		{
			name: "namespaced",
			args: args{
				addon: cosmov1alpha1.UserAddon{
					Template: cosmov1alpha1.UserAddonTemplateRef{
						Name: "tmpl",
					},
				},
				username: "tom",
			},
			want: &cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "useraddon-tmpl",
					Namespace: "cosmo-user-tom",
				},
			},
		},
		{
			name: "cluster",
			args: args{
				addon: cosmov1alpha1.UserAddon{
					Template: cosmov1alpha1.UserAddonTemplateRef{
						Name:          "ctmpl",
						ClusterScoped: true,
					},
				},
				username: "tom",
			},
			want: &cosmov1alpha1.ClusterInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name: "useraddon-tom-ctmpl",
				},
			},
		},
		{
			name: "empty",
			args: args{
				addon:    cosmov1alpha1.UserAddon{},
				username: "tom",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EmptyInstanceObject(tt.args.addon, tt.args.username); !equality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("EmptyInstanceObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInstanceName(t *testing.T) {
	type args struct {
		addonTmplName string
		userName      string
	}
	tests := []struct {
		name     string
		args     args
		wantName string
	}{
		{
			name: "namespaced",
			args: args{
				addonTmplName: "tmpl",
				userName:      "",
			},
			wantName: "useraddon-tmpl",
		},
		{
			name: "cluster",
			args: args{
				addonTmplName: "tmpl",
				userName:      "tom",
			},
			wantName: "useraddon-tom-tmpl",
		},
		{
			name: "long name",
			args: args{
				addonTmplName: "tmpltmpltmpltmpltmpltmpltmpltmpltmpltmpltmpltmpltmpl",
				userName:      "tom",
			},
			wantName: "useraddon-tom-tmpltmpltmpltmpltmpltmpltmpltmpltmpltmpltmpltmplt", //truncate sufix
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotName := InstanceName(tt.args.addonTmplName, tt.args.userName); gotName != tt.wantName {
				t.Errorf("InstanceName() = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}

func TestPatchUserAddonInstanceAsDesired(t *testing.T) {
	validScheme := runtime.NewScheme()
	utilruntime.Must(cosmov1alpha1.AddToScheme(validScheme))
	invalidScheme := runtime.NewScheme()

	type args struct {
		inst   cosmov1alpha1.InstanceObject
		addon  cosmov1alpha1.UserAddon
		user   cosmov1alpha1.User
		scheme *runtime.Scheme
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		want         cosmov1alpha1.InstanceObject
		wantOwnerref bool
	}{
		{
			name: "patch instance",
			args: args{
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "useraddon-tmpl",
						Namespace: "cosmo-user-tom",
					},
				},
				addon: cosmov1alpha1.UserAddon{
					Template: cosmov1alpha1.UserAddonTemplateRef{
						Name: "tmpl",
					},
					Vars: map[string]string{
						"VAR1": "VAL1",
					},
				},
				user: cosmov1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tom",
						UID:  "1qaz2wsx3edc",
					},
					Spec: cosmov1alpha1.UserSpec{
						// use selected addon in param not in user spec
						// Addons: []cosmov1alpha1.UserAddon{
						// 	{
						// 		Template: cosmov1alpha1.UserAddonTemplateRef{
						// 			Name: "tmpl",
						// 		},
						// 		Vars: map[string]string{
						// 			"VAR1": "VAL1",
						// 		},
						// 	},
						// },
					},
				},
				scheme: validScheme,
			},
			want: &cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "useraddon-tmpl",
					Namespace: "cosmo-user-tom",
					Labels: map[string]string{
						cosmov1alpha1.TemplateLabelKeyType: cosmov1alpha1.TemplateLabelEnumTypeUserAddon,
					},
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: "tmpl",
					},
					Vars: map[string]string{
						cosmov1alpha1.TemplateVarUserName: "tom",
						template.DefaultVarsNamespace:     "cosmo-user-tom",
						"VAR1":                            "VAL1",
					},
				},
			},
			wantOwnerref: true,
		},
		{
			name: "patch clusterinstance",
			args: args{
				inst: &cosmov1alpha1.ClusterInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name: "useraddon-tom-ctmpl",
					},
				},
				addon: cosmov1alpha1.UserAddon{
					Template: cosmov1alpha1.UserAddonTemplateRef{
						Name:          "ctmpl",
						ClusterScoped: true,
					},
				},
				user: cosmov1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tom",
						UID:  "1qaz2wsx3edc",
					},
					Spec: cosmov1alpha1.UserSpec{
						// use selected addon in param not in user spec
						// Addons: []cosmov1alpha1.UserAddon{
						// 	{
						// 		Template: cosmov1alpha1.UserAddonTemplateRef{
						// 			Name: "ctmpl",
						//          ClusterScoped: true,
						// 		},
						// 	},
						// },
					},
				},
			},
			want: &cosmov1alpha1.ClusterInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name: "useraddon-tom-ctmpl",
					Labels: map[string]string{
						cosmov1alpha1.TemplateLabelKeyType: cosmov1alpha1.TemplateLabelEnumTypeUserAddon,
					},
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: "ctmpl",
					},
					Vars: map[string]string{
						cosmov1alpha1.TemplateVarUserName: "tom",
						template.DefaultVarsNamespace:     "cosmo-user-tom",
					},
				},
			},
		},
		{
			name: "invalid scheme",
			args: args{
				inst: &cosmov1alpha1.ClusterInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name: "useraddon-tom-ctmpl",
					},
				},
				addon: cosmov1alpha1.UserAddon{
					Template: cosmov1alpha1.UserAddonTemplateRef{
						Name:          "ctmpl",
						ClusterScoped: true,
					},
				},
				user: cosmov1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tom",
						UID:  "1qaz2wsx3edc",
					},
					Spec: cosmov1alpha1.UserSpec{},
				},
				scheme: invalidScheme,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PatchUserAddonInstanceAsDesired(tt.args.inst, tt.args.addon, tt.args.user, tt.args.scheme)
			if (err != nil) != tt.wantErr {
				t.Errorf("PatchUserAddonInstanceAsDesired() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				ownerRef := tt.args.inst.GetOwnerReferences()
				tt.args.inst.SetOwnerReferences(nil)

				if !equality.Semantic.DeepEqual(tt.args.inst, tt.want) {
					t.Errorf("EmptyInstanceObject() = %v, want %v", tt.args.inst, tt.want)
				}

				if (ownerRef != nil) != tt.wantOwnerref {
					t.Errorf("EmptyInstanceObject() ownerRef = %v, wantOwnerref %v", ownerRef, tt.wantOwnerref)
				}
				if len(ownerRef) > 0 {
					if len(ownerRef) != 1 {
						t.Errorf("EmptyInstanceObject() ownerRef should be 1 but %v", len(ownerRef))
					}
					expectedRef := metav1.OwnerReference{
						APIVersion:         cosmov1alpha1.GroupVersion.String(),
						Kind:               "User",
						Name:               tt.args.user.GetName(),
						UID:                tt.args.user.GetUID(),
						BlockOwnerDeletion: pointer.BoolPtr(true),
						Controller:         pointer.BoolPtr(true),
					}
					if !equality.Semantic.DeepEqual(ownerRef[0], expectedRef) {
						t.Errorf("EmptyInstanceObject() owner ref = %v, want %v", ownerRef[0], expectedRef)
					}
				}

			}
		})
	}
}
