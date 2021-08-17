INSERT INTO customer
    (store_id, first_name, last_name, email, address_id)
VALUES
    (1, 'REGINA', 'TATE', 'regina_tate@email.com', 1)
;

SELECT EXISTS(
    SELECT 1
    FROM customer
    WHERE
        store_id = $id
        AND (store_id, first_name, last_name, email, address_id) = (1, 'REGINA', 'TATE', 'regina_tate@email.com', 1)
);

INSERT INTO customer
    (store_id, first_name, last_name, email, address_id)
VALUES
    (1, 'REGINA', 'TATE', 'regina_tate@email.com', 1)
ON CONFLICT DO NOTHING
;

-- assert last_insert_id is nothing

INSERT INTO customer
    (store_id, first_name, last_name, email, address_id)
VALUES
    (1, 'ANTHONY', 'CURTIS', 'anthony_curtis@email.com', 1)
ON CONFLICT (email) DO UPDATE SET
    store_id = EXCLUDED.store_id
    ,first_name = EXCLUDED.first_name
    ,last_name = EXCLUDED.last_name
    ,address_id = EXCLUDED.address_id
;

SELECT EXISTS(
    SELECT 1
    FROM customer
    WHERE
        store_id = $id
        AND (store_id, first_name, last_name, email, address_id) = (1, 'ANTHONY', 'CURTIS', 'anthony_curtis@email.com', 1)
);

INSERT INTO customer
    (store_id, first_name, last_name, email, address_id)
VALUES
    (1, 'Anthony', 'Curtis', 'anthony_curtis@email.com', 1)
ON CONFLICT (email) DO UPDATE SET
    store_id = EXCLUDED.store_id
    ,first_name = EXCLUDED.first_name
    ,last_name = EXCLUDED.last_name
    ,address_id = EXCLUDED.address_id
;

SELECT EXISTS(
    SELECT 1
    FROM customer
    WHERE
        store_id = $id
        AND (store_id, first_name, last_name, email, address_id) = (1, 'Anthony', 'Curtis', 'anthony_curtis@email.com', 1)
);

INSERT INTO customer
    (store_id, first_name, last_name, email, address_id)
VALUES
    (1, 'JULIA', 'HAYWARD', 'julia_hayward@email.com', 1)
    ,(1, 'DUNCAN', 'PEARSON', 'duncan_pearson@email.com', 1)
    ,(1, 'IDA', 'WATKINS', 'ida_watkins@email.com', 1)
    ,(1, 'THOMAS', 'BINDER', 'thomas_binder@email.com', 1)
RETURNING
    customer_id
;

SELECT
    EXISTS(SELECT 1 FROM customer WHERE /* JULIA HAYWARD */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* DUNCAN PEARSON */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* IDA WATKINS */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* THOMAS BINDER */)
;

INSERT INTO customer
    (store_id, first_name, last_name, email, address_id)
VALUES
    (1, 'JULIA', 'HAYWARD', 'julia_hayward@email.com', 1)
    ,(1, 'DUNCAN', 'PEARSON', 'duncan_pearson@email.com', 1)
    ,(1, 'IDA', 'WATKINS', 'ida_watkins@email.com', 1)
    ,(1, 'THOMAS', 'BINDER', 'thomas_binder@email.com', 1)
    ,(1, 'ASTRID', 'SILVA', 'astrid_silva@email.com', 1)
    ,(1, 'HARPER', 'CRAIG', 'harper_craig@email.com', 1)
    ,(1, 'SAMANTHA', 'STEVENSON', 'samantha_stevenson@email.com', 1)
    ,(1, 'PHILIP', 'REID', 'philip_reid@email.com', 1)
ON CONFLICT DO NOTHING
RETURNING
    customer_id
;

-- assert the only ids were returned for ASTRID SILVA, HARPER CRAIG, SAMANTHA STEVENSON and PHILIP REID

SELECT
    EXISTS(SELECT 1 FROM customer WHERE /* JULIA HAYWARD */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* DUNCAN PEARSON */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* IDA WATKINS */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* THOMAS BINDER */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* ASTRID SILVA */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* HARPER CRAIG */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* SAMANTHA STEVENSON */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* PHILIP REID */)
;

INSERT INTO customer
    (store_id, first_name, last_name, email, address_id)
VALUES
    (1, 'Julia', 'Hayward', 'julia_hayward@email.com', 1)
    ,(1, 'Duncan', 'Pearson', 'duncan_pearson@email.com', 1)
    ,(1, 'Ida', 'Watkins', 'ida_watkins@email.com', 1)
    ,(1, 'Thomas', 'Binder', 'thomas_binder@email.com', 1)
ON CONFLICT DO UPDATE SET
    store_id = EXCLUDED.store_id
    ,first_name = EXCLUDED.first_name
    ,last_name = EXCLUDED.last_name
    ,address_id = EXCLUDED.address_id
RETURNING
    customer_id
;

SELECT
    EXISTS(SELECT 1 FROM customer WHERE /* JULIA HAYWARD */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* DUNCAN PEARSON */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* IDA WATKINS */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* THOMAS BINDER */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* ASTRID SILVA */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* HARPER CRAIG */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* SAMANTHA STEVENSON */)
    AND EXISTS(SELECT 1 FROM customer WHERE /* PHILIP REID */)
;

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
    JOIN store ON store.store_id = inventory.store_id
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
