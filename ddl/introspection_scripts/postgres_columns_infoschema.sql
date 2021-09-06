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
    ,COALESCE(c.is_generated = 'ALWAYS', FALSE) AS generated_expr_stored
    ,COALESCE(c.collation_name, '') AS collation_name
    ,COALESCE(c.column_default, '') AS column_default
    ,'' AS column_comment
FROM
    information_schema.columns AS c
    JOIN information_schema.tables AS t USING (table_schema, table_name)
WHERE
    t.table_type = 'BASE TABLE'
    {{ if not .IncludeSystemCatalogs }}AND c.table_schema <> 'information_schema' AND c.table_schema NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND c.table_schema IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND c.table_schema NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND c.table_name IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND c.table_name NOT IN ({{ printList .WithTables }}){{ end }}
{{- if .SortOutput }}
ORDER BY
    c.table_schema
    ,c.table_name
    ,c.column_name
{{- end }}
;
