-- Get Tables

SELECT table_schema, table_name
FROM information_schema.tables
WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema');

-- Get Columns

SELECT
    table_schema                             -- TableSchema
    ,table_name                              -- TableName
    ,column_name                             -- ColumnName
    ,data_type AS column_type_1              -- ColumnType
    ,udt_name AS column_type_2               -- ColumnType
    ,COALESCE(numeric_precision, 0)          -- ColumnType
    ,COALESCE(numeric_scale, 0)              -- ColumnType
    -- N.A.                                  -- IsAutoincrement
    ,COALESCE(identity_generation::TEXT, '') AS identity -- Identity
    ,NOT is_nullable::BOOLEAN AS is_notnull  -- IsNotNull
    -- N.A.                                  -- OnUpdateCurrentTimestamp
    ,COALESCE(generation_expression, '') AS generated_expr -- GeneratedExpr
    ,CASE is_generated WHEN 'ALWAYS' THEN TRUE ELSE FALSE END AS generated_expr_stored -- GeneratedExprStored
    ,COALESCE(collation_name, '')            -- CollationName
    ,COALESCE(column_default, '')            -- ColumnDefault
FROM
    information_schema.columns
    JOIN information_schema.tables USING (table_schema, table_name)
WHERE
    tables.table_type = 'BASE TABLE'
    AND table_schema NOT IN ('pg_catalog', 'information_schema')
    AND table_name = 'actor'
;

-- Get Constraints

SELECT
    constraint_schema           -- ConstraintSchema
    ,constraint_name            -- ConstraintName
    ,constraint_type            -- ConstraintType
    ,table_schema               -- TableSchema
    ,table_name                 -- TableName
    ,string_agg(column_name,',') AS columns -- Columns
    ,NULL AS references_schema  -- ReferencesSchema
    ,NULL AS references_table   -- ReferencesTable
    ,NULL AS references_columns -- ReferencesColumns
    ,NULL AS on_update          -- OnUpdate
    ,NULL AS on_delete          -- OnDelete
    ,NULL AS match_option       -- MatchOption
    ,NULL AS check_expr         -- CheckExpr
    ,is_deferrable              -- IsDeferrable
    ,is_initially_deferred      -- IsInitiallyDeferred
FROM (
    SELECT
        tc.constraint_schema
        ,tc.constraint_name
        ,tc.constraint_type
        ,tc.table_schema
        ,tc.table_name
        ,kcu.column_name
        ,tc.is_deferrable::BOOLEAN
        ,tc.initially_deferred::BOOLEAN AS is_initially_deferred
    FROM
        information_schema.table_constraints AS tc
        JOIN information_schema.constraint_column_usage AS ccu USING (constraint_schema, constraint_name)
        JOIN information_schema.key_column_usage AS kcu USING (constraint_schema, constraint_name, column_name)
    WHERE
        tc.constraint_type IN ('PRIMARY KEY', 'UNIQUE')
        AND tc.table_schema = 'public'
        AND tc.table_name = 'customer'
    ORDER BY
        kcu.ordinal_position
) AS primary_key_unique_columns
GROUP BY
    constraint_schema
    ,constraint_name
    ,constraint_type
    ,table_schema
    ,table_name
    ,is_deferrable
    ,is_initially_deferred
UNION ALL
SELECT
    constraint_schema      -- ConstraintSchema
    ,constraint_name       -- ConstraintName
    ,constraint_type       -- ConstraintType
    ,table_schema          -- TableSchema
    ,table_name            -- TableName
    ,string_agg(column_name,',') AS columns -- ColumnName
    ,references_schema     -- ReferencesSchema
    ,references_table      -- ReferencesTable
    ,string_agg(references_column,',') AS references_columns -- ReferencesColumns
    ,on_update             -- OnUpdate
    ,on_delete             -- OnDelete
    ,match_option          -- MatchOption
    ,NULL AS check_expr    -- CheckExpr
    ,is_deferrable         -- IsDeferrable
    ,is_initially_deferred -- IsInitiallyDeferred
