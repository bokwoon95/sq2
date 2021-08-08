SELECT
    *
FROM (
SELECT
    table_name
    ,'PRIMARY KEY' AS constraint_type
    ,COALESCE(group_concat(column_name), 'ROWID') AS columns
    ,'' AS references_table
    ,'' AS references_columns
    ,'' AS update_rule
    ,'' AS delete_rule
FROM (
    SELECT
        tables.tbl_name AS table_name
        ,columns.name AS column_name
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
        CROSS JOIN pragma_table_info(tables.tbl_name) AS columns
    WHERE
        columns.pk > 0
    ORDER BY
        tables.tbl_name
        ,columns.pk
    ) AS primary_key_columns
GROUP BY
    table_name
UNION ALL
SELECT
    table_name
    ,'UNIQUE' AS constraint_type
    ,COALESCE(group_concat(column_name), '') AS columns
    ,'' AS references_table
    ,'' AS references_columns
    ,'' AS update_rule
    ,'' AS delete_rule
FROM (
    SELECT
        tables.tbl_name AS table_name
        ,il.name AS index_name
        ,ii.name AS column_name
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
    WHERE
        il."unique"
        AND il.origin = 'u'
    ORDER BY
        ii.seqno
    ) AS unique_columns
GROUP BY
    table_name
    ,index_name
UNION ALL
SELECT
    table_name
    ,'FOREIGN KEY' AS constraint_type
    ,COALESCE(group_concat(column_name), '') AS columns
    ,references_table
    ,COALESCE(group_concat(references_column), '') AS references_columns
    ,update_rule
    ,delete_rule
FROM (
    SELECT
        tables.tbl_name AS table_name
        ,fkl.id
        ,fkl."from" AS column_name
        ,fkl."table" AS references_table
        ,fkl."to" AS references_column
        ,fkl.on_update AS update_rule
        ,fkl.on_delete AS delete_rule
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
        CROSS JOIN pragma_foreign_key_list(tables.tbl_name) AS fkl
    ORDER BY
        fkl.seq
    ) AS foreign_key_columns
GROUP BY
    table_name
    ,id
    ,references_table
    ,update_rule
    ,delete_rule
) AS tmp
{{- if .SortOutput }}
ORDER BY
    table_name
    ,constraint_type
{{- end }}
;
