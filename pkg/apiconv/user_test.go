package apiconv

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

func TestC2D_Users(t *testing.T) {
	type args struct {
		users []cosmov1alpha1.User
		opts  []UserConvertOptions
	}
	tests := []struct {
		name string
		args args
		want []*dashv1alpha1.User
	}{
		{
			name: "OK",
			args: args{
				users: []cosmov1alpha1.User{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "testuser",
						},
						Spec: cosmov1alpha1.UserSpec{
							DisplayName: "testdisplay",
							AuthType:    cosmov1alpha1.UserAuthTypePasswordSecert,
						},
						Status: cosmov1alpha1.UserStatus{
							Namespace: cosmov1alpha1.ObjectRef{
								ObjectReference: corev1.ObjectReference{
									Name: "cosmo-user-testuser",
								},
							},
							Phase: "Ready",
						},
					},
				},
			},
			want: []*dashv1alpha1.User{
				{
					Name:        "testuser",
					DisplayName: "testdisplay",
					AuthType:    cosmov1alpha1.UserAuthTypePasswordSecert.String(),
					Status:      "Ready",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := C2D_Users(tt.args.users, tt.args.opts...)
			newGot := make([]string, len(got))
			for _, v := range got {
				newGot = append(newGot, v.String())
			}
			want := make([]string, len(tt.want))
			for _, v := range tt.want {
				want = append(want, v.String())
			}
			if !reflect.DeepEqual(newGot, want) {
				t.Errorf("C2D_Users() = %v, want %v\ndiff = %v", newGot, want, cmp.Diff(want, newGot))
			}
		})
	}
}

