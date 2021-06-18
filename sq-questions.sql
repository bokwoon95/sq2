-- Get Tables

SELECT tbl_name AS table_name
FROM sqlite_schema
WHERE "type" = 'table';

-- Get Columns

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

-- Get Constraints

SELECT
    -- N.A.                           -- ConstraintSchema
    '' AS constraint_name             -- ConstraintName (needs sql parsing)
    ,'PRIMARY KEY' AS constraint_type -- ConstraintType
    -- N.A.                           -- TableSchema
    -- already known                  -- TableName
    ,group_concat(column_name) AS columns -- Columns
    -- N.A.                           -- ReferencesSchema
    ,NULL AS references_table         -- ReferencesTable
    ,NULL AS references_columns       -- ReferencesColumns
    ,NULL AS on_update                -- OnUpdate
    ,NULL AS on_delete                -- OnDelete
    -- N.A.                           -- Match
    -- N.A.                           -- CheckExpr
    -- N.A.                           -- IsDeferrable
    -- N.A.                           -- IsInitiallyDeferred
FROM (
    SELECT name AS column_name
    FROM pragma_table_info('customer')
    WHERE pk > 0
    ORDER BY pk
) AS primary_key_columns
UNION ALL
SELECT
    -- N.A.                           -- ConstraintSchema
    '' AS constraint_name             -- ConstraintName (needs sql parsing)
    ,'FOREIGN KEY' AS constraint_type -- ConstraintType
    -- N.A.                           -- TableSchema
    -- already known                  -- TableName
    ,group_concat(column_name) AS columns -- Columns
    -- N.A.                           -- ReferencesSchema
    ,references_table                 -- ReferencesTable
    ,group_concat(references_column) AS references_columns -- ReferencesColumns
    ,on_update                        -- OnUpdate
    ,on_delete                        -- OnDelete
    -- N.A.                           -- Match
    -- N.A.                           -- CheckExpr
    -- needs sql parsing              -- IsDeferrable
    -- needs sql parsing              -- IsInitiallyDeferred
FROM (
    SELECT
        id
        ,"from" AS column_name
        ,"table" AS references_table
        ,"to" AS references_column
        ,on_update
        ,on_delete
    FROM
        pragma_foreign_key_list('customer')
    ORDER BY
        seq
) AS foreign_key_columns
GROUP BY
    references_table
    ,on_update
    ,on_delete
UNION ALL
SELECT
    -- N.A.                           -- ConstraintSchema
    constraint_name                   -- ConstraintName
    ,'UNIQUE' AS constraint_type      -- ConstraintType
    -- N.A.                           -- TableSchema
    -- already known                  -- TableName
    ,group_concat(column_name) AS columns -- Columns
    -- N.A.                           -- ReferencesSchema
    ,NULL AS references_table         -- ReferencesTable
    ,NULL AS references_columns       -- ReferencesColumns
    ,NULL AS on_update                -- OnUpdate
    ,NULL AS on_delete                -- OnDelete
    -- N.A.                           -- Match
    -- N.A.                           -- CheckExpr
    -- N.A.                           -- IsDeferrable
    -- N.A.                           -- IsInitiallyDeferred
FROM (
    SELECT il.name AS constraint_name, ii.name AS column_name
    FROM pragma_index_list('customer') AS il CROSS JOIN pragma_index_info(il.name) AS ii
    WHERE il."unique" AND il.origin = 'u'
    ORDER BY ii.seqno
) AS unique_columns
GROUP BY
    constraint_name
;

-- Get Indices

WITH indexed_columns AS (
    SELECT
        'customer' AS table_name
        ,il.name AS index_name
        ,il."unique" AS is_unique
        ,il.partial AS is_partial
        ,CASE ii.cid WHEN -1 THEN '' WHEN -2 THEN '' ELSE ii.name END AS column_name
    FROM
        pragma_index_list('film_actor') AS il
        CROSS JOIN pragma_index_info(il.name) AS ii
    WHERE
        il.origin = 'c'
    ORDER BY
        il.name
        ,ii.seqno
)
SELECT
    -- N.A.                                   -- IndexSchema
    index_name                                -- IndexName
    -- N.A.                                   -- IndexType
    ,is_unique                                -- IsUnique
    -- N.A.                                   -- TableSchema
    ,table_name                               -- TableName
    ,json_group_array(column_name) AS columns -- Columns
    -- needs sql parsing                      -- Exprs
    -- N.A.                                   -- Include
    -- needs sql parsing                      -- Predicate
    ,is_partial                               -- Predicate.Valid
FROM
    indexed_columns
GROUP BY
    table_name
    ,index_name
    ,is_unique
    ,is_partial
;
