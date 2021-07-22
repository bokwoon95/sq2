DROP VIEW IF EXISTS staff_list;
DROP VIEW IF EXISTS sales_by_store;
DROP VIEW IF EXISTS sales_by_film_category;
DROP VIEW IF EXISTS nicer_but_slower_film_list;
DROP VIEW IF EXISTS film_list;
DROP VIEW IF EXISTS customer_list;
DROP VIEW IF EXISTS actor_info;
DROP TABLE IF EXISTS dummy_table_2;
DROP TABLE IF EXISTS dummy_table;
DROP TABLE IF EXISTS payment;
DROP TABLE IF EXISTS rental;
DROP TABLE IF EXISTS inventory;
DROP TABLE IF EXISTS customer;
DROP TABLE IF EXISTS store;
DROP TABLE IF EXISTS staff;
DROP TABLE IF EXISTS film_category;
DROP TABLE IF EXISTS film_actor;
DROP TABLE IF EXISTS film_text;
DROP TABLE IF EXISTS film;
DROP TABLE IF EXISTS language;
DROP TABLE IF EXISTS address;
DROP TABLE IF EXISTS city;
DROP TABLE IF EXISTS country;
DROP TABLE IF EXISTS category;
DROP TABLE IF EXISTS actor;

CREATE TABLE IF NOT EXISTS actor (
    actor_id INTEGER PRIMARY KEY AUTOINCREMENT
    ,first_name TEXT NOT NULL
    ,last_name TEXT NOT NULL
    ,full_name TEXT GENERATED ALWAYS AS (first_name || ' ' || last_name) VIRTUAL
    ,full_name_reversed TEXT GENERATED ALWAYS AS (last_name || ' ' || first_name) STORED
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))
);

CREATE TABLE IF NOT EXISTS category (
    category_id INTEGER PRIMARY KEY
    ,name TEXT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))
);

CREATE TABLE IF NOT EXISTS country (
    country_id INTEGER PRIMARY KEY
    ,country TEXT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))
);

