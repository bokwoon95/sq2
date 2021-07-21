SELECT
    table_schema
    ,table_name
    ,'' AS sql
FROM
    information_schema.tables
WHERE
    table_type = 'BASE TABLE'
    AND table_schema NOT IN ('mysql', 'performance_schema', 'sys')