func TestC2D_User(t *testing.T) {
	type args struct {
		user cosmov1alpha1.User
		opts []UserConvertOptions
	}
	tests := []struct {
		name string
		args args
		want *dashv1alpha1.User
	}{
		{
			name: "OK",
			args: args{
				user: cosmov1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{
						Name: "testuser",
					},
					Spec: cosmov1alpha1.UserSpec{
						DisplayName: "testdisplay",
						AuthType:    cosmov1alpha1.UserAuthTypePasswordSecert,
					},
				},
			},
			want: &dashv1alpha1.User{
				Name:        "testuser",
				DisplayName: "testdisplay",
				AuthType:    cosmov1alpha1.UserAuthTypePasswordSecert.String(),
			},
		},
		{
			name: "WithRaw",
			args: args{
				user: cosmov1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{
						Name: "testuser",
					},
					Spec: cosmov1alpha1.UserSpec{
						DisplayName: "testdisplay",
						AuthType:    cosmov1alpha1.UserAuthTypePasswordSecert,
					},
				},
				opts: []UserConvertOptions{
					WithUserRaw(ptr.To(true)),
				},
			},
			want: &dashv1alpha1.User{
				Name:        "testuser",
				DisplayName: "testdisplay",
				AuthType:    cosmov1alpha1.UserAuthTypePasswordSecert.String(),
				Raw: ptr.To(`apiVersion: cosmo-workspace.github.io/v1alpha1
kind: User
metadata:
  creationTimestamp: null
  name: testuser
spec:
  authType: password-secret
  displayName: testdisplay
status:
  namespace: {}
`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := C2D_User(tt.args.user, tt.args.opts...); !reflect.DeepEqual(got.String(), tt.want.String()) {
				t.Errorf("C2D_User() raw diff %v", cmp.Diff(tt.want.Raw, got.Raw))
				t.Errorf("C2D_User() obj diff %v", cmp.Diff(tt.want.String(), got.String()))
				t.Errorf("C2D_User() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestC2S_UserRole(t *testing.T) {
	type args struct {
		apiRoles []cosmov1alpha1.UserRole
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "OK",
			args: args{
				apiRoles: []cosmov1alpha1.UserRole{
					{
						Name: "testrole",
					},
				},
			},
			want: []string{"testrole"},
		},
		{
			name: "OK",
			args: args{
				apiRoles: []cosmov1alpha1.UserRole{
					{
						Name: "testrole",
					},
					{
						Name: "testrole2",
					},
				},
			},
			want: []string{"testrole", "testrole2"},
		},
		{
			name: "Empty",
			args: args{
				apiRoles: []cosmov1alpha1.UserRole{},
			},
			want: []string{},
		},
		{
			name: "Empty",
			args: args{
				apiRoles: nil,
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := C2S_UserRole(tt.args.apiRoles); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("C2S_UserRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestS2C_UserRoles(t *testing.T) {
	type args struct {
		roles []string
	}
	tests := []struct {
		name string
		args args
		want []cosmov1alpha1.UserRole
	}{
		{
			name: "OK",
			args: args{
				roles: []string{"testrole"},
			},
			want: []cosmov1alpha1.UserRole{
				{
					Name: "testrole",
				},
			},
		},
		{
			name: "OK",
			args: args{
				roles: []string{"testrole", "testrole2"},
			},
			want: []cosmov1alpha1.UserRole{
				{
					Name: "testrole",
				},
				{
					Name: "testrole2",
				},
			},
		},
		{
			name: "Empty",
			args: args{
				roles: []string{},
			},
			want: []cosmov1alpha1.UserRole{},
		},
		{
			name: "Empty",
			args: args{
				roles: nil,
			},
			want: []cosmov1alpha1.UserRole{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := S2C_UserRoles(tt.args.roles); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("S2C_UserRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestD2C_UserAddons(t *testing.T) {
	type args struct {
		addons []*dashv1alpha1.UserAddon
	}
	tests := []struct {
		name string
		args args
		want []cosmov1alpha1.UserAddon
	}{
		{
			name: "OK",
			args: args{
				addons: []*dashv1alpha1.UserAddon{
					{
						Template:      "testaddon",
						ClusterScoped: true,
					},
					{
						Template: "testaddon2",
						Vars: map[string]string{
							"key1": "val1",
							"key2": "val2",
						},
					},
				},
			},
			want: []cosmov1alpha1.UserAddon{
				{
					Template: cosmov1alpha1.UserAddonTemplateRef{
						Name:          "testaddon",
						ClusterScoped: true,
					},
				},
				{
					Template: cosmov1alpha1.UserAddonTemplateRef{
						Name: "testaddon2",
					},
					Vars: map[string]string{
						"key1": "val1",
						"key2": "val2",
					},
				},
			},
		},
		{
			name: "Empty",
			args: args{
				addons: []*dashv1alpha1.UserAddon{},
			},
			want: []cosmov1alpha1.UserAddon{},
		},
		{
			name: "Empty",
			args: args{
				addons: nil,
			},
			want: []cosmov1alpha1.UserAddon{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := D2C_UserAddons(tt.args.addons); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("D2C_UserAddons() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestC2D_UserAddons(t *testing.T) {
	type args struct {
		addons []cosmov1alpha1.UserAddon
	}
	tests := []struct {
		name string
		args args
		want []*dashv1alpha1.UserAddon
	}{
		{
			name: "OK",
			args: args{
				addons: []cosmov1alpha1.UserAddon{
					{
						Template: cosmov1alpha1.UserAddonTemplateRef{
							Name:          "testaddon",
							ClusterScoped: true,
						},
					},
					{
						Template: cosmov1alpha1.UserAddonTemplateRef{
							Name: "testaddon2",
						},
						Vars: map[string]string{
							"key1": "val1",
							"key2": "val2",
						},
					},
				},
			},
			want: []*dashv1alpha1.UserAddon{
				{
					Template:      "testaddon",
					ClusterScoped: true,
				},
				{
					Template: "testaddon2",
					Vars: map[string]string{
						"key1": "val1",
						"key2": "val2",
					},
				},
			},
		},
		{
			name: "Empty",
			args: args{
				addons: []cosmov1alpha1.UserAddon{},
			},
			want: []*dashv1alpha1.UserAddon{},
		},
		{
			name: "Empty",
			args: args{
				addons: nil,
			},
			want: []*dashv1alpha1.UserAddon{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := C2D_UserAddons(tt.args.addons); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("C2D_UserAddons() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestS2D_UserAddons(t *testing.T) {
	type args struct {
		addons []string
	}
	tests := []struct {
		name    string
		args    args
		want    []*dashv1alpha1.UserAddon
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				addons: []string{
					"testaddon",
					"testaddon2:key1=val1,key2=val2",
				},
			},
			want: []*dashv1alpha1.UserAddon{
				{
					Template: "testaddon",
				},
				{
					Template: "testaddon2",
					Vars: map[string]string{
						"key1": "val1",
						"key2": "val2",
					},
				},
			},
		},
		{
			name: "OK",
			args: args{
				addons: []string{"testaddon:key1=val1"},
			},
			want: []*dashv1alpha1.UserAddon{{Template: "testaddon", Vars: map[string]string{"key1": "val1"}}},
		},
		{
			name: "OK",
			args: args{
				addons: []string{"testaddon:key1=val1,key2=val2"},
			},
			want: []*dashv1alpha1.UserAddon{{Template: "testaddon", Vars: map[string]string{"key1": "val1", "key2": "val2"}}},
		},
		{
			name: "OK",
			args: args{
				addons: []string{"testaddon=key1=val1"},
			},
			want: []*dashv1alpha1.UserAddon{{Template: "testaddon=key1=val1"}},
		},
		{
			name: "OK",
			args: args{
				addons: []string{"testaddon,key1=val1"},
			},
			want: []*dashv1alpha1.UserAddon{{Template: "testaddon,key1=val1"}},
		},
		{
			name: "NG",
			args: args{
				addons: []string{"testaddon:"},
			},
			wantErr: true,
		},
		{
			name: "NG",
			args: args{
				addons: []string{"testaddon:key1=val1,"},
			},
			wantErr: true,
		},
		{
			name: "NG",
			args: args{
				addons: []string{":key1=val1"},
			},
			wantErr: true,
		},
		{
			name: "NG",
			args: args{
				addons: []string{":"},
			},
			wantErr: true,
		},
		{
			name: "NG",
			args: args{
				addons: []string{"testaddon:"},
			},
			wantErr: true,
		},
		{
			name: "NG",
			args: args{
				addons: []string{"testaddon:key"},
			},
			wantErr: true,
		},
		{
			name: "Duplicate key",
			args: args{
				addons: []string{"testaddon:key1=val1,key1=val2"},
			},
			wantErr: true,
		},
		{
			name: "Empty",
			args: args{
				addons: []string{},
			},
			want: []*dashv1alpha1.UserAddon{},
		},
		{
			name: "Empty",
			args: args{
				addons: nil,
			},
			want: []*dashv1alpha1.UserAddon{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := S2D_UserAddons(tt.args.addons)
			if (err != nil) != tt.wantErr {
				t.Errorf("S2D_UserAddons() error = %v, wantErr %v, got %v", err, tt.wantErr, got)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("S2D_UserAddons() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestD2S_UserAddons(t *testing.T) {
	type args struct {
		addons []*dashv1alpha1.UserAddon
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "OK",
			args: args{
				addons: []*dashv1alpha1.UserAddon{
					{
						Template: "testaddon",
					},
					{
						Template:      "testaddon2",
						ClusterScoped: true,
						Vars: map[string]string{
							"key1": "val1",
							"key2": "val2",
						},
					},
				},
			},
			want: []string{"testaddon", "testaddon2:key1=val1,key2=val2"},
		},
		{
			name: "Empty",
			args: args{
				addons: []*dashv1alpha1.UserAddon{},
			},
			want: []string{},
		},
		{
			name: "Empty",
			args: args{
				addons: nil,
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := D2S_UserAddons(tt.args.addons); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("D2S_UserAddons() = %v, want %v", got, tt.want)
			}
		})
	}
}

func timeParse(t string) metav1.Time {
	tt, _ := time.Parse(time.RFC3339, t)
	return metav1.Time{Time: tt}
}

func TestK2D_Events(t *testing.T) {
	type args struct {
		events []eventsv1.Event
	}
	tests := []struct {
		name string
		args args
		want []*dashv1alpha1.Event
	}{
		{
			name: "OK",
			args: args{
				events: []eventsv1.Event{
					{
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: timeParse("2024-05-20T14:00:50Z"),
							Name:              "ws1.17d13738ea85aac8",
							Namespace:         "cosmo-user-tom",
							ResourceVersion:   "1043537",
							UID:               "3acc422e-ef39-4f1a-ab88-389f20a5e22d",
						},

						DeprecatedCount:          1,
						DeprecatedFirstTimestamp: timeParse("2024-05-20T14:00:50Z"),
						DeprecatedLastTimestamp:  timeParse("2024-05-20T14:00:50Z"),
						DeprecatedSource: corev1.EventSource{
							Component: "cosmo-workspace-controller",
						},
						Note:   "successfully reconciled. instance synced",
						Reason: "updated",
						Regarding: corev1.ObjectReference{
							APIVersion:      "cosmo-workspace.github.io/v1alpha1",
							Kind:            "Workspace",
							Name:            "ws1",
							Namespace:       "cosmo-user-tom",
							ResourceVersion: "1043534",
							UID:             "71ebf5c1-fa6a-4058-ab92-5f7b82667e41",
						},
						ReportingController: "cosmo-workspace-controller",
						Type:                "Normal",
					},
					{
						DeprecatedCount:          1,
						DeprecatedFirstTimestamp: timeParse("2024-05-20T14:00:50Z"),
						DeprecatedLastTimestamp:  timeParse("2024-05-20T14:00:50Z"),
						DeprecatedSource: corev1.EventSource{
							Component: "cosmo-instance-controller",
						},
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: timeParse("2024-05-20T14:00:50Z"),
							Name:              "ws1.17d13738ecc558d0",
							Namespace:         "cosmo-user-tom",
							ResourceVersion:   "1043539",
							UID:               "faee1826-0edb-4541-a836-d51a3004ab1e",
						},
						Note:   "Deployment ws1-workspace is not desired state, synced",
						Reason: "Synced",
						Regarding: corev1.ObjectReference{
							APIVersion:      "cosmo-workspace.github.io/v1alpha1",
							Kind:            "Instance",
							Name:            "ws1",
							Namespace:       "cosmo-user-tom",
							ResourceVersion: "1043536",
							UID:             "06a048d8-a763-4511-84ea-f8505862b4a1",
						},
						ReportingController: "cosmo-instance-controller",
						Type:                "Normal",
					},
					{
						DeprecatedCount:          1,
						DeprecatedFirstTimestamp: timeParse("2024-05-20T14:00:50Z"),
						DeprecatedLastTimestamp:  timeParse("2024-05-20T14:00:50Z"),
						DeprecatedSource: corev1.EventSource{
							Component: "deployment-controller",
						},
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: timeParse("2024-05-20T14:00:50Z"),
							Name:              "ws1-workspace.17d13738edffd9f0",
							Namespace:         "cosmo-user-tom",
							ResourceVersion:   "1043541",
							UID:               "af12d627-35c9-4980-81ff-d66f2b791055",
						},
						Note:   "Scaled down replica set ws1-workspace-66b8cd6764 to 0 from 1",
						Reason: "ScalingReplicaSet",
						Regarding: corev1.ObjectReference{
							APIVersion:      "apps/v1",
							Kind:            "Deployment",
							Name:            "ws1-workspace",
							Namespace:       "cosmo-user-tom",
							ResourceVersion: "1043538",
							UID:             "da76fc70-622b-4b3a-8bd9-7912bd5a356a",
						},
						ReportingController: "deployment-controller",
						Type:                "Normal",
					},
					{
						DeprecatedCount:          1,
						DeprecatedFirstTimestamp: timeParse("2024-05-20T14:00:50Z"),
						DeprecatedLastTimestamp:  timeParse("2024-05-20T14:00:50Z"),
						DeprecatedSource: corev1.EventSource{
							Component: "replicaset-controller",
						},
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: timeParse("2024-05-20T14:00:50Z"),
							Name:              "ws1-workspace-66b8cd6764.17d13738ef8a33c4",
							Namespace:         "cosmo-user-tom",
							ResourceVersion:   "1043544",
							UID:               "f91550ad-2dd5-4235-89bf-360f74b8ce0e",
						},
						Note:   "Deleted pod: ws1-workspace-66b8cd6764-fz2k7",
						Reason: "SuccessfulDelete",
						Regarding: corev1.ObjectReference{
							APIVersion:      "apps/v1",
							Kind:            "ReplicaSet",
							Name:            "ws1-workspace-66b8cd6764",
							Namespace:       "cosmo-user-tom",
							ResourceVersion: "1043540",
							UID:             "323d78cf-86f9-4644-b078-a10ee8662da0",
						},
						ReportingController: "replicaset-controller",
						Type:                "Normal",
					},
					{
						DeprecatedCount:          1,
						DeprecatedFirstTimestamp: timeParse("2024-05-20T14:00:50Z"),
						DeprecatedLastTimestamp:  timeParse("2024-05-20T14:00:50Z"),
						DeprecatedSource: corev1.EventSource{
							Component: "kubelet",
							Host:      "k3d-cosmo-server-0",
						},
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: timeParse("2024-05-20T14:00:50Z"),
							Name:              "ws1-workspace-66b8cd6764-fz2k7.17d13738efcc7a13",
							Namespace:         "cosmo-user-tom",
							ResourceVersion:   "1043545",
							UID:               "d2f21050-ae78-4894-b6b1-19fe3918b7c2",
							Annotations: map[string]string{
								cosmov1alpha1.EventAnnKeyInstanceName: "aaa",
								cosmov1alpha1.EventAnnKeyUserName:     "bbb",
							},
						},
						Note:   "Stopping container code-server",
						Reason: "Killing",
						Regarding: corev1.ObjectReference{
							APIVersion:      "v1",
							FieldPath:       "spec.containers{code-server}",
							Kind:            "Pod",
							Name:            "ws1-workspace-66b8cd6764-fz2k7",
							Namespace:       "cosmo-user-tom",
							ResourceVersion: "3298",
							UID:             "d944b617-8f50-4222-93a0-33b6b6eb5b89",
						},
						ReportingController: "kubelet",
						ReportingInstance:   "k3d-cosmo-server-0",
						Type:                "Normal",
					},
				},
			},
			want: []*dashv1alpha1.Event{
				{
					Id:        "ws1.17d13738ea85aac8",
					User:      "tom",
					EventTime: timestamppb.New(timeParse("2024-05-20T14:00:50Z").Time),
					Type:      "Normal",
					Note:      "successfully reconciled. instance synced",
					Reason:    "updated",
					Regarding: &dashv1alpha1.ObjectReference{
						ApiVersion: "cosmo-workspace.github.io/v1alpha1",
						Kind:       "Workspace",
						Name:       "ws1",
						Namespace:  "cosmo-user-tom",
					},
					ReportingController: "cosmo-workspace-controller",
					Series: &dashv1alpha1.EventSeries{
						Count:            1,
						LastObservedTime: timestamppb.New(timeParse("2024-05-20T14:00:50Z").Time),
					},
				},
				{
					Id:        "ws1.17d13738ecc558d0",
					User:      "tom",
					EventTime: timestamppb.New(timeParse("2024-05-20T14:00:50Z").Time),
					Type:      "Normal",
					Note:      "Deployment ws1-workspace is not desired state, synced",
					Reason:    "Synced",
					Regarding: &dashv1alpha1.ObjectReference{
						ApiVersion: "cosmo-workspace.github.io/v1alpha1",
						Kind:       "Instance",
						Name:       "ws1",
						Namespace:  "cosmo-user-tom",
					},
					Series: &dashv1alpha1.EventSeries{
						Count:            1,
						LastObservedTime: timestamppb.New(timeParse("2024-05-20T14:00:50Z").Time),
					},
					ReportingController: "cosmo-instance-controller",
				},
				{
					Id:        "ws1-workspace.17d13738edffd9f0",
					User:      "tom",
					EventTime: timestamppb.New(timeParse("2024-05-20T14:00:50Z").Time),
					Type:      "Normal",
					Note:      "Scaled down replica set ws1-workspace-66b8cd6764 to 0 from 1",
					Reason:    "ScalingReplicaSet",
					Regarding: &dashv1alpha1.ObjectReference{
						ApiVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "ws1-workspace",
						Namespace:  "cosmo-user-tom",
					},
					Series: &dashv1alpha1.EventSeries{
						Count:            1,
						LastObservedTime: timestamppb.New(timeParse("2024-05-20T14:00:50Z").Time),
					},

					ReportingController: "deployment-controller",
				},
				{
					Id:        "ws1-workspace-66b8cd6764.17d13738ef8a33c4",
					User:      "tom",
					EventTime: timestamppb.New(timeParse("2024-05-20T14:00:50Z").Time),
					Type:      "Normal",
					Note:      "Deleted pod: ws1-workspace-66b8cd6764-fz2k7",
					Reason:    "SuccessfulDelete",
					Regarding: &dashv1alpha1.ObjectReference{
						ApiVersion: "apps/v1",
						Kind:       "ReplicaSet",
						Name:       "ws1-workspace-66b8cd6764",
						Namespace:  "cosmo-user-tom",
					},
					Series: &dashv1alpha1.EventSeries{
						Count:            1,
						LastObservedTime: timestamppb.New(timeParse("2024-05-20T14:00:50Z").Time),
					},
					ReportingController: "replicaset-controller",
				},
				{
					Id:        "ws1-workspace-66b8cd6764-fz2k7.17d13738efcc7a13",
					EventTime: timestamppb.New(timeParse("2024-05-20T14:00:50Z").Time),
					Type:      "Normal",
					Note:      "Stopping container code-server",
					Reason:    "Killing",
					Regarding: &dashv1alpha1.ObjectReference{
						ApiVersion: "v1",
						Kind:       "Pod",
						Name:       "ws1-workspace-66b8cd6764-fz2k7",
						Namespace:  "cosmo-user-tom",
					},
					Series: &dashv1alpha1.EventSeries{
						Count:            1,
						LastObservedTime: timestamppb.New(timeParse("2024-05-20T14:00:50Z").Time),
					},
					ReportingController: "kubelet",
					RegardingWorkspace:  ptr.To("aaa"),
					User:                "bbb",
				},
			},
		},
		{
			name: "NewEvent",
			args: args{
				events: []eventsv1.Event{
					{
						EventTime: metav1.MicroTime(timeParse("2024-05-20T14:00:50Z")),
						Type:      "Normal",
						Note:      "successfully reconciled. instance synced",
						Reason:    "updated",
						Regarding: corev1.ObjectReference{
							APIVersion: "cosmo-workspace.github.io/v1alpha1",
							Kind:       "Workspace",
							Name:       "ws1",
							Namespace:  "cosmo-user-tom",
						},
						ReportingController: "cosmo-workspace-controller",
						Series: &eventsv1.EventSeries{
							Count:            3,
							LastObservedTime: metav1.MicroTime(timeParse("2024-05-20T14:30:50Z")),
						},
					},
				},
			},
			want: []*dashv1alpha1.Event{
				{
					EventTime: timestamppb.New(timeParse("2024-05-20T14:00:50Z").Time),
					Type:      "Normal",
					Note:      "successfully reconciled. instance synced",
					Reason:    "updated",
					Regarding: &dashv1alpha1.ObjectReference{
						ApiVersion: "cosmo-workspace.github.io/v1alpha1",
						Kind:       "Workspace",
						Name:       "ws1",
						Namespace:  "cosmo-user-tom",
					},
					ReportingController: "cosmo-workspace-controller",
					Series: &dashv1alpha1.EventSeries{
						Count:            3,
						LastObservedTime: timestamppb.New(timeParse("2024-05-20T14:30:50Z").Time),
					},
				},
			},
		},
		{
			name: "Empty",
			args: args{
				events: []eventsv1.Event{},
			},
			want: []*dashv1alpha1.Event{},
		},
		{
			name: "Empty",
			args: args{
				events: nil,
			},
			want: []*dashv1alpha1.Event{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := K2D_Events(tt.args.events); !reflect.DeepEqual(got, tt.want) {
				for i, g := range got {
					t.Errorf("K2D_Events() %d diff %v", i, cmp.Diff(tt.want[i].String(), g.String()))
				}
				t.Errorf("K2D_Events() = %v, want %v", got, tt.want)
			}
		})
	}
}
