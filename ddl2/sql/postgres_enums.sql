SELECT
    pg_namespace.nspname AS enum_schema
    ,pg_type.typname AS enum_name
    ,json_agg(pg_enum.enumlabel::TEXT ORDER BY pg_enum.enumsortorder) AS enum_values
FROM
    pg_catalog.pg_enum
    JOIN pg_catalog.pg_type ON pg_type.oid = pg_enum.enumtypid
    JOIN pg_catalog.pg_namespace ON pg_namespace.oid = pg_type.typnamespace
WHERE
    TRUE
    {{ if not .IncludeSystemSchemas }}AND pg_namespace.nspname <> 'information_schema' AND pg_namespace.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .IncludedSchemas }}AND pg_namespace.nspname IN ({{ listify .IncludedSchemas }}){{ end }}
    {{ if .ExcludedSchemas }}AND pg_namespace.nspname NOT IN ({{ listify .ExcludedSchemas }}){{ end }}
GROUP BY
    pg_namespace.nspname
    ,pg_type.typname
;
