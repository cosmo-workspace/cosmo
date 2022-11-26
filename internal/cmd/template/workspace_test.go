package template

import (
	"fmt"
	"reflect"
	"testing"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/yaml"

	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
)

func Test_completeWorkspaceConfig(t *testing.T) {
	type args struct {
		wsConfig *cosmov1alpha1.Config
		tmpl     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *cosmov1alpha1.Config
	}{
		{
			name: "validate",
			args: args{
				wsConfig: &cosmov1alpha1.Config{
					DeploymentName:      "workspace",
					ServiceName:         "workspace",
					IngressName:         "workspace",
					ServiceMainPortName: "main",
				},
				tmpl: `
apiVersion: v1
kind: Service
metadata:
  name: 'workspace'
spec:
  ports:
  - name: main
    port: 3000
    protocol: TCP
  - name: sub
    port: 3001
    protocol: TCP
  type: ClusterIP
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: 'workspace'
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: 'workspace'
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: theiaide/theia
        imagePullPolicy: IfNotPresent
        name: theia
        ports:
        - containerPort: 3000
          name: http
          protocol: TCP
        volumeMounts:
        - mountPath: /home/project
          name: data
      serviceAccountName: default
      volumes:
      - emptyDir: {}
        name: data
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: 'workspace'
spec:
  rules:
  - host: main-{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}
    http:
      paths:
      - backend:
          service:
            name: 'workspace'
            port:
              name: main
        path: /*
        pathType: Exact
`,
			},
			wantErr: false,
			want: &cosmov1alpha1.Config{
				DeploymentName:      "workspace",
				ServiceName:         "workspace",
				IngressName:         "workspace",
				ServiceMainPortName: "main",
			},
		},
		{
			name: "complete",
			args: args{
				wsConfig: &cosmov1alpha1.Config{},
				tmpl: `
apiVersion: v1
kind: Service
metadata:
  name: 'workspace'
spec:
  ports:
  - name: main
    port: 3000
    protocol: TCP
  type: ClusterIP
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: 'workspace'
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: 'workspace'
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: theiaide/theia
        imagePullPolicy: IfNotPresent
        name: theia
        ports:
        - containerPort: 3000
          name: http
          protocol: TCP
        volumeMounts:
        - mountPath: /home/project
          name: data
      serviceAccountName: default
      volumes:
      - emptyDir: {}
        name: data
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: 'workspace'
spec:
  rules:
  - host: main-{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}
    http:
      paths:
      - backend:
          service:
            name: 'workspace'
            port:
              name: main
        path: /*
        pathType: Exact
`,
			},
			wantErr: false,
			want: &cosmov1alpha1.Config{
				DeploymentName:      "workspace",
				ServiceName:         "workspace",
				IngressName:         "workspace",
				ServiceMainPortName: "main",
			},
		},
		{
			name: "complete error",
			args: args{
				wsConfig: &cosmov1alpha1.Config{},
				tmpl: `
apiVersion: v1
kind: Service
metadata:
  name: 'workspace'
spec:
  ports:
  - name: main
    port: 3000
    protocol: TCP
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: 'workspace2'
spec:
  ports:
  - name: main
    port: 3000
    protocol: TCP
  type: ClusterIP
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: 'workspace'
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: 'workspace'
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: theiaide/theia
        imagePullPolicy: IfNotPresent
        name: theia
        ports:
        - containerPort: 3000
          name: http
          protocol: TCP
        volumeMounts:
        - mountPath: /home/project
          name: data
      serviceAccountName: default
      volumes:
      - emptyDir: {}
        name: data
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: 'workspace'
spec:
  rules:
  - host: main-{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}
    http:
      paths:
      - backend:
          service:
            name: 'workspace'
            port:
              name: main
        path: /*
        pathType: Exact
`,
			},
			wantErr: true,
		},
		{
			name: "NG service name",
			args: args{
				wsConfig: &cosmov1alpha1.Config{
					DeploymentName:      "workspace",
					ServiceName:         "workspace",
					IngressName:         "workspace",
					ServiceMainPortName: "main",
				},
				tmpl: `
apiVersion: v1
kind: Service
metadata:
  name: 'NG'
spec:
  ports:
  - name: main
    port: 3000
    protocol: TCP
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: 'workspace'
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: theiaide/theia
        imagePullPolicy: IfNotPresent
        name: theia
        ports:
        - containerPort: 3000
          name: http
          protocol: TCP
        volumeMounts:
        - mountPath: /home/project
          name: data
      serviceAccountName: default
      volumes:
      - emptyDir: {}
        name: data
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: 'workspace'
spec:
  rules:
  - host: main-{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}
    http:
      paths:
      - backend:
          service:
            name: 'workspace'
            port:
              name: main
        path: /*
        pathType: Exact
`,
			},
			wantErr: true,
		},
		{
			name: "NG deploy name",
			args: args{
				wsConfig: &cosmov1alpha1.Config{
					DeploymentName:      "workspace",
					ServiceName:         "workspace",
					IngressName:         "workspace",
					ServiceMainPortName: "main",
				},
				tmpl: `
apiVersion: v1
kind: Service
metadata:
  name: 'workspace'
spec:
  ports:
  - name: main
    port: 3000
    protocol: TCP
  type: ClusterIP
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: 'workspace'
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: 'NG'
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: theiaide/theia
        imagePullPolicy: IfNotPresent
        name: theia
        ports:
        - containerPort: 3000
          name: http
          protocol: TCP
        volumeMounts:
        - mountPath: /home/project
          name: data
      serviceAccountName: default
      volumes:
      - emptyDir: {}
        name: data
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: 'workspace'
spec:
  rules:
  - host: main-{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}
    http:
      paths:
      - backend:
          service:
            name: 'workspace'
            port:
              name: main
        path: /*
        pathType: Exact
`,
			},
			wantErr: true,
		},
		{
			name: "NG ingress name",
			args: args{
				wsConfig: &cosmov1alpha1.Config{
					DeploymentName:      "workspace",
					ServiceName:         "workspace",
					IngressName:         "workspace",
					ServiceMainPortName: "main",
				},
				tmpl: `
apiVersion: v1
kind: Service
metadata:
  name: 'workspace'
spec:
  ports:
  - name: main
    port: 3000
    protocol: TCP
  type: ClusterIP
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: 'workspace'
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: 'workspace'
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: theiaide/theia
        imagePullPolicy: IfNotPresent
        name: theia
        ports:
        - containerPort: 3000
          name: http
          protocol: TCP
        volumeMounts:
        - mountPath: /home/project
          name: data
      serviceAccountName: default
      volumes:
      - emptyDir: {}
        name: data
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: 'ng'
spec:
  rules:
  - host: main-{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}
    http:
      paths:
      - backend:
          service:
            name: 'workspace'
            port:
              name: main
        path: /*
        pathType: Exact
`,
			},
			wantErr: true,
		},
		{
			name: "NG ingress backend service",
			args: args{
				wsConfig: &cosmov1alpha1.Config{
					DeploymentName:      "workspace",
					ServiceName:         "workspace",
					IngressName:         "workspace",
					ServiceMainPortName: "main",
				},
				tmpl: `
apiVersion: v1
kind: Service
metadata:
  name: 'workspace'
spec:
  ports:
  - name: main
    port: 3000
    protocol: TCP
  type: ClusterIP
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: 'workspace'
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: 'workspace'
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: theiaide/theia
        imagePullPolicy: IfNotPresent
        name: theia
        ports:
        - containerPort: 3000
          name: http
          protocol: TCP
        volumeMounts:
        - mountPath: /home/project
          name: data
      serviceAccountName: default
      volumes:
      - emptyDir: {}
        name: data
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: 'workspace'
spec:
  rules:
  - host: main-{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}
    http:
      paths:
      - backend:
          service:
            name: 'ng'
            port:
              number: 3000
        path: /*
        pathType: Exact
`,
			},
			wantErr: true,
		},
		{
			name: "no service",
			args: args{
				wsConfig: &cosmov1alpha1.Config{
					DeploymentName:      "workspace",
					ServiceName:         "workspace",
					IngressName:         "workspace-ing",
					ServiceMainPortName: "main",
				},
				tmpl: `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: 'workspace-pvc'
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: 'workspace'
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: theiaide/theia
        imagePullPolicy: IfNotPresent
        name: theia
        ports:
        - containerPort: 3000
          name: http
          protocol: TCP
        volumeMounts:
        - mountPath: /home/project
          name: data
      serviceAccountName: default
      volumes:
      - emptyDir: {}
        name: data
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: 'workspace-ing'
spec:
  rules:
  - host: main-{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}
    http:
      paths:
      - backend:
          service:
            name: 'workspace'
            port:
              number: 3000
        path: /*
        pathType: Exact
`,
			},
			wantErr: true,
		},
		{
			name: "no ingress, service LoadBalancer",
			args: args{
				wsConfig: &cosmov1alpha1.Config{},
				tmpl: `
apiVersion: v1
kind: Service
metadata:
  name: 'workspace'
spec:
  ports:
  - name: main
    port: 3000
    protocol: TCP
  type: LoadBalancer
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: 'workspace'
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: theiaide/theia
        imagePullPolicy: IfNotPresent
        name: theia
        ports:
        - containerPort: 3000
          name: http
          protocol: TCP
        volumeMounts:
        - mountPath: /home/project
          name: data
      serviceAccountName: default
      volumes:
      - emptyDir: {}
        name: data
`,
			},
			wantErr: false,
			want: &cosmov1alpha1.Config{
				DeploymentName:      "workspace",
				ServiceName:         "workspace",
				IngressName:         "",
				ServiceMainPortName: "main",
			},
		},
		{
			name: "no ingress, service NodePort",
			args: args{
				wsConfig: &cosmov1alpha1.Config{},
				tmpl: `
apiVersion: v1
kind: Service
metadata:
  name: 'workspace'
spec:
  ports:
  - name: main
    port: 3000
    protocol: TCP
  type: NodePort
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: 'workspace'
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: theiaide/theia
        imagePullPolicy: IfNotPresent
        name: theia
        ports:
        - containerPort: 3000
          name: http
          protocol: TCP
        volumeMounts:
        - mountPath: /home/project
          name: data
      serviceAccountName: default
      volumes:
      - emptyDir: {}
        name: data
`,
			},
			wantErr: false,
			want: &cosmov1alpha1.Config{
				DeploymentName:      "workspace",
				ServiceName:         "workspace",
				IngressName:         "",
				ServiceMainPortName: "main",
			},
		},
		{
			name: "no deployment",
			args: args{
				wsConfig: &cosmov1alpha1.Config{
					DeploymentName:      "workspace",
					ServiceName:         "workspace",
					IngressName:         "",
					ServiceMainPortName: "main",
				},
				tmpl: `
apiVersion: v1
kind: Service
metadata:
  name: 'workspace'
spec:
  ports:
  - name: main
    port: 3000
    protocol: TCP
  type: NodePort
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: 'workspace'
spec:
  rules:
  - host: main-{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}
    http:
      paths:
      - backend:
          service:
            name: 'workspace'
            port:
              number: 3000
        path: /*
        pathType: Exact
`,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := preTemplateBuild(tt.args.tmpl)
			if err != nil {
				t.Errorf("preTemplateBuild() error = %v", err)
			}
			if err := completeWorkspaceConfig(tt.args.wsConfig, u); (err != nil) != tt.wantErr {
				t.Errorf("completeWorkspaceConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(tt.args.wsConfig, tt.want) {
					t.Errorf("completeWorkspaceConfig() got = %v, want %v", tt.args.wsConfig, tt.want)
				}
			}
		})
	}
}

