{{/*
Expand the name of the chart.
*/}}
{{- define "artifact-registry.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "artifact-registry.fullname" -}}
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
{{- define "artifact-registry.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "artifact-registry.labels" -}}
helm.sh/chart: {{ include "artifact-registry.chart" . }}
{{ include "artifact-registry.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "artifact-registry.selectorLabels" -}}
app.kubernetes.io/name: {{ include "artifact-registry.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "artifact-registry.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "artifact-registry.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "artifact-registry.backend" -}}
{{- if (.Values.config.backend).host -}}
{{ .Values.config.backend.host }}
{{- else if .Values.registry.enabled -}}
{{ template "docker-registry.fullname" .Subcharts.registry }}:{{ .Values.registry.service.port }}
{{- else -}}
docker.io
{{- end }}
{{- end }}

{{- define "artifact-registry.aesKeySecretName" -}}
{{- if (((.Values.config).aes).secret).name -}}
{{ .Values.config.aes.secret.name }}
{{- else -}}
{{ include "artifact-registry.fullname" . }}-key
{{- end }}
{{- end }}

{{- define "artifact-registry.aesKeySecretKeyName" -}}
{{- if (((.Values.config).aes).secret).key -}}
{{ .Values.config.aes.secret.key }}
{{- else -}}
aes-key
{{- end }}
{{- end }}

{{/*
Renders a value that contains a template.
Usage:
{{ include "artifact-registry.renderTemplate" ( dict "value" .Values.path.to.the.Value "context" $) }}
*/}}
{{- define "artifact-registry.renderTemplate" -}}
{{- if typeIs "string" .value }}
    {{- tpl .value .context }}
{{- else }}
    {{- tpl (.value | toYaml) .context }}
{{- end }}
{{- end -}}
