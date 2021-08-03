SELECT
    table_schema
    ,table_name
    ,constraint_name
    ,constraint_type
    ,string_agg(column_name, ',' ORDER BY ordinal_position) AS columns
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
        tc.table_schema
        ,tc.table_name
        ,tc.constraint_name
        ,tc.constraint_type
        ,kcu.column_name
        ,tc.is_deferrable::BOOLEAN
        ,tc.initially_deferred::BOOLEAN AS is_initially_deferred
        ,kcu.ordinal_position
    FROM
        information_schema.table_constraints AS tc
        JOIN information_schema.constraint_column_usage AS ccu USING (constraint_schema, constraint_name)
        JOIN information_schema.key_column_usage AS kcu USING (constraint_schema, constraint_name, column_name)
    WHERE
        tc.constraint_type IN ('PRIMARY KEY', 'UNIQUE')
        {{ if not .IncludeSystemSchemas }}AND tc.table_schema <> 'information_schema' AND tc.table_schema NOT LIKE 'pg_%'{{ end }}
        {{ if .IncludedSchemas }}AND tc.table_schema IN ({{ listify .IncludedSchemas }}){{ end }}
        {{ if .ExcludedSchemas }}AND tc.table_schema NOT IN ({{ listify .ExcludedSchemas }}){{ end }}
        {{ if .IncludedTables }}AND tc.table_name IN ({{ listify .IncludedTables }}){{ end }}
        {{ if .ExcludedTables }}AND tc.table_name NOT IN ({{ listify .ExcludedTables }}){{ end }}
) AS primary_key_unique_columns
GROUP BY
    table_schema
    ,table_name
    ,constraint_name
    ,constraint_type
    ,is_deferrable
    ,is_initially_deferred
UNION ALL
SELECT
    table_schema
    ,table_name
    ,constraint_name
    ,constraint_type
    ,string_agg(column_name, ',' ORDER BY ordinal_position) AS columns
    ,'' AS exprs
    ,references_schema
    ,references_table
    ,string_agg(references_column, ',' ORDER BY ordinal_position) AS references_columns
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
        tc.constraint_schema
        ,tc.constraint_name
        ,tc.constraint_type
        ,tc.table_schema
        ,tc.table_name
        ,kcu.column_name
        ,ccu.table_schema AS references_schema
        ,ccu.table_name AS references_table
        ,ccu.column_name AS references_column
        ,rc.update_rule
        ,rc.delete_rule
        ,rc.match_option
        ,tc.is_deferrable::BOOLEAN
        ,tc.initially_deferred::BOOLEAN AS is_initially_deferred
        ,kcu.ordinal_position
    FROM
        information_schema.table_constraints AS tc
        JOIN information_schema.constraint_column_usage AS ccu USING (constraint_schema, constraint_name)
        JOIN information_schema.key_column_usage AS kcu USING (constraint_schema, constraint_name, column_name)
        LEFT JOIN information_schema.referential_constraints AS rc USING (constraint_schema, constraint_name)
    WHERE
        tc.constraint_type = 'FOREIGN KEY'
        {{ if not .IncludeSystemSchemas }}AND tc.table_schema <> 'information_schema' AND tc.table_schema NOT LIKE 'pg_%'{{ end }}
        {{ if .IncludedSchemas }}AND tc.table_schema IN ({{ listify .IncludedSchemas }}){{ end }}
        {{ if .ExcludedSchemas }}AND tc.table_schema NOT IN ({{ listify .ExcludedSchemas }}){{ end }}
        {{ if .IncludedTables }}AND tc.table_name IN ({{ listify .IncludedTables }}){{ end }}
        {{ if .ExcludedTables }}AND tc.table_name NOT IN ({{ listify .ExcludedTables }}){{ end }}
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
UNION ALL
SELECT
    table_namespace.nspname AS table_schema
    ,pg_class.relname AS table_name
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
    ,SUBSTR(pg_get_constraintdef(pg_constraint.oid, TRUE), 7) AS check_expr
    ,'' AS operators
    ,'' AS index_type
    ,'' AS predicate
    ,pg_constraint.condeferrable AS is_deferrable
    ,pg_constraint.condeferred AS is_initially_deferred
FROM
    pg_catalog.pg_constraint
    JOIN pg_catalog.pg_class ON pg_class.oid = pg_constraint.conrelid
    JOIN pg_catalog.pg_namespace AS table_namespace ON table_namespace.oid = pg_class.relnamespace
    JOIN pg_catalog.pg_namespace AS constraint_namespace ON constraint_namespace.oid = pg_constraint.connamespace
WHERE
    pg_constraint.contype = 'c'
    {{ if not .IncludeSystemSchemas }}AND table_namespace.nspname <> 'information_schema' AND table_namespace.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .IncludedSchemas }}AND table_namespace.nspname IN ({{ listify .IncludedSchemas }}){{ end }}
    {{ if .ExcludedSchemas }}AND table_namespace.nspname NOT IN ({{ listify .ExcludedSchemas }}){{ end }}
    {{ if .IncludedTables }}AND pg_class.relname IN ({{ listify .IncludedTables }}){{ end }}
    {{ if .ExcludedTables }}AND pg_class.relname NOT IN ({{ listify .ExcludedTables }}){{ end }}
;
