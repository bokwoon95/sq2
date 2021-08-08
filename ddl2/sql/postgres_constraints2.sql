SELECT
    table_namespace.nspname AS table_schema
    ,table_info.relname AS table_name
    ,pg_constraint.conname AS constraint_name
    ,CASE pg_constraint.contype
        WHEN 'p' THEN 'PRIMARY KEY'
        WHEN 'u' THEN 'UNIQUE'
        WHEN 'f' THEN 'FOREIGN KEY'
        WHEN 'c' THEN 'CHECK'
        WHEN 'x' THEN 'EXCLUDE'
    END AS constraint_type
    ,COALESCE(referenced_table_namespace.nspname, '') AS references_schema
    ,COALESCE(referenced_table_info.relname, '') AS references_table
    ,pg_get_constraintdef(pg_constraint.oid, TRUE) AS expr
    ,pg_constraint.conkey
    ,pg_constraint.confkey
    ,pg_constraint.conpfeqop
    ,pg_constraint.conppeqop
    ,pg_constraint.conffeqop
    ,pg_constraint.conexclop
FROM
    pg_constraint
    JOIN pg_class AS table_info ON table_info.oid = pg_constraint.conrelid
    JOIN pg_namespace AS table_namespace ON table_namespace.oid = table_info.relnamespace
    LEFT JOIN pg_class AS referenced_table_info ON referenced_table_info.oid = pg_constraint.confrelid
    LEFT JOIN pg_namespace AS referenced_table_namespace ON referenced_table_namespace.oid = referenced_table_info.relnamespace
WHERE
    TRUE
    AND table_namespace.nspname <> 'information_schema' AND table_namespace.nspname NOT LIKE 'pg_%'
;

-- PRIMARY KEY, UNIQUE
SELECT
    table_schema
    ,table_name
    ,constraint_name
    ,constraint_type
    ,string_agg(column_name, ',' ORDER BY seq) AS columns
    ,'' AS exprs
    ,'' AS references_schema
    ,'' AS references_table
    ,'' AS references_columns
    ,'' AS update_rule
    ,'' AS delete_rule
    ,'' AS match_option
    ,'' AS check_expr
    ,'' AS operators
    ,'' AS index_type
    ,'' AS predicate
    ,is_deferrable
    ,is_initially_deferred
FROM (
    SELECT
        schemas.nspname AS table_schema
        ,tables.relname AS table_name
        ,pg_constraint.conname AS constraint_name
        ,CASE pg_constraint.contype
            WHEN 'p' THEN 'PRIMARY KEY'
            WHEN 'u' THEN 'UNIQUE'
        END AS constraint_type
        ,columns.attname AS column_name
        ,pg_constraint.condeferrable AS is_deferrable
        ,pg_constraint.condeferred AS is_initially_deferred
        ,c.seq
    FROM
        pg_constraint
        JOIN pg_class AS tables ON tables.oid = pg_constraint.conrelid
        JOIN pg_namespace AS schemas ON schemas.oid = tables.relnamespace
        CROSS JOIN unnest(pg_constraint.conkey) WITH ORDINALITY AS c(oid, seq)
        JOIN pg_attribute AS columns ON columns.attrelid = pg_constraint.conrelid AND columns.attnum = c.oid
    WHERE
        pg_constraint.contype IN ('p', 'u')
        {{ if not .IncludeSystemCatalogs }}AND schemas.nspname <> 'information_schema' AND schemas.nspname NOT LIKE 'pg_%'{{ end }}
        {{ if .WithSchemas }}AND schemas.nspname IN ({{ printList .WithSchemas }}){{ end }}
        {{ if .WithoutSchemas }}AND schemas.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
        {{ if .WithTables }}AND tables.relname IN ({{ printList .WithTables }}){{ end }}
        {{ if .WithoutTables }}AND tables.relname NOT IN ({{ printList .WithoutTables }}){{ end }}
) AS primary_key_unique_columns
GROUP BY
    table_schema
    ,table_name
    ,constraint_name
    ,constraint_type
    ,is_deferrable
    ,is_initially_deferred
;

-- FOREIGN KEY
SELECT
    table_schema
    ,table_name
    ,constraint_name
    ,constraint_type
    ,string_agg(column_name, ',' ORDER BY seq) AS columns
    ,'' AS exprs
    ,references_schema
    ,references_table
    ,string_agg(references_column, ',' ORDER BY seq) AS references_columns
    ,update_rule
    ,delete_rule
    ,match_option
    ,'' AS check_expr
    ,'' AS operators
    ,'' AS index_type
    ,'' AS predicate
    ,is_deferrable
    ,is_initially_deferred
