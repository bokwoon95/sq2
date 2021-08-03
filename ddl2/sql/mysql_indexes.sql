SELECT
    table_schema
    ,table_name
    ,index_name
    ,index_type
    ,is_unique
    ,group_concat(column_name ORDER BY seq_in_index) AS columns
    ,group_concat(expr ORDER BY seq_in_index) AS exprs
FROM (
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
        index_name <> 'PRIMARY'
        {{ if not .IncludeSystemSchemas }}AND table_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys'){{ end }}
        {{ if .IncludedSchemas }}AND table_schema IN ({{ listify .IncludedSchemas }}){{ end }}
        {{ if .ExcludedSchemas }}AND table_schema NOT IN ({{ listify .ExcludedSchemas }}){{ end }}
        {{ if .IncludedTables }}AND table_name IN ({{ listify .IncludedTables }}){{ end }}
        {{ if .ExcludedTables }}AND table_name NOT IN ({{ listify .ExcludedTables }}){{ end }}
) AS indexed_columns
GROUP BY
    table_schema
    ,table_name
    ,index_name
    ,index_type
    ,is_unique
;
