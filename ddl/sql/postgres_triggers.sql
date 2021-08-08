SELECT
    schemas.nspname AS table_schema
    ,tables.relname AS table_name
    ,pg_trigger.tgname AS trigger_name
    ,COALESCE(pg_get_triggerdef(pg_trigger.oid, TRUE) || ';', '') AS sql
FROM
    pg_trigger
    JOIN pg_class AS tables ON tables.oid = pg_trigger.tgrelid
    JOIN pg_namespace AS schemas ON schemas.oid = tables.relnamespace
WHERE
    NOT pg_trigger.tgisinternal
    {{ if not .IncludeSystemCatalogs }}AND schemas.nspname <> 'information_schema' AND schemas.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND schemas.nspname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND schemas.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND tables.relname IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND tables.relname NOT IN ({{ printList .WithoutTables }}){{ end }}
{{- if .SortOutput }}
ORDER BY
    schemas.nspname
    ,tables.relname
    ,pg_trigger.tgname
{{- end }}
;
