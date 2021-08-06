SELECT
    tables.tbl_name AS table_name
    ,columns.name AS column_name
    ,columns."type" AS column_type
    ,columns."notnull" AS is_notnull
    ,COALESCE(columns.dflt_value, '') AS column_default
FROM (
    SELECT
        tbl_name
    FROM
        sqlite_schema
    WHERE
        "type" = 'table'
        {{ if not .IncludeSystemObjects }}AND tbl_name NOT LIKE 'sqlite_%' AND sql NOT LIKE 'CREATE TABLE ''%'{{ end }}
        {{ if .WithTables }}AND tbl_name IN ({{ printList .WithTables }}){{ end }}
        {{ if .WithoutTables }}AND tbl_name NOT IN ({{ printList .WithoutTables }}){{ end }}
    ) AS tables
    CROSS JOIN pragma_table_xinfo(tables.tbl_name) AS columns
;
