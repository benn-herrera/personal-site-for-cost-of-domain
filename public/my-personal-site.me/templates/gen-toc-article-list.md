::: post-list
{{- range .}}
::: post-item
[{{.Date}}]{.post-date}

[{{.Display}}](/writing/{{.Slug}}/){.post-title}

::: post-tags
{{range .Tags}}[{{.}}]{.tag} {{end}}
:::

:::
<!-- /post-item -->
{{- end}}
:::
<!-- /post-list -->
