SELECT
    tbl_name AS table_name
    ,name AS trigger_name
    ,sql || ';' AS sql
FROM
    sqlite_schema
WHERE
    "type" = 'trigger'
    {{ if not .IncludeSystemCatalogs }}AND tbl_name NOT LIKE 'sqlite_%' AND sql NOT LIKE 'CREATE TABLE ''%'{{ end }}
    {{ if .WithTables }}AND tbl_name IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND tbl_name NOT IN ({{ printList .WithoutTables }}){{ end }}
{{- if .SortOutput }}
ORDER BY
    tbl_name
    ,name
{{- end }}
;
