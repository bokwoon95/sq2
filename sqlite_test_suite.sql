------------
-- INSERT --
------------

-- Insert and get ID (sqlite uses both RETURNING and LastInsertID)
DELETE FROM country WHERE country IN ('Norway', 'Ireland', 'Iceland', 'Singapore', 'Denmark', 'Luxembourg', 'Slovenia', 'Czech Republic', 'Malta', 'Cyprus', 'Andorra', 'Qatar', 'Portugal', 'Croatia', 'Russia', 'Montenegro');

INSERT INTO country
    (country)
VALUES
    ('Norway')
    ,('Ireland')
    ,('Iceland')
    ,('Singapore')
    ,('Denmark')
    ,('Luxembourg')
    ,('Slovenia')
RETURNING country_id, country;

INSERT INTO country
    (country)
VALUES
    ('Norway')
    ,('Ireland')
    ,('Iceland')
    ,('Singapore')
    ,('Denmark')
    ,('Luxembourg')
    ,('Slovenia')
    ,('Czech Republic')
    ,('Malta')
    ,('Cyprus')
    ,('Andorra')
    ,('Qatar')
    ,('Portugal')
    ,('Croatia')
    ,('Russia')
    ,('Montenegro')
ON CONFLICT DO NOTHING
RETURNING country_id, country;

-- Insert the same row with the ID but ignore conflicts

-- Upsert a row, get ID (sqlite uses both RETURNING and LastInsertID)

-- Upsert the same row with ID, but change a column

-- Insert from SELECT
-- Customer 'MARY SMITH' rents the film 'ACADEMY DINOSAUR' from staff 'Mike
-- Hillyer' at Store 1 on 9th of August 2021 4pm. Write the query that creates a
-- new rental record representing that transaction.
INSERT INTO rental
    (inventory_id, customer_id, staff_id, rental_date)
SELECT
    inventory.inventory_id
    ,(
        SELECT customer_id
        FROM customer
        WHERE (first_name, last_name) = ('MARY', 'SMITH')
    ) AS customer_id
    ,(
        SELECT staff.staff_id
        FROM staff JOIN store ON store.store_id = staff.store_id
        WHERE store.store_id = 1 AND (staff.first_name, staff.last_name) = ('Mike', 'Hillyer')
    ) AS staff_id
    ,'2021-08-09 16:00:00' AS rental_date
FROM
    film
    JOIN inventory ON inventory.film_id = film.film_id
    JOIN store ON store.store_id = inventory.inventory_id
WHERE
    film.title = 'ACADEMY DINOSAUR'
    AND store.store_id = 1
    AND NOT EXISTS (
        SELECT 1
        FROM rental
        WHERE rental.inventory_id = inventory.inventory_id AND rental.return_date IS NULL
    )
ORDER BY
    inventory.inventory_id
LIMIT
    1
;

------------
-- UPDATE --
------------

-- Update with join

-- Update with returning (postgres, sqlite)

-- Update with limit (mysql)

-- Multi-table update (mysql)

------------
-- DELETE --
------------

-- Delete with join

-- Delete with returning (postgres, sqlite)

-- Delete with limit (mysql)

-- Multi-table delete (mysql)
