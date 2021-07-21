WITH generated_columns AS (
    SELECT txi.name, txi."type"
    FROM pragma_table_xinfo($1) AS txi
    LEFT JOIN pragma_table_info($1) AS ti ON ti.name = txi.name
    WHERE ti.name IS NULL AND txi."type" like '% GENERATED ALWAYS'
)
SELECT
    '' AS table_schema
    ,$1 AS table_name
    ,ti.name AS column_name
    ,COALESCE(gc."type", ti."type") AS column_type_1
    ,'' AS column_type_2
    ,0 AS numeric_precision
    ,0 AS numeric_scale
    ,FALSE AS is_autoincrement
    ,'' AS identity
    ,ti."notnull" AS is_notnull
    ,FALSE AS on_update_current_timestamp
    ,'' AS generated_expr
    ,FALSE AS generated_expr_stored
    ,'' AS collation_name
    ,COALESCE(ti.dflt_value, '') AS column_default
FROM
    pragma_table_xinfo($1) AS ti
    LEFT JOIN generated_columns AS gc ON gc.name = ti.name
;
