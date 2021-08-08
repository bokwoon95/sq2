SELECT
    schemas.nspname AS table_schema
    ,tables.relname AS table_name
FROM
    pg_class AS tables
    JOIN pg_namespace AS schemas ON schemas.oid = tables.relnamespace
WHERE
    tables.relkind = 'r'
    {{ if not .IncludeSystemCatalogs }}AND schemas.nspname <> 'information_schema' AND schemas.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND schemas.nspname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND schemas.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND tables.relname IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND tables.relname NOT IN ({{ printList .WithoutTables }}){{ end }}
{{- if .SortOutput }}
ORDER BY
    schemas.nspname
    ,tables.relname
{{- end }}
;
