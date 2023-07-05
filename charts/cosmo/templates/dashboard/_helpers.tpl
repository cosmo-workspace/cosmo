{{/*
Generate certificates for dashboard
*/}}
{{- define "cosmo.dashboard.gen-certs" -}}
{{- $altNames := list ( printf "%s.%s.svc" "cosmo-dashboard" .Release.Namespace ) ( printf "%s.%s.svc.cluster.local" "cosmo-dashboard" .Release.Namespace ) -}}
{{- $ca := genCA "cosmo-ca" 3650 -}}
{{- $cert := genSignedCert ( include "cosmo.fullname" . ) nil $altNames 3650 $ca -}}
{{/* fetch current certificates if exist */}}
{{- $currentData := (lookup "v1" "Secret" .Release.Namespace "dashboard-server-cert").data | default dict }}
caCert: {{ (get $currentData "ca.crt") | default ($ca.Cert | b64enc) }}
clientCert: {{ (get $currentData "tls.crt") | default ($cert.Cert | b64enc) }}
clientKey: {{ (get $currentData "tls.key") | default ($cert.Key | b64enc) }}
{{- end }}