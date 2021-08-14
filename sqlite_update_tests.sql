UPDATE film
SET description = description || ' starring THORA TEMPLE'
WHERE EXISTS (
    SELECT 1
    FROM
        film_actor
        JOIN actor ON actor.actor_id = film_actor.actor_id
    WHERE
        film_actor.film_id = film.film_id
        AND actor.first_name = 'THORA'
        AND actor.last_name = 'TEMPLE'
)
RETURNING film_id
;

UPDATE film
SET description = substr(description, 0, LENGTH(description) - LENGTH(' starring THORA TEMPLE') + 1)
WHERE EXISTS (
    SELECT 1
    FROM
        film_actor
        JOIN actor ON actor.actor_id = film_actor.actor_id
    WHERE
        film_actor.film_id = film.film_id
        AND actor.first_name = 'THORA'
        AND actor.last_name = 'TEMPLE'
)
RETURNING film_id
;

UPDATE film
WHERE EXISTS(
    SELECT 1
    FROM
)
RETURNING film_id
;

-- 318 BRIAN WYMAN has rented 12 films. We can use this for multi-table update
-- (update the customer and the film at the same time).

-- multi-update can update existing records in the database; multi-delete
-- cannot because it will cascade delete everything else. So we can only delete
-- the dummy data that we inserted in. We can use Norway as the country, 3
-- cities, and 9 addresses. Then delete them all at once.
