SELECT
    c.table_schema
    ,c.table_name
    ,c.column_name
    ,c.data_type AS column_type_1
    ,c.column_type AS column_type_2
    ,COALESCE(c.numeric_precision, 0) AS numeric_precision
    ,COALESCE(c.numeric_scale, 0) AS numeric_scale
    ,c.extra = 'auto_increment' AS is_autoincrement
    ,NOT c.is_nullable AS is_notnull
    ,c.extra = 'DEFAULT_GENERATED on update CURRENT_TIMESTAMP' AS on_update_current_timestamp
    ,COALESCE(c.generation_expression, '') AS generated_expr
    ,CASE c.extra WHEN 'STORED GENERATED' THEN TRUE ELSE FALSE END AS generated_expr_stored
    ,COALESCE(c.collation_name, '') AS collation_name
    ,COALESCE(c.column_default, '') AS column_default
FROM
    information_schema.columns AS c
    JOIN information_schema.tables AS t USING (table_schema, table_name)
WHERE
    t.table_type = 'BASE TABLE'
    {{ if not .IncludeSystemSchemas }}AND c.table_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys'){{ end }}
    {{ if .IncludedSchemas }}AND c.table_schema IN ({{ listify .IncludedSchemas }}){{ end }}
    {{ if .ExcludedSchemas }}AND c.table_schema NOT IN ({{ listify .ExcludedSchemas }}){{ end }}
    {{ if .IncludedTables }}AND c.table_name IN ({{ listify .IncludedTables }}){{ end }}
    {{ if .ExcludedTables }}AND c.table_name NOT IN ({{ listify .ExcludedTables }}){{ end }}
;
