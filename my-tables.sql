SET foreign_key_checks = 0;
DROP TABLE IF EXISTS dummy_table_2 CASCADE;
DROP TABLE IF EXISTS dummy_table CASCADE;
DROP TABLE IF EXISTS payment CASCADE;
DROP TABLE IF EXISTS rental CASCADE;
DROP TABLE IF EXISTS inventory CASCADE;
DROP TABLE IF EXISTS customer CASCADE;
DROP TABLE IF EXISTS store CASCADE;
DROP TABLE IF EXISTS staff CASCADE;
DROP TABLE IF EXISTS film_category CASCADE;
DROP TABLE IF EXISTS film_actor CASCADE;
DROP TABLE IF EXISTS film_text CASCADE;
DROP TABLE IF EXISTS film CASCADE;
DROP TABLE IF EXISTS language CASCADE;
DROP TABLE IF EXISTS address CASCADE;
DROP TABLE IF EXISTS city CASCADE;
DROP TABLE IF EXISTS country CASCADE;
DROP TABLE IF EXISTS category CASCADE;
DROP TABLE IF EXISTS actor CASCADE;
SET foreign_key_checks = 1;

CREATE TABLE actor (
    actor_id INT AUTO_INCREMENT PRIMARY KEY
    ,first_name VARCHAR(45) NOT NULL
    ,last_name VARCHAR(45) NOT NULL
    ,full_name VARCHAR(45) GENERATED ALWAYS AS (CONCAT(first_name, ' ', last_name)) VIRTUAL
    ,full_name_reversed VARCHAR(45) GENERATED ALWAYS AS (CONCAT(last_name, ' ', first_name)) STORED
    ,last_update TIMESTAMP DEFAULT (CURRENT_TIMESTAMP) ON UPDATE CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX actor_last_name_idx ON actor (last_name);

CREATE TABLE category (
    category_id INT AUTO_INCREMENT PRIMARY KEY
    ,name VARCHAR(25) NOT NULL
    ,last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE country (
    country_id INT AUTO_INCREMENT PRIMARY KEY
    ,country VARCHAR(50) NOT NULL
    ,last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE city (
    city_id INT AUTO_INCREMENT PRIMARY KEY
    ,city VARCHAR(50) NOT NULL
    ,country_id INT NOT NULL
    ,last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE city ADD CONSTRAINT city_country_id_fkey FOREIGN KEY (country_id) REFERENCES country (country_id) ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE INDEX city_country_id_idx ON city (country_id);

CREATE TABLE address (
    address_id INT AUTO_INCREMENT PRIMARY KEY
    ,address VARCHAR(50) NOT NULL
    ,address2 VARCHAR(50)
    ,district VARCHAR(20) NOT NULL
    ,city_id INT NOT NULL
    ,postal_code VARCHAR(10)
    ,phone VARCHAR(20) NOT NULL
    ,last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE address ADD CONSTRAINT address_city_id_fkey FOREIGN KEY (city_id) REFERENCES city (city_id) ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE INDEX address_city_id_idx ON address (city_id);

CREATE TABLE language (
    language_id INT AUTO_INCREMENT PRIMARY KEY
    ,name CHAR(20) NOT NULL
    ,last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE film (
    film_id INT AUTO_INCREMENT PRIMARY KEY
    ,title VARCHAR(255) NOT NULL
    ,description TEXT
    ,release_year INT
    ,language_id INT NOT NULL
    ,original_language_id INT
    ,rental_duration INT DEFAULT 3 NOT NULL
    ,rental_rate DECIMAL(4,2) DEFAULT 4.99 NOT NULL
    ,length INT
    ,replacement_cost DECIMAL(5,2) DEFAULT 19.99 NOT NULL
    ,rating ENUM('G','PG','PG-13','R','NC-17') DEFAULT 'G'
    ,special_features JSON
    ,last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL

    ,CONSTRAINT film_release_year_check CHECK (release_year >= 1901 AND release_year <= 2155)
);

ALTER TABLE film ADD CONSTRAINT film_language_id_fkey FOREIGN KEY (language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE film ADD CONSTRAINT film_original_language_id_fkey FOREIGN KEY (original_language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE INDEX film_title_idx ON film (title);

CREATE INDEX film_language_id_idx ON film (language_id);

CREATE INDEX film_original_language_id_idx ON film (original_language_id);

CREATE TABLE film_text (
  film_id INT NOT NULL PRIMARY KEY
  ,title VARCHAR(255) NOT NULL
  ,description TEXT
);

CREATE FULLTEXT INDEX film_text_title_description_idx ON film_text (title, description);

DELIMITER ;;
CREATE TRIGGER film_after_insert_trg AFTER INSERT ON film FOR EACH ROW BEGIN
    INSERT INTO film_text
        (film_id, title, description)
    VALUES
        (NEW.film_id, NEW.title, NEW.description)
    ;
END;;
CREATE TRIGGER film_after_update_trg AFTER UPDATE ON film FOR EACH ROW BEGIN
    IF OLD.title <> NEW.title OR OLD.description <> NEW.description THEN
        UPDATE
            film_text
        SET
            title = NEW.title
            ,description = NEW.description
            ,film_id = NEW.film_id
        WHERE
            film_id = OLD.film_id
        ;
    END IF;
END;;
CREATE TRIGGER film_after_delete_trg AFTER DELETE ON film FOR EACH ROW BEGIN
    DELETE FROM film_text WHERE film_id = OLD.film_id;
END;;
DELIMITER ;

CREATE TABLE film_actor (
    actor_id INT NOT NULL
    ,film_id INT NOT NULL
    ,last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE film_actor ADD CONSTRAINT film_actor_actor_id_fkey FOREIGN KEY (actor_id) REFERENCES actor (actor_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE film_actor ADD CONSTRAINT film_actor_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE UNIQUE INDEX film_actor_actor_id_film_id_idx ON film_actor (actor_id, film_id);

CREATE INDEX film_actor_film_id_idx ON film_actor (film_id);

CREATE TABLE film_category (
    film_id INT NOT NULL
    ,category_id INT NOT NULL
    ,last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE film_category ADD CONSTRAINT film_category_category_id_fkey FOREIGN KEY (category_id) REFERENCES category (category_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE film_category ADD CONSTRAINT film_category_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE TABLE staff (
    staff_id INT AUTO_INCREMENT PRIMARY KEY
    ,first_name VARCHAR(45) NOT NULL
    ,last_name VARCHAR(45) NOT NULL
    ,address_id INT NOT NULL
    ,email VARCHAR(50)
    ,store_id INT
    ,active BOOLEAN DEFAULT TRUE NOT NULL
    ,username VARCHAR(16) NOT NULL
    ,password VARCHAR(40)
    ,last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL
    ,picture BLOB
);

ALTER TABLE staff ADD CONSTRAINT staff_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE TABLE store (
    store_id INT AUTO_INCREMENT PRIMARY KEY
    ,manager_staff_id INT NOT NULL
    ,address_id INT NOT NULL
    ,last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE staff ADD CONSTRAINT staff_store_id_fkey FOREIGN KEY (store_id) REFERENCES store (store_id);

ALTER TABLE store ADD CONSTRAINT store_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE store ADD CONSTRAINT store_manager_staff_id_fkey FOREIGN KEY (manager_staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE UNIQUE INDEX store_manager_staff_id_idx ON store (manager_staff_id);

CREATE TABLE customer (
    customer_id INT AUTO_INCREMENT PRIMARY KEY
    ,store_id INT NOT NULL
    ,first_name VARCHAR(45) NOT NULL
    ,last_name VARCHAR(45) NOT NULL
    ,email VARCHAR(50) UNIQUE
    ,address_id INT NOT NULL
    ,active BOOLEAN DEFAULT TRUE NOT NULL
    ,create_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
    ,last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT customer_email_first_name_last_name_key UNIQUE (email, first_name, last_name)
);

ALTER TABLE customer ADD CONSTRAINT customer_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE customer ADD CONSTRAINT customer_store_id_fkey FOREIGN KEY (store_id) REFERENCES store (store_id) ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE INDEX customer_address_id_idx ON customer (address_id);

CREATE INDEX customer_store_id_idx ON customer (store_id);

CREATE INDEX customer_last_name_idx ON customer (last_name);

CREATE TABLE inventory (
    inventory_id INT AUTO_INCREMENT PRIMARY KEY
    ,film_id INT NOT NULL
    ,store_id INT NOT NULL
    ,last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE inventory ADD CONSTRAINT inventory_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE inventory ADD CONSTRAINT inventory_store_id_fkey FOREIGN KEY (store_id) REFERENCES store (store_id) ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE INDEX inventory_store_id_film_id_idx ON inventory (store_id, film_id);

CREATE TABLE rental (
    rental_id INT AUTO_INCREMENT PRIMARY KEY
    ,rental_date TIMESTAMP NOT NULL
    ,inventory_id INT NOT NULL
    ,customer_id INT NOT NULL
    ,return_date TIMESTAMP
    ,staff_id INT NOT NULL
    ,last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE rental ADD CONSTRAINT rental_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE rental ADD CONSTRAINT rental_inventory_id_fkey FOREIGN KEY (inventory_id) REFERENCES inventory (inventory_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE rental ADD CONSTRAINT rental_staff_id_fkey FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE UNIQUE INDEX rental_rental_date_inventory_id_customer_id_idx ON rental (rental_date, inventory_id, customer_id);

CREATE INDEX rental_inventory_id_idx ON rental (inventory_id);

CREATE INDEX rental_customer_id_idx ON rental (customer_id);

CREATE INDEX rental_staff_id_idx ON rental (staff_id);

CREATE TABLE payment (
    payment_id INT AUTO_INCREMENT PRIMARY KEY
    ,customer_id INT NOT NULL
    ,staff_id INT NOT NULL
    ,rental_id INT
    ,amount DECIMAL(5,2) NOT NULL
    ,payment_date TIMESTAMP NOT NULL
);

ALTER TABLE payment ADD CONSTRAINT payment_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE payment ADD CONSTRAINT payment_rental_id_fkey FOREIGN KEY (rental_id) REFERENCES rental (rental_id) ON UPDATE CASCADE ON DELETE SET NULL;

ALTER TABLE payment ADD CONSTRAINT payment_staff_id_fkey FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE INDEX payment_customer_id_idx ON payment (customer_id);

CREATE INDEX payment_staff_id_idx ON payment (staff_id);

CREATE TABLE dummy_table (
    id1 INT
    ,id2 VARCHAR(255)
    ,score INT
    ,color VARCHAR(50) COLLATE latin1_swedish_ci
    ,data JSON

    ,CONSTRAINT dummy_table_score_positive_check CHECK (score > 0)
    ,CONSTRAINT dummy_table_id1_id2_pkey PRIMARY KEY (id2, id1)
    ,CONSTRAINT dummy_table_score_color_key UNIQUE (score, color)
    ,CONSTRAINT dummy_table_score_id1_greater_than_check CHECK (score > id1)
);

CREATE INDEX dummy_table_score_color_data_idx ON dummy_table (score, (SUBSTR(color,1,2)), (CONCAT(color,' abcd')), (CAST(data->>'$.age' AS SIGNED)));

CREATE TABLE dummy_table_2 (
    id1 INT
    ,id2 VARCHAR(255)
);

ALTER TABLE dummy_table_2 ADD CONSTRAINT dummy_table_2_id2_id1_fkey FOREIGN KEY (id2, id1) REFERENCES dummy_table (id2, id1) ON UPDATE CASCADE ON DELETE RESTRICT;
