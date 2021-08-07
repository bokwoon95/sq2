SELECT
    extname AS extension_name
    ,extversion AS extension_version
FROM
    pg_extension
{{- if .SortOutput }}
ORDER BY
    extname
{{- end }}
;
