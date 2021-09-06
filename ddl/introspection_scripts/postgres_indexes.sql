SELECT
    table_schema
    ,table_name
    ,index_name
    ,index_type
    ,is_unique
    ,num_key_columns
    ,string_agg(column_name, ',' ORDER BY seq) AS columns
    ,COALESCE(pg_get_expr(exprs_oid, table_oid, TRUE), '') AS exprs
    ,COALESCE(pg_get_expr(predicate_oid, table_oid, TRUE), '') AS predicate
FROM (
    SELECT
        schemas.nspname AS table_schema
        ,tables.relname AS table_name
        ,indexes.relname AS index_name
        ,UPPER(pg_am.amname) AS index_type
        ,pg_index.indisunique AS is_unique
        ,pg_index.indnkeyatts AS num_key_columns
        ,COALESCE(columns.attname, '') AS column_name
        ,pg_index.indrelid AS table_oid
        ,pg_index.indexprs AS exprs_oid
        ,pg_index.indpred AS predicate_oid
        ,c.seq
    FROM
        pg_index
        JOIN pg_class AS indexes ON indexes.oid = pg_index.indexrelid
        JOIN pg_class AS tables ON tables.oid = pg_index.indrelid
        JOIN pg_namespace AS schemas ON schemas.oid = tables.relnamespace
        JOIN pg_am ON pg_am.oid = indexes.relam
        CROSS JOIN unnest(pg_index.indkey) WITH ORDINALITY AS c(oid, seq)
        LEFT JOIN pg_attribute AS columns ON columns.attrelid = pg_index.indrelid AND columns.attnum = c.oid
    WHERE
        NOT EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE pg_constraint.conname = indexes.relname
        )
        {{ if not .IncludeSystemCatalogs }}AND schemas.nspname <> 'information_schema' AND schemas.nspname NOT LIKE 'pg_%'{{ end }}
        {{ if .WithSchemas }}AND schemas.nspname IN ({{ printList .WithSchemas }}){{ end }}
        {{ if .WithoutSchemas }}AND schemas.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
        {{ if .WithTables }}AND tables.relname IN ({{ printList .WithTables }}){{ end }}
        {{ if .WithoutTables }}AND tables.relname NOT IN ({{ printList .WithoutTables }}){{ end }}
) AS indexed_columns
GROUP BY
    table_schema
    ,table_name
    ,index_name
    ,index_type
    ,is_unique
    ,num_key_columns
    ,table_oid
    ,exprs_oid
    ,predicate_oid
{{- if .SortOutput }}
ORDER BY
    table_schema
    ,table_name
    ,index_name
{{- end }}
;
