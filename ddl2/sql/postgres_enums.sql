SELECT
    schemas.nspname AS enum_schema
    ,pg_type.typname AS enum_name
    ,json_agg(pg_enum.enumlabel::TEXT ORDER BY pg_enum.enumsortorder) AS enum_values
FROM
    pg_enum
    JOIN pg_type ON pg_type.oid = pg_enum.enumtypid
    JOIN pg_namespace AS schemas ON schemas.oid = pg_type.typnamespace
WHERE
    TRUE
    {{ if not .IncludeSystemCatalogs }}AND schemas.nspname <> 'information_schema' AND schemas.nspname NOT LIKE 'pg_%'{{ end }}
    {{ if .WithSchemas }}AND schemas.nspname IN ({{ printList .WithSchemas }}){{ end }}
    {{ if .WithoutSchemas }}AND schemas.nspname NOT IN ({{ printList .WithoutSchemas }}){{ end }}
GROUP BY
    schemas.nspname
    ,pg_type.typname
{{- if .SortOutput }}
ORDER BY
    schemas.nspname
    ,pg_type.typname
{{- end }}
;
