DROP VIEW IF EXISTS staff_list;
DROP VIEW IF EXISTS sales_by_store;
DROP VIEW IF EXISTS sales_by_film_category;
DROP VIEW IF EXISTS nicer_but_slower_film_list;
DROP VIEW IF EXISTS film_list;
DROP VIEW IF EXISTS customer_list;
DROP VIEW IF EXISTS actor_info;

CREATE VIEW actor_info AS
SELECT
    a.actor_id
    ,a.first_name
    ,a.last_name
    ,jsonb_object_agg(c.name, (
        SELECT
            jsonb_agg(f.title)
        FROM
            film AS f
            JOIN film_category AS fc ON f.film_id = fc.film_id
            JOIN film_actor AS fa ON f.film_id = fa.film_id
        WHERE
            fc.category_id = c.category_id AND fa.actor_id = a.actor_id
        GROUP BY
            fa.actor_id
    )) AS film_info
FROM
    actor AS a
    LEFT JOIN film_actor AS fa ON a.actor_id = fa.actor_id
    LEFT JOIN film_category AS fc ON fa.film_id = fc.film_id
    LEFT JOIN category AS c ON fc.category_id = c.category_id
GROUP BY
    a.actor_id
    ,a.first_name
    ,a.last_name
;

CREATE VIEW customer_list AS
SELECT
    cu.customer_id AS id
    ,cu.first_name || ' ' || cu.last_name AS name
    ,a.address
    ,a.postal_code AS "zip code"
    ,a.phone
    ,city.city
    ,country.country
    ,CASE WHEN cu.active THEN 'active' ELSE '' END AS notes
    ,cu.store_id AS sid
FROM
    customer AS cu
    JOIN address AS a ON cu.address_id = a.address_id
    JOIN city ON a.city_id = city.city_id
    JOIN country ON city.country_id = country.country_id
;

CREATE VIEW film_list AS
SELECT
    film.film_id AS fid
    ,film.title
    ,film.description
    ,category.name AS category
    ,film.rental_rate AS price
    ,film.length
    ,film.rating
    ,jsonb_agg(actor.first_name || ' ' || actor.last_name) AS actors
FROM
    category
    LEFT JOIN film_category ON category.category_id = film_category.category_id
    LEFT JOIN film ON film_category.film_id = film.film_id
    JOIN film_actor ON film.film_id = film_actor.film_id
    JOIN actor ON film_actor.actor_id = actor.actor_id
GROUP BY
    film.film_id
    ,film.title
    ,film.description
    ,category.name
    ,film.rental_rate
    ,film.length
    ,film.rating
;

CREATE VIEW nicer_but_slower_film_list AS
SELECT
    film.film_id AS fid
    ,film.title
    ,film.description
    ,category.name AS category
    ,film.rental_rate AS price
    ,film.length
    ,film.rating
    ,jsonb_agg(
        UPPER(SUBSTRING(actor.first_name, 1, 1))
        || LOWER(SUBSTRING(actor.first_name, 2))
        || ' '
        || UPPER(SUBSTRING(actor.last_name, 1, 1))
        || LOWER(SUBSTRING(actor.last_name, 2))
    ) AS actors
FROM
    category
    LEFT JOIN film_category ON category.category_id = film_category.category_id
    LEFT JOIN film ON film_category.film_id = film.film_id
    JOIN film_actor ON film.film_id = film_actor.film_id
    JOIN actor ON film_actor.actor_id = actor.actor_id
GROUP BY
    film.film_id
    ,film.title
    ,film.description
    ,category.name
    ,film.rental_rate
    ,film.length
    ,film.rating
;

CREATE VIEW sales_by_film_category AS
SELECT
    c.name AS category
    ,SUM(p.amount) AS total_sales
FROM
    payment p
    JOIN rental r ON p.rental_id = r.rental_id
    JOIN inventory i ON r.inventory_id = i.inventory_id
    JOIN film f ON i.film_id = f.film_id
    JOIN film_category fc ON f.film_id = fc.film_id
    JOIN category c ON fc.category_id = c.category_id
GROUP BY
    c.name
ORDER BY
    SUM(p.amount) DESC
;

CREATE VIEW sales_by_store AS
SELECT
    c.city || ',' || cy.country AS store
    ,m.first_name || ' ' || m.last_name AS manager
    ,SUM(p.amount) AS total_sales
FROM
    payment AS p
    JOIN rental AS r ON p.rental_id = r.rental_id
    JOIN inventory AS i ON r.inventory_id = i.inventory_id
    JOIN store AS s ON i.store_id = s.store_id
    JOIN address AS a ON s.address_id = a.address_id
    JOIN city AS c ON a.city_id = c.city_id
    JOIN country AS cy ON c.country_id = cy.country_id
    JOIN staff AS m ON s.manager_staff_id = m.staff_id
GROUP BY
    cy.country
    ,c.city
    ,s.store_id
    ,m.first_name
    ,m.last_name
ORDER BY
    cy.country
    ,c.city
;

CREATE VIEW staff_list AS
SELECT
    s.staff_id AS id
    ,s.first_name || ' ' || s.last_name AS name
    ,a.address
    ,a.postal_code AS "zip code"
    ,a.phone
    ,city.city
    ,country.country
    ,s.store_id AS sid
FROM
    staff AS s
    JOIN address AS a ON s.address_id = a.address_id
    JOIN city ON a.city_id = city.city_id
    JOIN country ON city.country_id = country.country_id
;
