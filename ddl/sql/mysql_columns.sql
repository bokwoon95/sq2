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
    columns.table_schema NOT IN ('mysql', 'performance_schema', 'sys')
    AND tables.table_type = 'BASE TABLE'
