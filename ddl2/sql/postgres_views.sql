SELECT
    schemaname AS view_schema
    ,viewname AS view_name
    ,FALSE AS is_materialized
    ,pg_get_viewdef(to_regclass(schemaname || '.' || viewname), true) AS sql
FROM
    pg_catalog.pg_views
WHERE
    TRUE
    {{ if not .IncludeSystemSchemas }}AND schemaname <> 'information_schema' AND schemaname NOT LIKE 'pg_%'{{ end }}
    {{ if .IncludedSchemas }}AND schemaname IN ({{ listify .IncludedSchemas }}){{ end }}
    {{ if .ExcludedSchemas }}AND schemaname NOT IN ({{ listify .ExcludedSchemas }}){{ end }}
    {{ if .IncludedTables }}AND viewname IN ({{ listify .IncludedTables }}){{ end }}
    {{ if .ExcludedTables }}AND viewname NOT IN ({{ listify .ExcludedTables }}){{ end }}
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
    {{ if not .IncludeSystemSchemas }}AND schemaname <> 'information_schema' AND schemaname NOT LIKE 'pg_%'{{ end }}
    {{ if .IncludedSchemas }}AND schemaname IN ({{ listify .IncludedSchemas }}){{ end }}
    {{ if .ExcludedSchemas }}AND schemaname NOT IN ({{ listify .ExcludedSchemas }}){{ end }}
    {{ if .IncludedTables }}AND matviewname IN ({{ listify .IncludedTables }}){{ end }}
    {{ if .ExcludedTables }}AND matviewname NOT IN ({{ listify .ExcludedTables }}){{ end }}
;
