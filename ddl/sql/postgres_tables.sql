SELECT
    table_schema
    ,table_name
    ,'' AS sql
FROM
    information_schema.tables
WHERE
    table_type = 'BASE TABLE'
    AND table_schema <> 'information_schema'
    AND table_schema NOT LIKE 'pg_%'
;
