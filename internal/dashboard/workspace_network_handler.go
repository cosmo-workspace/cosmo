package dashboard

import (
	"context"
	"sort"

	connect_go "github.com/bufbuild/connect-go"
	"k8s.io/utils/pointer"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

func (s *Server) UpsertNetworkRule(ctx context.Context, req *connect_go.Request[dashv1alpha1.UpsertNetworkRuleRequest]) (*connect_go.Response[dashv1alpha1.UpsertNetworkRuleResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	if err := userAuthentication(ctx, req.Msg.UserName); err != nil {
		return nil, ErrResponse(log, err)
	}

	m := req.Msg
	netRule, err := s.Klient.AddNetworkRule(ctx, m.WsName, m.UserName, m.NetworkRule.Name, m.NetworkRule.PortNumber, pointer.String(m.NetworkRule.Group), m.NetworkRule.HttpPath, m.NetworkRule.Public)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	rule := convertNetRuleTodashv1alpha1NetRule(*netRule)
	res := &dashv1alpha1.UpsertNetworkRuleResponse{
		Message:     "Successfully upserted network rule",
		NetworkRule: &rule,
	}
	return connect_go.NewResponse(res), nil
}

func (s *Server) DeleteNetworkRule(ctx context.Context, req *connect_go.Request[dashv1alpha1.DeleteNetworkRuleRequest]) (*connect_go.Response[dashv1alpha1.DeleteNetworkRuleResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	if err := userAuthentication(ctx, req.Msg.UserName); err != nil {
		return nil, ErrResponse(log, err)
	}

	m := req.Msg
	delRule, err := s.Klient.DeleteNetworkRule(ctx, m.WsName, m.UserName, m.NetworkRuleName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	rule := convertNetRuleTodashv1alpha1NetRule(*delRule)
	res := &dashv1alpha1.DeleteNetworkRuleResponse{
		Message:     "Successfully removed network rule",
		NetworkRule: &rule,
	}
	return connect_go.NewResponse(res), nil
}

func convertNetRulesTodashv1alpha1NetRules(netRules []cosmov1alpha1.NetworkRule, urlMap map[string]string, serviceMainPortName string) []*dashv1alpha1.NetworkRule {
	apirules := make([]*dashv1alpha1.NetworkRule, 0, len(netRules))
	for _, v := range netRules {
		if v.Name == serviceMainPortName {
			continue
		}

		r := convertNetRuleTodashv1alpha1NetRule(v)
		r.Url = urlMap[v.Name]

		apirules = append(apirules, &r)
	}
	sort.Slice(apirules, func(i, j int) bool { return apirules[i].Name < apirules[j].Name })

	return apirules
}

func convertNetRuleTodashv1alpha1NetRule(v cosmov1alpha1.NetworkRule) dashv1alpha1.NetworkRule {
	return dashv1alpha1.NetworkRule{
		Name:       v.Name,
		PortNumber: int32(v.PortNumber),
		Group:      *v.Group,
		HttpPath:   v.HTTPPath,
		Public:     v.Public,
	}
}
