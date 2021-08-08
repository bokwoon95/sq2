SELECT
    pg_namespace.nspname AS function_schema
    ,pg_proc.proname AS function_name
    ,pg_catalog.pg_get_functiondef(pg_proc.oid) AS sql
    ,pg_catalog.pg_get_function_arguments(pg_proc.oid) AS raw_args
    ,pg_catalog.pg_get_function_result(pg_proc.oid) AS return_type
FROM
    pg_catalog.pg_proc
    JOIN pg_catalog.pg_namespace ON pg_proc.pronamespace = pg_namespace.oid
WHERE
    pg_catalog.pg_function_is_visible(pg_proc.oid)
    AND NOT EXISTS (
        SELECT
            *
        FROM
            pg_catalog.pg_extension
            JOIN pg_catalog.pg_depend ON pg_depend.refobjid = pg_extension.oid
            JOIN pg_catalog.pg_proc AS extension_proc ON extension_proc.oid = pg_depend.objid
        WHERE
            pg_depend.deptype = 'e'
            AND extension_proc.oid = pg_proc.oid
    )
    {{ if not .IncludeSystemObjects }}AND pg_namespace.nspname <> 'information_schema' AND pg_namespace.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND pg_namespace.nspname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND pg_namespace.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithFunctions }}AND pg_proc.proname IN ({{ printList .WithFunctions }}){{ end }}
    {{ if .WithoutFunctions }}AND pg_proc.proname NOT IN ({{ printList .WithoutFunctions }}){{ end }}
{{- if .SortOutput }}
ORDER BY
    pg_namespace.nspname
    ,pg_proc.proname
{{- end }}
;
