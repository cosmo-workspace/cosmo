
HELM = ../bin/helm

helm:
	make -C ../ helm

chartsnap:
	go install github.com/cosmo/controller-testtools/cmd/chartsnap

test:
	chartsnap --chart cosmo --values cosmo/test

test-list:
	grep -R '{{[-|] if .*}}' cosmo/templates/* | grep .Values | awk -F':' '{print $$2}' | sed -n 's/.*\(.Values[^ ]*\).*/\1/p' | tr -d ')' | sort | uniq | awk -F'.Values.' '{print $$2}' > cosmo/test/if-values.list
