SELECT
    table_schema
    ,table_name
    ,column_name
    ,data_type AS column_type_1
    ,udt_name AS column_type_2
    ,COALESCE(numeric_precision, 0)
    ,COALESCE(numeric_scale, 0)
    ,COALESCE(identity_generation::TEXT, '') AS identity
    ,NOT is_nullable::BOOLEAN AS is_notnull
    ,COALESCE(generation_expression, '') AS generated_expr
    ,CASE is_generated WHEN 'ALWAYS' THEN TRUE ELSE FALSE END AS generated_expr_stored
    ,COALESCE(collation_name, '')
    ,COALESCE(column_default, '')
FROM
    information_schema.columns
    JOIN information_schema.tables USING (table_schema, table_name)
WHERE
    tables.table_type = 'BASE TABLE'
    AND table_schema NOT IN ('pg_catalog', 'information_schema')
;
