{{- define "cryptoutil-infra.labels" -}}
app.kubernetes.io/managed-by: Helm
app.kubernetes.io/part-of: cryptoutil
cu.io/run: {{ .root.Values.run | quote }}
cu.io/release: {{ .root.Values.helmRelease | quote }}
cu.io/component: {{ .component | quote }}
{{- end -}}

{{- define "cryptoutil-infra.selectorLabels" -}}
cu.io/run: {{ .root.Values.run | quote }}
cu.io/release: {{ .root.Values.helmRelease | quote }}
cu.io/component: {{ .component | quote }}
{{- end -}}

{{- define "cryptoutil-infra.probe" -}}
{{- if eq .type "http" }}
httpGet:
    path: {{ .path | quote }}
    port: {{ .port }}
    scheme: {{ default "HTTP" .scheme }}
{{- else }}
tcpSocket:
    port: {{ .port }}
{{- end }}
periodSeconds: {{ .periodSeconds }}
timeoutSeconds: {{ .timeoutSeconds }}
failureThreshold: {{ .failureThreshold }}
{{- end -}}
