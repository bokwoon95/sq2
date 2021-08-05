SELECT
    table_schema
    ,table_name
FROM
    information_schema.tables
WHERE
    table_type = 'BASE TABLE'
    {{ if not .IncludeSystemObjects }}AND table_schema <> 'information_schema' AND table_schema NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND table_schema IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND table_schema NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND table_name IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND table_name NOT IN ({{ printList .WithoutTables }}){{ end }}
;
