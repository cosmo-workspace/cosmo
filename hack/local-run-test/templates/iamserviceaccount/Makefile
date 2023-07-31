all: template

.PHONY: template
template:
	kustomize build kubernetes | cosmoctl tmpl gen --useraddon \
		--name team-a-serviceaccount \
	    --userroles=team-a-* \
		--disable-nameprefix \
		--cluster-scope \
		-o addon.yaml
