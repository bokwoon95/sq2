SET foreign_key_checks = 0;
DROP TABLE IF EXISTS db.dummy_table_2 CASCADE;
DROP TABLE IF EXISTS db.dummy_table CASCADE;
DROP TABLE IF EXISTS db.payment CASCADE;
DROP TABLE IF EXISTS db.rental CASCADE;
DROP TABLE IF EXISTS db.inventory CASCADE;
DROP TABLE IF EXISTS db.customer CASCADE;
DROP TABLE IF EXISTS db.store CASCADE;
DROP TABLE IF EXISTS db.staff CASCADE;
DROP TABLE IF EXISTS db.film_category CASCADE;
DROP TABLE IF EXISTS db.film_actor CASCADE;
DROP TABLE IF EXISTS db.film_text CASCADE;
DROP TABLE IF EXISTS db.film CASCADE;
DROP TABLE IF EXISTS db.language CASCADE;
DROP TABLE IF EXISTS db.address CASCADE;
DROP TABLE IF EXISTS db.city CASCADE;
DROP TABLE IF EXISTS db.country CASCADE;
DROP TABLE IF EXISTS db.category CASCADE;
DROP TABLE IF EXISTS db.actor CASCADE;
SET foreign_key_checks = 1;

CREATE TABLE db.actor (
    actor_id INT AUTO_INCREMENT
    ,first_name VARCHAR(45) NOT NULL
    ,last_name VARCHAR(45) NOT NULL
    ,full_name VARCHAR(45) GENERATED ALWAYS AS (CONCAT(first_name, ' ', last_name)) VIRTUAL
    ,full_name_reversed VARCHAR(45) GENERATED ALWAYS AS (CONCAT(last_name, ' ', first_name)) STORED
    ,last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT actor_actor_id_pkey PRIMARY KEY (actor_id)
);

CREATE INDEX actor_last_name_idx ON db.actor (last_name);

CREATE TABLE db.category (
    category_id INT AUTO_INCREMENT
    ,name VARCHAR(25) NOT NULL
    ,last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT category_category_id_pkey PRIMARY KEY (category_id)
);

CREATE TABLE db.country (
    country_id INT AUTO_INCREMENT
    ,country VARCHAR(50) NOT NULL
    ,last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT country_country_id_pkey PRIMARY KEY (country_id)
);

CREATE TABLE db.city (
    city_id INT AUTO_INCREMENT
    ,city VARCHAR(50) NOT NULL
    ,country_id INT NOT NULL
    ,last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT city_city_id_pkey PRIMARY KEY (city_id)
);

CREATE INDEX city_country_id_idx ON db.city (country_id);

CREATE TABLE db.address (
    address_id INT AUTO_INCREMENT
    ,address VARCHAR(50) NOT NULL
    ,address2 VARCHAR(50)
    ,district VARCHAR(20) NOT NULL
    ,city_id INT NOT NULL
    ,postal_code VARCHAR(10)
    ,phone VARCHAR(20) NOT NULL
    ,last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT address_address_id_pkey PRIMARY KEY (address_id)
);

CREATE INDEX address_city_id_idx ON db.address (city_id);

CREATE TABLE db.language (
    language_id INT AUTO_INCREMENT
    ,name CHAR(20) NOT NULL
    ,last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT language_language_id_pkey PRIMARY KEY (language_id)
);

CREATE TABLE db.film (
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
    ,last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT film_film_id_pkey PRIMARY KEY (film_id)
    ,CONSTRAINT film_release_year_check CHECK (release_year >= 1901 AND release_year <= 2155)
);

CREATE INDEX film_title_idx ON db.film (title);

CREATE INDEX film_language_id_idx ON db.film (language_id);

CREATE INDEX film_original_language_id_idx ON db.film (original_language_id);

CREATE TABLE db.film_text (
  film_id INT NOT NULL PRIMARY KEY
  ,title VARCHAR(255) NOT NULL
  ,description TEXT
);

CREATE FULLTEXT INDEX film_text_title_description_idx ON db.film_text (title, description);

