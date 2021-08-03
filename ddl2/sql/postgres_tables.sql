SELECT
    table_schema
    ,table_name
FROM
    information_schema.tables
WHERE
    table_type = 'BASE TABLE'
    {{ if not .IncludeSystemSchemas }}AND table_schema <> 'information_schema' AND table_schema NOT LIKE 'pg_%'{{ end }}
    {{ if .IncludedSchemas }}AND table_schema IN ({{ listify .IncludedSchemas }}){{ end }}
    {{ if .ExcludedSchemas }}AND table_schema NOT IN ({{ listify .ExcludedSchemas }}){{ end }}
    {{ if .IncludedTables }}AND table_name IN ({{ listify .IncludedTables }}){{ end }}
    {{ if .ExcludedTables }}AND table_name NOT IN ({{ listify .ExcludedTables }}){{ end }}
;
