SELECT
    pg_namespace.nspname AS function_schema
    ,pg_proc.proname AS function_name
    ,pg_get_functiondef(pg_proc.oid) AS sql
    ,pg_get_function_arguments(pg_proc.oid) AS raw_args
    ,pg_type.typname AS return_type
FROM
    pg_catalog.pg_proc
    JOIN pg_catalog.pg_namespace ON pg_proc.pronamespace = pg_namespace.oid
    JOIN pg_catalog.pg_type ON pg_type.oid = pg_proc.prorettype
WHERE
    TRUE
    {{ if not .IncludeSystemObjects }}AND pg_namespace.nspname <> 'information_schema' AND pg_namespace.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND pg_namespace.nspname IN ({{ listify .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND pg_namespace.nspname NOT IN ({{ listify .WithoutSchemas }}){{ end }}
;
