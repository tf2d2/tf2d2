{{- range .Shapes }}
{{- if (and (ne .ID "") (ne .Label "")) }}
{{ .ID }}: {
  label: {{ .Label }}
  {{- if ne (.Icon | urlToString) "" }}
  shape: image
  icon: {{ .Icon | urlToString }}
  {{- end}}
  style: {
    opacity: {{ .Opacity}}
    stroke-dash: {{ .StrokeDash }}
    stroke-width: {{ .StrokeWidth }}
    bold: {{ .Text.Bold }}
  }
}
{{- end }}
{{- end }}

{{- range .Connections }}
{{ .Src }} -> {{ .Dst }}
{{- end }}
