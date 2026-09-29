{{- define "rc-admission-webhook.fullname" -}}
{{ .Release.Name }}-rc-admission-webhook
{{- end -}}

{{- define "rc-admission-webhook.labels" -}}
app.kubernetes.io/name: rc-admission-webhook
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "rc-admission-webhook.selectorLabels" -}}
control-plane: controller-manager
app.kubernetes.io/name: rc-admission-webhook
{{- end -}}

{{- define "rc-admission-webhook.serviceName" -}}
{{ include "rc-admission-webhook.fullname" . }}-webhook
{{- end -}}
