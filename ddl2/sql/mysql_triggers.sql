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
    {{- if not .IncludeSystemSchemas }}AND event_object_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys'){{ end }}
    {{- if .IncludedSchemas }}AND event_object_schema IN ({{ listify .IncludedSchemas }}){{ end }}
    {{- if .ExcludedSchemas }}AND event_object_schema NOT IN ({{ listify .ExcludedSchemas }}){{ end }}
    {{- if .IncludedTables }}AND event_object_table IN ({{ listify .IncludedTables }}){{ end }}
    {{- if .ExcludedTables }}AND event_object_table NOT IN ({{ listify .ExcludedTables }}){{ end }}
;
