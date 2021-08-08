SELECT
    table_schema
    ,table_name
FROM
    pg_catalog.pg_class AS table_info
    JOIN pg_catalog.pg_namespace AS table_namespace ON table_namespace.oid = table_info.relnamespace
    JOIN pg_catalog.pg_attribute AS column_info
WHERE
    table_type = 'BASE TABLE'
    {{ if not .IncludeSystemCatalogs }}AND table_schema <> 'information_schema' AND table_schema NOT LIKE 'pg_%'{{ end }}
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
