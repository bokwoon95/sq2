SELECT
    table_schema AS view_schema
    ,table_name AS view_name
    ,view_definition AS "sql"
FROM
    information_schema.views
WHERE
    TRUE
    {{ if not .IncludeSystemObjects }}AND table_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys'){{ end }}
    {{ if .WithSchemas }}AND table_schema IN ({{ listify .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND table_schema NOT IN ({{ listify .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND table_name IN ({{ listify .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND table_name NOT IN ({{ listify .WithoutTables }}){{ end }}
;