CREATE TABLE db.film_actor (
    actor_id INT NOT NULL
    ,film_id INT NOT NULL
    ,last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX film_actor_actor_id_film_id_idx ON db.film_actor (actor_id, film_id);

CREATE INDEX film_actor_film_id_idx ON db.film_actor (film_id);

CREATE TABLE db.film_category (
    film_id INT NOT NULL
    ,category_id INT NOT NULL
    ,last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE db.staff (
    staff_id INT AUTO_INCREMENT
    ,first_name VARCHAR(45) NOT NULL
    ,last_name VARCHAR(45) NOT NULL
    ,address_id INT NOT NULL
    ,email VARCHAR(50)
    ,store_id INT
    ,active BOOLEAN NOT NULL DEFAULT TRUE
    ,username VARCHAR(16) NOT NULL
    ,password VARCHAR(40)
    ,last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ,picture BLOB

    ,CONSTRAINT staff_staff_id_pkey PRIMARY KEY (staff_id)
);

CREATE TABLE db.store (
    store_id INT AUTO_INCREMENT
    ,manager_staff_id INT NOT NULL
    ,address_id INT NOT NULL
    ,last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT store_store_id_pkey PRIMARY KEY (store_id)
);

CREATE UNIQUE INDEX store_manager_staff_id_idx ON db.store (manager_staff_id);

CREATE TABLE db.customer (
    customer_id INT AUTO_INCREMENT
    ,store_id INT NOT NULL
    ,first_name VARCHAR(45) NOT NULL
    ,last_name VARCHAR(45) NOT NULL
    ,email VARCHAR(50)
    ,address_id INT NOT NULL
    ,active BOOLEAN NOT NULL DEFAULT TRUE
    ,data JSON
    ,create_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    ,last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT customer_email_first_name_last_name_key UNIQUE (email, first_name, last_name)
    ,CONSTRAINT customer_customer_id_pkey PRIMARY KEY (customer_id)
    ,CONSTRAINT customer_email_key UNIQUE (email)
);

CREATE INDEX customer_store_id_idx ON db.customer (store_id);

CREATE INDEX customer_last_name_idx ON db.customer (last_name);

CREATE INDEX customer_address_id_idx ON db.customer (address_id);

CREATE TABLE db.inventory (
    inventory_id INT AUTO_INCREMENT
    ,film_id INT NOT NULL
    ,store_id INT NOT NULL
    ,last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT inventory_inventory_id_pkey PRIMARY KEY (inventory_id)
);

CREATE INDEX inventory_store_id_film_id_idx ON db.inventory (store_id, film_id);

CREATE TABLE db.rental (
    rental_id INT AUTO_INCREMENT
    ,rental_date DATETIME NOT NULL
    ,inventory_id INT NOT NULL
    ,customer_id INT NOT NULL
    ,return_date TIMESTAMP
    ,staff_id INT NOT NULL
    ,last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

    ,CONSTRAINT rental_rental_id_pkey PRIMARY KEY (rental_id)
);

CREATE UNIQUE INDEX rental_rental_date_inventory_id_customer_id_idx ON db.rental (rental_date, inventory_id, customer_id);

CREATE INDEX rental_inventory_id_idx ON db.rental (inventory_id);

CREATE INDEX rental_customer_id_idx ON db.rental (customer_id);

CREATE INDEX rental_staff_id_idx ON db.rental (staff_id);

CREATE TABLE db.payment (
    payment_id INT AUTO_INCREMENT
    ,customer_id INT NOT NULL
    ,staff_id INT NOT NULL
    ,rental_id INT
    ,amount DECIMAL(5,2) NOT NULL
    ,payment_date TIMESTAMP NOT NULL

    ,CONSTRAINT payment_payment_id_pkey PRIMARY KEY (payment_id)
);

CREATE INDEX payment_customer_id_idx ON db.payment (customer_id);

CREATE INDEX payment_staff_id_idx ON db.payment (staff_id);

CREATE TABLE db.dummy_table (
    id1 INT
    ,id2 VARCHAR(255)
    ,score INT
    ,color VARCHAR(50) DEFAULT 'red' COLLATE latin1_swedish_ci
    ,data JSON

    ,CONSTRAINT dummy_table_id1_id2_pkey PRIMARY KEY (id1, id2)
    ,CONSTRAINT dummy_table_score_color_key UNIQUE (score, color)
    ,CONSTRAINT dummy_table_score_positive_check CHECK (score > 0)
    ,CONSTRAINT dummy_table_score_id1_greater_than_check CHECK (score > id1)
);

CREATE INDEX dummy_table_complex_expr_idx ON db.dummy_table (score, (SUBSTR(color, 1, 2)), (CONCAT(color, ' abcd')), (CAST(data->>'$.age' AS SIGNED)));

CREATE TABLE db.dummy_table_2 (
    id1 INT
    ,id2 VARCHAR(255)
);

ALTER TABLE db.city ADD CONSTRAINT city_country_id_fkey FOREIGN KEY (country_id) REFERENCES country (country_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.address ADD CONSTRAINT address_city_id_fkey FOREIGN KEY (city_id) REFERENCES city (city_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.film ADD CONSTRAINT film_language_id_fkey FOREIGN KEY (language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.film ADD CONSTRAINT film_original_language_id_fkey FOREIGN KEY (original_language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.film_actor ADD CONSTRAINT film_actor_actor_id_fkey FOREIGN KEY (actor_id) REFERENCES actor (actor_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.film_actor ADD CONSTRAINT film_actor_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.film_category ADD CONSTRAINT film_category_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.film_category ADD CONSTRAINT film_category_category_id_fkey FOREIGN KEY (category_id) REFERENCES category (category_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.staff ADD CONSTRAINT staff_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.staff ADD CONSTRAINT staff_store_id_fkey FOREIGN KEY (store_id) REFERENCES store (store_id);

ALTER TABLE db.store ADD CONSTRAINT store_manager_staff_id_fkey FOREIGN KEY (manager_staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.store ADD CONSTRAINT store_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.customer ADD CONSTRAINT customer_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.inventory ADD CONSTRAINT inventory_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.inventory ADD CONSTRAINT inventory_store_id_fkey FOREIGN KEY (store_id) REFERENCES store (store_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.rental ADD CONSTRAINT rental_inventory_id_fkey FOREIGN KEY (inventory_id) REFERENCES inventory (inventory_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.rental ADD CONSTRAINT rental_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.rental ADD CONSTRAINT rental_staff_id_fkey FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.payment ADD CONSTRAINT payment_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.payment ADD CONSTRAINT payment_staff_id_fkey FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.payment ADD CONSTRAINT payment_rental_id_fkey FOREIGN KEY (rental_id) REFERENCES rental (rental_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE db.dummy_table_2 ADD CONSTRAINT dummy_table_2_id1_id2_fkey FOREIGN KEY (id1, id2) REFERENCES dummy_table (id1, id2) ON UPDATE CASCADE ON DELETE RESTRICT;
