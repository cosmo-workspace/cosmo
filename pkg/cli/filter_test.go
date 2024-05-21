package cli

import (
	"reflect"
	"testing"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_opBool(t *testing.T) {
	type args struct {
		op string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "eq",
			args: args{op: OperatorEqual},
			want: true,
		},
		{
			name: "ne",
			args: args{op: OperatorNotEqual},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := opBool(tt.args.op); got != tt.want {
				t.Errorf("opBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFilters(t *testing.T) {
	type args struct {
		filterExpressions []string
	}
	tests := []struct {
		name    string
		args    args
		want    []Filter
		wantErr bool
	}{
		{
			name: "eq",
			args: args{
				filterExpressions: []string{"key1==value1", "key2==value2", "key3!=value3"},
			},
			want: []Filter{
				{
					Key:      "key1",
					Value:    "value1",
					Operator: OperatorEqual,
				},
				{
					Key:      "key2",
					Value:    "value2",
					Operator: OperatorEqual,
				},
				{
					Key:      "key3",
					Value:    "value3",
					Operator: OperatorNotEqual,
				},
			},
		},
		{
			name: "error",
			args: args{
				filterExpressions: []string{"key1==value1", "key2==value2", "key3=value3"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFilters(tt.args.filterExpressions)
			if !reflect.DeepEqual(got, tt.want) || (err != nil) != tt.wantErr {
				t.Errorf("ParseFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseFilterExpression(t *testing.T) {
	type args struct {
		exp string
		op  string
	}
	tests := []struct {
		name string
		args args
		want *Filter
	}{
		{
			name: "eq",
			args: args{
				exp: "key==value",
				op:  OperatorEqual,
			},
			want: &Filter{
				Key:      "key",
				Value:    "value",
				Operator: OperatorEqual,
			},
		},
		{
			name: "ne",
			args: args{
				exp: "key!=value",
				op:  OperatorNotEqual,
			},
			want: &Filter{
				Key:      "key",
				Value:    "value",
				Operator: OperatorNotEqual,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseFilterExpression(tt.args.exp, tt.args.op); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFilterExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoFilter(t *testing.T) {
	type args struct {
		objects             []cosmov1alpha1.User
		objectFilterKeyFunc func(cosmov1alpha1.User) []string
		f                   Filter
	}
	tests := []struct {
		name string
		args args
		want []cosmov1alpha1.User
	}{
		{
			name: "eq",
			args: args{
				objects: []cosmov1alpha1.User{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "user1"},
						Spec: cosmov1alpha1.UserSpec{
							Addons: []cosmov1alpha1.UserAddon{
								{Template: cosmov1alpha1.UserAddonTemplateRef{Name: "addon1"}},
								{Template: cosmov1alpha1.UserAddonTemplateRef{Name: "addon2"}},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "user2"},
						Spec: cosmov1alpha1.UserSpec{
							Addons: []cosmov1alpha1.UserAddon{
								{Template: cosmov1alpha1.UserAddonTemplateRef{Name: "addon1"}},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "user3"},
						Spec: cosmov1alpha1.UserSpec{
							Addons: []cosmov1alpha1.UserAddon{
								{Template: cosmov1alpha1.UserAddonTemplateRef{Name: "addon2"}},
							},
						},
					},
				},
				objectFilterKeyFunc: func(u cosmov1alpha1.User) []string {
					addons := make([]string, 0, len(u.Spec.Addons))
					for _, a := range u.Spec.Addons {
						addons = append(addons, a.Template.Name)
					}
					return addons
				},
				f: Filter{
					Value:    "addon2",
					Operator: OperatorEqual,
				},
			},
			want: []cosmov1alpha1.User{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "user1"},
					Spec: cosmov1alpha1.UserSpec{
						Addons: []cosmov1alpha1.UserAddon{
							{Template: cosmov1alpha1.UserAddonTemplateRef{Name: "addon1"}},
							{Template: cosmov1alpha1.UserAddonTemplateRef{Name: "addon2"}},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "user3"},
					Spec: cosmov1alpha1.UserSpec{
						Addons: []cosmov1alpha1.UserAddon{
							{Template: cosmov1alpha1.UserAddonTemplateRef{Name: "addon2"}},
						},
					},
				},
			},
		},
		{
			name: "ne",
			args: args{
				objects: []cosmov1alpha1.User{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "user1"},
						Spec: cosmov1alpha1.UserSpec{
							Addons: []cosmov1alpha1.UserAddon{
								{Template: cosmov1alpha1.UserAddonTemplateRef{Name: "addon1"}},
								{Template: cosmov1alpha1.UserAddonTemplateRef{Name: "addon2"}},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "user2"},
						Spec: cosmov1alpha1.UserSpec{
							Addons: []cosmov1alpha1.UserAddon{
								{Template: cosmov1alpha1.UserAddonTemplateRef{Name: "addon1"}},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "user3"},
						Spec: cosmov1alpha1.UserSpec{
							Addons: []cosmov1alpha1.UserAddon{
								{Template: cosmov1alpha1.UserAddonTemplateRef{Name: "addon2"}},
							},
						},
					},
				},
				objectFilterKeyFunc: func(u cosmov1alpha1.User) []string {
					addons := make([]string, 0, len(u.Spec.Addons))
					for _, a := range u.Spec.Addons {
						addons = append(addons, a.Template.Name)
					}
					return addons
				},
				f: Filter{
					Value:    "addon2",
					Operator: OperatorNotEqual,
				},
			},
			want: []cosmov1alpha1.User{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "user2"},
					Spec: cosmov1alpha1.UserSpec{
						Addons: []cosmov1alpha1.UserAddon{
							{Template: cosmov1alpha1.UserAddonTemplateRef{Name: "addon1"}},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DoFilter(tt.args.objects, tt.args.objectFilterKeyFunc, tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DoFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
