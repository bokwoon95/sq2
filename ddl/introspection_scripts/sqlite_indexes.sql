SELECT
    table_name
    ,index_name
    ,is_unique
    ,group_concat(column_name) AS columns
    ,sql
FROM (
    SELECT
        tables.tbl_name AS table_name
        ,il.name AS index_name
        ,il."unique" AS is_unique
        ,CASE ii.cid WHEN -1 THEN '' WHEN -2 THEN '' ELSE ii.name END AS column_name
        ,ii.seqno
        ,m.sql
    FROM (
        SELECT
            tbl_name
        FROM
            sqlite_schema
        WHERE
            "type" = 'table'
            {{ if not .IncludeSystemCatalogs }}AND tbl_name NOT LIKE 'sqlite_%' AND sql NOT LIKE 'CREATE TABLE ''%'{{ end }}
            {{ if .WithTables }}AND tbl_name IN ({{ printList .WithTables }}){{ end }}
            {{ if .WithoutTables }}AND tbl_name NOT IN ({{ printList .WithoutTables }}){{ end }}
        ) AS tables
        CROSS JOIN pragma_index_list(tables.tbl_name) AS il
        CROSS JOIN pragma_index_info(il.name) AS ii
        JOIN sqlite_schema AS m ON m."type" = 'index' AND m.tbl_name = tables.tbl_name AND m.name = il.name
    WHERE
        il.origin = 'c'
    ORDER BY
        il.name
        ,ii.seqno
    ) AS indexed_columns
GROUP BY
    table_name
    ,index_name
    ,is_unique
    ,sql
{{- if .SortOutput }}
ORDER BY
    table_name
    ,index_name
{{- end }}
;
