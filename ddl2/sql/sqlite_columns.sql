SELECT
    tables.tbl_name AS table_name
    ,columns.name AS column_name
    ,columns."type" AS column_type
    ,columns."notnull" AS is_notnull
    ,COALESCE(columns.dflt_value, '') AS column_default
FROM
    (SELECT tbl_name FROM sqlite_schema WHERE "type" = 'table' AND tbl_name <> 'sqlite_sequence' AND sql NOT LIKE 'CREATE TABLE ''%') AS tables
    CROSS JOIN pragma_table_xinfo(tables.tbl_name) AS columns
WHERE
    TRUE
    {{ if .IncludedTables }}AND tables.tbl_name IN ({{ listify .IncludedTables }}){{ end }}
    {{ if .ExcludedTables }}AND tables.tbl_name NOT IN ({{ listify .ExcludedTables }}){{ end }}
;
