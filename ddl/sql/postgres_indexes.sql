WITH index_columns AS (
    SELECT
        table_namespace.nspname AS table_schema
        ,table_info.relname AS table_name
        ,index_info.relname AS index_name
        ,UPPER(pg_am.amname) AS index_type
        ,pg_index.indisunique AS is_unique
        ,pg_index.indnkeyatts AS num_key_columns
        ,COALESCE(pg_attribute.attname, '') AS column_name
        ,COALESCE(pg_get_expr(pg_index.indexprs, pg_index.indrelid, true), '') AS exprs
        ,COALESCE(pg_get_expr(pg_index.indpred, pg_index.indrelid, true), '') AS predicate
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
        table_namespace.nspname <> 'information_schema'
        AND table_namespace.nspname NOT LIKE 'pg_%'
)
SELECT
    table_schema
    ,table_name
    ,index_name
    ,index_type
    ,is_unique
    ,num_key_columns
    ,string_agg(column_name, ',' ORDER BY seq_in_index) AS columns
    ,exprs
    ,predicate
    ,'' AS sql
FROM
    index_columns
GROUP BY
    table_schema
    ,table_name
    ,index_name
    ,index_type
    ,is_unique
    ,num_key_columns
    ,exprs
    ,predicate
;
