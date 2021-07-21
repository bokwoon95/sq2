SELECT
    table_schema
    ,table_name
    ,constraint_name
    ,constraint_type
    ,columns
    ,exprs
    ,references_schema
    ,references_table
    ,references_columns
    ,update_rule
    ,delete_rule
    ,match_option
    ,check_expr
    ,operators
    ,index_type
    ,predicate
    ,is_deferrable
    ,is_initially_deferred
FROM (
SELECT
    tc.table_schema
    ,tc.table_name
    ,tc.constraint_name
    ,tc.constraint_type
    ,group_concat(kcu.column_name ORDER BY kcu.ordinal_position) AS columns
    ,'' AS exprs
    ,COALESCE(kcu.referenced_table_schema, '') AS references_schema
    ,COALESCE(kcu.referenced_table_name, '') AS references_table
    ,COALESCE(group_concat(kcu.referenced_column_name ORDER BY kcu.ordinal_position), '') AS references_columns
    ,COALESCE(rc.update_rule, '') AS update_rule
    ,COALESCE(rc.delete_rule, '') AS delete_rule
    ,COALESCE(rc.match_option, '') AS match_option
    ,'' AS check_expr
    ,'' AS operators
    ,'' AS index_type
    ,'' AS predicate
    ,FALSE AS is_deferrable
    ,FALSE AS is_initially_deferred
FROM
    information_schema.table_constraints AS tc
    JOIN information_schema.key_column_usage AS kcu USING (constraint_schema, constraint_name, table_name)
    LEFT JOIN information_schema.referential_constraints AS rc USING (constraint_schema, constraint_name)
WHERE
    tc.table_schema NOT IN ('mysql', 'performance_schema', 'sys')
    AND tc.constraint_type IN ('PRIMARY KEY', 'FOREIGN KEY', 'UNIQUE')
GROUP BY
    tc.table_schema
    ,tc.table_name
    ,tc.constraint_name
    ,tc.constraint_type
    ,kcu.referenced_table_schema
    ,kcu.referenced_table_name
    ,rc.update_rule
    ,rc.delete_rule
    ,rc.match_option
UNION ALL
SELECT
    tc.table_schema
    ,tc.table_name
    ,tc.constraint_name
    ,tc.constraint_type
    ,'' AS columns
    ,'' AS exprs
    ,'' AS references_schema
    ,'' AS references_table
    ,'' AS references_columns
    ,'' AS update_rule
    ,'' AS delete_rule
    ,'' AS match_option
    ,check_clause AS check_expr
    ,'' AS operators
    ,'' AS index_type
    ,'' AS predicate
    ,FALSE AS is_deferrable
    ,FALSE AS is_initially_deferred
FROM
    information_schema.table_constraints AS tc
    JOIN information_schema.check_constraints AS cc USING (constraint_schema, constraint_name)
WHERE
    tc.table_schema NOT IN ('mysql', 'performance_schema', 'sys')
) AS tmp
;
