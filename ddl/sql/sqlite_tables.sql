SELECT
    '' AS table_schema
    ,tbl_name AS table_name
    ,sql
FROM
    sqlite_schema
WHERE
    "type" = 'table'
    AND tbl_name <> 'sqlite_sequence'
    AND sql NOT LIKE 'CREATE TABLE ''%'
;
