SELECT
    table_namespace.nspname AS table_schema
    ,table_info.relname AS table_name
    ,pg_trigger.tgname AS trigger_name
    ,COALESCE(pg_get_triggerdef(pg_trigger.oid, TRUE) || ';', '') AS sql
    ,'' AS action_timing
    ,'' AS event_manipulation
FROM
    pg_catalog.pg_trigger
    JOIN pg_catalog.pg_class AS table_info ON table_info.oid = pg_trigger.tgrelid
    JOIN pg_catalog.pg_namespace AS table_namespace ON table_namespace.oid = table_info.relnamespace
WHERE
    NOT pg_trigger.tgisinternal
    AND table_namespace.nspname <> 'information_schema'
    AND table_namespace.nspname NOT LIKE 'pg_%'
;
