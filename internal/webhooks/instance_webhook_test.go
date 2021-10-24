package webhooks

import (
	"testing"

	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

func Test_fixServiceNameInIngressBackend(t *testing.T) {
	type args struct {
		ingRules []netv1.IngressRule
		instName string
	}
	tests := []struct {
		name string
		args args
		want []netv1.IngressRule
	}{
		{
			name: "OK",
			args: args{
				ingRules: []netv1.IngressRule{
					{
						Host: "example.com",
						IngressRuleValue: netv1.IngressRuleValue{
							HTTP: &netv1.HTTPIngressRuleValue{
								Paths: []netv1.HTTPIngressPath{
									{
										Backend: netv1.IngressBackend{
											Service: &netv1.IngressServiceBackend{
												Name: "test-svc",
											},
										},
									},
								},
							},
						},
					},
					{
						Host: "example.com",
						IngressRuleValue: netv1.IngressRuleValue{
							HTTP: &netv1.HTTPIngressRuleValue{
								Paths: []netv1.HTTPIngressPath{
									{
										Backend: netv1.IngressBackend{
											Service: &netv1.IngressServiceBackend{
												Name: "test-svc2",
											},
										},
									},
								},
							},
						},
					},
				},
				instName: "instance",
			},
			want: []netv1.IngressRule{
				{
					Host: "example.com",
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Backend: netv1.IngressBackend{
										Service: &netv1.IngressServiceBackend{
											Name: "instance-test-svc",
										},
									},
								},
							},
						},
					},
				},
				{
					Host: "example.com",
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Backend: netv1.IngressBackend{
										Service: &netv1.IngressServiceBackend{
											Name: "instance-test-svc2",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixIngressBackendName(tt.args.ingRules, tt.args.instName)
			if !equality.Semantic.DeepEqual(tt.args.ingRules, tt.want) {
				t.Error(tt.args, tt.want)
			}
		})
	}
}
