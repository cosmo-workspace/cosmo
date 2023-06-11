package auth

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"

	"github.com/go-ldap/ldap/v3"
)

// LdapAuthorizer authorize with cosmo user's password secret
type LdapAuthorizer struct {
	URL               string // URI to your LDAP server/Domain Controller. ldap[s]://host_or_ip[:port]
	BaseDN            string // Base DN used for all LDAP queries. ex: dc=example,dc=com
	UserNameAttribute string // Prepended to the base DN for user queries. sAMAccountname, cn, uid, etc.
	StartTLS          bool   // Enables StartTLS functionality
	tlsConfig         *tls.Config
}

func NewLdapAuthorizer(ldapUrl, baseDN, userNameAttribute string, tlsConfig *tls.Config, startTLS bool) (*LdapAuthorizer, error) {

	u, err := url.Parse(ldapUrl)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "ldaps" || startTLS {
		if tlsConfig == nil {
			return nil, errors.New("tlsConfig is required")
		}
	}
	if u.Scheme == "ldaps" && startTLS {
		return nil, errors.New("'ldaps://' and 'startTLS' cannot be used together")
	}

	if tlsConfig != nil {
		tlsConfig = tlsConfig.Clone()
	}
	return &LdapAuthorizer{
		URL:               ldapUrl,
		BaseDN:            baseDN,
		UserNameAttribute: userNameAttribute,
		StartTLS:          startTLS,
		tlsConfig:         tlsConfig,
	}, nil
}

func (a *LdapAuthorizer) Authorize(ctx context.Context, msg AuthRequest) (bool, error) {

	conn, err := ldap.DialURL(a.URL, ldap.DialWithTLSConfig(a.tlsConfig))
	if err != nil {
		return false, err
	}
	defer conn.Close()

	if a.StartTLS {
		conn.StartTLS(a.tlsConfig)
	}

	userDN := fmt.Sprintf("%s=%s,%s", a.UserNameAttribute, msg.GetUserName(), a.BaseDN)
	err = conn.Bind(userDN, msg.GetPassword())
	return err == nil, err
}
