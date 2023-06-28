package webhooks

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

var _ = Describe("Workspace webhook", func() {
	wsConfig := cosmov1alpha1.Config{
		DeploymentName:      "ws-dep",
		ServiceName:         "ws-svc",
		ServiceMainPortName: "mainPort",
	}

	tmpl := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "code-server-test2",
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: cosmov1alpha1.TemplateLabelEnumTypeWorkspace,
			},
			Annotations: map[string]string{
				cosmov1alpha1.WorkspaceTemplateAnnKeyDeploymentName:  wsConfig.DeploymentName,
				cosmov1alpha1.WorkspaceTemplateAnnKeyServiceName:     wsConfig.ServiceName,
				cosmov1alpha1.WorkspaceTemplateAnnKeyServiceMainPort: wsConfig.ServiceMainPortName,
			},
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RawYaml: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  labels:
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: code-server-test2
  name: ws-ing
  namespace: '{{NAMESPACE}}'
spec:
  rules:
  - host: 'main-{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}'
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: ws-svc
            port: 
              number: 8080
---
apiVersion: v1
kind: Service
metadata:
  labels:
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: code-server-test2
  name: ws-svc
  namespace: '{{NAMESPACE}}'
spec:
  ports:
  - name: main
    port: 8080
    protocol: TCP
  selector:
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: code-server-test2
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: code-server-test2
  name: ws-dep
  namespace: '{{NAMESPACE}}'
spec:
  replicas: 1
  selector:
    matchLabels:
      cosmo-workspace.github.io/instance: '{{INSTANCE}}'
      cosmo-workspace.github.io/template: code-server-test2
  template:
    metadata:
      labels:
        cosmo-workspace.github.io/instance: '{{INSTANCE}}'
        cosmo-workspace.github.io/template: code-server-test2
    spec:
      containers:
      - image: 'code-server:{{IMAGE_TAG}}'
        name: code-server-test2
        ports:
        - containerPort: 8080
          name: main
          protocol: TCP
`,
			RequiredVars: []cosmov1alpha1.RequiredVarSpec{
				{
					Var: "{{DOMAIN}}",
				},
				{
					Var:     "{{IMAGE_TAG}}",
					Default: "latest",
				},
			},
		},
	}

	noWsLabelTmpl := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "code-server-nowslabel",
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: "nowslabel",
			},
			Annotations: map[string]string{
				cosmov1alpha1.WorkspaceTemplateAnnKeyDeploymentName:  wsConfig.DeploymentName,
				cosmov1alpha1.WorkspaceTemplateAnnKeyServiceName:     wsConfig.ServiceName,
				cosmov1alpha1.WorkspaceTemplateAnnKeyServiceMainPort: wsConfig.ServiceMainPortName,
			},
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RawYaml: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  labels:
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: code-server-nowslabel
  name: ws-ing
  namespace: '{{NAMESPACE}}'
spec:
  rules:
  - host: 'main-{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}'
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: ws-svc
            port: 
              number: 8080
---
apiVersion: v1
kind: Service
metadata:
  labels:
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: code-server-nowslabel
  name: ws-svc
  namespace: '{{NAMESPACE}}'
spec:
  ports:
  - name: main
    port: 8080
    protocol: TCP
  selector:
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: code-server-nowslabel
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: code-server-nowslabel
  name: ws-dep
  namespace: '{{NAMESPACE}}'
spec:
  replicas: 1
  selector:
    matchLabels:
      cosmo-workspace.github.io/instance: '{{INSTANCE}}'
      cosmo-workspace.github.io/template: code-server-nowslabel
  template:
    metadata:
      labels:
        cosmo-workspace.github.io/instance: '{{INSTANCE}}'
        cosmo-workspace.github.io/template: code-server-nowslabel
    spec:
      containers:
      - image: 'code-server:{{IMAGE_TAG}}'
        name: code-server-test2
        ports:
        - containerPort: 8080
          name: main
          protocol: TCP
`,
		},
	}

	Context("when creating workspace", func() {
		It("should pass with defaulting networking", func() {
			ctx := context.Background()

			var err error
			err = k8sClient.Create(ctx, &tmpl)
			Expect(err).ShouldNot(HaveOccurred())

			ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "cosmo-user-testuser-ws"}}
			err = k8sClient.Create(ctx, &ns)
			Expect(err).ShouldNot(HaveOccurred())

			rep := pointer.Int64(1)
			ws := cosmov1alpha1.Workspace{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workspace",
					APIVersion: cosmov1alpha1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testws1",
					Namespace: "cosmo-user-testuser-ws",
				},
				Spec: cosmov1alpha1.WorkspaceSpec{
					Template: cosmov1alpha1.TemplateRef{Name: tmpl.GetName()},
					Vars: map[string]string{
						"DOMAIN":    "example.com",
						"IMAGE_TAG": "latest",
					},
					Replicas: rep,
				},
			}

			err = k8sClient.Create(ctx, &ws)
			Expect(err).ShouldNot(HaveOccurred())

			var createdWs cosmov1alpha1.Workspace
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: ws.GetName(), Namespace: ws.GetNamespace()}, &createdWs)
			}, time.Second*10).Should(Succeed())

			Expect(ObjectSnapshot(&createdWs)).Should(MatchSnapShot())
		})
	})

	Context("when creating workspace without replica", func() {
		It("should pass and defaulting replica", func() {
			ctx := context.Background()

			ws := cosmov1alpha1.Workspace{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Workspace",
					APIVersion: cosmov1alpha1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testws3",
					Namespace: "cosmo-user-testuser-ws",
				},
				Spec: cosmov1alpha1.WorkspaceSpec{
					Template: cosmov1alpha1.TemplateRef{Name: tmpl.GetName()},
					Vars: map[string]string{
						"DOMAIN":    "example.com",
						"IMAGE_TAG": "latest",
					},
				},
			}

			err := k8sClient.Create(ctx, &ws)
			Expect(err).ShouldNot(HaveOccurred())

			var createdWs cosmov1alpha1.Workspace
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: ws.GetName(), Namespace: ws.GetNamespace()}, &createdWs)
			}, time.Second*10).Should(Succeed())

			Expect(ObjectSnapshot(&createdWs)).Should(MatchSnapShot())
		})
	})

	Context("when creating workspace without workspace label", func() {
		It("should deny", func() {
			ctx := context.Background()

			var err error
			err = k8sClient.Create(ctx, &noWsLabelTmpl)
			Expect(err).ShouldNot(HaveOccurred())

			ws := cosmov1alpha1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testws5",
					Namespace: "cosmo-user-testuser-ws",
				},
				Spec: cosmov1alpha1.WorkspaceSpec{
					Template: cosmov1alpha1.TemplateRef{Name: noWsLabelTmpl.GetName()},
					Vars: map[string]string{
						"DOMAIN":    "example.com",
						"IMAGE_TAG": "latest",
					},
				},
			}

			err = k8sClient.Create(ctx, &ws)
			Expect(err).To(MatchSnapShot())
		})
	})

	DescribeTable("when creating workspace",
		func(netRules []cosmov1alpha1.NetworkRule) {
			ws := cosmov1alpha1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testws6",
					Namespace: "cosmo-user-testuser-ws",
				},
				Spec: cosmov1alpha1.WorkspaceSpec{
					Template: cosmov1alpha1.TemplateRef{Name: tmpl.GetName()},
					Vars:     map[string]string{"DOMAIN": "example.com", "IMAGE_TAG": "latest"},
					Network:  netRules,
				},
			}
			err := k8sClient.Create(context.Background(), &ws)
			Expect(err).To(MatchSnapShot())
		},
		Entry("❌ fail with invalid port number", []cosmov1alpha1.NetworkRule{
			{
				PortNumber:       0,
				CustomHostPrefix: "a23456789012345",
				HTTPPath:         "",
				Public:           false,
			},
		}),

		Entry("❌ fail with invalid port name", []cosmov1alpha1.NetworkRule{
			{
				CustomHostPrefix: "a234567890123456",
				PortNumber:       1,
				HTTPPath:         "",
				Public:           false,
			},
		}),
		Entry("❌ fail with duplicated network rule name", []cosmov1alpha1.NetworkRule{
			{
				CustomHostPrefix: "nw1",
				PortNumber:       1111,
			},
			{
				CustomHostPrefix: "nw1",
				PortNumber:       2222,
			},
		}),
		Entry("❌ fail with duplicated network rule group and path", []cosmov1alpha1.NetworkRule{
			{
				CustomHostPrefix: "nw1",
				PortNumber:       1111,
				HTTPPath:         "/",
			},
			{
				CustomHostPrefix: "nw2",
				PortNumber:       2222,
				HTTPPath:         "/",
			},
		}),
		Entry("❌ fail with duplicated network rule host and path", []cosmov1alpha1.NetworkRule{
			{
				CustomHostPrefix: "nw1",
				PortNumber:       1111,
				HTTPPath:         "/",
				Public:           false,
			},
			{
				CustomHostPrefix: "nw2",
				PortNumber:       2222,
				HTTPPath:         "/",
			},
		}),
	)

	Context("when creating workspace within non user namespace", func() {
		It("should deny", func() {
			ctx := context.Background()

			var err error

			ws := cosmov1alpha1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testws7",
					Namespace: "cosmo-user-xxxxxx",
				},
				Spec: cosmov1alpha1.WorkspaceSpec{
					Template: cosmov1alpha1.TemplateRef{Name: tmpl.GetName()},
					Vars:     map[string]string{"DOMAIN": "example.com", "IMAGE_TAG": "latest"},
					Network: []cosmov1alpha1.NetworkRule{
						{
							CustomHostPrefix: "a23456789012345",
							PortNumber:       0,
							HTTPPath:         "",
							Public:           false,
						},
					},
				},
			}

			err = k8sClient.Create(ctx, &ws)
			Expect(err).To(MatchSnapShot())
		})
	})
})

