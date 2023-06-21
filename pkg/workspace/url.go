package workspace

import (
	"reflect"
	"strconv"
	"strings"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

const (
	// common
	URLVarPortName      = "{{PORT_NAME}}" // For compatibility
	URLVarNetRuleName   = "{{NETRULE_NAME}}"
	URLVarPortNumber    = "{{PORT_NUMBER}}"
	URLVarNetRuleGroup  = "{{NETRULE_GROUP}}"
	URLVarInstanceName  = "{{INSTANCE}}"
	URLVarWorkspaceName = "{{WORKSPACE}}"
	URLVarNamespace     = "{{NAMESPACE}}"
	URLVarUserName      = "{{USER_NAME}}"

	// for network type LoadBalancer service
	URLVarLoadBalancer = "{{LOAD_BALANCER}}"

	// for network type NodePort service
	URLVarNodePortNumber = "{{NODEPORT_NUMBER}}"
)

// e.g. http://localhost:{{PORT_NUMBER}}
// e.g. https://{{NETRULE_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain
type URLBase string

type URLVars struct {
	NetworkRuleName string
	PortNumber      string
	NetRuleGroup    string

	InstanceName  string
	WorkspaceName string
	UserName      string
	Namespace     string

	NodePortNumber string
	LoadBalancer   string

	IngressPath string
}

func NewURLVars(netRule cosmov1alpha1.NetworkRule) URLVars {
	netRule.Default()

	v := URLVars{}
	v.NetworkRuleName = netRule.Name
	v.PortNumber = strconv.Itoa(int(netRule.PortNumber))
	v.IngressPath = netRule.HTTPPath
	v.NetRuleGroup = *netRule.Group

	return v
}

func (u URLBase) GenURL(v URLVars) string {
	v.setDefault()

	url := string(u)

	url = strings.ReplaceAll(url, URLVarNetRuleName, v.NetworkRuleName)
	url = strings.ReplaceAll(url, URLVarPortName, v.NetworkRuleName)
	url = strings.ReplaceAll(url, URLVarPortNumber, v.PortNumber)
	url = strings.ReplaceAll(url, URLVarNetRuleGroup, v.NetRuleGroup)
	url = strings.ReplaceAll(url, URLVarInstanceName, v.InstanceName)
	url = strings.ReplaceAll(url, URLVarWorkspaceName, v.WorkspaceName)
	url = strings.ReplaceAll(url, URLVarNamespace, v.Namespace)
	url = strings.ReplaceAll(url, URLVarUserName, v.UserName)
	url = strings.ReplaceAll(url, URLVarNodePortNumber, v.NodePortNumber)
	url = strings.ReplaceAll(url, URLVarLoadBalancer, v.LoadBalancer)

	url += v.IngressPath

	return url
}

func GenerateIngressHost(r cosmov1alpha1.NetworkRule, name, namespace string, urlBase URLBase) string {
	urlvar := NewURLVars(r)
	urlvar.InstanceName = name
	urlvar.WorkspaceName = name
	urlvar.Namespace = namespace
	urlvar.UserName = cosmov1alpha1.UserNameByNamespace(namespace)

	ingUrl := urlBase.GenURL(urlvar)

	return extractHost(ingUrl)
}

func extractHost(url string) string {
	// http://localhost:8080/
	// http://localhost/

	// remove proto
	s := strings.Split(url, "://")
	if len(s) != 2 {
		return ""
	}
	noProto := s[1]

	// remove path
	s = strings.Split(noProto, "/")
	hostWithPort := s[0]

	// remove port
	s = strings.Split(hostWithPort, ":")
	return s[0]
}

// Default sets "undefined" to empty properties
func (v *URLVars) setDefault() {
	if v.IngressPath == "" {
		v.IngressPath = "/"
	}

	if !strings.HasPrefix(v.IngressPath, "/") {
		v.IngressPath = "/" + v.IngressPath
	}

	if v.PortNumber == "" {
		v.PortNumber = "0"
	}

	if v.NodePortNumber == "" {
		v.NodePortNumber = "0"
	}

	if v.NetRuleGroup == "" {
		v.NetRuleGroup = v.NetworkRuleName
	}

	val := reflect.Indirect(reflect.ValueOf(v))
	vt := val.Type()
	for i := 0; i < vt.NumField(); i++ {
		if val.Field(i).String() == "" {
			val.Field(i).SetString("undefined")
		}
	}
}

func (v *URLVars) Dump() []interface{} {
	rv := reflect.ValueOf(*v)
	rt := rv.Type()
	keyAndVals := make([]interface{}, rt.NumField()*2)

	for i := 0; i < rt.NumField(); i++ {
		keyAndVals[i*2] = rt.Field(i).Name
		keyAndVals[i*2+1] = rv.Field(i).Interface()
	}

	return keyAndVals
}
