SELECT
    table_schema
    ,table_name
    ,column_name
    ,data_type AS column_type_1
    ,column_type AS column_type_2
    ,COALESCE(numeric_precision, 0) AS numeric_precision
    ,COALESCE(numeric_scale, 0) AS numeric_scale
    ,extra = 'auto_increment' AS is_autoincrement
    ,'' AS identity
    ,NOT is_nullable AS is_notnull
    ,extra = 'DEFAULT_GENERATED on update CURRENT_TIMESTAMP' AS on_update_current_timestamp
    ,COALESCE(generation_expression, '') AS generated_expr
    ,CASE extra WHEN 'STORED GENERATED' THEN TRUE ELSE FALSE END AS generated_expr_stored
    ,COALESCE(collation_name, '') AS collation_name
    ,COALESCE(column_default, '') AS column_default
FROM
    information_schema.columns
    JOIN information_schema.tables USING (table_schema, table_name)
WHERE
    tables.table_type = 'BASE TABLE'
    {{- if not .IncludeSystemSchemas }}
    AND columns.table_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys')
    {{- end }}
    {{- if not .CustomPredicate }}
    AND {{ .CustomPredicate }}
    {{- end }}
    -- user provides something like "{schema} NOT IN ({1}, {2})", "schema_migrations", "schema_versions"
    -- alternatively: "{schema} NOT IN ({})", []string{"schema_migrations", "schema_versions"}
    -- sq.Param("schema", sq.Literal("columns.table_schema")) will be appended
    -- to the end of the args so that it doesn't mess with any ordinal params.
    -- available params: {tableSchema}, {tableName}, {columnName}, {columnType1}, {columnType2}

-- MySQLColumns()
