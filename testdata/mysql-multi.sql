SELECT *
FROM address AS a JOIN city AS ci ON ci.city_id = a.city_id JOIN country AS co ON co.country_id = ci.country_id
WHERE a.address_id = 632;

SELECT *
FROM address AS a JOIN city AS ci ON ci.city_id = a.city_id JOIN country AS co ON co.country_id = ci.country_id
WHERE ci.city_id = 609;

SELECT *
FROM address AS a JOIN city AS ci ON ci.city_id = a.city_id JOIN country AS co ON co.country_id = ci.country_id
WHERE co.country_id = 112;

------------
-- DELETE --
------------

-- [OK] No join, with aliases
DELETE FROM address AS a
WHERE
    a.address_id = 632
;

-- [FAIL] naive join, with aliases (syntax error)
DELETE FROM address AS a
    JOIN city AS ci ON ci.city_id = a.city_id
    JOIN country AS co ON co.country_id = ci.country_id
WHERE
    ci.city_id = 609
;

-- [FAIL] naive join, with no aliases (syntax error)
DELETE FROM address
    JOIN city ON city.city_id = address.city_id
    JOIN country ON country.country_id = city.country_id
WHERE
    city.city_id = 609
;

-- [FAIL] USING join, with aliases, FromTable non-repeated (syntax error)
DELETE FROM address AS a
    USING city AS ci
    JOIN country AS co ON co.country_id = ci.country_id
WHERE
    ci.city_id = a.city_id
    AND ci.city_id = 609
;

-- [FAIL] USING join, with no aliases, FromTable non-repeated (unknown table address)
DELETE FROM address
    USING city
    JOIN country ON country.country_id = city.country_id
WHERE
    city.city_id = address.city_id
    AND city.city_id = 609
;

-- [OK] USING join, with no aliases on FromTable, FromTable repeated
DELETE FROM address
    USING address
    JOIN city AS ci ON ci.city_id = address.city_id
    JOIN country AS co ON co.country_id = ci.country_id
WHERE
    AND ci.city_id = 609
;

-- [OK] USING join, with aliases, FromTable repeated
DELETE FROM a
    USING address AS a
    JOIN city AS ci ON ci.city_id = a.city_id
    JOIN country AS co ON co.country_id = ci.country_id
WHERE
    AND ci.city_id = 609
;

-- [OK] multi-table delete
-- Warning! SELECT ROW_COUNT(); is unreliable after a multi-table delete. Do not rely on it!
DELETE FROM
    a, ci, co
USING
    address AS a
    JOIN city AS ci ON ci.city_id = a.city_id
    JOIN country AS co ON co.country_id = ci.country_id
WHERE
    co.country_id = 112
;

-- run this before and after the multi-table delete to verify that all 9
-- addresses associated with country_id 112 have been deleted
SELECT * from address where address_id > 623;

------------
-- UPDATE --
------------

-- [OK] USING join, with no aliases
UPDATE address
    JOIN city ON city.city_id = address.city_id
    JOIN country ON country.country_id = city.country_id
SET
    address.address = '3 CC Street (modified)'
    ,city.city = 'CITY C-C (modified)'
    ,country.country = 'Country C (modified)'
WHERE
    address.address_id = 632
;

-- [OK] USING join, with aliases
UPDATE address AS a
    JOIN city AS ci ON ci.city_id = a.city_id
    JOIN country AS co ON co.country_id = ci.country_id
SET
    a.address = '3 CC Street'
    ,ci.city = 'CITY C-C'
    ,co.country = 'Country C'
WHERE
    a.address_id = 632
;

-- rune this before and after the multi-table update to verify that all 9
-- addresses, cities and countries associated with country_id 112 have been
-- modified
SELECT *
FROM address AS a JOIN city AS ci ON ci.city_id = a.city_id JOIN country AS co ON co.country_id = ci.country_id
WHERE co.country_id = 112;
