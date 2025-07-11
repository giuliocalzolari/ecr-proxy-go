{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "ecr-proxy.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "ecr-proxy.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ecr-proxy.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{/*
Common labels
*/}}
{{- define "ecr-proxy.labels" -}}
app.kubernetes.io/name: {{ include "ecr-proxy.name" . }}
helm.sh/chart: {{ include "ecr-proxy.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{ with .Values.extraLabels }}
{{ toYaml . }}
{{- end }}
{{- end -}}



{{/*
Create the name of the service account to use
*/}}
{{- define "ecr-proxy.serviceAccountName" -}}
{{ if .Values.serviceAccount.create -}}
{{- coalesce .Values.serviceAccount.name "default" -}}
{{- else -}}
{{- "default" -}}
{{- end -}}
{{- end -}}




{{/*
Lookup a secret value by name and key.
{{ include "ecr-proxy.secretLookup" ( dict  "src" .Values.password "default" (randAlphaNum 10) "secretName" "your-secret-name" "key" "your-secret-key" "ns" "your-ns" ) }}
*/}}
{{- define "ecr-proxy.secretLookup" -}}
{{- $sec := .src | default .default -}}
{{- $existingSecret := (lookup "v1" "Secret" .ns .secretName ) -}}
{{- if $existingSecret  -}}
    {{- if (hasKey $existingSecret.data .key)  -}}
        {{- $sec1 := index $existingSecret.data .key | b64dec -}}
        {{/* "(empty .src)" is when the user supply manually the input */}}
        {{- if (empty .src) -}}
            {{- $sec = $sec1 -}}
        {{- end -}}
    {{- end -}}
{{- end -}}
{{- $sec -}}
{{- end -}}
