WITH indexed_columns AS (
    SELECT
        table_schema
        ,table_name
        ,index_name
        ,index_type
        ,NOT non_unique AS is_unique
        ,COALESCE(column_name, '') AS column_name
        ,COALESCE(expression, '') AS expr
        ,seq_in_index
    FROM
        information_schema.statistics
    WHERE
        statistics.table_schema NOT IN ('mysql', 'performance_schema', 'sys')
        AND statistics.index_name <> 'PRIMARY'
)
SELECT
    table_schema
    ,table_name
    ,index_name
    ,index_type
    ,is_unique
    ,COUNT(column_name) AS num_key_columns
    ,group_concat(column_name ORDER BY seq_in_index) AS columns
    ,group_concat(expr ORDER BY seq_in_index) AS exprs
    ,'' AS predicate
    ,FALSE AS is_partial
FROM
    indexed_columns
GROUP BY
    table_schema
    ,table_name
    ,index_name
    ,index_type
    ,is_unique
;
