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
			name: "❌ flags are required",
			args: []string{
				"--ldap-url=xxxx",
			},
		},
		{
			name: "❌ cookie-hashkey is minimum 16 characters",
			args: []string{
				"--cookie-hashkey=123456789012345",
				"--cookie-blockkey=1234567890123456",
				"--insecure",
			},
		},
		{
			name: "❌ cookie-blockkey is minimum 16 characters",
			args: []string{
				"--cookie-hashkey=1234567890123456",
				"--cookie-blockkey=yyy",
				"--insecure",
			},
		},
		{
			name: "❌ ldap-url is invalid",
			args: []string{
				"--cookie-hashkey=1234567890123456",
				"--cookie-blockkey=1234567890123456",
				"--ldap-url=ldap://sss  ssss",
				"--ldap-basedn=xxxx",
				"--ldap-user-attr=yyyy",
				"--insecure",
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
