package auth

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

// LdapAuthorizer authorize with cosmo user's password secret
type LdapAuthorizer struct {
	// for connection
	URL                string // URI to your LDAP server/Domain Controller. ldap[s]://host_or_ip[:port]
	StartTLS           bool   // Enables StartTLS functionality
	TlsConfig          *tls.Config
	BindDN             string // [for Bind mode] ex: uid=%s,ou=users,dc=example,dc=com. "%s" is replaced by the user name.
	SearchBindDN       string // [for Search mode] The domain name to bind. ex: cn=admin,dc=example,dc=com
	SearchBindPassword string // [for Search mode] The password to bind.
	SearchBaseDN       string // [for Search mode] Base DN used for all LDAP queries. ex: dc=example,dc=com
	SearchFilter       string // [for Search mode] ex: "(sAMAccountname=%s)"  "%s" is replaced by the user name.
}

func (a *LdapAuthorizer) Authorize(ctx context.Context, msg AuthRequest) (bool, error) {

	conn, err := ldap.DialURL(a.URL, ldap.DialWithTLSConfig(a.TlsConfig))
	if err != nil {
		return false, err
	}
	defer conn.Close()

	if a.StartTLS {
		if err := conn.StartTLS(a.TlsConfig); err != nil {
			return false, err
		}
	}
	if a.SearchFilter == "" {
		return a.checkWithBindMode(ctx, conn, msg)
	} else {
		return a.checkWithSearchMode(ctx, conn, msg)
	}
}

func (a *LdapAuthorizer) checkWithBindMode(ctx context.Context, conn *ldap.Conn, msg AuthRequest) (bool, error) {
	userDN := fmt.Sprintf(a.BindDN, msg.GetUserName())
	err := conn.Bind(userDN, msg.GetPassword())
	return err == nil, err
}

func (a *LdapAuthorizer) checkWithSearchMode(ctx context.Context, conn *ldap.Conn, msg AuthRequest) (bool, error) {

	if a.SearchBindDN != "" && a.SearchBindPassword != "" {
		if err := conn.Bind(a.SearchBindDN, a.SearchBindPassword); err != nil {
			return false, err
		}
	} else {
		_ = conn.UnauthenticatedBind("")
	}

	searchFilter := fmt.Sprintf(a.SearchFilter, msg.GetUserName())

	search := ldap.NewSearchRequest(
		a.SearchBaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		searchFilter,
		[]string{"dn", "cn", "sAMAccountName"},
		nil,
	)

	result, err := conn.Search(search)
	if err != nil {
		return false, err
	}
	if len(result.Entries) < 1 {
		return false, fmt.Errorf("not found user")
	} else if len(result.Entries) > 1 {
		return false, fmt.Errorf(fmt.Sprintf("found too many user (%d)", len(result.Entries)))
	}

	userDN := result.Entries[0].DN
	err = conn.Bind(userDN, msg.GetPassword())
	return err == nil, err
}
