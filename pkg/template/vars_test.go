package template

import (
	"reflect"
	"testing"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTemplateBuilder_ReplaceDefaultVars(t *testing.T) {
	type fields struct {
		data string
		inst *cosmov1alpha1.Instance
	}
	tests := []struct {
		name   string
		fields fields
		want   *TemplateBuilder
	}{
		{
			name: "OK",
			fields: fields{
				data: "{{INSTANCE}}-{{NAMESPACE}}-{{TEMPLATE}}",
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
			},
			want: &TemplateBuilder{
				data: "cs1-cosmo-user-tom-code-server",
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
			},
		},
		{
			name: "without brackets",
			fields: fields{
				data: "{{INSTANCE}}-{{NAMESPACE}}-{{TEMPLATE}}",
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
						Vars:     map[string]string{"TEST": "OK"},
					},
				},
			},
			want: &TemplateBuilder{
				data: "cs1-cosmo-user-tom-code-server",
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
						Vars:     map[string]string{"TEST": "OK"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TemplateBuilder{
				data: tt.fields.data,
				inst: tt.fields.inst,
			}
			if got := tr.ReplaceDefaultVars(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TemplateBuilder.ReplaceDefaultVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateBuilder_ReplaceCustomVars(t *testing.T) {
	type fields struct {
		data string
		inst *cosmov1alpha1.Instance
	}
	tests := []struct {
		name   string
		fields fields
		want   *TemplateBuilder
	}{
		{
			name: "OK",
			fields: fields{
				data: "{{INSTANCE}}-{{NAMESPACE}}-{{TEMPLATE}}-{{TEST}}",
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
			},
			want: &TemplateBuilder{
				data: "{{INSTANCE}}-{{NAMESPACE}}-{{TEMPLATE}}-OK",
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TemplateBuilder{
				data: tt.fields.data,
				inst: tt.fields.inst,
			}
			if got := tr.ReplaceCustomVars(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TemplateBuilder.ReplaceCustomVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidCustomVars(t *testing.T) {
	type args struct {
		varString string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				varString: "{{INSTACE}}",
			},
			wantErr: false,
		},
		{
			name: "Invalid",
			args: args{
				varString: "INSTACE",
			},
			wantErr: true,
		},
		{
			name: "Invalid sufix",
			args: args{
				varString: "{{INSTACE)",
			},
			wantErr: true,
		},
		{
			name: "Invalid prefix",
			args: args{
				varString: "$(INSTACE}}",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidCustomVars(tt.args.varString); (err != nil) != tt.wantErr {
				t.Errorf("ValidCustomVars() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
