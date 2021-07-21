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

CREATE TABLE actor (
    actor_id INTEGER PRIMARY KEY AUTOINCREMENT
    ,first_name TEXT NOT NULL
    ,last_name TEXT NOT NULL
    ,full_name TEXT GENERATED ALWAYS AS (first_name || ' ' || last_name) VIRTUAL
    ,full_name_reversed TEXT GENERATED ALWAYS AS (last_name || ' ' || first_name) STORED
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))
);

CREATE INDEX actor_last_name_idx ON actor (last_name);

CREATE TABLE category (
    category_id INTEGER PRIMARY KEY
    ,name TEXT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))
);

CREATE TABLE country (
    country_id INTEGER PRIMARY KEY
    ,country TEXT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))
);

CREATE TABLE city (
    city_id INTEGER PRIMARY KEY
    ,city TEXT NOT NULL
    ,country_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,CONSTRAINT city_country_id_fkey FOREIGN KEY (country_id) REFERENCES country (country_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX city_country_id_idx ON city (country_id);

CREATE TABLE address (
    address_id INTEGER PRIMARY KEY
    ,address TEXT NOT NULL
    ,address2 TEXT
    ,district TEXT NOT NULL
    ,city_id INT NOT NULL
    ,postal_code TEXT
    ,phone TEXT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,CONSTRAINT address_city_id_fkey FOREIGN KEY (city_id) REFERENCES city (city_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX address_city_id_idx ON address (city_id);

CREATE TABLE language (
    language_id INTEGER PRIMARY KEY
    ,name TEXT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))
);

CREATE TABLE film (
    film_id INTEGER PRIMARY KEY
    ,title TEXT NOT NULL
    ,description TEXT DEFAULT '1 + 2 % 3'
    ,release_year INT
    ,language_id INT NOT NULL
    ,original_language_id INT
    ,rental_duration INT NOT NULL DEFAULT (1 + 2 % 3)
    ,rental_rate DECIMAL(4,2) NOT NULL DEFAULT 4.99
    ,length INT
    ,replacement_cost DECIMAL(5,2) NOT NULL DEFAULT 19.99
    ,rating TEXT DEFAULT 'G'
    ,special_features JSON
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,CONSTRAINT film_language_id_fkey FOREIGN KEY (language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,CONSTRAINT film_original_language_id_fkey FOREIGN KEY (original_language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,CONSTRAINT film_release_year_check CHECK (release_year >= 1901 AND release_year <= 2155)
    ,CONSTRAINT film_rating_check CHECK (rating IN ('G','PG','PG-13','R','NC-17'))
);

CREATE INDEX film_title_idx ON film (title);

CREATE INDEX film_language_id_idx ON film (language_id);

CREATE INDEX film_original_language_id_idx ON film (original_language_id);

CREATE VIRTUAL TABLE film_text USING fts5 (
    title
    ,description
    ,content='film'
    ,content_rowid='film_id'
);

CREATE TABLE film_actor (
    actor_id INT NOT NULL
    ,film_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,CONSTRAINT film_actor_actor_id_fkey FOREIGN KEY (actor_id) REFERENCES actor (actor_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,CONSTRAINT film_actor_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX film_actor_actor_id_film_id_idx ON film_actor (actor_id, film_id);

CREATE INDEX film_actor_film_id_idx ON film_actor (film_id);

CREATE TABLE film_category (
    film_id INT NOT NULL
    ,category_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,CONSTRAINT film_category_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,CONSTRAINT film_category_category_id_fkey FOREIGN KEY (category_id) REFERENCES category (category_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TABLE staff (
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

    ,CONSTRAINT staff_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,CONSTRAINT staff_store_id_fkey FOREIGN KEY (store_id) REFERENCES store (store_id)
);

CREATE TABLE store (
    store_id INTEGER PRIMARY KEY
    ,manager_staff_id INT NOT NULL
    ,address_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,CONSTRAINT store_manager_staff_id_fkey FOREIGN KEY (manager_staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,CONSTRAINT store_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX store_manager_staff_id_idx ON store (manager_staff_id);

CREATE TABLE customer (
    customer_id INTEGER PRIMARY KEY
    ,store_id INT NOT NULL
    ,first_name TEXT NOT NULL
    ,last_name TEXT NOT NULL
    ,email TEXT
    ,address_id INT NOT NULL
    ,active BOOLEAN NOT NULL DEFAULT TRUE
    ,data JSON
    ,create_date DATETIME NOT NULL DEFAULT (DATETIME('now'))
    ,last_update DATETIME DEFAULT (DATETIME('now'))

    ,CONSTRAINT customer_email_first_name_last_name_key UNIQUE (email, first_name, last_name)
    ,CONSTRAINT customer_email_key UNIQUE (email)
    ,CONSTRAINT customer_address_id_fkey FOREIGN KEY (address_id) REFERENCES address ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX customer_store_id_idx ON customer (store_id);

CREATE INDEX customer_last_name_idx ON customer (last_name);

CREATE INDEX customer_address_id_idx ON customer (address_id);

CREATE TABLE inventory (
    inventory_id INTEGER PRIMARY KEY
    ,film_id INT NOT NULL
    ,store_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,CONSTRAINT inventory_film_id_fkey FOREIGN KEY (film_id) REFERENCES film ON UPDATE CASCADE ON DELETE RESTRICT
    ,CONSTRAINT inventory_store_id_fkey FOREIGN KEY (store_id) REFERENCES store ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX inventory_store_id_film_id_idx ON inventory (store_id, film_id);

CREATE TABLE rental (
    rental_id INTEGER PRIMARY KEY
    ,rental_date DATETIME NOT NULL
    ,inventory_id INT NOT NULL
    ,customer_id INT NOT NULL
    ,return_date DATETIME
    ,staff_id INT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT (DATETIME('now'))

    ,CONSTRAINT rental_inventory_id_fkey FOREIGN KEY (inventory_id) REFERENCES inventory (inventory_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,CONSTRAINT rental_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,CONSTRAINT rental_staff_id_fkey FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX rental_rental_date_inventory_id_customer_id_idx ON rental (rental_date, inventory_id, customer_id);

CREATE INDEX rental_inventory_id_idx ON rental (inventory_id);

CREATE INDEX rental_customer_id_idx ON rental (customer_id);

CREATE INDEX rental_staff_id_idx ON rental (staff_id);

CREATE TABLE payment (
    payment_id INTEGER PRIMARY KEY
    ,customer_id INT NOT NULL
    ,staff_id INT NOT NULL
    ,rental_id INT
    ,amount DECIMAL(5,2) NOT NULL
    ,payment_date DATETIME NOT NULL

    ,CONSTRAINT payment_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,CONSTRAINT payment_staff_id_fkey FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,CONSTRAINT payment_rental_id_fkey FOREIGN KEY (rental_id) REFERENCES rental (rental_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX payment_customer_id_idx ON payment (customer_id);

CREATE INDEX payment_staff_id_idx ON payment (staff_id);

CREATE TABLE dummy_table (
    id1 INT
    ,id2 TEXT
    ,score INT
    ,color TEXT DEFAULT 'red' COLLATE nocase
    ,data JSON

    ,CONSTRAINT dummy_table_id1_id2_pkey PRIMARY KEY (id1, id2)
    ,CONSTRAINT dummy_table_score_color_key UNIQUE (score, color)
    ,CONSTRAINT dummy_table_score_positive_check CHECK (score > 0)
    ,CONSTRAINT dummy_table_score_id1_greater_than_check CHECK (score > id1)
);

CREATE INDEX dummy_table_complex_expr_idx ON dummy_table (score, (SUBSTR(color, 1, 2)), (color || ' abcd'), (CAST(JSON_EXTRACT(data, '$.age') AS INT))) WHERE color = 'red';

CREATE TABLE dummy_table_2 (
    id1 INT
    ,id2 TEXT

    ,CONSTRAINT dummy_table_2_id1_id2_fkey FOREIGN KEY (id1, id2) REFERENCES dummy_table (id1, id2) ON UPDATE CASCADE ON DELETE RESTRICT
);
