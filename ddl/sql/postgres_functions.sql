SELECT
    pg_namespace.nspname AS function_schema
    ,pg_proc.proname AS function_name
    ,pg_get_functiondef(pg_proc.oid) AS sql
    ,pg_get_function_arguments(pg_proc.oid) AS raw_args
FROM
    pg_proc
    JOIN pg_namespace on pg_proc.pronamespace = pg_namespace.oid
WHERE
    pg_namespace.nspname not in ('pg_catalog', 'information_schema')
;
