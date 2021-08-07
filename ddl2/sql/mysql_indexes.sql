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
        {{ if not .IncludeSystemObjects }}AND table_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys'){{ end }}
        {{ if .WithSchemas }}AND table_schema IN ({{ printList .WithSchemas }}){{ end }}
        {{ if .WithoutSchemas }}AND table_schema NOT IN ({{ printList .WithoutSchemas }}){{ end }}
        {{ if .WithTables }}AND table_name IN ({{ printList .WithTables }}){{ end }}
        {{ if .WithoutTables }}AND table_name NOT IN ({{ printList .WithoutTables }}){{ end }}
) AS indexed_columns
GROUP BY
    table_schema
    ,table_name
    ,index_name
    ,index_type
    ,is_unique
{{- if .SortOutput }}
ORDER BY
    table_schema
    ,table_name
    ,index_name
{{- end }}
;
