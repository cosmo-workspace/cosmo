package auth

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

// LdapAuthorizer authorize with cosmo user's password secret
type LdapAuthorizer struct {
	URL               string // ldap[s]://host_or_ip[:port]
	BaseDN            string // dc=example,dc=com
	UserNameAttribute string // sAMAccountname, cn, uid, etc.
}

func NewLdapAuthorizer(url, baseDN, userNameAttribute string) *LdapAuthorizer {
	return &LdapAuthorizer{
		URL:               url,
		BaseDN:            baseDN,
		UserNameAttribute: userNameAttribute,
	}
}

func (a *LdapAuthorizer) Authorize(ctx context.Context, msg AuthRequest) (bool, error) {

	conn, err := ldap.DialURL(a.URL, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	if err != nil {
		return false, err
	}
	defer conn.Close()

	userDN := fmt.Sprintf("%s=%s,%s", a.UserNameAttribute, msg.GetUserName(), a.BaseDN)
	err = conn.Bind(userDN, msg.GetPassword())
	return err == nil, err
}
