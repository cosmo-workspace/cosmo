package kubeutil

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func TestPodStatusReason(t *testing.T) {
	type args struct {
		podYAML string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Running",
			want: "Running",
			args: args{
				podYAML: `
apiVersion: v1
kind: Pod
metadata:
  name: code-server
  namespace: default
spec:
  containers:
  - name: dind
  - name: code-server
  - name: istio-proxy
  initContainers:
  - name: init-chmod-data
  - name: istio-init
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T00:33:41Z"
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T00:33:54Z"
    status: "True"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T00:33:54Z"
    status: "True"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-10T05:34:00Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - name: code-server
    ready: true
    restartCount: 5
    started: true
    state:
      running:
        startedAt: "2022-04-16T00:33:43Z"
  - name: dind
    ready: true
    restartCount: 5
    started: true
    state:
      running:
        startedAt: "2022-04-16T00:33:42Z"
  - name: istio-proxy
    ready: true
    restartCount: 5
    started: true
    state:
      running:
        startedAt: "2022-04-16T00:33:43Z"
  initContainerStatuses:
  - name: init-chmod-data
    ready: true
    restartCount: 5
    state:
      terminated:
        exitCode: 0
        finishedAt: "2022-04-16T00:33:40Z"
        reason: Completed
        startedAt: "2022-04-16T00:33:12Z"
  - name: istio-init
    ready: true
    state:
      terminated:
        exitCode: 0
        finishedAt: "2022-04-16T00:33:41Z"
        reason: Completed
        startedAt: "2022-04-16T00:33:41Z"
  phase: Running
`,
			},
		},
		{
			name: "Running",
			want: "Running",
			args: args{
				podYAML: `
apiVersion: v1
kind: Pod
metadata:
  name: code-server
  namespace: default
spec:
  containers:
  - name: dind
  - name: code-server
  - name: istio-proxy
  initContainers:
  - name: init-chmod-data
  - name: istio-init
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T00:33:41Z"
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T00:33:54Z"
    status: "True"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T00:33:54Z"
    status: "True"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-10T05:34:00Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - name: code-server
    ready: true
    restartCount: 5
    started: true
    state:
      running:
        startedAt: "2022-04-16T00:33:43Z"
  - name: dind
    ready: true
    restartCount: 5
    started: true
    state:
      running:
        startedAt: "2022-04-16T00:33:42Z"
  - name: istio-proxy
    ready: true
    restartCount: 5
    started: true
    state:
      running:
        startedAt: "2022-04-16T00:33:43Z"
  initContainerStatuses:
  - name: init-chmod-data
    ready: true
    restartCount: 5
    state:
      terminated:
        exitCode: 0
        finishedAt: "2022-04-16T00:33:40Z"
        reason: Completed
        startedAt: "2022-04-16T00:33:12Z"
  - name: istio-init
    ready: true
    state:
      terminated:
        exitCode: 0
        finishedAt: "2022-04-16T00:33:41Z"
        reason: Completed
        startedAt: "2022-04-16T00:33:41Z"
  phase: Running
`,
			},
		},
		{
			name: "Init:ErrImageNeverPull",
			want: "Init:ErrImageNeverPull",
			args: args{
				podYAML: `
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: default
spec:
  containers:
  - name: nginx
  - name: istio-proxy
  initContainers:
  - name: wait-data
  - name: istio-init
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-03-05T14:10:50Z"
    message: 'containers with incomplete status: [wait-data]'
    reason: ContainersNotInitialized
    status: "False"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-03-05T14:10:38Z"
    message: 'containers with unready status: [nginx istio-proxy]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-03-05T14:10:38Z"
    message: 'containers with unready status: [nginx istio-proxy]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2021-09-06T03:18:27Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - name: istio-proxy
    ready: false
    restartCount: 58
    started: false
    state:
      terminated:
        exitCode: 0
        finishedAt: "2022-03-02T16:20:27Z"
        reason: Completed
        startedAt: "2022-03-02T04:57:25Z"
  - name: nginx
    ready: false
    started: false
    state:
      terminated:
        exitCode: 0
        finishedAt: "2022-03-02T16:20:22Z"
        reason: Completed
        startedAt: "2022-03-02T04:57:25Z"
  initContainerStatuses:
  - name: wait-data
    ready: false
    state:
      waiting:
        message: Container image "" is not present with pull policy
          of Never
        reason: ErrImageNeverPull
  - name: istio-init
    ready: true
    state:
      terminated:
        exitCode: 0
        finishedAt: "2022-03-02T04:57:24Z"
        reason: Completed
        startedAt: "2022-03-02T04:57:23Z"
  phase: Running
`,
			},
		},
		{
			name: "CrashLoopBackOff",
			want: "CrashLoopBackOff",
			args: args{
				podYAML: `
apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  containers:
  - name: busybox
  - name: istio-proxy
  initContainers:
  - name: istio-init
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T02:58:48Z"
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T02:58:46Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T02:58:46Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T02:58:46Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - name: busybox
    ready: false
    restartCount: 1
    started: false
    state:
      waiting:
        message: back-off 10s restarting failed container=busybox pod=busybox_default(c4e3ba1c-0e18-4a5c-a56d-014bb34425aa)
        reason: CrashLoopBackOff
  - name: istio-proxy
    ready: true
    restartCount: 0
    started: true
    state:
      running:
        startedAt: "2022-04-16T02:58:53Z"
  initContainerStatuses:
  - name: istio-init
    ready: true
    restartCount: 0
    state:
      terminated:
        exitCode: 0
        finishedAt: "2022-04-16T02:58:47Z"
        reason: Completed
        startedAt: "2022-04-16T02:58:47Z"
  phase: Running
`,
			},
		},
		{
			name: "PodInitializing",
			want: "PodInitializing",
			args: args{
				podYAML: `
apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  containers:
  - name: busybox
  - name: istio-proxy
  initContainers:
  - name: istio-init
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:19:43Z"
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:19:41Z"
    message: 'containers with unready status: [busybox istio-proxy]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:19:41Z"
    message: 'containers with unready status: [busybox istio-proxy]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T03:19:41Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - name: busybox
    ready: false
    restartCount: 0
    started: false
    state:
      waiting:
        reason: PodInitializing
  - name: istio-proxy
    ready: false
    restartCount: 0
    started: false
    state:
      waiting:
        reason: PodInitializing
  initContainerStatuses:
  - name: istio-init
    ready: true
    restartCount: 0
    state:
      terminated:
        exitCode: 0
        finishedAt: "2022-04-16T03:19:42Z"
        reason: Completed
        startedAt: "2022-04-16T03:19:42Z"
  phase: Pending
`,
			},
		},
		{
			name: "Terminating",
			want: "Terminating",
			args: args{
				podYAML: `
apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
  deletionGracePeriodSeconds: 30
  deletionTimestamp: "2022-04-16T03:29:42Z"
spec:
  containers:
  - name: busybox
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
  - lastState:
      terminated:
        exitCode: 0
        finishedAt: "2022-04-16T03:27:58Z"
        reason: Completed
        startedAt: "2022-04-16T03:27:58Z"
    name: busybox
    ready: false
    restartCount: 4
    started: false
    state:
      waiting:
        message: back-off 1m20s restarting failed container=busybox pod=busybox_test(33e56430-a2d7-4d67-b89b-ec05bfe10682)
        reason: CrashLoopBackOff
  phase: Running`,
			},
		},
		{
			name: "Status reason",
			want: "Hello",
			args: args{
				podYAML: `apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  containers:
  - name: busybox
status:
  reason: Hello`,
			},
		},
		{
			name: "Still running",
			want: "Running:1/2",
			args: args{
				podYAML: `
apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  containers:
  - command:
    - sleep
    - infinity
    image: busybox
    name: busybox
    readinessProbe:
      exec:
        command:
        - sh
      initialDelaySeconds: 30
  - command:
    - sleep
    - infinity
    image: busybox
    name: busybox2
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:35:47Z"
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:35:47Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:35:47Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:35:47Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - name: busybox
    ready: false
    restartCount: 0
    started: true
    state:
      running:
        startedAt: "2022-04-16T16:35:50Z"
  - name: busybox2
    ready: true
    restartCount: 0
    started: true
    state:
      running:
        startedAt: "2022-04-16T16:35:53Z"
  phase: Running
`,
			},
		},
		{
			name: "1 of 2 containers is down",
			want: "Running:1/2",
			args: args{
				podYAML: `
apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  containers:
  - image: busybox
    name: busybox
  - command:
    - sleep
    - infinity
    image: busybox
    name: busybox2
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:39:44Z"
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:39:44Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:39:44Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:39:44Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - name: busybox
    ready: false
    restartCount: 0
    started: false
    state:
      terminated:
        exitCode: 0
        finishedAt: "2022-04-16T16:39:47Z"
        reason: Completed
        startedAt: "2022-04-16T16:39:47Z"
  - name: busybox2
    ready: true
    restartCount: 0
    started: true
    state:
      running:
        startedAt: "2022-04-16T16:39:50Z"
  phase: Running`,
			},
		},
		{
			name: "Failed without reason",
			want: "ExitCode:1",
			args: args{
				podYAML: `apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  containers:
  - command:
    - sh
    - -c
    - exit 1
    image: busybox
    name: busybox
  restartPolicy: Never
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with incomplete status: [busybox-init]'
    reason: ContainersNotInitialized
    status: "False"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - name: busybox
    ready: false
    restartCount: 1
    state:
      terminated:
        exitCode: 1
        finishedAt: "2022-04-16T16:14:05Z"
        startedAt: "2022-04-16T16:14:05Z"
  phase: Pending`,
			},
		},

		{
			name: "Failed without reason but signal",
			want: "Signal:9",
			args: args{
				podYAML: `apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  containers:
  - command:
    - sh
    - -c
    - exit 1
    image: busybox
    name: busybox
  restartPolicy: Never
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with incomplete status: [busybox-init]'
    reason: ContainersNotInitialized
    status: "False"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - name: busybox
    ready: false
    restartCount: 1
    state:
      terminated:
        signal: 9
        exitCode: 9
        finishedAt: "2022-04-16T16:14:05Z"
        startedAt: "2022-04-16T16:14:05Z"
  phase: Pending`,
			},
		},
		{
			name: "Init",
			want: "Init:0/1",
			args: args{
				podYAML: `
apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  containers:
  - command:
    - sleep
    - infinity
    name: busybox
    image: busybox
  initContainers:
  - command:
    - sleep
    - 30
    name: busybox-init
    image: busybox
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:01:56Z"
    message: 'containers with incomplete status: [busybox-init]'
    reason: ContainersNotInitialized
    status: "False"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:01:56Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:01:56Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:01:56Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - image: busybox
    name: busybox
    ready: false
    restartCount: 0
    started: false
    state:
      waiting:
        reason: PodInitializing
  initContainerStatuses:
  - name: busybox-init
    ready: false
    restartCount: 0
    state:
      running:
        startedAt: "2022-04-16T16:01:59Z"
  phase: Pending`,
			},
		},
		{
			name: "Init CrashLoopBackOff",
			want: "Init:CrashLoopBackOff",
			args: args{
				podYAML: `apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  containers:
  - command:
    - sleep
    - infinity
    image: busybox
    name: busybox
  initContainers:
  - command:
    - sh
    - -c
    - exit 1
    image: busybox
    name: busybox-init
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with incomplete status: [busybox-init]'
    reason: ContainersNotInitialized
    status: "False"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - name: busybox
    ready: false
    restartCount: 0
    started: false
    state:
      waiting:
        reason: PodInitializing
  initContainerStatuses:
  - name: busybox-init
    ready: false
    restartCount: 1
    state:
      waiting:
        message: back-off 10s restarting failed container=busybox-init pod=busybox_test(f41bcd7f-3d63-44ec-b160-e539ebf5b8b9)
        reason: CrashLoopBackOff
  phase: Pending`,
			},
		},
		{
			name: "Init failed",
			want: "Init:Error",
			args: args{
				podYAML: `apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  containers:
  - command:
    - sleep
    - infinity
    image: busybox
    name: busybox
  initContainers:
  - command:
    - sh
    - -c
    - exit 1
    image: busybox
    name: busybox-init
  restartPolicy: Never
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with incomplete status: [busybox-init]'
    reason: ContainersNotInitialized
    status: "False"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - name: busybox
    ready: false
    restartCount: 0
    started: false
    state:
      waiting:
        reason: PodInitializing
  initContainerStatuses:
  - name: busybox-init
    ready: false
    restartCount: 1
    state:
      terminated:
        exitCode: 1
        finishedAt: "2022-04-16T16:14:05Z"
        reason: Error
        startedAt: "2022-04-16T16:14:05Z"
  phase: Pending`,
			},
		},
		{
			name: "Init failed empty reason",
			want: "Init:ExitCode:1",
			args: args{
				podYAML: `apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  containers:
  - command:
    - sleep
    - infinity
    image: busybox
    name: busybox
  initContainers:
  - command:
    - sh
    - -c
    - exit 1
    image: busybox
    name: busybox-init
  restartPolicy: Never
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with incomplete status: [busybox-init]'
    reason: ContainersNotInitialized
    status: "False"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - name: busybox
    ready: false
    restartCount: 0
    started: false
    state:
      waiting:
        reason: PodInitializing
  initContainerStatuses:
  - name: busybox-init
    ready: false
    restartCount: 1
    state:
      terminated:
        exitCode: 1
        finishedAt: "2022-04-16T16:14:05Z"
        startedAt: "2022-04-16T16:14:05Z"
  phase: Pending`,
			},
		},

		{
			name: "Init failed empty reason but signal",
			want: "Init:Signal:9",
			args: args{
				podYAML: `apiVersion: v1
kind: Pod
metadata:
  name: busybox
  namespace: default
spec:
  containers:
  - command:
    - sleep
    - infinity
    image: busybox
    name: busybox
  initContainers:
  - command:
    - sh
    - -c
    - exit 1
    image: busybox
    name: busybox-init
  restartPolicy: Never
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with incomplete status: [busybox-init]'
    reason: ContainersNotInitialized
    status: "False"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    message: 'containers with unready status: [busybox]'
    reason: ContainersNotReady
    status: "False"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2022-04-16T16:06:46Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - name: busybox
    ready: false
    restartCount: 0
    started: false
    state:
      waiting:
        reason: PodInitializing
  initContainerStatuses:
  - name: busybox-init
    ready: false
    restartCount: 1
    state:
      terminated:
        signal: 9
        exitCode: 9
        finishedAt: "2022-04-16T16:14:05Z"
        startedAt: "2022-04-16T16:14:05Z"
  phase: Pending`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pod corev1.Pod
			err := yaml.Unmarshal([]byte(tt.args.podYAML), &pod)
			if err != nil {
				t.Error(err)
			}

			if got := PodStatusReason(pod); got != tt.want {
				t.Errorf("PodStatusReason() = %v, want %v", got, tt.want)
			}
		})
	}
}
