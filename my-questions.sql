-- Get tables

SELECT table_schema, table_name
FROM information_schema.tables
WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('mysql', 'performance_schema', 'sys');

-- Get columns

SELECT
    table_schema                                       -- TableSchema
    ,table_name                                        -- TableName
    ,column_name                                       -- ColumnName
    ,data_type AS column_type_1                        -- ColumnType
    ,column_type AS column_type_2                      -- ColumnType
    ,numeric_precision                                 -- ColumnType
    ,numeric_scale                                     -- ColumnType
    ,extra = 'auto_increment' AS is_autoincrement      -- IsAutoincrement
    -- N.A.                                            -- IsIdentity
    ,NOT is_nullable AS is_notnull                     -- IsNotNull
    ,extra = 'DEFAULT_GENERATED on update CURRENT_TIMESTAMP' AS on_update_current_timestamp -- OnUpdateCurrentTimestamp
    ,CASE generation_expression
        WHEN '' THEN NULL
        ELSE generation_expression
    END AS generated_expr -- GeneratedExpr
    ,CASE extra
        WHEN 'STORED GENERATED' THEN TRUE
        WHEN 'VIRTUAL GENERATED' THEN FALSE
        ELSE NULL
    END AS generated_expr_stored -- GeneratedExprStored
    ,collation_name -- Collation
    ,column_default -- ColumnDefault
FROM
    information_schema.columns
WHERE
    table_schema = 'db'
    AND table_name = 'actor'
;

-- Get constraints

SELECT
    tc.constraint_schema         -- ConstraintSchema
    ,tc.constraint_name          -- ConstraintName
    ,tc.constraint_type          -- ConstraintType
    ,tc.table_schema             -- TableSchema
    ,tc.table_name               -- TableName
    ,group_concat(kcu.column_name ORDER BY kcu.ordinal_position) AS columns -- Columns
    ,kcu.referenced_table_schema AS references_schema -- ReferencesSchema
    ,kcu.referenced_table_name AS references_table    -- ReferencesTable
    ,group_concat(kcu.referenced_column_name ORDER BY kcu.ordinal_position) AS references_columns -- ReferencesColumns
    ,rc.update_rule AS on_update -- OnUpdate
    ,rc.delete_rule AS on_delete -- Ondelete
    ,rc.match_option             -- MatchOption
    ,NULL AS check_expr          -- CheckExpr
    -- N.A.                      -- IsDeferrable
    -- N.A.                      -- IsInitiallyDeferred
FROM
    information_schema.table_constraints AS tc
    JOIN information_schema.key_column_usage AS kcu USING (constraint_schema, constraint_name, table_name)
    LEFT JOIN information_schema.referential_constraints AS rc USING (constraint_schema, constraint_name)
WHERE
    tc.constraint_type IN ('PRIMARY KEY', 'FOREIGN KEY', 'UNIQUE')
    AND tc.table_schema = 'db'
    AND tc.table_name = 'customer'
GROUP BY
    tc.constraint_schema
    ,tc.constraint_name
    ,tc.constraint_type
    ,tc.table_schema
    ,tc.table_name
    ,kcu.referenced_table_schema
    ,kcu.referenced_table_name
    ,rc.update_rule
    ,rc.delete_rule
    ,rc.match_option
UNION ALL
SELECT
    tc.constraint_schema        -- ConstraintSchema
    ,tc.constraint_name         -- ConstraintName
    ,tc.constraint_type         -- ConstraintType
    ,tc.table_schema            -- TableSchema
    ,tc.table_name              -- TableName
    ,NULL AS columns            -- Columns
    ,NULL AS references_schema  -- ReferencesSchema
    ,NULL AS references_table   -- ReferencesTable
    ,NULL AS references_columns -- ReferencesColumns
    ,NULL AS on_update          -- OnUpdate
    ,NULL AS on_delete          -- Ondelete
    ,NULL AS match_option       -- MatchOption
    ,check_clause AS check_expr -- CheckExpr
    -- N.A.                     -- IsDeferrable
    -- N.A.                     -- IsInitiallyDeferred
FROM
    information_schema.table_constraints AS tc
    JOIN information_schema.check_constraints AS cc USING (constraint_schema, constraint_name)
WHERE
    tc.table_schema = 'db'
    AND tc.table_name = 'customer'
;

-- Get indices

WITH indexed_columns AS (
    SELECT
        index_schema
        ,index_name
        ,index_type
        ,NOT non_unique AS is_unique
        ,table_schema
        ,table_name
        ,COALESCE(column_name, '') AS column_name
        ,COALESCE(expression, '') AS expr
        ,seq_in_index
    FROM
        information_schema.statistics
    WHERE
        statistics.index_name <> 'PRIMARY'
        AND statistics.table_schema = 'db'
        AND statistics.table_name = 'rental'
    ORDER BY
        index_name
        ,seq_in_index
)
SELECT
    index_schema  -- IndexSchema
    ,index_name   -- IndexName
    ,index_type   -- IndexType
    ,is_unique    -- IsUnique
    ,table_schema -- TableSchema
    ,table_name   -- TableName
    ,group_concat(column_name) AS columns -- Columns
    ,group_concat(expr) AS exprs          -- Exprs
FROM
    indexed_columns
GROUP BY
    index_schema
    ,index_name
    ,index_type
    ,is_unique
    ,table_schema
    ,table_name
;

-- Deterministic version that uses group_concat(... ORDER BY ...)
WITH indexed_columns AS (
    SELECT
        index_schema
        ,index_name
        ,index_type
        ,NOT non_unique AS is_unique
        ,table_schema
        ,table_name
        ,COALESCE(column_name, '') AS column_name
        ,COALESCE(expression, '') AS expr
        ,seq_in_index
    FROM
        information_schema.statistics
    WHERE
        statistics.index_name <> 'PRIMARY'
        AND statistics.table_schema = 'db'
        AND statistics.table_name = 'rental'
)
SELECT
    index_schema  -- IndexSchema
    ,index_name   -- IndexName
    ,index_type   -- IndexType
    ,is_unique    -- IsUnique
    ,table_schema -- TableSchema
    ,table_name   -- TableName
    ,group_concat(column_name ORDER BY seq_in_index) AS columns -- Columns
    ,group_concat(expr ORDER BY seq_in_index) AS exprs          -- Exprs
FROM
    indexed_columns
GROUP BY
    index_schema
    ,index_name
    ,index_type
    ,is_unique
    ,table_schema
    ,table_name
;
