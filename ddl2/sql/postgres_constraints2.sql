SELECT
    *
FROM
    pg_catalog.pg_constraint
    JOIN pg_catalog.pg_class ON pg_class.oid = pg_constraint.conrelid
    JOIN pg_catalog.pg_namespace ON pg_namespace.oid = pg_class.relnamespace
WHERE
    TRUE
    AND pg_namespace.nspname <> 'information_schema' AND pg_namespace.nspname NOT LIKE 'pg_%'
;
