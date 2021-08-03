SELECT
    ,name AS view_name
    ,sql
FROM
    sqlite_schema
WHERE
    "type" = 'view'
;
