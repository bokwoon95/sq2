CREATE SCHEMA IF NOT EXISTS db;

CREATE TABLE IF NOT EXISTS db.actor (
    actor_id INT NOT NULL AUTO_INCREMENT
    ,first_name VARCHAR(45) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,last_name VARCHAR(45) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,full_name VARCHAR(255) NOT NULL GENERATED ALWAYS AS (concat(`first_name`,_utf8mb4' ',`last_name`)) VIRTUAL COLLATE utf8mb4_0900_ai_ci
    ,full_name_reversed VARCHAR(255) NOT NULL GENERATED ALWAYS AS (concat(`last_name`,_utf8mb4' ',`first_name`)) STORED COLLATE utf8mb4_0900_ai_ci
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (actor_id)
    ,INDEX actor_last_name_idx (last_name)
);

CREATE TABLE IF NOT EXISTS db.address (
    address_id INT NOT NULL AUTO_INCREMENT
    ,address VARCHAR(50) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,address2 VARCHAR(50) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,district VARCHAR(20) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,city_id INT NOT NULL
    ,postal_code VARCHAR(10) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,phone VARCHAR(20) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (address_id)
    ,INDEX address_city_id_idx (city_id)
);

CREATE TABLE IF NOT EXISTS db.category (
    category_id INT NOT NULL AUTO_INCREMENT
    ,name VARCHAR(25) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (category_id)
);

CREATE TABLE IF NOT EXISTS db.city (
    city_id INT NOT NULL AUTO_INCREMENT
    ,city VARCHAR(50) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,country_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (city_id)
    ,INDEX city_country_id_idx (country_id)
);

CREATE TABLE IF NOT EXISTS db.country (
    country_id INT NOT NULL AUTO_INCREMENT
    ,country VARCHAR(50) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (country_id)
);

CREATE TABLE IF NOT EXISTS db.customer (
    customer_id INT NOT NULL AUTO_INCREMENT
    ,store_id INT NOT NULL
    ,first_name VARCHAR(45) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,last_name VARCHAR(45) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,email VARCHAR(50) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,address_id INT NOT NULL
    ,active TINYINT(1) NOT NULL DEFAULT 1
    ,data JSON NOT NULL
    ,create_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT customer_email_first_name_last_name_key UNIQUE (email, first_name, last_name)
    ,CONSTRAINT customer_email_key UNIQUE (email)
    ,PRIMARY KEY (customer_id)
    ,INDEX customer_address_id_idx (address_id)
    ,INDEX customer_last_name_idx (last_name)
    ,INDEX customer_store_id_idx (store_id)
);

CREATE TABLE IF NOT EXISTS db.film (
    film_id INT NOT NULL AUTO_INCREMENT
    ,title VARCHAR(255) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,description TEXT NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,release_year INT NOT NULL
    ,language_id INT NOT NULL
    ,original_language_id INT NOT NULL
    ,rental_duration INT NOT NULL DEFAULT 3
    ,rental_rate DECIMAL(4,2) NOT NULL DEFAULT 4.99
    ,length INT NOT NULL
    ,replacement_cost DECIMAL(5,2) NOT NULL DEFAULT 19.99
    ,rating ENUM('G','PG','PG-13','R','NC-17') NOT NULL DEFAULT G COLLATE utf8mb4_0900_ai_ci
    ,special_features JSON NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT film_release_year_check CHECK ((`release_year` >= 1901) and (`release_year` <= 2155))
    ,PRIMARY KEY (film_id)
    ,INDEX film_language_id_idx (language_id)
    ,INDEX film_original_language_id_idx (original_language_id)
    ,INDEX film_title_idx (title)
);

CREATE TABLE IF NOT EXISTS db.film_actor (
    film_id INT NOT NULL
    ,actor_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT film_actor_actor_id_film_id_idx UNIQUE (actor_id, film_id)
    ,INDEX film_actor_film_id_idx (film_id)
);

CREATE TABLE IF NOT EXISTS db.film_actor_review (
    film_id INT NOT NULL
    ,actor_id INT NOT NULL
    ,review_title VARCHAR(50) NOT NULL COLLATE latin1_swedish_ci
    ,review_body VARCHAR(255) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,metadata JSON NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ,delete_date DATETIME NOT NULL

    ,CONSTRAINT film_actor_review_check CHECK (length(`review_body`) > length(`review_title`))
    ,PRIMARY KEY (film_id, actor_id)
    ,INDEX film_actor_review_misc (film_id, (substr(`review_body`,2,10)), (concat(`review_title`,_latin1' abcd')), (cast(json_unquote(json_extract(`metadata`,_utf8mb4'$.score')) as signed)))
);

