SELECT '' AS table_schema, tbl_name AS table_name
FROM sqlite_schema
WHERE "type" = 'table';
