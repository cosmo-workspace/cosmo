
HELM = ../bin/helm
KUSTOMIZE = ../bin/kustomize

all: crd cosmo-username-headers-addon.yaml webhook.yaml

helm:
	make -C ../ helm

helm-dependency-update:
	cd cosmo; ../$(HELM) dependency update

chartsnap:
	-helm plugin install https://github.com/jlandowner/helm-chartsnap

TEST_VALUES ?= cosmo/test

test: chartsnap helm helm-dependency-update
	helm chartsnap -c cosmo --values $(TEST_VALUES) -n cosmo-system $(CHARTSNAP_OPT)

test-list:
	grep -R '{{[-|] if .*}}' cosmo/templates/* | grep .Values | awk -F':' '{print $$2}' | sed -n 's/.*\(.Values[^ ]*\).*/\1/p' | tr -d ')' | sort | uniq | awk -F'.Values.' '{print $$2}' > cosmo/test/if-values.list

crd:
	cp ../config/crd/bases/* cosmo/crds/

ADDON_YAML ?= cosmo/templates/cosmo-username-headers-addon.yaml

cosmo-username-headers-addon.yaml:
	echo '{{- if .Values.installDefaultUserAddon }}' > $(ADDON_YAML)
	cat ../config/user-addon/traefik-middleware/cosmo-username-headers-addon.yaml \
		| sed -e 's/{{INSTANCE}}/{{ print "{{INSTANCE}}" }}/g' \
		| sed -e 's/{{TEMPLATE}}/{{ print "{{TEMPLATE}}" }}/g' \
		| sed -e 's/{{NAMESPACE}}/{{ print "{{NAMESPACE}}" }}/g' \
		| sed -e 's/{{USER_NAME}}/{{ print "{{USER_NAME}}" }}/g' >> $(ADDON_YAML)
	echo '{{- end }}' >> $(ADDON_YAML)

WEBHOOK_CHART_YAML ?= cosmo/templates/controller-manager/webhook.yaml
webhook.yaml:
	$(KUSTOMIZE) build ../config/webhook-chart \
		| sed -e 's/namespace: system/namespace: {{ .Release.Namespace }}/g' \
		| sed -z 's;apiVersion: v1\nkind: Service\nmetadata:\n  name: cosmo-webhook-service\n  namespace: {{ .Release.Namespace }}\nspec:\n  ports:\n  - port: 443\n    targetPort: 9443\n  selector:\n    control-plane: controller-manager\n---;{{ $$tls := fromYaml ( include "cosmo.webhook.gen-certs" . ) }}\n---\n{{- if $$.Values.controllerManager.webhook.enabled }};g' \
		| sed -z 's;metadata:\n;metadata:\n{{- if $$.Values.certManager.enabled }}\n  annotations:\n    cert-manager.io/inject-ca-from: {{ .Release.Namespace }}/cosmo-webhook-cert\n  {{- end }}\n  labels:\n    {{- include "cosmo.labels" . | nindent 4 }}\n;g' \
		| sed -z 's;clientConfig:;clientConfig:\n    caBundle: {{ if not $$.Values.certManager.enabled -}}{{ $$tls.caCert }}{{- else -}}Cg=={{ end }};g' > $(WEBHOOK_CHART_YAML)
	echo '{{- end }}' >> $(WEBHOOK_CHART_YAML)
