SELECT
    table_schema
    ,table_name
    ,column_name
    ,data_type AS column_type_1
    ,udt_name AS column_type_2
    ,COALESCE(numeric_precision, 0) AS numeric_precision
    ,COALESCE(numeric_scale, 0) AS numeric_scale
    ,FALSE AS is_autoincrement
    ,COALESCE(identity_generation::TEXT, '') AS is_identity
    ,NOT is_nullable::BOOLEAN AS is_notnull
    ,FALSE AS on_update_current_timestamp
    ,COALESCE(generation_expression, '') AS generated_expr
    ,CASE is_generated WHEN 'ALWAYS' THEN TRUE ELSE FALSE END AS generated_expr_stored
    ,COALESCE(collation_name, '') AS collation_name
    ,COALESCE(column_default, '') AS column_default
FROM
    information_schema.columns
    JOIN information_schema.tables USING (table_schema, table_name)
WHERE
    tables.table_type = 'BASE TABLE'
    AND table_schema NOT IN ('pg_catalog', 'information_schema')
;
