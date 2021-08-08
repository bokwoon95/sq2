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
    {{ if not .IncludeSystemCatalogs }}AND event_object_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys'){{ end }}
    {{ if .WithSchemas }}AND event_object_schema IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND event_object_schema NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND event_object_table IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND event_object_table NOT IN ({{ printList .WithoutTables }}){{ end }}
{{- if .SortOutput }}
ORDER BY
    table_schema
    ,table_name
    ,trigger_name
{{- end }}
;