func TestNetworkRulesByService(t *testing.T) {
	type args struct {
		svcPorts []corev1.ServicePort
	}
	tests := []struct {
		name string
		args args
		want []cosmov1alpha1.NetworkRule
	}{
		{
			name: "✅ OK",
			args: args{
				svcPorts: []corev1.ServicePort{
					{
						Name:     "main",
						Port:     int32(7777),
						Protocol: corev1.ProtocolTCP,
					},
					{
						Name:       "main2",
						Port:       int32(7778),
						Protocol:   corev1.ProtocolTCP,
						TargetPort: intstr.FromInt(32001),
					},
				},
			},
			want: []cosmov1alpha1.NetworkRule{
				{
					CustomHostPrefix: "main",
					PortNumber:       7777,
				},
				{
					CustomHostPrefix: "main2",
					PortNumber:       7778,
					TargetPortNumber: pointer.Int32(32001),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := networkRulesByServicePorts(tt.args.svcPorts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NetworkRulesByService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_appendNetworkRuleIfNotExist(t *testing.T) {
	type args struct {
		ws      *cosmov1alpha1.Workspace
		netRule cosmov1alpha1.NetworkRule
	}
	tests := []struct {
		name string
		args args
		want *cosmov1alpha1.Workspace
	}{
		{
			name: "✅ OK",
			args: args{
				ws: &cosmov1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ws1",
						Namespace: "cosmo-user-xxx",
					},
					Spec: cosmov1alpha1.WorkspaceSpec{
						Network: []cosmov1alpha1.NetworkRule{
							{
								CustomHostPrefix: "main",
								Protocol:         "http",
								PortNumber:       8080,
								HTTPPath:         "/",
							},
						},
					},
				},
				netRule: cosmov1alpha1.NetworkRule{
					Protocol:   "http",
					PortNumber: 8081,
					HTTPPath:   "/",
				},
			},
			want: &cosmov1alpha1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ws1",
					Namespace: "cosmo-user-xxx",
				},
				Spec: cosmov1alpha1.WorkspaceSpec{
					Network: []cosmov1alpha1.NetworkRule{
						{
							CustomHostPrefix: "main",
							Protocol:         "http",
							PortNumber:       8080,
							HTTPPath:         "/",
						},
						{
							Protocol:   "http",
							PortNumber: 8081,
							HTTPPath:   "/",
						},
					},
				},
			},
		},
		{
			name: "✅ not append",
			args: args{
				ws: &cosmov1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ws1",
						Namespace: "cosmo-user-xxx",
					},
					Spec: cosmov1alpha1.WorkspaceSpec{
						Network: []cosmov1alpha1.NetworkRule{
							{
								CustomHostPrefix: "main",
								Protocol:         "http",
								PortNumber:       8080,
								HTTPPath:         "/",
							},
						},
					},
				},
				netRule: cosmov1alpha1.NetworkRule{
					CustomHostPrefix: "main",
					Protocol:         "http",
					PortNumber:       8080,
					HTTPPath:         "/",
					Public:           true,
				},
			},
			want: &cosmov1alpha1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ws1",
					Namespace: "cosmo-user-xxx",
				},
				Spec: cosmov1alpha1.WorkspaceSpec{
					Network: []cosmov1alpha1.NetworkRule{
						{
							CustomHostPrefix: "main",
							Protocol:         "http",
							PortNumber:       8080,
							HTTPPath:         "/",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appendNetworkRuleIfNotExist(tt.args.ws, tt.args.netRule)
			got, _ := json.Marshal(tt.args.ws)
			want, _ := json.Marshal(tt.want)
			if string(got) != string(want) {
				t.Errorf("appendNetworkRuleIfNotExist() = %v, want %v", string(got), string(want))
			}
		})
	}
}

func Test_sortNetworkRule(t *testing.T) {
	type args struct {
		netRules []cosmov1alpha1.NetworkRule
		cfg      cosmov1alpha1.Config
	}
	tests := []struct {
		name string
		args args
		want []cosmov1alpha1.NetworkRule
	}{
		{
			name: "✅ OK",
			args: args{
				netRules: []cosmov1alpha1.NetworkRule{
					{
						Protocol:   "http",
						PortNumber: 8080,
						HTTPPath:   "/",
					},
					{
						CustomHostPrefix: "main",
						Protocol:         "http",
						PortNumber:       8080,
						HTTPPath:         "/",
					},
					{
						Protocol:   "http",
						PortNumber: 8081,
						HTTPPath:   "/dev",
					},
					{
						CustomHostPrefix: "bbb",
						Protocol:         "http",
						PortNumber:       8080,
						HTTPPath:         "/",
					},
					{
						CustomHostPrefix: "aaa",
						Protocol:         "http",
						PortNumber:       8080,
						HTTPPath:         "/",
					},
				},
				cfg: cosmov1alpha1.Config{
					ServiceMainPortName: "main",
				},
			},
			want: []cosmov1alpha1.NetworkRule{
				// main
				{
					CustomHostPrefix: "main",
					Protocol:         "http",
					PortNumber:       8080,
					HTTPPath:         "/",
				},
				// no name and longer path
				{
					Protocol:   "http",
					PortNumber: 8081,
					HTTPPath:   "/dev",
				},
				// no name and shorter path
				{
					Protocol:   "http",
					PortNumber: 8080,
					HTTPPath:   "/",
				},
				// order of name
				{
					CustomHostPrefix: "aaa",
					Protocol:         "http",
					PortNumber:       8080,
					HTTPPath:         "/",
				},
				{
					CustomHostPrefix: "bbb",
					Protocol:         "http",
					PortNumber:       8080,
					HTTPPath:         "/",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorted := sortNetworkRule(tt.args.netRules, tt.args.cfg)
			got, _ := json.MarshalIndent(sorted, " ", "  ")
			want, _ := json.MarshalIndent(tt.want, " ", "  ")
			if string(got) != string(want) {
				t.Errorf("sortNetworkRule() got = %v", string(got))
				t.Errorf("diff: %v", cmp.Diff(string(got), string(want)))
			}
		})
	}
}
