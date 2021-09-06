SELECT
    *
FROM (
SELECT
    tc.table_schema
    ,tc.table_name
    ,tc.constraint_name
    ,tc.constraint_type
    ,group_concat(kcu.column_name ORDER BY kcu.ordinal_position) AS columns
    ,COALESCE(kcu.referenced_table_schema, '') AS references_schema
    ,COALESCE(kcu.referenced_table_name, '') AS references_table
    ,COALESCE(group_concat(kcu.referenced_column_name ORDER BY kcu.ordinal_position), '') AS references_columns
    ,COALESCE(rc.update_rule, '') AS update_rule
    ,COALESCE(rc.delete_rule, '') AS delete_rule
    ,CASE rc.match_option WHEN 'NONE' THEN '' ELSE COALESCE(CONCAT('MATCH ', rc.match_option), '') END AS match_option
    ,'' AS check_expr
FROM
    information_schema.table_constraints AS tc
    JOIN information_schema.key_column_usage AS kcu USING (constraint_schema, constraint_name, table_name)
    LEFT JOIN information_schema.referential_constraints AS rc USING (constraint_schema, constraint_name)
WHERE
    tc.constraint_type IN ('PRIMARY KEY', 'FOREIGN KEY', 'UNIQUE')
    {{ if not .IncludeSystemCatalogs }}AND tc.table_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys'){{ end }}
    {{ if .WithSchemas }}AND tc.table_schema IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND tc.table_schema NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND tc.table_name IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND tc.table_name NOT IN ({{ printList .WithoutTables }}){{ end }}
GROUP BY
    tc.table_schema
    ,tc.table_name
    ,tc.constraint_name
    ,tc.constraint_type
    ,kcu.referenced_table_schema
    ,kcu.referenced_table_name
    ,rc.update_rule
    ,rc.delete_rule
    ,rc.match_option
UNION ALL
SELECT
    tc.table_schema
    ,tc.table_name
    ,tc.constraint_name
    ,tc.constraint_type
    ,'' AS columns
    ,'' AS references_schema
    ,'' AS references_table
    ,'' AS references_columns
    ,'' AS update_rule
    ,'' AS delete_rule
    ,'' AS match_option
    ,cc.check_clause AS check_expr
FROM
    information_schema.table_constraints AS tc
    JOIN information_schema.check_constraints AS cc USING (constraint_schema, constraint_name)
WHERE
    {{ if not .IncludeSystemCatalogs }}table_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys'){{ end }}
    {{ if .WithSchemas }}AND tc.table_schema IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND tc.table_schema NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND tc.table_name IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND tc.table_name NOT IN ({{ printList .WithoutTables }}){{ end }}
) AS tmp
{{- if .SortOutput }}
ORDER BY
    table_schema
    ,table_name
    ,constraint_name
{{- end }}
;
