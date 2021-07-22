SELECT
    event_object_schema AS table_schema
    ,event_object_table AS table_name
    ,trigger_name
    ,action_statement AS "sql"
    ,action_timing
    ,event_manipulation
FROM
    information_schema.triggers
WHERE
    event_object_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys')
;
