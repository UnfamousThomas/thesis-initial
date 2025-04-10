{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "crd-chart.operatorFullname" -}}
{{- $releaseName := .Values.operator.releaseName }}
{{- $nameOverride := .Values.operator.nameOverride }}
{{- $fullnameOverride := .Values.operator.fullnameOverride }}

{{- if $fullnameOverride }}
{{- $fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default "thesis-operator" $nameOverride }}
{{- if contains $name $releaseName }}
{{- $releaseName | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" $releaseName $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}
