package apiconv

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	cosmowebauthn "github.com/cosmo-workspace/cosmo/pkg/auth/webauthn"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

func C2D_Credentials(creds []cosmowebauthn.Credential) []*dashv1alpha1.Credential {
	ret := make([]*dashv1alpha1.Credential, len(creds))
	for i, cred := range creds {
		ret[i] = &dashv1alpha1.Credential{
			Id:          cred.Base64URLEncodedId,
			DisplayName: cred.DisplayName,
			Timestamp:   timestamppb.New(time.Unix(cred.Timestamp, 0)),
		}
	}
	return ret
}
