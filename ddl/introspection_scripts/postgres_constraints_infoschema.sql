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
        {{ if not .IncludeSystemCatalogs }}AND tc.table_schema <> 'information_schema' AND tc.table_schema NOT LIKE 'pg_%'{{ end }}
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
        ,CASE rc.match_option WHEN 'NONE' THEN '' ELSE 'MATCH ' || match_option END AS match_option
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
        {{ if not .IncludeSystemCatalogs }}AND tc.table_schema <> 'information_schema' AND tc.table_schema NOT LIKE 'pg_%'{{ end }}
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
    pg_constraint
    JOIN pg_class ON pg_class.oid = pg_constraint.conrelid
    JOIN pg_namespace ON pg_namespace.oid = pg_class.relnamespace
WHERE
    pg_constraint.contype = 'c'
    {{ if not .IncludeSystemCatalogs }}AND pg_namespace.nspname <> 'information_schema' AND pg_namespace.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND pg_namespace.nspname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND pg_namespace.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND pg_class.relname IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND pg_class.relname NOT IN ({{ printList .WithoutTables }}){{ end }}
UNION ALL
SELECT
    table_schema
    ,table_name
    ,constraint_name
    ,'EXCLUDE' AS constraint_type
    ,string_agg(column_name, ',' ORDER BY seq) AS columns
    ,COALESCE(pg_get_expr(exprs_oid, table_oid, TRUE), '') AS exprs
    ,'' AS references_schema
    ,'' AS references_table
    ,'' AS references_columns
    ,'' AS update_rule
    ,'' AS delete_rule
    ,'' AS match_option
    ,'' AS check_expr
    ,string_agg(operator, ',' ORDER BY seq) AS operators
    ,exclusion_index_type
    ,COALESCE(pg_get_expr(predicate_oid, table_oid, TRUE), '') AS predicate
    ,is_deferrable
    ,is_initially_deferred
FROM (
    SELECT
        schemas.nspname AS table_schema
        ,tables.relname AS table_name
        ,pg_constraint.conname AS constraint_name
        ,COALESCE(columns.attname, '') AS column_name
        ,pg_operator.oprname AS operator
        ,UPPER(pg_am.amname) AS exclusion_index_type
        ,pg_constraint.condeferrable AS is_deferrable
        ,pg_constraint.condeferred AS is_initially_deferred
        ,pg_index.indrelid AS table_oid
        ,pg_index.indexprs AS exprs_oid
        ,pg_index.indpred AS predicate_oid
        ,c.seq
    FROM
        pg_constraint
        JOIN pg_class AS tables ON tables.oid = pg_constraint.conrelid
        JOIN pg_class AS indexes ON indexes.oid = pg_constraint.conindid
        JOIN pg_namespace AS schemas ON schemas.oid = tables.relnamespace
        JOIN pg_index ON pg_index.indexrelid = indexes.oid
        JOIN pg_am ON pg_am.oid = indexes.relam
        CROSS JOIN unnest(pg_index.indkey) WITH ORDINALITY AS c(oid, seq)
        LEFT JOIN pg_attribute AS columns ON columns.attrelid = pg_index.indrelid AND columns.attnum = c.oid
        JOIN unnest(pg_constraint.conexclop) WITH ORDINALITY AS o(oid, seq) ON o.seq = c.seq
        JOIN pg_operator ON pg_operator.oid = o.oid
    WHERE
        pg_constraint.contype = 'x'
        {{ if not .IncludeSystemCatalogs }}AND schemas.nspname <> 'information_schema' AND schemas.nspname NOT LIKE 'pg_%'{{ end }}
        {{ if .WithSchemas }}AND schemas.nspname IN ({{ printList .WithSchemas }}){{ end }}
        {{ if .WithoutSchemas }}AND schemas.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
        {{ if .WithTables }}AND tables.relname IN ({{ printList .WithTables }}){{ end }}
        {{ if .WithoutTables }}AND tables.relname NOT IN ({{ printList .WithoutTables }}){{ end }}
) AS exclude_columns
GROUP BY
    table_schema
    ,table_name
    ,constraint_name
    ,exclusion_index_type
    ,is_deferrable
    ,is_initially_deferred
    ,table_oid
    ,exprs_oid
    ,predicate_oid
) AS tmp
{{- if .SortOutput }}
ORDER BY
    table_schema
    ,table_name
    ,constraint_name
{{- end }}
;
