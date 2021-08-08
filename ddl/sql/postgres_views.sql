SELECT
    *
FROM (
SELECT
    schemaname AS view_schema
    ,viewname AS view_name
    ,FALSE AS is_materialized
    ,pg_get_viewdef(to_regclass(schemaname || '.' || viewname), true) AS sql
FROM
    pg_views
WHERE
    TRUE
    {{ if not .IncludeSystemCatalogs }}AND schemaname <> 'information_schema' AND schemaname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND schemaname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND schemaname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND viewname IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND viewname NOT IN ({{ printList .WithoutTables }}){{ end }}
UNION ALL
SELECT
    schemaname AS view_schema
    ,matviewname AS view_name
    ,TRUE AS is_materialized
    ,pg_get_viewdef(to_regclass(schemaname || '.' || matviewname), true) AS sql
FROM
    pg_matviews
WHERE
    TRUE
    {{ if not .IncludeSystemCatalogs }}AND schemaname <> 'information_schema' AND schemaname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND schemaname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND schemaname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND matviewname IN ({{ printList .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND matviewname NOT IN ({{ printList .WithoutTables }}){{ end }}
) AS tmp
{{- if .SortOutput }}
ORDER BY
    view_schema
    ,view_name
{{- end }}
;
