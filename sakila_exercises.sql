------------
-- SELECT --
------------

-- Find the distinct actor last names ordered by last name. Show only the top 4
-- results.
SELECT DISTINCT last_name FROM actor ORDER BY last_name LIMIT 4;

-- Check if there is any actor with first name 'SCARLETT' or last name
-- 'JOHANSSON'.
SELECT EXISTS(SELECT 1 FROM actor WHERE first_name = 'SCARLETT' OR last_name = 'JOHANSSON');

-- Find the number of distinct actor last names.
SELECT COUNT(DISTINCT last_name) FROM actor;

-- Find all actors whose last name contain the letters 'GEN', ordered by
-- actor_id. Return the actor_id, first_name and last_name.
SELECT actor_id, first_name, last_name FROM actor WHERE last_name LIKE '%GEN%' ORDER BY actor_id;

-- Find actor last names that only once in the database, ordered by last_name.
-- Show only the top 10 results.
SELECT last_name FROM actor GROUP BY last_name HAVING COUNT(*) = 1 ORDER BY last_name LIMIT 10;

-- Find the cities of the countries Egypt, Greece and Puerto Rico, ordered by
-- country_name and city_name. Return the country_name and city_name.
SELECT
    country.country
    ,city.city
FROM
    city
    JOIN country ON country.country_id = city.country_id
WHERE
    country.country IN ('Egypt', 'Greece', 'Puerto Rico')
ORDER BY
    country.country, city.city
;

-- List films with their price categories, ordered by title. The categories
-- are: price < 2.99 is 'discount', price >= 2.99 and price < 4.99 is
-- 'regular', price >= 4.99 is 'premium'. Return the title, price, and price
-- category. Show only the top 10 results.
SELECT
    title
    ,price
    ,CASE
        WHEN price < 2.99 THEN 'discount'
        WHEN price >= 2.99 AND price < 4.99 THEN 'regular'
        ELSE 'premium'
    END AS price_category
FROM
    film
ORDER BY
    title
LIMIT
    10
;

-- Find the actors who have appeared in the most films ordered by descending
-- film count. Return the first name, last name and film count. Show only the
-- top 10 results.
SELECT
    actor.actor_id
    ,actor.first_name
    ,actor.last_name
    ,COUNT(*) AS film_count
FROM
    actor
    JOIN film_actor ON film_actor.actor_id = actor.actor_id
GROUP BY
    actor.actor_id
ORDER BY
    film_count DESC
LIMIT
    10
;

-- https://stackoverflow.com/questions/67080935/how-can-i-get-the-desired-results-from-the-sakila-database-using-sql
-- Find the films whose total number of actors is above the average, ordered by
-- title. Return the film title and the actor count. Show only the top 10
-- results.
WITH film_stats AS (
    SELECT film_id, COUNT(*) AS actor_count
    FROM film_actor
    GROUP BY film_id
)
SELECT
    film.title
    ,film_stats.actor_count
FROM
    film_stats
    JOIN film ON film.film_id = film_stats.film_id
WHERE
    film_stats.actor_count > (SELECT AVG(actor_count) FROM film_stats)
ORDER BY
    film.title
LIMIT
    10
;

------------
-- INSERT --
------------

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