FROM (
    SELECT
        tc.constraint_schema
        ,tc.constraint_name
        ,tc.constraint_type
        ,tc.table_schema
        ,tc.table_name
        ,kcu.column_name
        ,ccu.table_schema AS references_schema
        ,ccu.table_name AS references_table
        ,ccu.column_name AS references_column
        ,rc.update_rule AS on_update
        ,rc.delete_rule AS on_delete
        ,rc.match_option
        ,tc.is_deferrable::BOOLEAN
        ,tc.initially_deferred::BOOLEAN AS is_initially_deferred
    FROM
        information_schema.table_constraints AS tc
        JOIN information_schema.constraint_column_usage AS ccu USING (constraint_schema, constraint_name)
        JOIN information_schema.key_column_usage AS kcu USING (constraint_schema, constraint_name, column_name)
        LEFT JOIN information_schema.referential_constraints AS rc USING (constraint_schema, constraint_name)
    WHERE
        tc.constraint_type = 'FOREIGN KEY'
        AND tc.table_schema = 'public'
        AND tc.table_name = 'customer'
    ORDER BY
        kcu.ordinal_position
) AS foreign_key_columns
GROUP BY
    constraint_schema
    ,constraint_name
    ,constraint_type
    ,table_schema
    ,table_name
    ,references_schema
    ,references_table
    ,on_update
    ,on_delete
    ,match_option
    ,is_deferrable
    ,is_initially_deferred
UNION ALL
SELECT
    constraint_namespace.nspname AS constraint_schema            -- ConstraintSchema
    ,pg_constraint.conname AS constraint_name                    -- ConstraintName
    ,'CHECK' AS constraint_type                                  -- ConstraintType
    ,table_namespace.nspname AS table_schema                     -- TableSchema
    ,pg_class.relname AS table_name                              -- TableName
    ,NULL AS columns                                             -- Columns
    ,NULL AS references_schema                                   -- ReferencesSchema
    ,NULL AS references_table                                    -- ReferencesTable
    ,NULL AS references_columns                                  -- ReferencesColumns
    ,NULL AS on_update                                           -- OnUpdate
    ,NULL AS on_delete                                           -- OnDelete
    ,NULL AS match_option                                        -- MatchOption
    ,SUBSTR(pg_get_constraintdef(pg_constraint.oid, TRUE), 7) AS check_expr -- CheckExpr
    ,pg_constraint.condeferrable AS is_deferrable                -- IsDeferrable
    ,pg_constraint.condeferred AS is_initially_deferred          -- IsInitiallyDeferred
FROM
    pg_catalog.pg_constraint
    JOIN pg_catalog.pg_class ON pg_class.oid = pg_constraint.conrelid
    JOIN pg_catalog.pg_namespace AS table_namespace ON table_namespace.oid = pg_class.relnamespace
    JOIN pg_catalog.pg_namespace AS constraint_namespace ON constraint_namespace.oid = pg_constraint.connamespace
WHERE
    pg_constraint.contype = 'c'
    AND table_namespace.nspname = 'public'
    AND pg_class.relname = 'customer'
;

-- Get Indices

