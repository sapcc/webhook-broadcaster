{{/* vim: set filetype=gotexttmpl: */}}

{{/*
Expand the name of the chart.
*/}}
{{- define "webhook-broadcaster.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "webhook-broadcaster.fullname" -}}
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
webhook-broadcaster.app-labels prints the standard Helm labels to be applied to k8s
objects *indirectly* created by Helm (eg. to Pods created by
Deployments/StatefulSets, etc.).
*/}}
{{- define "webhook-broadcaster.app-labels" -}}
app: {{ include "webhook-broadcaster.name" . }}
release: {{ .Release.Name | quote }}
{{- include "webhook-broadcaster.additional-labels" . }}
{{- end -}}

{{/*
webhook-broadcaster.helm-labels prints the standard Helm labels to be applied to k8s
objects *directly* created by Helm (ie. NOT to Pods created by
Deployments/StatefulSets, etc.).
*/}}
{{- define "webhook-broadcaster.helm-labels" -}}
{{- include "webhook-broadcaster.app-labels" . }}
chart: {{ include "webhook-broadcaster.chart" . }}
heritage: {{ .Release.Service | quote }}
{{- end -}}

{{/*
webhook-broadcaster.additional-labels prints additional, custom labels to be applied on
all k8s objects created by directly or indirectly by Helm.
*/}}
{{- define "webhook-broadcaster.additional-labels" -}}
{{- range $key, $val := .Values.additionalLabels }}
{{ $key }}: {{ $val }}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "webhook-broadcaster.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}