CREATE TABLE IF NOT EXISTS db.film_category (
    film_id INT NOT NULL
    ,category_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS db.film_text (
    film_id INT NOT NULL
    ,title VARCHAR(255) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,description TEXT NOT NULL COLLATE utf8mb4_0900_ai_ci

    ,PRIMARY KEY (film_id)
    ,FULLTEXT INDEX film_text_title_description_idx (title, description)
);

CREATE TABLE IF NOT EXISTS db.inventory (
    inventory_id INT NOT NULL AUTO_INCREMENT
    ,film_id INT NOT NULL
    ,store_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (inventory_id)
    ,INDEX inventory_store_id_film_id_idx (store_id, film_id)
);

CREATE TABLE IF NOT EXISTS db.language (
    language_id INT NOT NULL AUTO_INCREMENT
    ,name CHAR(20) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (language_id)
);

CREATE TABLE IF NOT EXISTS db.payment (
    payment_id INT NOT NULL AUTO_INCREMENT
    ,customer_id INT NOT NULL
    ,staff_id INT NOT NULL
    ,rental_id INT NOT NULL
    ,amount DECIMAL(5,2) NOT NULL
    ,payment_date DATETIME NOT NULL

    ,PRIMARY KEY (payment_id)
    ,INDEX payment_customer_id_idx (customer_id)
    ,INDEX payment_staff_id_idx (staff_id)
);

CREATE TABLE IF NOT EXISTS db.rental (
    rental_id INT NOT NULL AUTO_INCREMENT
    ,rental_date DATETIME NOT NULL
    ,inventory_id INT NOT NULL
    ,customer_id INT NOT NULL
    ,return_date DATETIME NOT NULL
    ,staff_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (rental_id)
    ,CONSTRAINT rental_rental_date_inventory_id_customer_id_idx UNIQUE (rental_date, inventory_id, customer_id)
    ,INDEX rental_customer_id_idx (customer_id)
    ,INDEX rental_inventory_id_idx (inventory_id)
    ,INDEX rental_staff_id_idx (staff_id)
);

CREATE TABLE IF NOT EXISTS db.staff (
    staff_id INT NOT NULL AUTO_INCREMENT
    ,first_name VARCHAR(45) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,last_name VARCHAR(45) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,address_id INT NOT NULL
    ,email VARCHAR(50) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,store_id INT NOT NULL
    ,active TINYINT(1) NOT NULL DEFAULT 1
    ,username VARCHAR(16) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,password VARCHAR(40) NOT NULL COLLATE utf8mb4_0900_ai_ci
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ,picture BLOB NOT NULL

    ,PRIMARY KEY (staff_id)
);

CREATE TABLE IF NOT EXISTS db.store (
    store_id INT NOT NULL AUTO_INCREMENT
    ,manager_staff_id INT NOT NULL
    ,address_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,PRIMARY KEY (store_id)
    ,CONSTRAINT store_manager_staff_id_idx UNIQUE (manager_staff_id)
);

CREATE OR REPLACE VIEW db.actor_info AS select `a`.`actor_id` AS `actor_id`,`a`.`first_name` AS `first_name`,`a`.`last_name` AS `last_name`,json_objectagg(`c`.`name`,(select json_arrayagg(`f`.`title`) from ((`db`.`film` `f` join `db`.`film_category` `fc` on((`fc`.`film_id` = `f`.`film_id`))) join `db`.`film_actor` `fa` on((`fa`.`film_id` = `f`.`film_id`))) where ((`fc`.`category_id` = `c`.`category_id`) and (`fa`.`actor_id` = `a`.`actor_id`)) group by `fa`.`actor_id`)) AS `film_info` from (((`db`.`actor` `a` left join `db`.`film_actor` `fa` on((`fa`.`actor_id` = `a`.`actor_id`))) left join `db`.`film_category` `fc` on((`fc`.`film_id` = `fa`.`film_id`))) left join `db`.`category` `c` on((`c`.`category_id` = `fc`.`category_id`))) group by `a`.`actor_id`,`a`.`first_name`,`a`.`last_name`;

CREATE OR REPLACE VIEW db.customer_list AS select `cu`.`customer_id` AS `id`,concat(`cu`.`first_name`,' ',`cu`.`last_name`) AS `name`,`a`.`address` AS `address`,`a`.`postal_code` AS `zip code`,`a`.`phone` AS `phone`,`db`.`city`.`city` AS `city`,`db`.`country`.`country` AS `country`,(case when `cu`.`active` then 'active' else '' end) AS `notes`,`cu`.`store_id` AS `sid` from (((`db`.`customer` `cu` join `db`.`address` `a` on((`a`.`address_id` = `cu`.`address_id`))) join `db`.`city` on((`db`.`city`.`city_id` = `a`.`city_id`))) join `db`.`country` on((`db`.`country`.`country_id` = `db`.`city`.`country_id`)));

CREATE OR REPLACE VIEW db.film_list AS select `db`.`film`.`film_id` AS `fid`,`db`.`film`.`title` AS `title`,`db`.`film`.`description` AS `description`,`db`.`category`.`name` AS `category`,`db`.`film`.`rental_rate` AS `price`,`db`.`film`.`length` AS `length`,`db`.`film`.`rating` AS `rating`,json_arrayagg(concat(`db`.`actor`.`first_name`,' ',`db`.`actor`.`last_name`)) AS `actors` from ((((`db`.`category` left join `db`.`film_category` on((`db`.`film_category`.`category_id` = `db`.`category`.`category_id`))) left join `db`.`film` on((`db`.`film`.`film_id` = `db`.`film_category`.`film_id`))) join `db`.`film_actor` on((`db`.`film_actor`.`film_id` = `db`.`film`.`film_id`))) join `db`.`actor` on((`db`.`actor`.`actor_id` = `db`.`film_actor`.`actor_id`))) group by `db`.`film`.`film_id`,`db`.`film`.`title`,`db`.`film`.`description`,`db`.`category`.`name`,`db`.`film`.`rental_rate`,`db`.`film`.`length`,`db`.`film`.`rating`;

CREATE OR REPLACE VIEW db.full_address AS select `db`.`country`.`country_id` AS `country_id`,`db`.`city`.`city_id` AS `city_id`,`db`.`address`.`address_id` AS `address_id`,`db`.`country`.`country` AS `country`,`db`.`city`.`city` AS `city`,`db`.`address`.`address` AS `address`,`db`.`address`.`address2` AS `address2`,`db`.`address`.`district` AS `district`,`db`.`address`.`postal_code` AS `postal_code`,`db`.`address`.`phone` AS `phone`,`db`.`address`.`last_update` AS `last_update` from ((`db`.`address` join `db`.`city` on((`db`.`city`.`city_id` = `db`.`address`.`city_id`))) join `db`.`country` on((`db`.`country`.`country_id` = `db`.`city`.`country_id`)));

CREATE OR REPLACE VIEW db.nicer_but_slower_film_list AS select `db`.`film`.`film_id` AS `fid`,`db`.`film`.`title` AS `title`,`db`.`film`.`description` AS `description`,`db`.`category`.`name` AS `category`,`db`.`film`.`rental_rate` AS `price`,`db`.`film`.`length` AS `length`,`db`.`film`.`rating` AS `rating`,json_arrayagg(concat(upper(substr(`db`.`actor`.`first_name`,1,1)),lower(substr(`db`.`actor`.`first_name`,2)),' ',upper(substr(`db`.`actor`.`last_name`,1,1)),lower(substr(`db`.`actor`.`last_name`,2)))) AS `actors` from ((((`db`.`category` left join `db`.`film_category` on((`db`.`film_category`.`category_id` = `db`.`category`.`category_id`))) left join `db`.`film` on((`db`.`film`.`film_id` = `db`.`film_category`.`film_id`))) join `db`.`film_actor` on((`db`.`film_actor`.`film_id` = `db`.`film`.`film_id`))) join `db`.`actor` on((`db`.`actor`.`actor_id` = `db`.`film_actor`.`actor_id`))) group by `db`.`film`.`film_id`,`db`.`film`.`title`,`db`.`film`.`description`,`db`.`category`.`name`,`db`.`film`.`rental_rate`,`db`.`film`.`length`,`db`.`film`.`rating`;

CREATE OR REPLACE VIEW db.sales_by_film_category AS select `c`.`name` AS `category`,sum(`p`.`amount`) AS `total_sales` from (((((`db`.`payment` `p` join `db`.`rental` `r` on((`r`.`rental_id` = `p`.`rental_id`))) join `db`.`inventory` `i` on((`i`.`inventory_id` = `r`.`inventory_id`))) join `db`.`film` `f` on((`f`.`film_id` = `i`.`film_id`))) join `db`.`film_category` `fc` on((`fc`.`film_id` = `f`.`film_id`))) join `db`.`category` `c` on((`c`.`category_id` = `fc`.`category_id`))) group by `c`.`name` order by sum(`p`.`amount`) desc;

CREATE OR REPLACE VIEW db.sales_by_store AS select concat(`ci`.`city`,',',`co`.`country`) AS `store`,concat(`m`.`first_name`,' ',`m`.`last_name`) AS `manager`,sum(`p`.`amount`) AS `total_sales` from (((((((`db`.`payment` `p` join `db`.`rental` `r` on((`r`.`rental_id` = `p`.`rental_id`))) join `db`.`inventory` `i` on((`i`.`inventory_id` = `r`.`inventory_id`))) join `db`.`store` `s` on((`s`.`store_id` = `i`.`store_id`))) join `db`.`address` `a` on((`a`.`address_id` = `s`.`address_id`))) join `db`.`city` `ci` on((`ci`.`city_id` = `a`.`city_id`))) join `db`.`country` `co` on((`co`.`country_id` = `ci`.`country_id`))) join `db`.`staff` `m` on((`m`.`staff_id` = `s`.`manager_staff_id`))) group by `co`.`country`,`ci`.`city`,`s`.`store_id`,`m`.`first_name`,`m`.`last_name` order by `co`.`country`,`ci`.`city`;

CREATE OR REPLACE VIEW db.staff_list AS select `s`.`staff_id` AS `id`,concat(`s`.`first_name`,' ',`s`.`last_name`) AS `name`,`a`.`address` AS `address`,`a`.`postal_code` AS `zip code`,`a`.`phone` AS `phone`,`ci`.`city` AS `city`,`co`.`country` AS `country`,`s`.`store_id` AS `sid` from (((`db`.`staff` `s` join `db`.`address` `a` on((`a`.`address_id` = `s`.`address_id`))) join `db`.`city` `ci` on((`ci`.`city_id` = `a`.`city_id`))) join `db`.`country` `co` on((`co`.`country_id` = `ci`.`country_id`)));

-- DELIMITER ;;

CREATE TRIGGER film_after_delete_trg AFTER DELETE ON db.film FOR EACH ROW BEGIN
    DELETE FROM film_text WHERE film_id = OLD.film_id;
END; -- ;;

CREATE TRIGGER film_after_insert_trg AFTER INSERT ON db.film FOR EACH ROW BEGIN
    INSERT INTO film_text (film_id, title, description) VALUES (NEW.film_id, NEW.title, NEW.description);
END; -- ;;

CREATE TRIGGER film_after_update_trg AFTER UPDATE ON db.film FOR EACH ROW BEGIN
    IF OLD.title <> NEW.title OR OLD.description <> NEW.description THEN
        UPDATE film_text
        SET title = NEW.title, description = NEW.description, film_id = NEW.film_id
        WHERE film_id = OLD.film_id;
    END IF;
END; -- ;;

-- DELIMITER ;

ALTER TABLE db.address
    ADD CONSTRAINT address_city_id_fkey FOREIGN KEY (city_id) REFERENCES db.city (city_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.city
    ADD CONSTRAINT city_country_id_fkey FOREIGN KEY (country_id) REFERENCES db.country (country_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.customer
    ADD CONSTRAINT customer_address_id_fkey FOREIGN KEY (address_id) REFERENCES db.address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.film
    ADD CONSTRAINT film_language_id_fkey FOREIGN KEY (language_id) REFERENCES db.language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT film_original_language_id_fkey FOREIGN KEY (original_language_id) REFERENCES db.language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.film_actor
    ADD CONSTRAINT film_actor_actor_id_fkey FOREIGN KEY (actor_id) REFERENCES db.actor (actor_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT film_actor_film_id_fkey FOREIGN KEY (film_id) REFERENCES db.film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.film_actor_review
    ADD CONSTRAINT film_actor_review_film_id_actor_id_fkey FOREIGN KEY (film_id, actor_id) REFERENCES db.film_actor (film_id, actor_id) ON UPDATE CASCADE ON DELETE NO ACTION;

ALTER TABLE db.film_category
    ADD CONSTRAINT film_category_category_id_fkey FOREIGN KEY (category_id) REFERENCES db.category (category_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT film_category_film_id_fkey FOREIGN KEY (film_id) REFERENCES db.film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.inventory
    ADD CONSTRAINT inventory_film_id_fkey FOREIGN KEY (film_id) REFERENCES db.film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT inventory_store_id_fkey FOREIGN KEY (store_id) REFERENCES db.store (store_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.payment
    ADD CONSTRAINT payment_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES db.customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT payment_rental_id_fkey FOREIGN KEY (rental_id) REFERENCES db.rental (rental_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT payment_staff_id_fkey FOREIGN KEY (staff_id) REFERENCES db.staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.rental
    ADD CONSTRAINT rental_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES db.customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT rental_inventory_id_fkey FOREIGN KEY (inventory_id) REFERENCES db.inventory (inventory_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT rental_staff_id_fkey FOREIGN KEY (staff_id) REFERENCES db.staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.staff
    ADD CONSTRAINT staff_address_id_fkey FOREIGN KEY (address_id) REFERENCES db.address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT staff_store_id_fkey FOREIGN KEY (store_id) REFERENCES db.store (store_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE db.store
    ADD CONSTRAINT store_address_id_fkey FOREIGN KEY (address_id) REFERENCES db.address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT store_manager_staff_id_fkey FOREIGN KEY (manager_staff_id) REFERENCES db.staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT;
