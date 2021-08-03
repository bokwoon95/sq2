SELECT
    tbl_name AS table_name
    ,sql
FROM
    sqlite_schema
WHERE
    "type" = 'table'
    AND tbl_name <> 'sqlite_sequence'
    AND sql NOT LIKE 'CREATE TABLE ''%'
    {{ if .IncludedTables }}AND tbl_name IN ({{ listify .IncludedTables }}){{ end }}
    {{ if .ExcludedTables }}AND tbl_name NOT IN ({{ listify .ExcludedTables }}){{ end }}
;
