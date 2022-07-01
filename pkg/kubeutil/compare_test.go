package kubeutil

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/cosmo-workspace/cosmo/pkg/template"
)

func TestLooseDeepEqual(t *testing.T) {
	var nilUnst *unstructured.Unstructured
	RegisterTestingT(t)
	type args struct {
		x    string
		y    string
		xObj Comparable
		yObj Comparable
		opts []DeepEqualOption
	}
	diffBuf := &bytes.Buffer{}
	tests := []struct {
		name     string
		args     args
		want     bool
		wantDiff string
	}{
		{
			name: "Equal",
			want: true,
			args: args{
				x: `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2022-04-16T03:26:11Z"
  labels:
    run: busybox
  managedFields:
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:labels:
          .: {}
          f:run: {}
      f:spec:
        f:containers:
          k:{"name":"busybox"}:
            .: {}
            f:image: {}
            f:imagePullPolicy: {}
            f:name: {}
            f:resources: {}
            f:terminationMessagePath: {}
            f:terminationMessagePolicy: {}
        f:dnsPolicy: {}
        f:enableServiceLinks: {}
        f:restartPolicy: {}
        f:schedulerName: {}
        f:securityContext: {}
        f:terminationGracePeriodSeconds: {}
    manager: kubectl-run
    operation: Update
    time: "2022-04-16T03:26:11Z"
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:status:
        f:conditions:
          k:{"type":"ContainersReady"}:
            .: {}
            f:lastProbeTime: {}
            f:lastTransitionTime: {}
            f:message: {}
            f:reason: {}
            f:status: {}
            f:type: {}
          k:{"type":"Initialized"}:
            .: {}
            f:lastProbeTime: {}
            f:lastTransitionTime: {}
            f:status: {}
            f:type: {}
          k:{"type":"Ready"}:
            .: {}
            f:lastProbeTime: {}
            f:lastTransitionTime: {}
            f:message: {}
            f:reason: {}
            f:status: {}
            f:type: {}
        f:containerStatuses: {}
        f:hostIP: {}
        f:phase: {}
        f:podIP: {}
        f:podIPs:
          .: {}
          k:{"ip":"1.1.1.1"}:
            .: {}
            f:ip: {}
        f:startTime: {}
    manager: k3s
    operation: Update
    time: "2022-04-16T03:26:15Z"
  name: busybox
  namespace: test
  resourceVersion: "19378787"
  selfLink: /api/v1/namespaces/test/pods/busybox
  uid: 33e56430-a2d7-4d67-b89b-ec05bfe10682
spec:
  containers:
  - image: busybox
    imagePullPolicy: Always
    name: busybox
    resources: {}
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-nsv49
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  nodeName: ursula-t3600
  preemptionPolicy: PreemptLowerPriority
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  tolerations:
  - effect: NoExecute
    key: node.kubernetes.io/not-ready
    operator: Exists
    tolerationSeconds: 300
  - effect: NoExecute
    key: node.kubernetes.io/unreachable
    operator: Exists
    tolerationSeconds: 300
  volumes:
  - name: default-token-nsv49
    secret:
      defaultMode: 420
      secretName: default-token-nsv49
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - containerID: docker://XXXXXXXX
    image: busybox:latest
    imageID: docker-pullable://busybox@sha256:d2b53584f580310186df7a2055ce3ff83cc0df6caacf1e3489bff8cf5d0af5d8
    lastState: {}
    name: busybox
    ready: false
    restartCount: 0
    started: false
    state:
      terminated:
        containerID: docker://XXXXXXXX
        exitCode: 0
        finishedAt: "2022-04-16T03:26:15Z"
        reason: Completed
        startedAt: "2022-04-16T03:26:15Z"
  hostIP: 127.0.0.1
  phase: Running
  podIP: 1.1.1.1
  podIPs:
  - ip: 1.1.1.1
  qosClass: BestEffort
  startTime: "2022-04-16T03:26:11Z"`,
				y: `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2022-04-16T03:26:11Z"
  labels:
    run: busybox
  managedFields:
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:labels:
          .: {}
          f:run: {}
      f:spec:
        f:containers:
          k:{"name":"busybox"}:
            .: {}
            f:image: {}
            f:imagePullPolicy: {}
            f:name: {}
            f:resources: {}
            f:terminationMessagePath: {}
            f:terminationMessagePolicy: {}
        f:dnsPolicy: {}
        f:enableServiceLinks: {}
        f:restartPolicy: {}
        f:schedulerName: {}
        f:securityContext: {}
        f:terminationGracePeriodSeconds: {}
    manager: kubectl-run
    operation: Update
    time: "2022-04-16T03:26:11Z"
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:status:
        f:conditions:
          k:{"type":"ContainersReady"}:
            .: {}
            f:lastProbeTime: {}
            f:lastTransitionTime: {}
            f:message: {}
            f:reason: {}
            f:status: {}
            f:type: {}
          k:{"type":"Initialized"}:
            .: {}
            f:lastProbeTime: {}
            f:lastTransitionTime: {}
            f:status: {}
            f:type: {}
          k:{"type":"Ready"}:
            .: {}
            f:lastProbeTime: {}
            f:lastTransitionTime: {}
            f:message: {}
            f:reason: {}
            f:status: {}
            f:type: {}
        f:containerStatuses: {}
        f:hostIP: {}
        f:phase: {}
        f:podIP: {}
        f:podIPs:
          .: {}
          k:{"ip":"1.1.1.1"}:
            .: {}
            f:ip: {}
        f:startTime: {}
    manager: k3s
    operation: Update
    time: "2022-04-16T03:26:15Z"
  name: busybox
  namespace: test
  resourceVersion: "19378787"
  selfLink: /api/v1/namespaces/test/pods/busybox
  uid: 33e56430-a2d7-4d67-b89b-ec05bfe10682
spec:
  containers:
  - image: busybox
    imagePullPolicy: Always
    name: busybox
    resources: {}
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-nsv49
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  nodeName: ursula-t3600
  preemptionPolicy: PreemptLowerPriority
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  tolerations:
  - effect: NoExecute
    key: node.kubernetes.io/not-ready
    operator: Exists
    tolerationSeconds: 300
  - effect: NoExecute
    key: node.kubernetes.io/unreachable
    operator: Exists
    tolerationSeconds: 300
  volumes:
  - name: default-token-nsv49
    secret:
      defaultMode: 420
      secretName: default-token-nsv49
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - containerID: docker://XXXXXXXX
    image: busybox:latest
    imageID: docker-pullable://busybox@sha256:d2b53584f580310186df7a2055ce3ff83cc0df6caacf1e3489bff8cf5d0af5d8
    lastState: {}
    name: busybox
    ready: false
    restartCount: 0
    started: false
    state:
      terminated:
        containerID: docker://XXXXXXXX
        exitCode: 0
        finishedAt: "2022-04-16T03:26:15Z"
        reason: Completed
        startedAt: "2022-04-16T03:26:15Z"
  hostIP: 127.0.0.1
  phase: Running
  podIP: 1.1.1.1
  podIPs:
  - ip: 1.1.1.1
  qosClass: BestEffort
  startTime: "2022-04-16T03:26:11Z"`,
			},
		},
		{
			name: "Not equal with print diff",
			want: false,
			wantDiff: `
- 					"imagePullPolicy": string("Always"),
+ 					"imagePullPolicy": string("Never"),`,
			args: args{
				opts: []DeepEqualOption{
					WithPrintDiff(diffBuf),
				},
				x: `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2022-04-16T03:26:11Z"
  labels:
    run: busybox
  managedFields:
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:labels:
          .: {}
          f:run: {}
      f:spec:
        f:containers:
          k:{"name":"busybox"}:
            .: {}
            f:image: {}
            f:imagePullPolicy: {}
            f:name: {}
            f:resources: {}
            f:terminationMessagePath: {}
            f:terminationMessagePolicy: {}
        f:dnsPolicy: {}
        f:enableServiceLinks: {}
        f:restartPolicy: {}
        f:schedulerName: {}
        f:securityContext: {}
        f:terminationGracePeriodSeconds: {}
    manager: kubectl-run
    operation: Update
    time: "2022-04-16T03:26:11Z"
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:status:
        f:conditions:
          k:{"type":"ContainersReady"}:
            .: {}
            f:lastProbeTime: {}
            f:lastTransitionTime: {}
            f:message: {}
            f:reason: {}
            f:status: {}
            f:type: {}
          k:{"type":"Initialized"}:
            .: {}
            f:lastProbeTime: {}
            f:lastTransitionTime: {}
            f:status: {}
            f:type: {}
          k:{"type":"Ready"}:
            .: {}
            f:lastProbeTime: {}
            f:lastTransitionTime: {}
            f:message: {}
            f:reason: {}
            f:status: {}
            f:type: {}
        f:containerStatuses: {}
        f:hostIP: {}
        f:phase: {}
        f:podIP: {}
        f:podIPs:
          .: {}
          k:{"ip":"1.1.1.1"}:
            .: {}
            f:ip: {}
        f:startTime: {}
    manager: k3s
    operation: Update
    time: "2022-04-16T03:26:15Z"
  name: busybox
  namespace: test
  resourceVersion: "19378787"
  selfLink: /api/v1/namespaces/test/pods/busybox
  uid: 33e56430-a2d7-4d67-b89b-ec05bfe10682
spec:
  containers:
  - image: busybox
    imagePullPolicy: Always
    name: busybox
    resources: {}
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-nsv49
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  nodeName: ursula-t3600
  preemptionPolicy: PreemptLowerPriority
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  tolerations:
  - effect: NoExecute
    key: node.kubernetes.io/not-ready
    operator: Exists
    tolerationSeconds: 300
  - effect: NoExecute
    key: node.kubernetes.io/unreachable
    operator: Exists
    tolerationSeconds: 300
  volumes:
  - name: default-token-nsv49
    secret:
      defaultMode: 420
      secretName: default-token-nsv49
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - containerID: docker://XXXXXXXX
    image: busybox:latest
    imageID: docker-pullable://busybox@sha256:d2b53584f580310186df7a2055ce3ff83cc0df6caacf1e3489bff8cf5d0af5d8
    lastState: {}
    name: busybox
    ready: false
    restartCount: 0
    started: false
    state:
      terminated:
        containerID: docker://XXXXXXXX
        exitCode: 0
        finishedAt: "2022-04-16T03:26:15Z"
        reason: Completed
        startedAt: "2022-04-16T03:26:15Z"
  hostIP: 127.0.0.1
  phase: Running
  podIP: 1.1.1.1
  podIPs:
  - ip: 1.1.1.1
  qosClass: BestEffort
  startTime: "2022-04-16T03:26:11Z"`,
				y: `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2022-04-16T03:26:11Z"
  labels:
    run: busybox
  managedFields:
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:labels:
          .: {}
          f:run: {}
      f:spec:
        f:containers:
          k:{"name":"busybox"}:
            .: {}
            f:image: {}
            f:imagePullPolicy: {}
            f:name: {}
            f:resources: {}
            f:terminationMessagePath: {}
            f:terminationMessagePolicy: {}
        f:dnsPolicy: {}
        f:enableServiceLinks: {}
        f:restartPolicy: {}
        f:schedulerName: {}
        f:securityContext: {}
        f:terminationGracePeriodSeconds: {}
    manager: kubectl-run
    operation: Update
    time: "2022-04-16T03:26:11Z"
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:status:
        f:conditions:
          k:{"type":"ContainersReady"}:
            .: {}
            f:lastProbeTime: {}
            f:lastTransitionTime: {}
            f:message: {}
            f:reason: {}
            f:status: {}
            f:type: {}
          k:{"type":"Initialized"}:
            .: {}
            f:lastProbeTime: {}
            f:lastTransitionTime: {}
            f:status: {}
            f:type: {}
          k:{"type":"Ready"}:
            .: {}
            f:lastProbeTime: {}
            f:lastTransitionTime: {}
            f:message: {}
            f:reason: {}
            f:status: {}
            f:type: {}
        f:containerStatuses: {}
        f:hostIP: {}
        f:phase: {}
        f:podIP: {}
        f:podIPs:
          .: {}
          k:{"ip":"1.1.1.1"}:
            .: {}
            f:ip: {}
        f:startTime: {}
    manager: k3s
    operation: Update
    time: "2022-04-16T03:26:15Z"
  name: busybox
  namespace: test
  resourceVersion: "19378787"
  selfLink: /api/v1/namespaces/test/pods/busybox
  uid: 33e56430-a2d7-4d67-b89b-ec05bfe10682
spec:
  containers:
  - image: busybox
    imagePullPolicy: Never
    name: busybox
    resources: {}
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-nsv49
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  nodeName: ursula-t3600
  preemptionPolicy: PreemptLowerPriority
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  tolerations:
  - effect: NoExecute
    key: node.kubernetes.io/not-ready
    operator: Exists
    tolerationSeconds: 300
  - effect: NoExecute
    key: node.kubernetes.io/unreachable
    operator: Exists
    tolerationSeconds: 300
  volumes:
  - name: default-token-nsv49
    secret:
      defaultMode: 420
      secretName: default-token-nsv49
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:26:11Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - containerID: docker://XXXXXXXX
    image: busybox:latest
    imageID: docker-pullable://busybox@sha256:d2b53584f580310186df7a2055ce3ff83cc0df6caacf1e3489bff8cf5d0af5d8
    lastState: {}
    name: busybox
    ready: false
    restartCount: 0
    started: false
    state:
      terminated:
        containerID: docker://XXXXXXXX
        exitCode: 0
        finishedAt: "2022-04-16T03:26:15Z"
        reason: Completed
        startedAt: "2022-04-16T03:26:15Z"
  hostIP: 127.0.0.1
  phase: Running
  podIP: 1.1.1.1
  podIPs:
  - ip: 1.1.1.1
  qosClass: BestEffort
  startTime: "2022-04-16T03:26:11Z"`,
			},
		},
		{
			name: "fix gvk",
			args: args{
				xObj: &corev1.Pod{
					TypeMeta: v1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Pod",
					},
					ObjectMeta: v1.ObjectMeta{
						Name: "app",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "app",
							},
						},
					},
				},
				yObj: &corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name: "app",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "app",
							},
						},
					},
				},
				opts: []DeepEqualOption{
					WithPrintDiff(diffBuf),
					WithFixGVK(scheme.Scheme),
				},
			},
			want: true,
			wantDiff: `
- 		Kind:       "Pod",
+ 		Kind:       "",
- 		APIVersion: "v1",
+ 		APIVersion: "",`,
		},
		{
			name: "one is nil",
			args: args{
				yObj: &corev1.Pod{},
			},
			want: false,
		},
		{
			name: "both nil",
			want: true,
		},
		{
			name: "one empty unstructured",
			args: args{
				xObj: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Pod",
					},
				},
				yObj: nilUnst,
			},
			want: false,
		},
		{
			name: "both empty unstructured",
			args: args{
				xObj: nilUnst,
				yObj: nilUnst,
			},
			want: true,
		},
	}
	diffLineReg := regexp.MustCompile("^[-|+].*$")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var x, y Comparable
			var err error
			if tt.args.x != "" {
				_, x, err = template.StringToUnstructured(tt.args.x)
				Expect(err).ShouldNot(HaveOccurred())
			} else {
				x = tt.args.xObj
			}
			if tt.args.y != "" {
				_, y, err = template.StringToUnstructured(tt.args.y)
				Expect(err).ShouldNot(HaveOccurred())
			} else {
				y = tt.args.yObj
			}

			equal := LooseDeepEqual(x, y, tt.args.opts...)
			Expect(equal).Should(BeEquivalentTo(tt.want))

			if tt.args.opts != nil {
				for _, v := range tt.args.opts {
					if _, ok := v.(printDiff); ok {
						d, err := io.ReadAll(diffBuf)
						Expect(err).ShouldNot(HaveOccurred())

						diff := []string{""}
						for _, line := range strings.Split(string(d), "\n") {
							if diffLineReg.MatchString(line) {
								// replace U+00a0 to space because cmp.Diff sometimes contains U+00a0
								diff = append(diff, strings.ReplaceAll(line, " ", " "))
							}
						}
						Expect(strings.Join(diff, "\n")).Should(BeEquivalentTo(tt.wantDiff))
					}
				}
			}
		})
	}
}

func TestIsGVKEqual(t *testing.T) {
	type args struct {
		a schema.GroupVersionKind
		b schema.GroupVersionKind
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Equal",
			args: args{
				a: schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Ingress",
				},
				b: schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Ingress",
				},
			},
			want: true,
		},
		{
			name: "Not Equal",
			args: args{
				a: schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Ingress",
				},
				b: schema.GroupVersionKind{
					Group:   "extentions",
					Version: "v1",
					Kind:    "Ingress",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGVKEqual(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("IsGVKEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