func Test_deploymentAuthProxyPatch(t *testing.T) {
	type args struct {
		injectDeploymentName string
		authProxyImage       string
		secretName           string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "OK",
			args: args{
				injectDeploymentName: "workspace",
				authProxyImage:       "cosmo-auth-proxy:latest",
				secretName:           "cosmo-auth-proxy-cert",
			},
			want: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: workspace
spec:
  template:
    spec:
      containers:
      - args:
        - --tls-cert=/app/cert/tls.crt
        - --tls-key=/app/cert/tls.key
        env:
        - name: COSMO_AUTH_PROXY_INSTANCE
          value: '{{INSTANCE}}'
        - name: COSMO_AUTH_PROXY_NAMESPACE
          value: '{{NAMESPACE}}'
        image: cosmo-auth-proxy:latest
        name: cosmo-auth-proxy
        volumeMounts:
        - mountPath: /app/cert
          name: cert
          readOnly: true
      volumes:
      - name: cert
        secret:
          defaultMode: 420
          secretName: cosmo-auth-proxy-cert
`,
		},
		{
			name: "insecure",
			args: args{
				injectDeploymentName: "workspace",
				authProxyImage:       "cosmo-auth-proxy:latest",
			},
			want: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: workspace
spec:
  template:
    spec:
      containers:
      - args:
        - --insecure
        env:
        - name: COSMO_AUTH_PROXY_INSTANCE
          value: '{{INSTANCE}}'
        - name: COSMO_AUTH_PROXY_NAMESPACE
          value: '{{NAMESPACE}}'
        image: cosmo-auth-proxy:latest
        name: cosmo-auth-proxy
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deploymentAuthProxyPatch(tt.args.injectDeploymentName, tt.args.authProxyImage, tt.args.secretName)
			var want appsv1apply.DeploymentApplyConfiguration
			err := yaml.Unmarshal([]byte(tt.want), &want)
			if err != nil {
				t.Errorf("deploymentAuthProxyPatch() Unmarshal err = %v", err)
			}
			if !equality.Semantic.DeepEqual(got, &want) {
				t.Errorf("deploymentAuthProxyPatch() = %v, want %v", got, want)
				fmt.Println(string(StructToYaml(got)))
			}
		})
	}
}
