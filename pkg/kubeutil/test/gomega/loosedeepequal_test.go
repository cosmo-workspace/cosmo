package gomega

import (
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("TestBeLooseDeepEqual", func() {
	When("x == y", func() {
		It("should matched", func() {
			x := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "nginx",
					Namespace:       "default",
					ResourceVersion: "1",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:alpine",
						},
						{
							Name:  "nginx2",
							Image: "nginx:alpine",
						},
					},
				},
			}
			y := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "nginx",
					Namespace:       "default",
					ResourceVersion: "2",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:alpine",
						},
						{
							Name:  "nginx2",
							Image: "nginx:alpine",
						},
					},
				},
			}
			Expect(x).Should(BeLooseDeepEqual(y))
		})
	})

	When("x != y", func() {
		It("should not matched", func() {
			x := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "nginx",
					Namespace:       "default",
					ResourceVersion: "1",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:alpine",
						},
						{
							Name:  "nginx2",
							Image: "nginx:alpine",
						},
					},
				},
			}
			y := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "nginx",
					Namespace:       "default",
					ResourceVersion: "2",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:alpine",
						},
						{
							Name:  "nginx3",
							Image: "nginx:alpine",
						},
					},
				},
			}
			Expect(x).ShouldNot(BeLooseDeepEqual(y))
		})
	})
})

func TestLooseDeepEqualMatcher_Match(t *testing.T) {
	type fields struct {
		Expected interface{}
	}
	type args struct {
		actual interface{}
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantSuccess bool
		wantErr     bool
	}{
		{
			name: "Match",
			fields: fields{
				Expected: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "nginx",
						Namespace:       "default",
						ResourceVersion: "1",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "nginx",
								Image: "nginx:alpine",
							},
							{
								Name:  "nginx2",
								Image: "nginx:alpine",
							},
						},
					},
				},
			},
			args: args{
				actual: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "nginx",
						Namespace:       "default",
						ResourceVersion: "2",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "nginx",
								Image: "nginx:alpine",
							},
							{
								Name:  "nginx2",
								Image: "nginx:alpine",
							},
						},
					},
				},
			},
			wantSuccess: true,
			wantErr:     false,
		},
		{
			name: "Not Match",
			fields: fields{
				Expected: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "nginx",
						Namespace:       "default",
						ResourceVersion: "1",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "nginx",
								Image: "nginx:alpine",
							},
							{
								Name:  "nginx2",
								Image: "nginx:alpine",
							},
						},
					},
				},
			},
			args: args{
				actual: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "nginx",
						Namespace:       "default",
						ResourceVersion: "2",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "nginx",
								Image: "nginx:alpine",
							},
							{
								Name:  "nginx3",
								Image: "nginx:alpine",
							},
						},
					},
				},
			},
			wantSuccess: false,
			wantErr:     false,
		},
		{
			name: "Actual nil",
			fields: fields{
				Expected: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "nginx",
						Namespace:       "default",
						ResourceVersion: "1",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "nginx",
								Image: "nginx:alpine",
							},
							{
								Name:  "nginx2",
								Image: "nginx:alpine",
							},
						},
					},
				},
			},
			args: args{
				actual: nil,
			},
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name: "Expect nil",
			fields: fields{
				Expected: nil,
			},
			args: args{
				actual: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "nginx",
						Namespace:       "default",
						ResourceVersion: "1",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "nginx",
								Image: "nginx:alpine",
							},
							{
								Name:  "nginx2",
								Image: "nginx:alpine",
							},
						},
					},
				},
			},
			wantSuccess: false,
			wantErr:     true,
		},
		{
			name: "Both nil",
			fields: fields{
				Expected: nil,
			},
			args: args{
				actual: nil,
			},
			wantSuccess: false,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := &LooseDeepEqualMatcher{
				Expected: tt.fields.Expected,
			}
			gotSuccess, err := matcher.Match(tt.args.actual)
			if (err != nil) != tt.wantErr {
				t.Errorf("LooseDeepEqualMatcher.Match() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotSuccess != tt.wantSuccess {
				t.Errorf("LooseDeepEqualMatcher.Match() = %v, want %v", gotSuccess, tt.wantSuccess)
			}
		})
	}
}

