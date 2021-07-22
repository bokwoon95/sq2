SELECT
    *
FROM
    information_schema.routines
    JOIN information_schema.parameters ON 
        parameters.specific_schema = routines.routine_schema
        AND parameters.specific_name
        -- uhhh need to GROUP BY group_concat the params first before joining it back to the routines
