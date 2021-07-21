WITH indexed_columns AS (
    SELECT
        $1 AS table_name
        ,il.name AS index_name
        ,il."unique" AS is_unique
        ,il.partial AS is_partial
        ,CASE ii.cid WHEN -1 THEN '' WHEN -2 THEN '' ELSE ii.name END AS column_name
    FROM
        pragma_index_list($1) AS il
        CROSS JOIN pragma_index_info(il.name) AS ii
    WHERE
        il.origin = 'c'
    ORDER BY
        il.name
        ,ii.seqno
)
SELECT
    '' AS table_schema
    ,table_name
    ,index_name
    ,'' AS index_type
    ,is_unique
    ,COUNT(column_name) AS num_key_columns
    ,group_concat(column_name) AS columns
    ,'' AS exprs
    ,'' AS predicate
    ,is_partial
FROM
    indexed_columns
GROUP BY
    table_name
    ,index_name
    ,is_unique
    ,is_partial
;
