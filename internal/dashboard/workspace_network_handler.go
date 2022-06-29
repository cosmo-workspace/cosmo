package dashboard

import (
	"context"
	"net/http"
	"sort"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"k8s.io/utils/pointer"
)

func (s *Server) PutNetworkRule(ctx context.Context, userId string, workspaceName string, networkRuleName string, req dashv1alpha1.UpsertNetworkRuleRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "workspaceName", workspaceName, "networkRuleName", networkRuleName, "req", req)

	netRule, err := s.Klient.AddNetworkRule(ctx, workspaceName, userId, networkRuleName, int(req.PortNumber), pointer.String(req.Group), req.HttpPath, req.Public)
	if err != nil {
		return ErrorResponse(log, err)
	}

	res := &dashv1alpha1.UpsertNetworkRuleResponse{
		Message:     "Successfully upserted network rule",
		NetworkRule: convertNetRuleTodashv1alpha1NetRule(*netRule),
	}
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) DeleteNetworkRule(ctx context.Context, userId string, workspaceName string, networkRuleName string) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "workspaceName", workspaceName, "networkRuleName", networkRuleName)

	delRule, err := s.Klient.DeleteNetworkRule(ctx, workspaceName, userId, networkRuleName)
	if err != nil {
		return ErrorResponse(log, err)
	}

	res := &dashv1alpha1.RemoveNetworkRuleResponse{
		Message:     "Successfully removed network rule",
		NetworkRule: convertNetRuleTodashv1alpha1NetRule(*delRule),
	}
	return NormalResponse(http.StatusOK, res)
}

func convertNetRulesTodashv1alpha1NetRules(netRules []wsv1alpha1.NetworkRule, urlMap map[string]string, serviceMainPortName string) []dashv1alpha1.NetworkRule {
	apirules := make([]dashv1alpha1.NetworkRule, 0, len(netRules))
	for _, v := range netRules {
		if v.PortName == serviceMainPortName {
			continue
		}

		r := convertNetRuleTodashv1alpha1NetRule(v)
		r.Url = urlMap[v.PortName]

		apirules = append(apirules, r)
	}
	sort.Slice(apirules, func(i, j int) bool { return apirules[i].PortName < apirules[j].PortName })

	return apirules
}

func convertNetRuleTodashv1alpha1NetRule(v wsv1alpha1.NetworkRule) dashv1alpha1.NetworkRule {
	return dashv1alpha1.NetworkRule{
		PortName:   v.PortName,
		PortNumber: int32(v.PortNumber),
		Group:      *v.Group,
		HttpPath:   v.HTTPPath,
		Public:     v.Public,
	}
}
