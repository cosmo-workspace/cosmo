{{/*
Expand the name of the chart.
*/}}
{{- define "cosmo-dashboard.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "cosmo-dashboard.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "cosmo-dashboard.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "cosmo-dashboard.labels" -}}
helm.sh/chart: {{ include "cosmo-dashboard.chart" . }}
{{ include "cosmo-dashboard.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "cosmo-dashboard.selectorLabels" -}}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/name: {{ include "cosmo-dashboard.name" . }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "cosmo-dashboard.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "cosmo-dashboard.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Generate certificates
*/}}
{{- define "cosmo-dashboard.gen-certs" -}}
{{- $altNames := list ( printf "%s.%s.svc" "cosmo-dashboard" .Release.Namespace ) ( printf "%s.%s.svc.cluster.local" "cosmo-dashboard" .Release.Namespace ) ( printf "%s" .Values.dnsName ) -}}
{{- $ca := genCA "cosmo-dashboard-ca" 3650 -}}
{{- $cert := genSignedCert ( include "cosmo-dashboard.fullname" . ) nil $altNames 3650 $ca -}}
caCert: {{ $ca.Cert | b64enc }}
clientCert: {{ $cert.Cert | b64enc }}
clientKey: {{ $cert.Key | b64enc }}
{{- end -}}