func TestLooseDeepEqualMatcher_FailureMessage(t *testing.T) {
	type fields struct {
		Expected interface{}
	}
	type args struct {
		actual interface{}
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantMessage string
	}{
		{
			name: "Not Match",
			fields: fields{
				Expected: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "nginx",
						Namespace:       "default",
						ResourceVersion: "1",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "nginx",
								Image: "nginx:alpine",
							},
							{
								Name:  "nginx2",
								Image: "nginx:alpine",
							},
						},
					},
				},
			},
			args: args{
				actual: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "nginx",
						Namespace:       "default",
						ResourceVersion: "2",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "nginx",
								Image: "nginx:alpine",
							},
							{
								Name:  "nginx3",
								Image: "nginx:alpine",
							},
						},
					},
				},
			},
			wantMessage: `Actual
    <*v1.Pod>: {
        TypeMeta: {Kind: "", APIVersion: ""},
        ObjectMeta: {
            Name: "nginx",
            GenerateName: "",
            Namespace: "default",
            SelfLink: "",
            UID: "",
            ResourceVersion: "",
            Generation: 0,
            CreationTimestamp: {
                Time: 0001-01-01T00:00:00Z,
            },
            DeletionTimestamp: nil,
            DeletionGracePeriodSeconds: nil,
            Labels: nil,
            Annotations: nil,
            OwnerReferences: nil,
            Finalizers: nil,
            ManagedFields: nil,
        },
        Spec: {
            Volumes: nil,
            InitContainers: nil,
            Containers: [
                {
                    Name: "nginx",
                    Image: "nginx:alpine",
                    Command: nil,
                    Args: nil,
                    WorkingDir: "",
                    Ports: nil,
                    EnvFrom: nil,
                    Env: nil,
                    Resources: {Limits: nil, Requests: nil, Claims: nil},
                    VolumeMounts: nil,
                    VolumeDevices: nil,
                    LivenessProbe: nil,
                    ReadinessProbe: nil,
                    StartupProbe: nil,
                    Lifecycle: nil,
                    TerminationMessagePath: "",
                    TerminationMessagePolicy: "",
                    ImagePullPolicy: "",
                    SecurityContext: nil,
                    Stdin: false,
                    StdinOnce: false,
                    TTY: false,
                },
                {
                    Name: "nginx3",
                    Image: "nginx:alpine",
                    Command: nil,
                    Args: nil,
                    WorkingDir: "",
                    Ports: nil,
                    EnvFrom: nil,
                    Env: nil,
                    Resources: {Limits: nil, Requests: nil, Claims: nil},
                    VolumeMounts: nil,
                    VolumeDevices: nil,
                    LivenessProbe: nil,
                    ReadinessProbe: nil,
                    StartupProbe: nil,
                    Lifecycle: nil,
                    TerminationMessagePath: "",
                    TerminationMessagePolicy: "",
                    ImagePullPolicy: "",
                    SecurityContext: nil,
                    Stdin: false,
                    StdinOnce: false,
                    TTY: false,
                },
            ],
            EphemeralContainers: nil,
            RestartPolicy: "",
            TerminationGracePeriodSeconds: nil,
            ActiveDeadlineSeconds: nil,
            DNSPolicy: "",
            NodeSelector: nil,
            ServiceAccountName: "",
            DeprecatedServiceAccount: "",
            AutomountServiceAccountToken: nil,
            NodeName: "",
            HostNetwork: false,
            HostPID: false,
            HostIPC: false,
            ShareProcessNamespace: nil,
            SecurityContext: nil,
            ImagePullSecrets: nil,
            Hostname: "",
            Subdomain: "",
            Affinity: nil,
            SchedulerName: "",
            Tolerations: nil,
            HostAliases: nil,
            PriorityClassName: "",
            Priority: nil,
            DNSConfig: nil,
            ReadinessGates: nil,
            RuntimeClassName: nil,
            EnableServiceLinks: nil,
            PreemptionPolicy: nil,
            Overhead: nil,
            TopologySpreadConstraints: nil,
            SetHostnameAsFQDN: nil,
            OS: nil,
            HostUsers: nil,
            SchedulingGates: nil,
            ResourceClaims: nil,
        },
        Status: {
            Phase: "",
            Conditions: nil,
            Message: "",
            Reason: "",
            NominatedNodeName: "",
            HostIP: "",
            PodIP: "",
            PodIPs: nil,
            StartTime: nil,
            InitContainerStatuses: nil,
            ContainerStatuses: nil,
            QOSClass: "",
            EphemeralContainerStatuses: nil,
        },
    }
shouled be equal to
    <*v1.Pod>: {
        TypeMeta: {Kind: "", APIVersion: ""},
        ObjectMeta: {
            Name: "nginx",
            GenerateName: "",
            Namespace: "default",
            SelfLink: "",
            UID: "",
            ResourceVersion: "",
            Generation: 0,
            CreationTimestamp: {
                Time: 0001-01-01T00:00:00Z,
            },
            DeletionTimestamp: nil,
            DeletionGracePeriodSeconds: nil,
            Labels: nil,
            Annotations: nil,
            OwnerReferences: nil,
            Finalizers: nil,
            ManagedFields: nil,
        },
        Spec: {
            Volumes: nil,
            InitContainers: nil,
            Containers: [
                {
                    Name: "nginx",
                    Image: "nginx:alpine",
                    Command: nil,
                    Args: nil,
                    WorkingDir: "",
                    Ports: nil,
                    EnvFrom: nil,
                    Env: nil,
                    Resources: {Limits: nil, Requests: nil, Claims: nil},
                    VolumeMounts: nil,
                    VolumeDevices: nil,
                    LivenessProbe: nil,
                    ReadinessProbe: nil,
                    StartupProbe: nil,
                    Lifecycle: nil,
                    TerminationMessagePath: "",
                    TerminationMessagePolicy: "",
                    ImagePullPolicy: "",
                    SecurityContext: nil,
                    Stdin: false,
                    StdinOnce: false,
                    TTY: false,
                },
                {
                    Name: "nginx2",
                    Image: "nginx:alpine",
                    Command: nil,
                    Args: nil,
                    WorkingDir: "",
                    Ports: nil,
                    EnvFrom: nil,
                    Env: nil,
                    Resources: {Limits: nil, Requests: nil, Claims: nil},
                    VolumeMounts: nil,
                    VolumeDevices: nil,
                    LivenessProbe: nil,
                    ReadinessProbe: nil,
                    StartupProbe: nil,
                    Lifecycle: nil,
                    TerminationMessagePath: "",
                    TerminationMessagePolicy: "",
                    ImagePullPolicy: "",
                    SecurityContext: nil,
                    Stdin: false,
                    StdinOnce: false,
                    TTY: false,
                },
            ],
            EphemeralContainers: nil,
            RestartPolicy: "",
            TerminationGracePeriodSeconds: nil,
            ActiveDeadlineSeconds: nil,
            DNSPolicy: "",
            NodeSelector: nil,
            ServiceAccountName: "",
            DeprecatedServiceAccount: "",
            AutomountServiceAccountToken: nil,
            NodeName: "",
            HostNetwork: false,
            HostPID: false,
            HostIPC: false,
            ShareProcessNamespace: nil,
            SecurityContext: nil,
            ImagePullSecrets: nil,
            Hostname: "",
            Subdomain: "",
            Affinity: nil,
            SchedulerName: "",
            Tolerations: nil,
            HostAliases: nil,
            PriorityClassName: "",
            Priority: nil,
            DNSConfig: nil,
            ReadinessGates: nil,
            RuntimeClassName: nil,
            EnableServiceLinks: nil,
            PreemptionPolicy: nil,
            Overhead: nil,
            TopologySpreadConstraints: nil,
            SetHostnameAsFQDN: nil,
            OS: nil,
            HostUsers: nil,
            SchedulingGates: nil,
            ResourceClaims: nil,
        },
        Status: {
            Phase: "",
            Conditions: nil,
            Message: "",
            Reason: "",
            NominatedNodeName: "",
            HostIP: "",
            PodIP: "",
            PodIPs: nil,
            StartTime: nil,
            InitContainerStatuses: nil,
            ContainerStatuses: nil,
            QOSClass: "",
            EphemeralContainerStatuses: nil,
        },
    }
diff:     <string>:   &v1.Pod{
      	TypeMeta:   {},
      	ObjectMeta: {Name: "nginx", Namespace: "default"},
      	Spec: v1.PodSpec{
      		Volumes:        nil,
      		InitContainers: nil,
      		Containers: []v1.Container{
      			{Name: "nginx", Image: "nginx:alpine"},
      			{
    - 				Name:    "nginx3",
    + 				Name:    "nginx2",
      				Image:   "nginx:alpine",
      				Command: nil,
      				... // 19 identical fields
      			},
      		},
      		EphemeralContainers: nil,
      		RestartPolicy:       "",
      		... // 34 identical fields
      	},
      	Status: {},
      }
    `,
		},
	}
	memRe := regexp.MustCompile(`\<\*v1\.Pod \| .*\>`)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := &LooseDeepEqualMatcher{
				Expected: tt.fields.Expected,
			}
			gotMessage := matcher.FailureMessage(tt.args.actual)

			// replace space
			gotMessage = strings.ReplaceAll(gotMessage, " ", " ")

			// replace memory expression
			gotMessage = string(memRe.ReplaceAll([]byte(gotMessage), []byte(`<*v1.Pod>`)))

			if gotMessage != tt.wantMessage {
				t.Errorf("LooseDeepEqualMatcher.FailureMessage() diff = %s", cmp.Diff(gotMessage, tt.wantMessage))
			}
		})
	}
}

