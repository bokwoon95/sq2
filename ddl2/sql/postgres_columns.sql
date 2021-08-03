SELECT
    c.table_schema
    ,c.table_name
    ,c.column_name
    ,c.data_type AS column_type_1
    ,c.udt_name AS column_type_2
    ,COALESCE(c.numeric_precision, 0) AS numeric_precision
    ,COALESCE(c.numeric_scale, 0) AS numeric_scale
    ,COALESCE(c.identity_generation::TEXT, '') AS is_identity
    ,NOT c.is_nullable::BOOLEAN AS is_notnull
    ,COALESCE(c.generation_expression, '') AS generated_expr
    ,CASE c.is_generated WHEN 'ALWAYS' THEN TRUE ELSE FALSE END AS generated_expr_stored
    ,COALESCE(c.collation_name, '') AS collation_name
    ,COALESCE(c.column_default, '') AS column_default
FROM
    information_schema.columns AS c
    JOIN information_schema.tables AS t USING (table_schema, table_name)
WHERE
    t.table_type = 'BASE TABLE'
    {{ if not .IncludeSystemSchemas }}AND columns.table_schema <> 'information_schema' AND columns.table_schema NOT LIKE 'pg_%'{{ end }}
    {{ if .IncludedSchemas }}AND c.table_schema IN ({{ listify .IncludedSchemas }}){{ end }}
    {{ if .ExcludedSchemas }}AND c.table_schema NOT IN ({{ listify .ExcludedSchemas }}){{ end }}
    {{ if .IncludedTables }}AND c.table_name IN ({{ listify .IncludedTables }}){{ end }}
    {{ if .ExcludedTables }}AND c.table_name NOT IN ({{ listify .IncludedTables }}){{ end }}
;
