all: template

.PHONY: template
template:
	kustomize build kubernetes | cosmoctl tmpl gen --useraddon \
		--name gryffindor-serviceaccount \
	    --userroles=gryffindor,gryffindor-admin \
		--disable-nameprefix \
		--cluster-scope \
		--required-useraddons resource-limitter \
		-o addon.yaml
