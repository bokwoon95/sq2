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
    FROM
        (SELECT tbl_name FROM sqlite_schema WHERE "type" = 'table' AND tbl_name <> 'sqlite_sequence' AND sql NOT LIKE 'CREATE TABLE ''%') AS tables
        CROSS JOIN pragma_index_list(tables.tbl_name) AS il
        JOIN sqlite_schema AS m ON m."type" = 'index' AND m.tbl_name = tables.tbl_name AND m.name = il.name
        CROSS JOIN pragma_index_info(il.name) AS ii
    WHERE
        il.origin = 'c'
        {{ if .IncludedTables }}AND tables.tbl_name IN ({{ listify .IncludedTables }}){{ end }}
        {{ if .ExcludedTables }}AND tables.tbl_name NOT IN ({{ listify .ExcludedTables }}){{ end }}
    ORDER BY
        il.name
        ,ii.seqno
) AS indexed_columns
GROUP BY
    table_name
    ,index_name
    ,is_unique
    ,sql
;
