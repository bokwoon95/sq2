WITH generated_columns AS (
    SELECT txi.name, txi."type"
    FROM pragma_table_xinfo('actor') AS txi
    LEFT JOIN pragma_table_info('actor') AS ti ON ti.name = txi.name
    WHERE ti.name IS NULL AND txi."type" like '% GENERATED ALWAYS'
)
SELECT
    -- N.A.                                        -- TableSchema
    -- already known                               -- TableName
    ti.name AS column_name                         -- ColumnName
    ,COALESCE(gc."type", ti."type") AS column_type -- ColumnType
    -- needs sql parsing                           -- IsAutoincrement
    -- N.A.                                        -- IsIdentity
    ,ti."notnull" AS is_notnull                    -- IsNotNull
    -- N.A.                                        -- OnUpdateCurrentTimestamp
    -- needs sql parsing                           -- GeneratedExpr
    ,gc.name IS NOT NULL AS is_generated           -- GeneratedExpr.Valid
    -- needs sql parsing                           -- GeneratedExprStored
    -- needs sql parsing                           -- CollationName
    ,ti.dflt_value AS column_default               -- ColumnDefault
FROM
    pragma_table_xinfo('actor') AS ti
    LEFT JOIN generated_columns AS gc ON gc.name = ti.name
;
