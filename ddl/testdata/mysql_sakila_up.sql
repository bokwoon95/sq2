CREATE TABLE IF NOT EXISTS actor (
    actor_id INT AUTO_INCREMENT
    ,first_name VARCHAR(45) NOT NULL
    ,last_name VARCHAR(45) NOT NULL
    ,full_name VARCHAR(255) GENERATED ALWAYS AS (CONCAT(first_name, ' ', last_name)) VIRTUAL
    ,full_name_reversed VARCHAR(255) GENERATED ALWAYS AS (CONCAT(last_name, ' ', first_name)) STORED
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (actor_id)
    ,INDEX actor_last_name_idx (last_name)
);

CREATE TABLE IF NOT EXISTS address (
    address_id INT AUTO_INCREMENT
    ,address VARCHAR(50) NOT NULL
    ,address2 VARCHAR(50)
    ,district VARCHAR(20) NOT NULL
    ,city_id INT NOT NULL
    ,postal_code VARCHAR(10)
    ,phone VARCHAR(20) NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (address_id)
    ,INDEX address_city_id_idx (city_id)
);

CREATE TABLE IF NOT EXISTS category (
    category_id INT AUTO_INCREMENT
    ,name VARCHAR(25) NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (category_id)
);

CREATE TABLE IF NOT EXISTS city (
    city_id INT AUTO_INCREMENT
    ,city VARCHAR(50) NOT NULL
    ,country_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (city_id)
    ,INDEX city_country_id_idx (country_id)
);

CREATE TABLE IF NOT EXISTS country (
    country_id INT AUTO_INCREMENT
    ,country VARCHAR(50) NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (country_id)
);

CREATE TABLE IF NOT EXISTS customer (
    customer_id INT AUTO_INCREMENT
    ,store_id INT NOT NULL
    ,first_name VARCHAR(45) NOT NULL
    ,last_name VARCHAR(45) NOT NULL
    ,email VARCHAR(50)
    ,address_id INT NOT NULL
    ,active BOOLEAN NOT NULL DEFAULT TRUE
    ,data JSON
    ,create_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT customer_email_first_name_last_name_key UNIQUE (email, first_name, last_name)
    ,PRIMARY KEY (customer_id)
    ,CONSTRAINT customer_email_key UNIQUE (email)
    ,INDEX customer_store_id_idx (store_id)
    ,INDEX customer_last_name_idx (last_name)
    ,INDEX customer_address_id_idx (address_id)
);

CREATE TABLE IF NOT EXISTS film (
    film_id INT AUTO_INCREMENT
    ,title VARCHAR(255) NOT NULL
    ,description TEXT
    ,release_year INT
    ,language_id INT NOT NULL
    ,original_language_id INT
    ,rental_duration INT NOT NULL DEFAULT 3
    ,rental_rate DECIMAL(4,2) NOT NULL DEFAULT 4.99
    ,length INT
    ,replacement_cost DECIMAL(5,2) NOT NULL DEFAULT 19.99
    ,rating ENUM('G','PG','PG-13','R','NC-17') DEFAULT 'G'
    ,special_features JSON
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (film_id)
    ,CONSTRAINT film_release_year_check CHECK (release_year >= 1901 AND release_year <= 2155)
    ,INDEX film_title_idx (title)
    ,INDEX film_language_id_idx (language_id)
    ,INDEX film_original_language_id_idx (original_language_id)
);

CREATE TABLE IF NOT EXISTS film_actor (
    film_id INT NOT NULL
    ,actor_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,UNIQUE INDEX film_actor_actor_id_film_id_idx (actor_id, film_id)
    ,INDEX film_actor_film_id_idx (film_id)
);

CREATE TABLE IF NOT EXISTS film_actor_review (
    film_id INT
    ,actor_id INT
    ,review_title VARCHAR(50) NOT NULL DEFAULT '' COLLATE latin1_swedish_ci
    ,review_body VARCHAR(255) NOT NULL DEFAULT ''
    ,metadata JSON
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ,delete_date DATETIME

    ,PRIMARY KEY (film_id, actor_id)
    ,CONSTRAINT film_actor_review_check CHECK (LENGTH(review_body) > LENGTH(review_title))
    ,INDEX film_actor_review_misc (film_id, (SUBSTR(review_body, 2, 10)), (CONCAT(review_title, ' abcd')), (CAST(metadata->>'$.score' AS SIGNED)))
);

