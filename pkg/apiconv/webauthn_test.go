package apiconv

import (
	"reflect"
	"testing"
	"time"

	cosmowebauthn "github.com/cosmo-workspace/cosmo/pkg/auth/webauthn"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/go-webauthn/webauthn/webauthn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestC2D_Credentials(t *testing.T) {
	now := time.Now()
	nowpb := timestamppb.New(now)
	nowpb.Nanos = 0
	type args struct {
		creds []cosmowebauthn.Credential
	}
	tests := []struct {
		name string
		args args
		want []*dashv1alpha1.Credential
	}{
		{
			name: "normal",
			args: args{
				creds: []cosmowebauthn.Credential{
					{
						Base64URLEncodedId: "YWJjZGVmZwo=",
						DisplayName:        "abcdefg",
						Timestamp:          now.Unix(),
						Cred:               webauthn.Credential{},
					},
				},
			},
			want: []*dashv1alpha1.Credential{
				{
					Id:          "YWJjZGVmZwo=",
					DisplayName: "abcdefg",
					Timestamp:   nowpb,
				},
			},
		},
		{
			name: "empty",
			args: args{
				creds: []cosmowebauthn.Credential{},
			},
			want: []*dashv1alpha1.Credential{},
		},
		{
			name: "empty",
			args: args{
				creds: nil,
			},
			want: []*dashv1alpha1.Credential{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := C2D_Credentials(tt.args.creds); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("C2D_Credentials() = %v, want %v", got, tt.want)
			}
		})
	}
}
