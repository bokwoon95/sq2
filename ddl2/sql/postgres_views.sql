SELECT
    schemaname AS view_schema
    ,viewname AS view_name
    ,FALSE AS is_materialized
    ,pg_get_viewdef(to_regclass(schemaname || '.' || viewname), true) AS sql
FROM
    pg_catalog.pg_views
WHERE
    TRUE
    {{ if not .IncludeSystemObjects }}AND schemaname <> 'information_schema' AND schemaname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND schemaname IN ({{ listify .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND schemaname NOT IN ({{ listify .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND viewname IN ({{ listify .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND viewname NOT IN ({{ listify .WithoutTables }}){{ end }}
UNION ALL
SELECT
    schemaname AS view_schema
    ,matviewname AS view_name
    ,TRUE AS is_materialized
    ,pg_get_viewdef(to_regclass(schemaname || '.' || matviewname), true) AS sql
FROM
    pg_catalog.pg_matviews
WHERE
    TRUE
    {{ if not .IncludeSystemObjects }}AND schemaname <> 'information_schema' AND schemaname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND schemaname IN ({{ listify .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND schemaname NOT IN ({{ listify .WithoutSchemas }}){{ end }}
    {{ if .WithTables }}AND matviewname IN ({{ listify .WithTables }}){{ end }}
    {{ if .WithoutTables }}AND matviewname NOT IN ({{ listify .WithoutTables }}){{ end }}
;
