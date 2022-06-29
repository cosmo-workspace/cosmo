package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/wsnet"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/utils/pointer"
)

func (s *Server) PutNetworkRule(ctx context.Context, userId string, workspaceName string, networkRuleName string, req dashv1alpha1.UpsertNetworkRuleRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "workspaceName", workspaceName, "networkRuleName", networkRuleName, "req", req)

	ws := workspaceFromContext(ctx)
	if ws == nil {
		log.Info("workspace not found in context")
		return ErrorResponse_old(http.StatusInternalServerError, "")
	}
	before := ws.DeepCopy()

	netRule := wsv1alpha1.NetworkRule{
		PortName:   networkRuleName,
		PortNumber: int(req.PortNumber),
		Group:      pointer.String(req.Group),
		HTTPPath:   req.HttpPath,
		Public:     req.Public,
	}
	log.Debug().Info("upserting network rule", "ws", ws.Name, "namespace", ws.Namespace, "netRule", netRule)

	var err error
	ws.Spec.Network, err = wsnet.UpsertNetRule(ws.Spec.Network, netRule)
	if err != nil {
		message := err.Error()
		log.Error(err, message, "userid", userId, "workspace", ws.Name, "netRuleName", networkRuleName)
		return ErrorResponse_old(http.StatusBadRequest, message)
	}

	log.DebugAll().Info("NetworkRule upserted", "ws", ws, "namespace", ws.Namespace, "netRule", netRule)
	log.DebugAll().PrintObjectDiff(before, ws)

	if equality.Semantic.DeepEqual(before, ws) {
		log.Info("no change", "userid", userId, "workspace", ws.Name, "netRuleName", networkRuleName)
		return ErrorResponse_old(http.StatusBadRequest, "no change in network rules")
	}

	if err := s.Klient.Update(ctx, ws); err != nil {
		message := "failed to upsert network rule"
		log.Error(err, message, "userid", userId, "workspace", ws.Name, "netRuleName", networkRuleName)
		return ErrorResponse_old(http.StatusInternalServerError, message)
	}

	res := &dashv1alpha1.UpsertNetworkRuleResponse{}
	res.Message = "Successfully upserted network rule"
	res.NetworkRule = convertNetRuleTodashv1alpha1NetRule(netRule)
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) DeleteNetworkRule(ctx context.Context, userId string, workspaceName string, networkRuleName string) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "workspaceName", workspaceName, "networkRuleName", networkRuleName)

	ws := workspaceFromContext(ctx)
	if ws == nil {
		log.Info("workspace not found in context")
		return ErrorResponse_old(http.StatusInternalServerError, "")
	}

	before := ws.DeepCopy()

	if networkRuleName == ws.Status.Config.ServiceMainPortName {
		return ErrorResponse_old(http.StatusBadRequest, "main port cannot be removed")
	}

	var delRule *wsv1alpha1.NetworkRule
	for _, v := range ws.Spec.Network {
		if v.PortName == networkRuleName {
			delRule = v.DeepCopy()
		}
	}
	if delRule == nil {
		message := fmt.Sprintf("port name %s is not found", networkRuleName)
		log.Info(message, "userid", userId, "workspace", ws.Name, "netRuleName", networkRuleName)
		return ErrorResponse_old(http.StatusBadRequest, message)
	}

	ws.Spec.Network = wsnet.RemoveNetworkOverrideByName(ws.Spec.Network, *delRule)
	log.DebugAll().Info("NetworkRule removed", "ws", ws, "userid", userId, "netRuleName", networkRuleName)

	log.DebugAll().PrintObjectDiff(before, ws)

	if err := s.Klient.Update(ctx, ws); err != nil {
		message := "failed to remove network rule"
		log.Error(err, message, "userid", userId, "workspace", ws.Name, "netRuleName", networkRuleName)
		return ErrorResponse_old(http.StatusInternalServerError, message)
	}

	res := &dashv1alpha1.RemoveNetworkRuleResponse{}
	res.Message = "Successfully removed network rule"
	res.NetworkRule = convertNetRuleTodashv1alpha1NetRule(*delRule)
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
