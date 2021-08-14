-- Q1) Find all distinct actor last names ordered by last name. Show only the
-- top 5 results.
SELECT DISTINCT last_name FROM actor ORDER BY last_name LIMIT 5;

-- Q2) Find if there is any actor with first name 'SCARLETT' or last name
-- 'JOHANSSON'.
SELECT EXISTS(SELECT 1 FROM actor WHERE first_name = 'SCARLETT' OR last_name = 'JOHANSSON');

-- Q3) Find the number of distinct actor last names.
SELECT COUNT(DISTINCT last_name) FROM actor;

-- Q4) Find all actors whose last name contain the letters 'GEN', ordered by
-- actor_id.
SELECT
    actor_id
    ,first_name
    ,last_name
    ,last_update
FROM
    actor
WHERE
    last_name LIKE '%GEN%'
ORDER BY
    actor_id
;

-- Q5) Find actor last names that occur only once in the database, ordered by
-- last_name. Show only the top 5 results.
SELECT last_name FROM actor GROUP BY last_name HAVING COUNT(*) = 1 ORDER BY last_name LIMIT 5;

-- Q6) Find all the cities (and their country) of the where country is either
-- 'Egypt', 'Greece' or 'Puerto Rico', ordered by country_name and city_name.
SELECT
    country.country_id
    ,country.country
    ,country.last_update AS country_last_update
    ,city.city_id
    ,city.city
    ,city.last_update AS city_last_update
FROM
    city
    JOIN country ON country.country_id = city.country_id
WHERE
    country.country IN ('Egypt', 'Greece', 'Puerto Rico')
ORDER BY
    country.country, city.city
;

-- Q7) List films together with their audience and length type, ordered by
-- title.
-- The audiences are:
-- - rating = 'G' then 'family'
-- - rating = 'PG' or rating = 'PG-13' then 'teens'
-- - rating = 'R' or rating = 'NC-17' then 'adults'
-- The length types are:
-- - length <= 60 then 'short'
-- - length > 60 and length <= 120 then 'medium'
-- - length > 120 then 'long'
-- Return the film_title, rating, length, audience and length_type. Show only the
-- top 10 results.
SELECT
    film_id
    ,title
    ,description
    ,release_year
    ,rental_duration
    ,rental_rate
    ,length
    ,replacement_cost
    ,rating
    ,special_features
    ,last_update
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

-- Q8) Find the films whose total number of actors is above the average,
-- ordered by descending actor count. Show only the top 10 results.
WITH film_stats AS (
    SELECT film_id, COUNT(*) AS actor_count
    FROM film_actor
    GROUP BY film_id
)
SELECT
    film.film_id
    ,film.title
    ,film.description
    ,film.release_year
    ,film.rental_duration
    ,film.rental_rate
    ,film.length
    ,film.replacement_cost
    ,film.rating
    ,film.special_features
    ,film.last_update
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

-- Q9) List the film categories and their total revenue (rounded to 2
-- decimals), ordered by descending revenue. Include the rank of that category
-- and the quartile it belongs to, relative to the other categories.
SELECT
    category.category_id
    ,category.name
    ,category.last_update
    ,ROUND(SUM(payment.amount), 2) AS revenue
    ,RANK() OVER (ORDER BY SUM(payment.amount) DESC) AS rank
    ,NTILE(4) OVER (ORDER BY SUM(payment.amount) ASC) AS quartile
FROM
    category
    JOIN film_category ON category.category_id = film_category.category_id
    JOIN inventory ON film_category.film_id = inventory.film_id
    JOIN rental ON inventory.inventory_id = rental.inventory_id
    JOIN payment ON rental.rental_id = payment.rental_id
GROUP BY
    category.category_id
    ,category.name
    ,category.last_update
ORDER BY
    revenue DESC
;

-- Q10) Find the total number of 'Horror' films, 'Action' films, 'Comedy' films
-- and 'Sci-Fi' films rented out every month between '2005-03-01' and
-- '2006-02-01', ordered by month. Months with 0 rentals should also be
-- included.
WITH RECURSIVE dates (date_value) AS (
    SELECT '2005-03-01'::DATE
    UNION ALL
    SELECT (date_value + '1 month'::INTERVAL)::DATE FROM dates WHERE date_value < '2006-02-01'
)
SELECT
    to_char(dates.date_value, 'YYYY FMMonth') AS month
    ,COUNT(CASE category.name WHEN 'Horror' THEN 1 END) AS horror_count
    ,COUNT(CASE category.name WHEN 'Action' THEN 1 END) AS action_count
    ,COUNT(CASE category.name WHEN 'Comedy' THEN 1 END) AS comedy_count
    ,COUNT(CASE category.name WHEN 'Sci-Fi' THEN 1 END) AS scifi_count
FROM
    dates
    LEFT JOIN rental ON to_char(rental.rental_date, 'YYYY FMMonth') = to_char(dates.date_value, 'YYYY FMMonth')
    LEFT JOIN film_category ON film_category.film_id = rental.inventory_id
    LEFT JOIN category ON category.category_id = film_category.category_id
GROUP BY
    dates.date_value
ORDER BY
    dates.date_value
;
