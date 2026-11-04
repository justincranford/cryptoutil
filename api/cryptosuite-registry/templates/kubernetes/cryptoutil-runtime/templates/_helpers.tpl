{{- define "cryptoutil-runtime.labels" -}}
app.kubernetes.io/managed-by: Helm
app.kubernetes.io/part-of: cryptoutil
cu.io/run: {{ .Values.run | quote }}
cu.io/release: {{ .Values.helmRelease | quote }}
{{- end -}}

{{- define "cryptoutil-runtime.selectorLabels" -}}
cu.io/run: {{ .root.Values.run | quote }}
cu.io/release: {{ .root.Values.helmRelease | quote }}
cu.io/psid: {{ .workload.psid | quote }}
cu.io/pool: {{ .workload.pool | quote }}
cu.io/workload: {{ .workload.name | quote }}
{{- end -}}
