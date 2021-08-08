SELECT
    pg_attribute.attname AS column_name
    ,pg_catalog.format_type(pg_attribute.atttypid, pg_attribute.atttypmod) AS data_type
FROM
    pg_catalog.pg_attribute
    INNER JOIN pg_catalog.pg_class ON pg_class.oid = pg_attribute.attrelid
    INNER JOIN pg_catalog.pg_namespace ON pg_namespace.oid = pg_class.relnamespace
WHERE
    pg_attribute.attnum > 0
    AND NOT pg_attribute.attisdropped
    AND pg_namespace.nspname = 'public'
    AND pg_class.relname = 'actor'
ORDER BY
    attnum ASC
;