CREATE TABLE IF NOT EXISTS film_category (
    film_id INT NOT NULL
    ,category_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS film_text (
    film_id INT NOT NULL
    ,title VARCHAR(255)
    ,description TEXT

    ,PRIMARY KEY (film_id)
    ,FULLTEXT INDEX film_text_title_description_idx (title, description)
);

CREATE TABLE IF NOT EXISTS inventory (
    inventory_id INT AUTO_INCREMENT
    ,film_id INT NOT NULL
    ,store_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (inventory_id)
    ,INDEX inventory_store_id_film_id_idx (store_id, film_id)
);

CREATE TABLE IF NOT EXISTS language (
    language_id INT AUTO_INCREMENT
    ,name CHAR(20) NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (language_id)
);

CREATE TABLE IF NOT EXISTS payment (
    payment_id INT AUTO_INCREMENT
    ,customer_id INT NOT NULL
    ,staff_id INT NOT NULL
    ,rental_id INT
    ,amount DECIMAL(5,2) NOT NULL
    ,payment_date DATETIME NOT NULL

    ,PRIMARY KEY (payment_id)
    ,INDEX payment_customer_id_idx (customer_id)
    ,INDEX payment_staff_id_idx (staff_id)
);

CREATE TABLE IF NOT EXISTS rental (
    rental_id INT AUTO_INCREMENT
    ,rental_date DATETIME NOT NULL
    ,inventory_id INT NOT NULL
    ,customer_id INT NOT NULL
    ,return_date DATETIME
    ,staff_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (rental_id)
    ,UNIQUE INDEX rental_rental_date_inventory_id_customer_id_idx (rental_date, inventory_id, customer_id)
    ,INDEX rental_inventory_id_idx (inventory_id)
    ,INDEX rental_customer_id_idx (customer_id)
    ,INDEX rental_staff_id_idx (staff_id)
);

CREATE TABLE IF NOT EXISTS staff (
    staff_id INT AUTO_INCREMENT
    ,first_name VARCHAR(45) NOT NULL
    ,last_name VARCHAR(45) NOT NULL
    ,address_id INT NOT NULL
    ,email VARCHAR(50)
    ,store_id INT
    ,active BOOLEAN NOT NULL DEFAULT TRUE
    ,username VARCHAR(16) NOT NULL
    ,password VARCHAR(40)
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ,picture BLOB

    ,PRIMARY KEY (staff_id)
);

CREATE TABLE IF NOT EXISTS store (
    store_id INT AUTO_INCREMENT
    ,manager_staff_id INT NOT NULL
    ,address_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (store_id)
    ,UNIQUE INDEX store_manager_staff_id_idx (manager_staff_id)
);

CREATE OR REPLACE VIEW actor_info AS SELECT a.actor_id, a.first_name, a.last_name, json_objectagg(c.name, (SELECT json_arrayagg(f.title) FROM film AS f JOIN film_category AS fc ON fc.film_id = f.film_id JOIN film_actor AS fa ON fa.film_id = f.film_id WHERE fc.category_id = c.category_id AND fa.actor_id = a.actor_id GROUP BY fa.actor_id)) AS film_info FROM actor AS a LEFT JOIN film_actor AS fa ON fa.actor_id = a.actor_id LEFT JOIN film_category AS fc ON fc.film_id = fa.film_id LEFT JOIN category AS c ON c.category_id = fc.category_id GROUP BY a.actor_id, a.first_name, a.last_name;

CREATE OR REPLACE VIEW customer_list AS SELECT cu.customer_id AS id, CONCAT(cu.first_name, ' ', cu.last_name) AS name, a.address, a.postal_code AS `zip code`, a.phone, city.city, country.country, CASE WHEN cu.active THEN 'active' ELSE '' END AS notes, cu.store_id AS sid FROM customer AS cu JOIN address AS a ON a.address_id = cu.address_id JOIN city ON city.city_id = a.city_id JOIN country ON country.country_id = city.country_id;

CREATE OR REPLACE VIEW film_list AS SELECT film.film_id AS fid, film.title, film.description, category.name AS category, film.rental_rate AS price, film.length, film.rating, json_arrayagg(CONCAT(actor.first_name, ' ', actor.last_name)) AS actors FROM category LEFT JOIN film_category ON film_category.category_id = category.category_id LEFT JOIN film ON film.film_id = film_category.film_id JOIN film_actor ON film_actor.film_id = film.film_id JOIN actor ON actor.actor_id = film_actor.actor_id GROUP BY film.film_id, film.title, film.description, category.name, film.rental_rate, film.length, film.rating;

CREATE OR REPLACE VIEW full_address AS SELECT country.country_id, city.city_id, address.address_id, country.country, city.city, address.address, address.address2, address.district, address.postal_code, address.phone, address.last_update FROM address JOIN city ON city.city_id = address.city_id JOIN country ON country.country_id = city.country_id;

CREATE OR REPLACE VIEW nicer_but_slower_film_list AS SELECT film.film_id AS fid, film.title, film.description, category.name AS category, film.rental_rate AS price, film.length, film.rating, json_arrayagg(CONCAT(UPPER(SUBSTRING(actor.first_name, 1, 1)), LOWER(SUBSTRING(actor.first_name, 2)), ' ', UPPER(SUBSTRING(actor.last_name, 1, 1)), LOWER(SUBSTRING(actor.last_name, 2)))) AS actors FROM category LEFT JOIN film_category ON film_category.category_id = category.category_id LEFT JOIN film ON film.film_id = film_category.film_id JOIN film_actor ON film_actor.film_id = film.film_id JOIN actor ON actor.actor_id = film_actor.actor_id GROUP BY film.film_id, film.title, film.description, category.name, film.rental_rate, film.length, film.rating;

CREATE OR REPLACE VIEW sales_by_film_category AS SELECT c.name AS category, SUM(p.amount) AS total_sales FROM payment AS p JOIN rental AS r ON r.rental_id = p.rental_id JOIN inventory AS i ON i.inventory_id = r.inventory_id JOIN film AS f ON f.film_id = i.film_id JOIN film_category AS fc ON fc.film_id = f.film_id JOIN category AS c ON c.category_id = fc.category_id GROUP BY c.name ORDER BY SUM(p.amount) DESC;

CREATE OR REPLACE VIEW sales_by_store AS SELECT CONCAT(ci.city, ',', co.country) AS store, CONCAT(m.first_name, ' ', m.last_name) AS manager, SUM(p.amount) AS total_sales FROM payment AS p JOIN rental AS r ON r.rental_id = p.rental_id JOIN inventory AS i ON i.inventory_id = r.inventory_id JOIN store AS s ON s.store_id = i.store_id JOIN address AS a ON a.address_id = s.address_id JOIN city AS ci ON ci.city_id = a.city_id JOIN country AS co ON co.country_id = ci.country_id JOIN staff AS m ON m.staff_id = s.manager_staff_id GROUP BY co.country, ci.city, s.store_id, m.first_name, m.last_name ORDER BY co.country, ci.city;

CREATE OR REPLACE VIEW staff_list AS SELECT s.staff_id AS id, CONCAT(s.first_name, ' ', s.last_name) AS name, a.address, a.postal_code AS `zip code`, a.phone, ci.city, co.country, s.store_id AS sid FROM staff AS s JOIN address AS a ON a.address_id = s.address_id JOIN city AS ci ON ci.city_id = a.city_id JOIN country AS co ON co.country_id = ci.country_id;

-- DELIMITER ;;

CREATE TRIGGER film_after_insert_trg AFTER INSERT ON film FOR EACH ROW BEGIN
    INSERT INTO film_text (film_id, title, description) VALUES (NEW.film_id, NEW.title, NEW.description);
END; -- ;;

CREATE TRIGGER film_after_update_trg AFTER UPDATE ON film FOR EACH ROW BEGIN
    IF OLD.title <> NEW.title OR OLD.description <> NEW.description THEN
        UPDATE film_text
        SET title = NEW.title, description = NEW.description, film_id = NEW.film_id
        WHERE film_id = OLD.film_id;
    END IF;
END; -- ;;

CREATE TRIGGER film_after_delete_trg AFTER DELETE ON film FOR EACH ROW BEGIN
    DELETE FROM film_text WHERE film_id = OLD.film_id;
END; -- ;;

-- DELIMITER ;

ALTER TABLE address
    ADD CONSTRAINT address_city_id_fkey FOREIGN KEY (city_id) REFERENCES city (city_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE city
    ADD CONSTRAINT city_country_id_fkey FOREIGN KEY (country_id) REFERENCES country (country_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE customer
    ADD CONSTRAINT customer_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE film
    ADD CONSTRAINT film_language_id_fkey FOREIGN KEY (language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT film_original_language_id_fkey FOREIGN KEY (original_language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE film_actor
    ADD CONSTRAINT film_actor_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT film_actor_actor_id_fkey FOREIGN KEY (actor_id) REFERENCES actor (actor_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE film_actor_review
    ADD CONSTRAINT film_actor_review_film_id_actor_id_fkey FOREIGN KEY (film_id, actor_id) REFERENCES film_actor (film_id, actor_id) ON UPDATE CASCADE;

ALTER TABLE film_category
    ADD CONSTRAINT film_category_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT film_category_category_id_fkey FOREIGN KEY (category_id) REFERENCES category (category_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE inventory
    ADD CONSTRAINT inventory_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT inventory_store_id_fkey FOREIGN KEY (store_id) REFERENCES store (store_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE payment
    ADD CONSTRAINT payment_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT payment_staff_id_fkey FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT payment_rental_id_fkey FOREIGN KEY (rental_id) REFERENCES rental (rental_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE rental
    ADD CONSTRAINT rental_inventory_id_fkey FOREIGN KEY (inventory_id) REFERENCES inventory (inventory_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT rental_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT rental_staff_id_fkey FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE staff
    ADD CONSTRAINT staff_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT staff_store_id_fkey FOREIGN KEY (store_id) REFERENCES store (store_id);

ALTER TABLE store
    ADD CONSTRAINT store_manager_staff_id_fkey FOREIGN KEY (manager_staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT store_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT;
