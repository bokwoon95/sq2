SELECT
    schemas.nspname AS table_schema
    ,tables.relname AS table_name
    ,columns.attname AS column_name
    ,UPPER(format_type(columns.atttypid, columns.atttypmod)) AS column_type
    -- https://stackoverflow.com/a/3351120 precision and scale calculation
    ,CASE columns.atttypid
        WHEN 21 /*int2*/ THEN 16
        WHEN 23 /*int4*/ THEN 32
        WHEN 20 /*int8*/ THEN 64
        WHEN 1700 /*numeric*/ THEN
            CASE
                WHEN columns.atttypmod = -1 THEN 0
                ELSE ((columns.atttypmod - 4) >> 16) & 65535
            END
        WHEN 700 /*float4*/ THEN 24 /*FLT_MANT_DIG*/
        WHEN 701 /*float8*/ THEN 53 /*DBL_MANT_DIG*/
        ELSE 0
    END AS numeric_precision
    ,CASE
        WHEN columns.atttypid IN (21, 23, 20) THEN 0
        WHEN columns.atttypid IN (1700) THEN
            CASE
                WHEN columns.atttypmod = -1 THEN 0
                ELSE (columns.atttypmod - 4) & 65535
            END
        ELSE 0
    END AS numeric_scale
    ,CASE columns.attidentity
        WHEN 'd' THEN 'BY DEFAULT AS IDENTITY'
        WHEN 'a' THEN 'ALWAYS AS IDENTITY'
        ELSE ''
    END AS identity
    ,columns.attnotnull AS is_notnull
    ,CASE columns.attgenerated
        WHEN 's' THEN COALESCE(pg_get_expr(pg_attrdef.adbin, pg_attrdef.adrelid, TRUE), '')
        ELSE ''
    END AS generated_expr
    ,COALESCE(columns.attgenerated = 's', FALSE) AS generated_expr_stored
    ,CASE pg_collation.collname
        WHEN 'default' THEN ''
        ELSE COALESCE(pg_collation.collname, '')
    END AS collation_name
    ,CASE columns.attgenerated
        WHEN 's' THEN ''
        ELSE COALESCE(pg_get_expr(pg_attrdef.adbin, pg_attrdef.adrelid, TRUE), '')
    END AS column_default
    ,COALESCE(comments.description, '') AS column_comment
FROM
    pg_attribute AS columns
    JOIN pg_class AS tables ON tables.relkind = 'r' AND tables.oid = columns.attrelid
    JOIN pg_namespace AS schemas ON schemas.oid = tables.relnamespace
    LEFT JOIN pg_attrdef ON pg_attrdef.adrelid = tables.oid AND pg_attrdef.adnum = columns.attnum
    LEFT JOIN pg_collation ON pg_collation.oid = columns.attcollation
    LEFT JOIN pg_description AS comments ON comments.objoid = tables.oid AND comments.objsubid = columns.attnum
WHERE
    columns.attnum > 0
    AND NOT columns.attisdropped
    {{ if not .IncludeSystemCatalogs }}AND schemas.nspname <> 'information_schema' AND schemas.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND schemas.nspname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND schemas.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND tables.relname IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND tables.relname NOT IN ({{ printList .WithTables }}){{ end }}
{{- if .SortOutput }}
ORDER BY
    schemas.nspname
    ,tables.relname
    ,columns.attnum
{{- end }}
;
