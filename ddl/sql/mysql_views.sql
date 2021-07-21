SELECT
    table_schema AS view_schema
    ,table_name AS view_name
    ,FALSE AS is_materialized
    ,view_definition AS "sql"
FROM
    information_schema.views
WHERE
    table_schema NOT IN ('mysql', 'performance_schema', 'sys')
;
