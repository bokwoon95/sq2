SELECT
    tbl_name AS table_name
    ,name AS trigger_name
    ,sql || ';' AS sql
FROM
    sqlite_schema
WHERE
    "type" = 'trigger'
;
