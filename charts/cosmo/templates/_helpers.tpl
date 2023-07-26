{{/*
Expand the name of the chart.
*/}}
{{- define "cosmo.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "cosmo.fullname" -}}
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
{{- define "cosmo.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "cosmo.labels" -}}
helm.sh/chart: {{ include "cosmo.chart" . }}
{{ include "cosmo.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "cosmo.selectorLabels" -}}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/name: {{ include "cosmo.name" . }}
{{- end }}

{{/*
Required domain
*/}}
{{- define "cosmo.domain" -}}
{{- $domain := required "'.Values.domain' is required" .Values.domain -}}
{{ $domain }}
{{- end }}

{{/*
Dashboad URL
*/}}
{{- define "cosmo.dashboard.signinUrl" -}}
{{ if not .Values.dashboard.tls.enabled -}}http{{- else -}}https{{ end }}://{{ .Values.dashboard.ingressRoute.host }}.{{ .Values.domain }}/#/signin
{{- end }}