FROM (
    SELECT
        schemas1.nspname AS table_schema
        ,tables1.relname AS table_name
        ,pg_constraint.conname AS constraint_name
        ,'FOREIGN KEY' constraint_type
        ,columns1.attname AS column_name
        ,schemas2.nspname AS references_schema
        ,tables2.relname AS references_table
        ,columns2.attname AS references_column
        ,CASE pg_constraint.confupdtype
            WHEN 'a' THEN 'NO ACTION'
            WHEN 'r' THEN 'RESTRICT'
            WHEN 'c' THEN 'CASCADE'
            WHEN 'n' THEN 'SET NULL'
            WHEN 'd' THEN 'SET DEFAULT'
        END AS update_rule
        ,CASE pg_constraint.confdeltype
            WHEN 'a' THEN 'NO ACTION'
            WHEN 'r' THEN 'RESTRICT'
            WHEN 'c' THEN 'CASCADE'
            WHEN 'n' THEN 'SET NULL'
            WHEN 'd' THEN 'SET DEFAULT'
        END AS delete_rule
        ,CASE pg_constraint.confmatchtype
            WHEN 'f' THEN 'MATCH FULL'
            WHEN 'p' THEN 'MATCH PARTIAL'
            ELSE ''
        END AS match_option
        ,pg_constraint.condeferrable AS is_deferrable
        ,pg_constraint.condeferred AS is_initially_deferred
        ,c1.seq
    FROM
        pg_constraint
        JOIN pg_class AS tables1 ON tables1.oid = pg_constraint.conrelid
        JOIN pg_class AS tables2 ON tables2.oid = pg_constraint.confrelid
        JOIN pg_namespace AS schemas1 ON schemas1.oid = tables1.relnamespace
        JOIN pg_namespace AS schemas2 ON schemas2.oid = tables2.relnamespace
        CROSS JOIN unnest(pg_constraint.conkey) WITH ORDINALITY AS c1(oid, seq)
        JOIN unnest(pg_constraint.confkey) WITH ORDINALITY AS c2(oid, seq) ON c2.seq = c1.seq
        JOIN pg_attribute AS columns1 ON columns1.attrelid = pg_constraint.conrelid AND columns1.attnum = c1.oid
        JOIN pg_attribute AS columns2 ON columns2.attrelid = pg_constraint.confrelid AND columns2.attnum = c2.oid
    WHERE
        pg_constraint.contype = 'f'
        {{ if not .IncludeSystemCatalogs }}AND schemas1.nspname <> 'information_schema' AND schemas1.nspname NOT LIKE 'pg_%'{{ end }}
        {{ if .WithSchemas }}AND schemas1.nspname IN ({{ printList .WithSchemas }}){{ end }}
        {{ if .WithoutSchemas }}AND schemas1.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
        {{ if .WithTables }}AND tables1.relname IN ({{ printList .WithTables }}){{ end }}
        {{ if .WithoutTables }}AND tables1.relname NOT IN ({{ printList .WithoutTables }}){{ end }}
) AS foreign_key_columns
GROUP BY
    table_schema
    ,table_name
    ,constraint_name
    ,constraint_type
    ,references_schema
    ,references_table
    ,update_rule
    ,delete_rule
    ,match_option
    ,is_deferrable
    ,is_initially_deferred
;

-- CHECK
SELECT
    schemas.nspname AS table_schema
    ,tables.relname AS table_name
    ,pg_constraint.conname AS constraint_name
    ,'CHECK' AS constraint_type
    ,'' AS columns
    ,'' AS exprs
    ,'' AS references_schema
    ,'' AS references_table
    ,'' AS references_columns
    ,'' AS update_rule
    ,'' AS delete_rule
    ,'' AS match_option
    ,pg_get_constraintdef(pg_constraint.oid, TRUE) AS check_expr
    ,'' AS operators
    ,'' AS index_type
    ,'' AS predicate
    ,pg_constraint.condeferrable AS is_deferrable
    ,pg_constraint.condeferred AS is_initially_deferred
FROM
    pg_constraint
    JOIN pg_class AS tables ON tables.oid = pg_constraint.conrelid
    JOIN pg_namespace AS schemas ON schemas.oid = tables.relnamespace
WHERE
    pg_constraint.contype = 'c'
    {{ if not .IncludeSystemCatalogs }}AND schemas.nspname <> 'information_schema' AND schemas.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND schemas.nspname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND schemas.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND tables.relname IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND tables.relname NOT IN ({{ printList .WithoutTables }}){{ end }}
;

-- EXCLUDE
SELECT
    schemas.nspname AS table_schema
    ,tables.relname AS table_name
    ,pg_constraint.conname AS constraint_name
    ,'CHECK' AS constraint_type
    ,'' AS columns
    ,'' AS exprs
    ,'' AS references_schema
    ,'' AS references_table
    ,'' AS references_columns
    ,'' AS update_rule
    ,'' AS delete_rule
    ,'' AS match_option
    ,pg_get_constraintdef(pg_constraint.oid, TRUE) AS check_expr
    ,'' AS operators
    ,'' AS index_type
    ,'' AS predicate
    ,pg_constraint.condeferrable AS is_deferrable
    ,pg_constraint.condeferred AS is_initially_deferred
FROM
    pg_constraint
    JOIN pg_class AS tables ON tables.oid = pg_constraint.conrelid
    JOIN pg_class AS indexes ON indexes.oid = pg_constraint.conindid
    JOIN pg_namespace AS schemas ON schemas.oid = tables.relnamespace
    JOIN pg_index ON pg_index.indexrelid = indexes.oid
    JOIN pg_am ON pg_am.oid = indexes.relam
    CROSS JOIN unnest(pg_index.indkey) WITH ORDINALITY AS c(oid, seq)
WHERE
    pg_constraint.contype = 'c'
    {{ if not .IncludeSystemCatalogs }}AND schemas.nspname <> 'information_schema' AND schemas.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND schemas.nspname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND schemas.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND tables.relname IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND tables.relname NOT IN ({{ printList .WithoutTables }}){{ end }}
;
