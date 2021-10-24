package wsnet

import (
	"fmt"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
)

// UpsertNetRule update or insert network rule
func UpsertNetRule(netRules []wsv1alpha1.NetworkRule, r wsv1alpha1.NetworkRule) ([]wsv1alpha1.NetworkRule, error) {
	r = *r.DeepCopy()

	if len(netRules) == 0 {
		return []wsv1alpha1.NetworkRule{r}, nil
	}

	index := -1
	for i, netRule := range netRules {
		if netRule.PortName == r.PortName {
			index = i

		} else if netRule.PortNumber == r.PortNumber {
			return nil, fmt.Errorf("port %d is already used", r.PortNumber)
		}
	}
	if index >= 0 {
		netRules[index] = r
	} else {
		netRules = append(netRules, r)
	}

	return netRules, nil
}

// RemoveNetworkOverrideByName removes the ingress rule and service port from instance.spec.override.network.ingress.rules and service.ports.
func RemoveNetworkOverrideByName(netRules []wsv1alpha1.NetworkRule, r wsv1alpha1.NetworkRule) []wsv1alpha1.NetworkRule {
	if len(netRules) == 0 {
		return netRules
	}

	index := -1
	for i, netRule := range netRules {
		if netRule.PortName == r.PortName {
			index = i
		}
	}
	if index >= 0 {
		return netRules[:index+copy(netRules[index:], netRules[index+1:])]

	} else {
		return netRules
	}
}
