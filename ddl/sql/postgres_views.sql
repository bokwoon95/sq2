SELECT
    view_schema
    ,view_name
    ,is_materialized
    ,sql
FROM (
SELECT
    schemaname AS view_schema
    ,viewname AS view_name
    ,FALSE AS is_materialized
    ,pg_get_viewdef(to_regclass(schemaname||'.'||viewname), true) AS sql
FROM
    pg_catalog.pg_views
UNION ALL
SELECT
    schemaname AS view_schema
    ,matviewname AS view_name
    ,TRUE AS is_materialized
    ,pg_get_viewdef(to_regclass(schemaname||'.'||matviewname), true) AS sql
FROM
    pg_catalog.pg_matviews
) AS tmp
WHERE
    view_schema <> 'information_schema'
    AND view_schema NOT LIKE 'pg_%'
;
