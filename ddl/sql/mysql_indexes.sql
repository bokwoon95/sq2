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
        statistics.index_name <> 'PRIMARY'
        AND statistics.table_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys')
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
    ,'' AS "sql"
FROM
    indexed_columns
GROUP BY
    table_schema
    ,table_name
    ,index_name
    ,index_type
    ,is_unique
;
