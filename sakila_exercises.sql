------------
-- SELECT --
------------

-- TODO: need a formalized structure for testing the following queries. Results
-- must always be the same, no matter the dialect. Only for
-- insert/update/delete are there some dialect-specific queries.

-- TODO: also need OLTP queries where you scan into a nested structure. Either
-- aggregate manually in the application or aggregate into json and unmarshal
-- in the application.

-- Find all distinct actor last names ordered by last name. Show only the top 4
-- results.
SELECT DISTINCT last_name FROM actor ORDER BY last_name LIMIT 5;

-- Find if there is any actor with first name 'SCARLETT' or last name
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

-- Find all the cities of the countries Egypt, Greece and Puerto Rico, ordered by
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

-- List films with their length classification, ordered by title. The classifications are:
-- length <= 60 then 'short', length > 60 and length <= 120 then 'medium',
-- length > 120 then 'long'. Return the title, length, and length_classification. Show
-- only the top 10 results.
SELECT
    title
    ,rating
    ,length
    ,CASE rating
        WHEN 'G' THEN 'family'
        WHEN 'PG' THEN 'teens'
        WHEN 'PG-13' THEN 'teens'
        WHEN 'R' THEN 'adults'
        WHEN 'NC-17' THEN 'adults'
    END AS audience
    ,CASE
        WHEN length <= 60 THEN 'short'
        WHEN length > 60 AND length <= 120 THEN 'medium'
        ELSE 'long'
    END AS length_type
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
    ,SUM(CASE category.name WHEN 'Horror' THEN 1 END) AS horror_count
    ,SUM(CASE category.name WHEN 'Action' THEN 1 END) AS action_count
    ,SUM(CASE category.name WHEN 'Comedy' THEN 1 END) AS comedy_count
    ,SUM(CASE category.name WHEN 'Sci-Fi' THEN 1 END) AS scifi_count
FROM
    rental
    JOIN months ON months.num = strftime('%m', rental.rental_date)
    JOIN film_category ON film_category.film_id = rental.inventory_id
    JOIN category ON category.category_id = film_category.category_id
GROUP BY
    months.name
;

-- Find customers who have rented the most items, ordered by descending rental
-- count. Return the first_name, last_name and rental_count. Show only the top
-- 10 results.
SELECT
    customer.first_name
    ,customer.last_name
    ,COUNT(*) AS rental_count
FROM
    customer
    JOIN rental ON rental.customer_id = customer.customer_id
GROUP BY
    customer.customer_id
ORDER BY
    rental_count DESC
LIMIT
    10
;

-- https://stackoverflow.com/questions/67080935/how-can-i-get-the-desired-results-from-the-sakila-database-using-sql
-- Find the films whose total number of actors is above the average, ordered by
-- descending actor count, ascending film title. Return the film title and the actor_count. Show only
-- the top 10 results.
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
    film_stats.actor_count DESC
    ,film.title ASC
LIMIT
    10
;

-- Window function example
-- CASE usage means I can drop the other query with 'intended_audience'
-- List the film categories and their total revenue (rounded to nearest
-- integer), ordered by descending revenue. Return the name, revenue, the
-- rank of that category and the quartile it belongs to (relative to the other
-- categories).
SELECT
    category.name
    ,ROUND(SUM(payment.amount)) AS revenue
    ,RANK() OVER (ORDER BY SUM(payment.amount) DESC) AS rank
    ,NTILE(4) OVER (ORDER BY SUM(payment.amount) ASC) AS quartile
FROM
    category
    JOIN film_category ON category.category_id = film_category.category_id
    JOIN inventory ON film_category.film_id = inventory.film_id
    JOIN rental ON inventory.inventory_id = rental.inventory_id
    JOIN payment ON rental.rental_id = payment.rental_id
GROUP BY
    name
ORDER BY
    revenue DESC
;

-- Recursive CTE (union) example
WITH RECURSIVE dates (date_value) AS (
    SELECT DATE('2005-03-01')
    UNION ALL
    SELECT DATE(date_value, '+1 month') FROM dates WHERE date_value < '2006-02-01'
)
,months (num, name) AS (
    VALUES ('01', 'January'), ('02', 'February'), ('03', 'March'),
        ('04', 'April'), ('05', 'May'), ('06', 'June'),
        ('07', 'July'), ('08', 'August'), ('09', 'September'),
        ('10', 'October'), ('11', 'November'), ('12', 'December')
)
SELECT
    strftime('%Y', dates.date_value) || ' ' || months.name AS rental_month
    ,COUNT(rental.rental_id) AS rental_count
FROM
    dates
    JOIN months ON months.num = strftime('%m', dates.date_value)
    LEFT JOIN rental ON strftime('%Y %m', rental.rental_date) = strftime('%Y %m', dates.date_value)
GROUP BY
    strftime('%Y %m', dates.date_value)
ORDER BY
    dates.date_value
;

-- Find the total number of 'Horror' films, 'Action' films, 'Comedy' films and
-- 'Sci-Fi' films rented out every month between '2005-03-01' and '2006-02-01',
-- ordered by month. Months with 0 rentals should also be included. Return the
-- month, horror_count, action_count, comedy_count and scifi_count.
WITH RECURSIVE dates (date_value) AS (
    SELECT DATE('2005-03-01')
    UNION ALL
    SELECT DATE(date_value, '+1 month') FROM dates WHERE date_value < '2006-02-01'
)
,months (num, name) AS (
    VALUES ('01', 'January'), ('02', 'February'), ('03', 'March'),
        ('04', 'April'), ('05', 'May'), ('06', 'June'),
        ('07', 'July'), ('08', 'August'), ('09', 'September'),
        ('10', 'October'), ('11', 'November'), ('12', 'December')
)
SELECT
    strftime('%Y', dates.date_value) || ' ' || months.name AS month
    ,COUNT(CASE category.name WHEN 'Horror' THEN 1 END) AS horror_count
    ,COUNT(CASE category.name WHEN 'Action' THEN 1 END) AS action_count
    ,COUNT(CASE category.name WHEN 'Comedy' THEN 1 END) AS comedy_count
    ,COUNT(CASE category.name WHEN 'Sci-Fi' THEN 1 END) AS scifi_count
    ,COUNT(*) OVER () AS count
FROM
    dates
    JOIN months ON months.num = strftime('%m', dates.date_value)
    LEFT JOIN rental ON strftime('%Y %m', rental.rental_date) = strftime('%Y %m', dates.date_value)
    LEFT JOIN film_category ON film_category.film_id = rental.inventory_id
    LEFT JOIN category ON category.category_id = film_category.category_id
GROUP BY
    dates.date_value
ORDER BY
    dates.date_value
;

------------
-- INSERT --
------------

-- data modifying queries should not use a common query testing framework, because operations must be rolled back. Better to keep everything constrained to one function.

-- Insert and get ID (sqlite uses both RETURNING and LastInsertID)

-- Insert the same row with the ID but ignore

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
