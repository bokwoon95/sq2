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
    '' AS table_schema
    ,$1 AS table_name
    ,'' AS constraint_name
    ,'PRIMARY KEY' AS constraint_type
    ,COALESCE(group_concat(column_name), 'ROWID') AS columns
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
    ,FALSE AS is_deferrable
    ,FALSE AS is_initially_deferred
FROM (
    SELECT name AS column_name
    FROM pragma_table_info($1)
    WHERE pk > 0
    ORDER BY pk
) AS primary_key_columns
UNION ALL
SELECT
    '' AS table_schema
    ,$1 AS table_name
    ,'' AS constraint_name
    ,'FOREIGN KEY' AS constraint_type
    ,group_concat(column_name) AS columns
    ,'' AS exprs
    ,'' AS references_schema
    ,references_table
    ,group_concat(references_column) AS references_columns
    ,update_rule
    ,delete_rule
    ,'' AS match_option
    ,'' AS check_expr
    ,'' AS operators
    ,'' AS index_type
    ,'' AS predicate
    ,FALSE AS is_deferrable
    ,FALSE AS is_initially_deferred
FROM (
    SELECT
        id
        ,"from" AS column_name
        ,"table" AS references_table
        ,"to" AS references_column
        ,on_update AS update_rule
        ,on_delete AS delete_rule
    FROM
        pragma_foreign_key_list($1)
    ORDER BY
        seq
) AS foreign_key_columns
GROUP BY
    id
    ,references_table
    ,update_rule
    ,delete_rule
    ,match_option
UNION ALL
SELECT
    '' AS table_schema
    ,$1 AS table_name
    ,'' AS constraint_name
    ,'UNIQUE' AS constraint_type
    ,group_concat(column_name) AS columns
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
    ,FALSE AS is_deferrable
    ,FALSE AS is_initially_deferred
FROM (
    SELECT il.name AS index_name, ii.name AS column_name
    FROM pragma_index_list($1) AS il CROSS JOIN pragma_index_info(il.name) AS ii
    WHERE il."unique" AND il.origin = 'u'
    ORDER BY ii.seqno
) AS unique_columns
GROUP BY
    index_name
) AS tmp
;
