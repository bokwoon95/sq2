SELECT
    table_schema
    ,table_name
FROM
    information_schema.tables
WHERE
    table_type = 'BASE TABLE'
    {{ if not .IncludeSystemCatalogs }}AND table_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys'){{ end }}
    {{ if .WithSchemas }}AND table_schema IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND table_schema NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND table_name IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND table_name NOT IN ({{ printList .WithoutTables }}){{ end }}
{{- if .SortOutput }}
ORDER BY
    table_schema
    ,table_name
{{- end }}
;
