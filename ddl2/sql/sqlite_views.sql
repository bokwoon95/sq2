SELECT
    '' AS view_schema
    ,name AS view_name
    ,FALSE AS is_materialized
    ,sql
FROM
    sqlite_schema
WHERE
    "type" = 'view'
;
