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

-- List films with their film classification, ordered by title. The classifications are:
-- length <= 60 then 'short', length > 60 and length <= 100 then 'medium',
-- length > 100 then 'long'. Return the title, length, and film_classification. Show
-- only the top 10 results.
SELECT
    title
    ,length
    ,CASE
        WHEN length <= 60 THEN 'short'
        WHEN length > 60 AND length <= 100 THEN 'medium'
        ELSE 'long'
    END AS film_classification
FROM
    film
ORDER BY
    title
LIMIT
    10
;

-- List films with their target audience, ordered by title. The target
-- audiences are: rating = 'G' then 'family', rating = 'PG' or rating = 'PG-13'
-- then 'teens', rating = 'R' or rating = 'NC-17' then 'adults'. Return the
-- title, rating, and target_audience.
SELECT
    title
    ,rating
    ,CASE rating
        WHEN 'G' THEN 'family'
        WHEN 'PG' THEN 'teens'
        WHEN 'PG-13' THEN 'teens'
        WHEN 'R' THEN 'adults'
        WHEN 'NC-17' THEN 'adults'
    END AS intended_audience
FROM
    film
ORDER BY
    title
LIMIT
    10
;

-- https://towardsdatascience.com/sql-tricks-for-data-scientists-53298467dd5
-- sqlite: case strftime('%m', rental_date) when 01 then 'January' when 02 then 'February' ... end
-- postgres: to_char(rental_date, 'Month')
-- mysql: monthname(rental_date)
WITH months (num, name) AS (
    VALUES ('01', 'January'), ('02', 'February'), ('03', 'March'),
        ('04', 'April'), ('05', 'May'), ('06', 'June'),
        ('07', 'July'), ('08', 'August'), ('09', 'September'),
        ('10', 'October'), ('11', 'November'), ('12', 'December')
)
SELECT
    months.name AS month
    ,SUM(category.name = 'Horror') AS horror_count
    ,SUM(category.name = 'Action') AS action_count
    ,SUM(category.name = 'Comedy') AS comedy_count
    ,SUM(category.name = 'Sci-Fi') AS scifi_count
FROM
    rental
    JOIN months ON months.num = strftime('%m', rental.rental_date)
    JOIN film_category ON film_category.film_id = rental.inventory_id
    JOIN category ON category.category_id = film_category.category_id
GROUP BY
    months.name
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

-- Recursive CTE (union) example

-- Window function example
-- CASE usage means I can drop the other query with 'intended_audience'
SELECT
    name
    ,SUM(amount) AS summ
    ,CASE NTILE(4) OVER (ORDER BY SUM(amount) DESC)
        WHEN 1 THEN 'Q4'
        WHEN 2 THEN 'Q3'
        WHEN 3 THEN 'Q2'
        WHEN 4 THEN 'Q1'
    END AS quartile
FROM
    category
    JOIN film_category ON category.category_id = film_category.category_id
    JOIN inventory ON film_category.film_id = inventory.film_id
    JOIN rental ON inventory.inventory_id = rental.inventory_id
    JOIN payment ON rental.rental_id = payment.rental_id
GROUP BY
    name
ORDER BY
    summ DESC
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
