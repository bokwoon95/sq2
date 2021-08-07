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
    {{ if not .IncludeSystemObjects }}AND pg_namespace.nspname <> 'information_schema' AND pg_namespace.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND pg_namespace.nspname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND pg_namespace.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
GROUP BY
    pg_namespace.nspname
    ,pg_type.typname
{{- if .SortOutput }}
ORDER BY
    pg_namespace.nspname
    ,pg_type.typname
{{- end }}
;