CREATE TABLE IF NOT EXISTS city (
    city_id INTEGER PRIMARY KEY
    ,city TEXT NOT NULL
    ,country_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,FOREIGN KEY (country_id) REFERENCES country (country_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS address (
    address_id INTEGER PRIMARY KEY
    ,address TEXT NOT NULL
    ,address2 TEXT
    ,district TEXT NOT NULL
    ,city_id INT NOT NULL
    ,postal_code TEXT
    ,phone TEXT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,FOREIGN KEY (city_id) REFERENCES city (city_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS language (
    language_id INTEGER PRIMARY KEY
    ,name TEXT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))
);

CREATE TABLE IF NOT EXISTS film (
    film_id INTEGER PRIMARY KEY
    ,title TEXT NOT NULL
    ,description TEXT
    ,release_year INT
    ,language_id INT NOT NULL
    ,original_language_id INT
    ,rental_duration INT NOT NULL DEFAULT 3
    ,rental_rate DECIMAL(4,2) NOT NULL DEFAULT 4.99
    ,length INT
    ,replacement_cost DECIMAL(5,2) NOT NULL DEFAULT 19.99
    ,rating TEXT DEFAULT 'G'
    ,special_features JSON
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,FOREIGN KEY (language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (original_language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,CHECK (release_year >= 1901 AND release_year <= 2155)
    ,CHECK (rating IN ('G','PG','PG-13','R','NC-17'))
);

CREATE VIRTUAL TABLE IF NOT EXISTS film_text USING fts5 (
    title
    ,description
    ,content='film'
    ,content_rowid='film_id'
);

CREATE TABLE IF NOT EXISTS film_actor (
    actor_id INT NOT NULL
    ,film_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,FOREIGN KEY (actor_id) REFERENCES actor (actor_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS film_category (
    film_id INT NOT NULL
    ,category_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (category_id) REFERENCES category (category_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS staff (
    staff_id INTEGER PRIMARY KEY
    ,first_name TEXT NOT NULL
    ,last_name TEXT NOT NULL
    ,address_id INT NOT NULL
    ,email TEXT
    ,store_id INT
    ,active BOOLEAN NOT NULL DEFAULT TRUE
    ,username TEXT NOT NULL
    ,password TEXT
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))
    ,picture BLOB

    ,FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (store_id) REFERENCES store (store_id)
);

CREATE TABLE IF NOT EXISTS store (
    store_id INTEGER PRIMARY KEY
    ,manager_staff_id INT NOT NULL
    ,address_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,FOREIGN KEY (manager_staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS customer (
    customer_id INTEGER PRIMARY KEY
    ,store_id INT NOT NULL
    ,first_name TEXT NOT NULL
    ,last_name TEXT NOT NULL
    ,email TEXT
    ,address_id INT NOT NULL
    ,active BOOLEAN NOT NULL DEFAULT TRUE
    ,create_date DATETIME NOT NULL DEFAULT (DATETIME('now'))
    ,last_update DATETIME DEFAULT (DATETIME('now'))

    ,UNIQUE (email, first_name, last_name)
    ,UNIQUE (email)
    ,FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS inventory (
    inventory_id INTEGER PRIMARY KEY
    ,film_id INT NOT NULL
    ,store_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (store_id) REFERENCES store (store_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS rental (
    rental_id INTEGER PRIMARY KEY
    ,rental_date DATETIME NOT NULL
    ,inventory_id INT NOT NULL
    ,customer_id INT NOT NULL
    ,return_date DATETIME
    ,staff_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,FOREIGN KEY (inventory_id) REFERENCES inventory (inventory_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS payment (
    payment_id INTEGER PRIMARY KEY
    ,customer_id INT NOT NULL
    ,staff_id INT NOT NULL
    ,rental_id INT
    ,amount DECIMAL(5,2) NOT NULL
    ,payment_date DATETIME NOT NULL

    ,FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (rental_id) REFERENCES rental (rental_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE VIEW IF NOT EXISTS actor_info AS
SELECT
    a.actor_id
    ,a.first_name
    ,a.last_name
    ,json_group_object(c.name, (
        SELECT
            json_group_array(f.title)
        FROM
            film AS f
            JOIN film_category AS fc ON fc.film_id = f.film_id
            JOIN film_actor AS fa ON fa.film_id = f.film_id
        WHERE
            fc.category_id = c.category_id
            AND fa.actor_id = a.actor_id
        GROUP BY
            fa.actor_id
    )) AS film_info
FROM
    actor AS a
    LEFT JOIN film_actor AS fa ON fa.actor_id = a.actor_id
    LEFT JOIN film_category AS fc ON fc.film_id = fa.film_id
    LEFT JOIN category AS c ON c.category_id = fc.category_id
GROUP BY
    a.actor_id
    ,a.first_name
    ,a.last_name
;

CREATE VIEW IF NOT EXISTS customer_list AS
SELECT
    cu.customer_id AS id
    ,cu.first_name || ' ' || cu.last_name AS name
    ,a.address, a.postal_code AS "zip code"
    ,a.phone
    ,city.city
    ,country.country,
    CASE
        WHEN cu.active THEN 'active'
        ELSE ''
    END AS notes
    ,cu.store_id AS sid
FROM
    customer AS cu
    JOIN address AS a ON a.address_id = cu.address_id
    JOIN city ON city.city_id = a.city_id
    JOIN country ON country.country_id = city.country_id
;

CREATE VIEW IF NOT EXISTS film_list AS
SELECT
    film.film_id AS fid
    ,film.title
    ,film.description
    ,category.name AS category
    ,film.rental_rate AS price
    ,film.length
    ,film.rating
    ,json_group_array(actor.first_name || ' ' || actor.last_name) AS actors
FROM
    category
    LEFT JOIN film_category ON film_category.category_id = category.category_id
    LEFT JOIN film ON film.film_id = film_category.film_id
    JOIN film_actor ON film_actor.film_id = film.film_id
    JOIN actor ON actor.actor_id = film_actor.actor_id
GROUP BY
    film.film_id
    ,film.title
    ,film.description
    ,category.name
    ,film.rental_rate
    ,film.length
    ,film.rating
;

CREATE VIEW IF NOT EXISTS nicer_but_slower_film_list AS
SELECT
    film.film_id AS fid
    ,film.title
    ,film.description
    ,category.name AS category
    ,film.rental_rate AS price
    ,film.length
    ,film.rating
    ,json_group_array(
        UPPER(SUBSTRING(actor.first_name, 1, 1))
        || LOWER(SUBSTRING(actor.first_name, 2))
        || ' '
        || UPPER(SUBSTRING(actor.last_name, 1, 1))
        || LOWER(SUBSTRING(actor.last_name, 2))
    ) AS actors
FROM
    category
    LEFT JOIN film_category ON film_category.category_id = category.category_id
    LEFT JOIN film ON film.film_id = film_category.film_id
    JOIN film_actor ON film_actor.film_id = film.film_id
    JOIN actor ON actor.actor_id = film_actor.actor_id
GROUP BY
    film.film_id
    ,film.title
    ,film.description
    ,category.name
    ,film.rental_rate
    ,film.length
    ,film.rating
;

CREATE VIEW IF NOT EXISTS sales_by_film_category AS
SELECT
    c.name AS category
    ,SUM(p.amount) AS total_sales
FROM
    payment AS p
    JOIN rental AS r ON r.rental_id = p.rental_id
    JOIN inventory AS i ON i.inventory_id = r.inventory_id
    JOIN film AS f ON f.film_id = i.film_id
    JOIN film_category AS fc ON fc.film_id = f.film_id
    JOIN category AS c ON c.category_id = fc.category_id
GROUP BY
    c.name
ORDER BY
    SUM(p.amount) DESC
;

CREATE VIEW IF NOT EXISTS sales_by_store AS
SELECT
    ci.city || ',' || co.country AS store
    ,m.first_name || ' ' || m.last_name AS manager
    ,SUM(p.amount) AS total_sales
FROM
    payment AS p
    JOIN rental AS r ON r.rental_id = p.rental_id
    JOIN inventory AS i ON i.inventory_id = r.inventory_id
    JOIN store AS s ON s.store_id = i.store_id
    JOIN address AS a ON a.address_id = s.address_id
    JOIN city AS ci ON ci.city_id = a.city_id
    JOIN country AS co ON co.country_id = ci.country_id
    JOIN staff AS m ON m.staff_id = s.manager_staff_id
GROUP BY
    co.country
    ,ci.city
    ,s.store_id
    ,m.first_name
    ,m.last_name
ORDER BY
    co.country
    ,ci.city
;

CREATE VIEW IF NOT EXISTS staff_list AS
SELECT
    s.staff_id AS id
    ,s.first_name || ' ' || s.last_name AS name
    ,a.address
    ,a.postal_code AS "zip code"
    ,a.phone
    ,ci.city
    ,co.country
    ,s.store_id AS sid
FROM
    staff AS s
    JOIN address AS a ON a.address_id = s.address_id
    JOIN city AS ci ON ci.city_id = a.city_id
    JOIN country AS co ON co.country_id = ci.country_id
;

CREATE INDEX IF NOT EXISTS actor_last_name_idx ON actor (last_name);

CREATE INDEX IF NOT EXISTS city_country_id_idx ON city (country_id);

CREATE INDEX IF NOT EXISTS address_city_id_idx ON address (city_id);

CREATE INDEX IF NOT EXISTS film_title_idx ON film (title);

CREATE INDEX IF NOT EXISTS film_language_id_idx ON film (language_id);

CREATE INDEX IF NOT EXISTS film_original_language_id_idx ON film (original_language_id);

CREATE UNIQUE INDEX IF NOT EXISTS film_actor_actor_id_film_id_idx ON film_actor (actor_id, film_id);

CREATE UNIQUE INDEX IF NOT EXISTS store_manager_staff_id_idx ON store (manager_staff_id);

CREATE INDEX IF NOT EXISTS customer_store_id_idx ON customer (store_id);

CREATE INDEX IF NOT EXISTS customer_last_name_idx ON customer (last_name);

CREATE INDEX IF NOT EXISTS customer_address_id_idx ON customer (address_id);

CREATE INDEX IF NOT EXISTS film_actor_film_id_idx ON film_actor (film_id);

CREATE INDEX IF NOT EXISTS inventory_store_id_film_id_idx ON inventory (store_id, film_id);

CREATE UNIQUE INDEX IF NOT EXISTS rental_rental_date_inventory_id_customer_id_idx ON rental (rental_date, inventory_id, customer_id);

CREATE INDEX IF NOT EXISTS rental_inventory_id_idx ON rental (inventory_id);

CREATE INDEX IF NOT EXISTS rental_customer_id_idx ON rental (customer_id);

CREATE INDEX IF NOT EXISTS rental_staff_id_idx ON rental (staff_id);

CREATE INDEX IF NOT EXISTS payment_customer_id_idx ON payment (customer_id);

CREATE INDEX IF NOT EXISTS payment_staff_id_idx ON payment (staff_id);

CREATE TRIGGER actor_last_update_after_update_trg AFTER UPDATE ON actor BEGIN
    UPDATE actor SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;

CREATE TRIGGER category_last_update_after_update_trg AFTER UPDATE ON category BEGIN
    UPDATE category SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;

CREATE TRIGGER country_last_update_after_update_trg AFTER UPDATE ON country BEGIN
    UPDATE country SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;

CREATE TRIGGER city_last_update_after_update_trg AFTER UPDATE ON city BEGIN
    UPDATE city SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;

CREATE TRIGGER address_last_update_after_update_trg AFTER UPDATE ON address BEGIN
    UPDATE address SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;

CREATE TRIGGER language_last_update_after_update_trg AFTER UPDATE ON language BEGIN
    UPDATE language SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;

CREATE TRIGGER film_last_update_after_update_trg AFTER UPDATE ON film BEGIN
    UPDATE film SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;

CREATE TRIGGER film_fts5_after_insert_trg AFTER INSERT ON film BEGIN
    INSERT INTO film_text (ROWID, title, description) VALUES (NEW.film_id, NEW.title, NEW.description);
END;

CREATE TRIGGER film_fts5_after_delete_trg AFTER DELETE ON film BEGIN
    INSERT INTO film_text (film_text, ROWID, title, description) VALUES ('delete', OLD.film_id, OLD.title, OLD.description);
END;

CREATE TRIGGER film_fts5_after_update_trg AFTER UPDATE ON film BEGIN
    INSERT INTO film_text (film_text, ROWID, title, description) VALUES ('delete', OLD.film_id, OLD.title, OLD.description);
    INSERT INTO film_text (film_text, ROWID, title, description) VALUES (NEW.film_id, NEW.title, NEW.description);
END;

CREATE TRIGGER film_actor_last_update_after_update_trg AFTER UPDATE ON film_actor BEGIN
    UPDATE film_actor SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;

CREATE TRIGGER film_category_last_update_after_update_trg AFTER UPDATE ON film_category BEGIN
    UPDATE film_category SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;

CREATE TRIGGER staff_last_update_after_update_trg AFTER UPDATE ON staff BEGIN
    UPDATE staff SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;

CREATE TRIGGER store_last_update_after_update_trg AFTER UPDATE ON store BEGIN
    UPDATE store SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;

CREATE TRIGGER customer_last_update_after_update_trg AFTER UPDATE ON customer BEGIN
    UPDATE customer SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;

CREATE TRIGGER inventory_last_update_after_update_trg AFTER UPDATE ON inventory BEGIN
    UPDATE inventory SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;

CREATE TRIGGER rental_last_update_after_update_trg AFTER UPDATE ON rental BEGIN
    UPDATE rental SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;
