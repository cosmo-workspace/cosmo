package template

import (
	"reflect"
	"testing"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
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
			name: "no service",
			args: args{
				wsConfig: &cosmov1alpha1.Config{
					DeploymentName:      "workspace",
					ServiceName:         "workspace",
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
				ServiceMainPortName: "main",
			},
		},
		{
			name: "no deployment",
			args: args{
				wsConfig: &cosmov1alpha1.Config{
					DeploymentName:      "workspace",
					ServiceName:         "workspace",
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
