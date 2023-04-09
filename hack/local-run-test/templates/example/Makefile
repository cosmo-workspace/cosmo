FROM_TAG=4.3.0

.PHONY: apply
apply: ## Apply template
	kubectl apply -f ../../../../example/code-server-ingress-no-storage-no-authproxy.yaml
	kubectl apply -f ../../../../example/code-server-ingress-no-storage.yaml
	kubectl apply -f ../../../../example/code-server-loadbalancer-no-storage.yaml
	kubectl apply -f ../../../../example/fixed-efs-pv-useraddon.yaml
	kubectl apply -f ../../../../example/theia.yaml
