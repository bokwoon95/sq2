SELECT
    *
FROM (
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
        {{ if not .IncludeSystemObjects }}AND tc.table_schema <> 'information_schema' AND tc.table_schema NOT LIKE 'pg_%'{{ end }}
        {{ if .WithSchemas }}AND tc.table_schema IN ({{ printList .WithSchemas }}){{ end }}
        {{ if .WithoutSchemas }}AND tc.table_schema NOT IN ({{ printList .WithoutSchemas }}){{ end }}
        {{ if .WithTables }}AND tc.table_name IN ({{ printList .WithTables }}){{ end }}
        {{ if .WithoutTables }}AND tc.table_name NOT IN ({{ printList .WithoutTables }}){{ end }}
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
        {{ if not .IncludeSystemObjects }}AND tc.table_schema <> 'information_schema' AND tc.table_schema NOT LIKE 'pg_%'{{ end }}
        {{ if .WithSchemas }}AND tc.table_schema IN ({{ printList .WithSchemas }}){{ end }}
        {{ if .WithoutSchemas }}AND tc.table_schema NOT IN ({{ printList .WithoutSchemas }}){{ end }}
        {{ if .WithTables }}AND tc.table_name IN ({{ printList .WithTables }}){{ end }}
        {{ if .WithoutTables }}AND tc.table_name NOT IN ({{ printList .WithoutTables }}){{ end }}
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
    pg_namespace.nspname AS table_schema
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
    ,pg_get_constraintdef(pg_constraint.oid, TRUE) AS check_expr
    ,'' AS operators
    ,'' AS index_type
    ,'' AS predicate
    ,pg_constraint.condeferrable AS is_deferrable
    ,pg_constraint.condeferred AS is_initially_deferred
FROM
    pg_catalog.pg_constraint
    JOIN pg_catalog.pg_class ON pg_class.oid = pg_constraint.conrelid
    JOIN pg_catalog.pg_namespace ON pg_namespace.oid = pg_class.relnamespace
WHERE
    pg_constraint.contype = 'c'
    {{ if not .IncludeSystemObjects }}AND pg_namespace.nspname <> 'information_schema' AND pg_namespace.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND pg_namespace.nspname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND pg_namespace.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND pg_class.relname IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND pg_class.relname NOT IN ({{ printList .WithoutTables }}){{ end }}
) AS tmp
{{- if .SortOutput }}
ORDER BY
    table_schema
    ,table_name
    ,constraint_name
{{- end }}
;
