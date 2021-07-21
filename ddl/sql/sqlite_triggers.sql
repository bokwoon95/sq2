SELECT
    '' AS table_schema
    ,tbl_name AS table_name
    ,name AS trigger_name
    ,sql || ';' AS sql
    ,'' AS action_timing
    ,'' AS event_manipulation
FROM
    sqlite_schema
WHERE
    "type" = 'trigger'
;