WITH index_columns AS (
    SELECT
        index_namespace.nspname AS index_schema
        ,index_info.relname AS index_name
        ,UPPER(pg_am.amname) AS index_type
        ,pg_index.indisunique AS is_unique
        ,table_namespace.nspname AS table_schema
        ,table_info.relname AS table_name
        ,pg_index.indnkeyatts AS num_key_columns
        ,COALESCE(pg_attribute.attname, '') AS column_name
        ,pg_get_expr(pg_index.indexprs, pg_index.indrelid, true) AS exprs
        ,pg_get_expr(pg_index.indpred, pg_index.indrelid, true) AS predicate
        ,ROW_NUMBER() OVER () AS n
    FROM
        pg_catalog.pg_index
        JOIN pg_catalog.pg_class AS index_info ON index_info.oid = pg_index.indexrelid
        JOIN pg_catalog.pg_class AS table_info ON table_info.oid = pg_index.indrelid
        JOIN pg_catalog.pg_namespace AS index_namespace ON index_namespace.oid = index_info.relnamespace
        JOIN pg_catalog.pg_namespace AS table_namespace ON table_namespace.oid = table_info.relnamespace
        JOIN pg_catalog.pg_am ON pg_am.oid = index_info.relam
        CROSS JOIN unnest(pg_index.indkey) AS column_oid
        LEFT JOIN pg_catalog.pg_attribute ON
            pg_attribute.attrelid = pg_index.indrelid
            AND pg_attribute.attnum = column_oid.column_oid
    WHERE
        table_namespace.nspname = 'public'
        AND table_info.relname = 'rental'
)
SELECT
    index_schema     -- IndexSchema
    ,index_name      -- IndexName
    ,index_type      -- IndexType
    ,is_unique       -- IsUnique
    ,table_schema    -- TableSchema
    ,table_name      -- TableName
    ,num_key_columns -- Columns, Include
    ,string_agg(column_name, ',' ORDER BY n) AS columns -- Columns, Exprs, Include
    ,exprs           -- Exprs
    ,predicate       -- Predicate
FROM
    index_columns
GROUP BY
    index_schema
    ,index_name
    ,index_type
    ,is_unique
    ,table_schema
    ,table_name
    ,num_key_columns
    ,exprs
    ,predicate
;

WITH index_columns AS (
    SELECT
        index_namespace.nspname AS index_schema
        ,index_info.relname AS index_name
        ,UPPER(pg_am.amname) AS index_type
        ,pg_index.indisunique AS is_unique
        ,table_namespace.nspname AS table_schema
        ,table_info.relname AS table_name
        ,pg_index.indnkeyatts AS num_key_columns
        ,COALESCE(pg_attribute.attname, '') AS column_name
        ,pg_get_expr(pg_index.indexprs, pg_index.indrelid, true) AS exprs
        ,pg_get_expr(pg_index.indpred, pg_index.indrelid, true) AS predicate
        ,columns.seq_in_index
    FROM
        pg_catalog.pg_index
        JOIN pg_catalog.pg_class AS index_info ON index_info.oid = pg_index.indexrelid
        JOIN pg_catalog.pg_class AS table_info ON table_info.oid = pg_index.indrelid
        JOIN pg_catalog.pg_namespace AS index_namespace ON index_namespace.oid = index_info.relnamespace
        JOIN pg_catalog.pg_namespace AS table_namespace ON table_namespace.oid = table_info.relnamespace
        JOIN pg_catalog.pg_am ON pg_am.oid = index_info.relam
        CROSS JOIN LATERAL unnest(pg_index.indkey) WITH ORDINALITY AS columns(column_oid, seq_in_index)
        LEFT JOIN pg_catalog.pg_attribute ON
            pg_attribute.attrelid = pg_index.indrelid
            AND pg_attribute.attnum = columns.column_oid
    WHERE
        table_namespace.nspname = 'public'
        AND table_info.relname = 'rental'
)
SELECT
    index_schema     -- IndexSchema
    ,index_name      -- IndexName
    ,index_type      -- IndexType
    ,is_unique       -- IsUnique
    ,table_schema    -- TableSchema
    ,table_name      -- TableName
    ,num_key_columns -- Columns, Include
    ,string_agg(column_name, ',' ORDER BY index_schema, index_name, seq_in_index) AS columns -- Columns, Exprs, Include
    ,exprs           -- Exprs
    ,predicate       -- Predicate
FROM
    index_columns
GROUP BY
    index_schema
    ,index_name
    ,index_type
    ,is_unique
    ,table_schema
    ,table_name
    ,num_key_columns
    ,exprs
    ,predicate
;
