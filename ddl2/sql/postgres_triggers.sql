SELECT
    table_namespace.nspname AS table_schema
    ,table_info.relname AS table_name
    ,pg_trigger.tgname AS trigger_name
    ,COALESCE(pg_get_triggerdef(pg_trigger.oid, TRUE) || ';', '') AS sql
FROM
    pg_trigger
    JOIN pg_class AS table_info ON table_info.oid = pg_trigger.tgrelid
    JOIN pg_namespace AS table_namespace ON table_namespace.oid = table_info.relnamespace
WHERE
    NOT pg_trigger.tgisinternal
    {{ if not .IncludeSystemCatalogs }}AND table_namespace.nspname <> 'information_schema' AND table_namespace.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND table_namespace.nspname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND table_namespace.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND table_info.relname IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND table_info.relname NOT IN ({{ printList .WithoutTables }}){{ end }}
{{- if .SortOutput }}
ORDER BY
    table_namespace.nspname
    ,table_info.relname
    ,pg_trigger.tgname
{{- end }}
;
