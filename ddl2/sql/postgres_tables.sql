SELECT
    table_namespace.nspname AS table_schema
    ,table_info.relname AS table_name
FROM
    pg_catalog.pg_namespace AS table_namespace
    JOIN pg_catalog.pg_class AS table_info ON table_info.relkind = 'r' AND table_info.relnamespace = table_namespace.oid 
WHERE
    TRUE
    {{ if not .IncludeSystemCatalogs }}AND table_namespace.nspname <> 'information_schema' AND table_namespace.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND table_namespace.nspname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND table_namespace.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND table_info.relname IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND table_info.relname NOT IN ({{ printList .WithoutTables }}){{ end }}
{{- if .SortOutput }}
ORDER BY
    table_namespace.nspname
    ,table_info.relname
{{- end }}
;
