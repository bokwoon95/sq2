SELECT
    schemas.nspname AS function_schema
    ,proc1.proname AS function_name
    ,pg_get_functiondef(proc1.oid) AS sql
    ,pg_get_function_arguments(proc1.oid) AS raw_args
    ,pg_get_function_result(proc1.oid) AS return_type
FROM
    pg_proc AS proc1
    JOIN pg_namespace AS schemas ON schemas.oid = proc1.pronamespace
WHERE
    pg_function_is_visible(proc1.oid)
    AND NOT EXISTS (
        SELECT
            *
        FROM
            pg_extension
            JOIN pg_depend ON pg_depend.refobjid = pg_extension.oid
            JOIN pg_proc AS proc2 ON proc2.oid = pg_depend.objid
        WHERE
            pg_depend.deptype = 'e'
            AND proc1.oid = proc2.oid
    )
    {{ if not .IncludeSystemCatalogs }}AND schemas.nspname <> 'information_schema' AND schemas.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND schemas.nspname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND schemas.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithFunctions }}AND proc1.proname IN ({{ printList .WithFunctions }}){{ end }}
    {{ if .WithoutFunctions }}AND proc1.proname NOT IN ({{ printList .WithoutFunctions }}){{ end }}
{{- if .SortOutput }}
ORDER BY
    schemas.nspname
    ,proc1.proname
{{- end }}
;
