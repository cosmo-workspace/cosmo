/*
Copyright © 2023 NAME HERE cosmo-workspace
*/
package dashboard

import (
	"bytes"
	"io"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func TestNewRootCmd(t *testing.T) {

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "❌ cookie-hashkey is required",
			args: []string{""},
		},
		{
			name: "❌ cookie-hashkey is minimum 16 characters",
			args: []string{
				"--cookie-hashkey=123456789012345",
				"--cookie-blockkey=1234567890123456",
			},
		},
		{
			name: "❌ cookie-blockkey is required",
			args: []string{
				"--cookie-hashkey=1234567890123456",
			},
		},
		{
			name: "❌ cookie-blockkey is minimum 16 characters",
			args: []string{
				"--cookie-hashkey=1234567890123456",
				"--cookie-blockkey=yyy",
			},
		},
		{
			name: "❌ tls-cert is required",
			args: []string{
				"--cookie-hashkey=1234567890123456",
				"--cookie-blockkey=1234567890123456",
				"--insecure=false",
				"--tls-key=",
				"--tls-cert=",
			},
		},
		{
			name: "❌ tls-key is required",
			args: []string{
				"--cookie-hashkey=1234567890123456",
				"--cookie-blockkey=1234567890123456",
				"--insecure=false",
				"--tls-key=",
				"--tls-cert=xxxxx",
			},
		},
		{
			name: "❌ ldap-basedn is required",
			args: []string{
				"--cookie-hashkey=1234567890123456",
				"--cookie-blockkey=1234567890123456",
				"--ldap-url=ldap://sssssss",
			},
		},
		{
			name: "❌ ldap-user-attr is required",
			args: []string{
				"--cookie-hashkey=1234567890123456",
				"--cookie-blockkey=1234567890123456",
				"--ldap-url=ldap://sssssss",
				"--ldap-basedn=xxxxx",
				"--ldap-user-attr=",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := bytes.NewBufferString("")
			o := &options{}
			rootCmd := NewRootCmd(o)
			rootCmd.SetOutput(outBuf)
			rootCmd.SetArgs(tt.args)
			_ = rootCmd.Execute()
			out, _ := io.ReadAll(outBuf)
			snaps.MatchSnapshot(t, string(out))
		})
	}
}
