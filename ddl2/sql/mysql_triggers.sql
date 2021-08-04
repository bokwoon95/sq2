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
    TRUE
    {{- if not .IncludeSystemObjects }}AND event_object_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys'){{ end }}
    {{- if .WithSchemas }}AND event_object_schema IN ({{ listify .WithSchemas }}){{ end }}
    {{- if .WithoutSchemas }}AND event_object_schema NOT IN ({{ listify .WithoutSchemas }}){{ end }}
    {{- if .WithTables }}AND event_object_table IN ({{ listify .WithTables }}){{ end }}
    {{- if .WithoutTables }}AND event_object_table NOT IN ({{ listify .WithoutTables }}){{ end }}
;
