SELECT
    c.table_schema
    ,c.table_name
    ,c.column_name
    ,UPPER(c.column_type) AS column_type
    ,COALESCE(c.numeric_precision, 0) AS numeric_precision
    ,COALESCE(c.numeric_scale, 0) AS numeric_scale
    ,COALESCE(c.extra = 'auto_increment', FALSE) AS is_autoincrement
    ,NOT c.is_nullable AS is_notnull
    ,COALESCE(c.extra = 'DEFAULT_GENERATED on update CURRENT_TIMESTAMP', FALSE) AS on_update_current_timestamp
    ,COALESCE(c.generation_expression, '') AS generated_expr
    ,COALESCE(c.extra = 'STORED GENERATED', FALSE) AS generated_expr_stored
    ,COALESCE(c.collation_name, '') AS collation_name
    ,COALESCE(c.column_default, '') AS column_default
    ,c.column_comment
FROM
    information_schema.columns AS c
    JOIN information_schema.tables AS t USING (table_schema, table_name)
WHERE
    t.table_type = 'BASE TABLE'
    {{ if not .IncludeSystemCatalogs }}AND c.table_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys'){{ end }}
    {{ if .WithSchemas }}AND c.table_schema IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND c.table_schema NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND c.table_name IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND c.table_name NOT IN ({{ printList .WithoutTables }}){{ end }}
{{- if .SortOutput }}
ORDER BY
    c.table_schema
    ,c.table_name
    ,c.ordinal_position
{{- end }}
;