func TestLooseDeepEqualMatcher_NegatedFailureMessage(t *testing.T) {
	type fields struct {
		Expected interface{}
	}
	type args struct {
		actual interface{}
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantMessage string
	}{
		{
			name: "Not Match",
			fields: fields{
				Expected: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "nginx",
						Namespace:       "default",
						ResourceVersion: "1",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "nginx",
								Image: "nginx:alpine",
							},
							{
								Name:  "nginx2",
								Image: "nginx:alpine",
							},
						},
					},
				},
			},
			args: args{
				actual: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "nginx",
						Namespace:       "default",
						ResourceVersion: "2",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "nginx",
								Image: "nginx:alpine",
							},
							{
								Name:  "nginx3",
								Image: "nginx:alpine",
							},
						},
					},
				},
			},
			wantMessage: `Expected
    <*v1.Pod>: {
        TypeMeta: {Kind: "", APIVersion: ""},
        ObjectMeta: {
            Name: "nginx",
            GenerateName: "",
            Namespace: "default",
            SelfLink: "",
            UID: "",
            ResourceVersion: "",
            Generation: 0,
            CreationTimestamp: {
                Time: 0001-01-01T00:00:00Z,
            },
            DeletionTimestamp: nil,
            DeletionGracePeriodSeconds: nil,
            Labels: nil,
            Annotations: nil,
            OwnerReferences: nil,
            Finalizers: nil,
            ManagedFields: nil,
        },
        Spec: {
            Volumes: nil,
            InitContainers: nil,
            Containers: [
                {
                    Name: "nginx",
                    Image: "nginx:alpine",
                    Command: nil,
                    Args: nil,
                    WorkingDir: "",
                    Ports: nil,
                    EnvFrom: nil,
                    Env: nil,
                    Resources: {Limits: nil, Requests: nil, Claims: nil},
                    VolumeMounts: nil,
                    VolumeDevices: nil,
                    LivenessProbe: nil,
                    ReadinessProbe: nil,
                    StartupProbe: nil,
                    Lifecycle: nil,
                    TerminationMessagePath: "",
                    TerminationMessagePolicy: "",
                    ImagePullPolicy: "",
                    SecurityContext: nil,
                    Stdin: false,
                    StdinOnce: false,
                    TTY: false,
                },
                {
                    Name: "nginx3",
                    Image: "nginx:alpine",
                    Command: nil,
                    Args: nil,
                    WorkingDir: "",
                    Ports: nil,
                    EnvFrom: nil,
                    Env: nil,
                    Resources: {Limits: nil, Requests: nil, Claims: nil},
                    VolumeMounts: nil,
                    VolumeDevices: nil,
                    LivenessProbe: nil,
                    ReadinessProbe: nil,
                    StartupProbe: nil,
                    Lifecycle: nil,
                    TerminationMessagePath: "",
                    TerminationMessagePolicy: "",
                    ImagePullPolicy: "",
                    SecurityContext: nil,
                    Stdin: false,
                    StdinOnce: false,
                    TTY: false,
                },
            ],
            EphemeralContainers: nil,
            RestartPolicy: "",
            TerminationGracePeriodSeconds: nil,
            ActiveDeadlineSeconds: nil,
            DNSPolicy: "",
            NodeSelector: nil,
            ServiceAccountName: "",
            DeprecatedServiceAccount: "",
            AutomountServiceAccountToken: nil,
            NodeName: "",
            HostNetwork: false,
            HostPID: false,
            HostIPC: false,
            ShareProcessNamespace: nil,
            SecurityContext: nil,
            ImagePullSecrets: nil,
            Hostname: "",
            Subdomain: "",
            Affinity: nil,
            SchedulerName: "",
            Tolerations: nil,
            HostAliases: nil,
            PriorityClassName: "",
            Priority: nil,
            DNSConfig: nil,
            ReadinessGates: nil,
            RuntimeClassName: nil,
            EnableServiceLinks: nil,
            PreemptionPolicy: nil,
            Overhead: nil,
            TopologySpreadConstraints: nil,
            SetHostnameAsFQDN: nil,
            OS: nil,
            HostUsers: nil,
            SchedulingGates: nil,
            ResourceClaims: nil,
        },
        Status: {
            Phase: "",
            Conditions: nil,
            Message: "",
            Reason: "",
            NominatedNodeName: "",
            HostIP: "",
            PodIP: "",
            PodIPs: nil,
            StartTime: nil,
            InitContainerStatuses: nil,
            ContainerStatuses: nil,
            QOSClass: "",
            EphemeralContainerStatuses: nil,
        },
    }
not to equal
    <*v1.Pod>: {
        TypeMeta: {Kind: "", APIVersion: ""},
        ObjectMeta: {
            Name: "nginx",
            GenerateName: "",
            Namespace: "default",
            SelfLink: "",
            UID: "",
            ResourceVersion: "",
            Generation: 0,
            CreationTimestamp: {
                Time: 0001-01-01T00:00:00Z,
            },
            DeletionTimestamp: nil,
            DeletionGracePeriodSeconds: nil,
            Labels: nil,
            Annotations: nil,
            OwnerReferences: nil,
            Finalizers: nil,
            ManagedFields: nil,
        },
        Spec: {
            Volumes: nil,
            InitContainers: nil,
            Containers: [
                {
                    Name: "nginx",
                    Image: "nginx:alpine",
                    Command: nil,
                    Args: nil,
                    WorkingDir: "",
                    Ports: nil,
                    EnvFrom: nil,
                    Env: nil,
                    Resources: {Limits: nil, Requests: nil, Claims: nil},
                    VolumeMounts: nil,
                    VolumeDevices: nil,
                    LivenessProbe: nil,
                    ReadinessProbe: nil,
                    StartupProbe: nil,
                    Lifecycle: nil,
                    TerminationMessagePath: "",
                    TerminationMessagePolicy: "",
                    ImagePullPolicy: "",
                    SecurityContext: nil,
                    Stdin: false,
                    StdinOnce: false,
                    TTY: false,
                },
                {
                    Name: "nginx2",
                    Image: "nginx:alpine",
                    Command: nil,
                    Args: nil,
                    WorkingDir: "",
                    Ports: nil,
                    EnvFrom: nil,
                    Env: nil,
                    Resources: {Limits: nil, Requests: nil, Claims: nil},
                    VolumeMounts: nil,
                    VolumeDevices: nil,
                    LivenessProbe: nil,
                    ReadinessProbe: nil,
                    StartupProbe: nil,
                    Lifecycle: nil,
                    TerminationMessagePath: "",
                    TerminationMessagePolicy: "",
                    ImagePullPolicy: "",
                    SecurityContext: nil,
                    Stdin: false,
                    StdinOnce: false,
                    TTY: false,
                },
            ],
            EphemeralContainers: nil,
            RestartPolicy: "",
            TerminationGracePeriodSeconds: nil,
            ActiveDeadlineSeconds: nil,
            DNSPolicy: "",
            NodeSelector: nil,
            ServiceAccountName: "",
            DeprecatedServiceAccount: "",
            AutomountServiceAccountToken: nil,
            NodeName: "",
            HostNetwork: false,
            HostPID: false,
            HostIPC: false,
            ShareProcessNamespace: nil,
            SecurityContext: nil,
            ImagePullSecrets: nil,
            Hostname: "",
            Subdomain: "",
            Affinity: nil,
            SchedulerName: "",
            Tolerations: nil,
            HostAliases: nil,
            PriorityClassName: "",
            Priority: nil,
            DNSConfig: nil,
            ReadinessGates: nil,
            RuntimeClassName: nil,
            EnableServiceLinks: nil,
            PreemptionPolicy: nil,
            Overhead: nil,
            TopologySpreadConstraints: nil,
            SetHostnameAsFQDN: nil,
            OS: nil,
            HostUsers: nil,
            SchedulingGates: nil,
            ResourceClaims: nil,
        },
        Status: {
            Phase: "",
            Conditions: nil,
            Message: "",
            Reason: "",
            NominatedNodeName: "",
            HostIP: "",
            PodIP: "",
            PodIPs: nil,
            StartTime: nil,
            InitContainerStatuses: nil,
            ContainerStatuses: nil,
            QOSClass: "",
            EphemeralContainerStatuses: nil,
        },
    }`,
		},
	}
	memRe := regexp.MustCompile(`\<\*v1\.Pod \| .*\>`)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := &LooseDeepEqualMatcher{
				Expected: tt.fields.Expected,
			}
			gotMessage := matcher.NegatedFailureMessage(tt.args.actual)

			// replace space
			gotMessage = strings.ReplaceAll(gotMessage, " ", " ")

			// replace memory expression
			gotMessage = string(memRe.ReplaceAll([]byte(gotMessage), []byte(`<*v1.Pod>`)))

			if gotMessage != tt.wantMessage {
				t.Errorf("LooseDeepEqualMatcher.NegatedFailureMessage() diff = %s", cmp.Diff(gotMessage, tt.wantMessage))
			}
		})
	}
}
