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
    FROM
        (SELECT tbl_name FROM sqlite_schema WHERE "type" = 'table' AND tbl_name <> 'sqlite_sequence' AND sql NOT LIKE 'CREATE TABLE ''%') AS tables
        CROSS JOIN pragma_table_info(tables.tbl_name) AS columns
    WHERE
        columns.pk > 0
        {{ if .IncludedTables }}AND tables.tbl_name IN ({{ listify .IncludedTables }}){{ end }}
        {{ if .ExcludedTables }}AND tables.tbl_name NOT IN ({{ listify .ExcludedTables }}){{ end }}
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
    FROM
        (SELECT tbl_name FROM sqlite_schema WHERE "type" = 'table' AND tbl_name <> 'sqlite_sequence' AND sql NOT LIKE 'CREATE TABLE ''%') AS tables
        CROSS JOIN pragma_index_list(tables.tbl_name) AS il
        CROSS JOIN pragma_index_info(il.name) AS ii
    WHERE
        il."unique"
        AND il.origin = 'u'
        {{ if .IncludedTables }}AND tables.tbl_name IN ({{ listify .IncludedTables }}){{ end }}
        {{ if .ExcludedTables }}AND tables.tbl_name NOT IN ({{ listify .ExcludedTables }}){{ end }}
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
    FROM
        (SELECT tbl_name FROM sqlite_schema WHERE "type" = 'table' AND tbl_name <> 'sqlite_sequence' AND sql NOT LIKE 'CREATE TABLE ''%') AS tables
        CROSS JOIN pragma_foreign_key_list(tables.tbl_name) AS fkl
    WHERE
        TRUE
        {{ if .IncludedTables }}AND tables.tbl_name IN ({{ listify .IncludedTables }}){{ end }}
        {{ if .ExcludedTables }}AND tables.tbl_name NOT IN ({{ listify .ExcludedTables }}){{ end }}
    ORDER BY
        fkl.seq
) AS foreign_key_columns
GROUP BY
    table_name
    ,id
    ,references_table
    ,update_rule
    ,delete_rule
;
