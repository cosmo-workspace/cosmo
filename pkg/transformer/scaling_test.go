package transformer

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/template"
)

func TestScalingTransformer_Transform(t *testing.T) {
	type fields struct {
		ScalingOverrideSpec []cosmov1alpha1.ScalingOverrideSpec
		instName            string
	}
	type args struct {
		obj string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Modify",
			fields: fields{
				ScalingOverrideSpec: []cosmov1alpha1.ScalingOverrideSpec{
					{
						Target: cosmov1alpha1.ObjectRef{
							APIVersion: metav1.GroupVersion{
								Group:   "apps",
								Version: "v1",
							}.String(),
							Kind: "Deployment",
							Name: "test-deployment",
						},
						Replicas: 0,
					},
				},
				instName: "instance",
			},

			args: args{
				obj: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: instance-test-deployment
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: theiaide/theia
        imagePullPolicy: IfNotPresent
        name: theia
        ports:
        - name: http
          containerPort: 3000
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
			want: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: instance-test-deployment
spec:
  replicas: 0
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
			wantErr: false,
		},
		{
			name: "Not modify gvk not match",
			fields: fields{
				ScalingOverrideSpec: []cosmov1alpha1.ScalingOverrideSpec{
					{
						Target: cosmov1alpha1.ObjectRef{
							APIVersion: metav1.GroupVersion{
								Group:   "apps",
								Version: "v1",
							}.String(),
							Kind: "StatefulSet",
							Name: "test-deployment",
						},
						Replicas: 0,
					},
				},
				instName: "instance",
			},
			args: args{
				obj: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: instance-test-deployment
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: theiaide/theia
        imagePullPolicy: IfNotPresent
        name: theia
        ports:
        - name: http
          containerPort: 3000
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
			want: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: instance-test-deployment
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
			wantErr: false,
		},
		{
			name: "Not modify name not match",
			fields: fields{
				ScalingOverrideSpec: []cosmov1alpha1.ScalingOverrideSpec{
					{
						Target: cosmov1alpha1.ObjectRef{
							APIVersion: metav1.GroupVersion{
								Group:   "apps",
								Version: "v1",
							}.String(),
							Kind: "Deployment",
							Name: "notfound",
						},
						Replicas: 0,
					},
				},
				instName: "instance",
			},
			args: args{
				obj: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: instance-test-deployment
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: theiaide/theia
        imagePullPolicy: IfNotPresent
        name: theia
        ports:
        - name: http
          containerPort: 3000
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
			want: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: instance-test-deployment
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
			wantErr: false,
		},
		{
			name: "Scaling spec nil",
			fields: fields{
				ScalingOverrideSpec: nil,
			},
			args: args{
				obj: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: instance-test-deployment
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: theiaide/theia
        imagePullPolicy: IfNotPresent
        name: theia
        ports:
        - name: http
          containerPort: 3000
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
			want: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: instance-test-deployment
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewScalingTransformer(tt.fields.ScalingOverrideSpec, tt.fields.instName)

			_, obj, err := template.StringToUnstructured(tt.args.obj)
			if err != nil {
				t.Errorf("appendServicePort() template.StringToUnstructured error = %v", err)
				return
			}

			obj, err = tr.Transform(obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScalingTransformer.Transform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := yaml.Marshal(obj)
			if err != nil {
				t.Errorf("ScalingTransformer.Transform() yaml.Marshal error = %v", err)
				return
			}
			if string(got) != tt.want {
				t.Errorf("ScalingTransformer.Transform() got = %v, want %v", string(got), tt.want)
				t.Errorf("ScalingTransformer.Transform() got = %v, want %v", got, []byte(tt.want))
			}
		})
	}
}
