SELECT
    table_namespace.nspname AS table_schema
    ,table_info.relname AS table_name
    ,column_info.attname AS column_name
    ,UPPER(format_type(column_info.atttypid, column_info.atttypmod)) AS column_type
    -- https://stackoverflow.com/a/3351120 precision and scale calculation
    ,CASE column_info.atttypid
        WHEN 21 /*int2*/ THEN 16
        WHEN 23 /*int4*/ THEN 32
        WHEN 20 /*int8*/ THEN 64
        WHEN 1700 /*numeric*/ THEN
            CASE
                WHEN column_info.atttypmod = -1 THEN 0
                ELSE ((column_info.atttypmod - 4) >> 16) & 65535
            END
        WHEN 700 /*float4*/ THEN 24 /*FLT_MANT_DIG*/
        WHEN 701 /*float8*/ THEN 53 /*DBL_MANT_DIG*/
        ELSE 0
    END AS numeric_precision
    ,CASE
        WHEN column_info.atttypid IN (21, 23, 20) THEN 0
        WHEN column_info.atttypid IN (1700) THEN
            CASE
                WHEN column_info.atttypmod = -1 THEN 0
                ELSE (column_info.atttypmod - 4) & 65535
            END
        ELSE 0
    END AS numeric_scale
    ,CASE column_info.attidentity
        WHEN 'd' THEN 'BY DEFAULT AS IDENTITY'
        WHEN 'a' THEN 'ALWAYS AS IDENTITY'
        ELSE ''
    END AS identity
    ,column_info.attnotnull AS is_notnull
    ,CASE column_info.attgenerated
        WHEN 's' THEN COALESCE(pg_get_expr(pg_attrdef.adbin, table_info.oid, TRUE), '')
        ELSE ''
    END AS generated_expr
    ,COALESCE(column_info.attgenerated = 's', FALSE) AS generated_expr_stored
    ,CASE pg_collation.collname
        WHEN 'default' THEN ''
        ELSE COALESCE(pg_collation.collname, '')
    END AS collation_name
    ,CASE column_info.attgenerated
        WHEN 's' THEN ''
        ELSE COALESCE(pg_get_expr(pg_attrdef.adbin, table_info.oid, TRUE), '')
    END AS column_default
FROM
    pg_attribute AS column_info
    JOIN pg_class AS table_info ON table_info.relkind = 'r' AND table_info.oid = column_info.attrelid
    JOIN pg_namespace AS table_namespace ON table_namespace.oid = table_info.relnamespace
    LEFT JOIN pg_attrdef ON pg_attrdef.adrelid = table_info.oid AND pg_attrdef.adnum = column_info.attnum
    LEFT JOIN pg_collation ON pg_collation.oid = column_info.attcollation
WHERE
    column_info.attnum > 0
    AND NOT column_info.attisdropped
    {{ if not .IncludeSystemCatalogs }}AND table_namespace.nspname <> 'information_schema' AND table_namespace.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND table_namespace.nspname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND table_namespace.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND table_info.relname IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND table_info.relname NOT IN ({{ printList .WithTables }}){{ end }}
{{- if .SortOutput }}
ORDER BY
    table_namespace.nspname
    ,table_info.relname
    ,column_info.attnum
{{- end }}
;
