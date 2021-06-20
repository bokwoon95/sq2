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

CREATE TRIGGER actor_last_updated_after_update_trg AFTER UPDATE ON actor BEGIN
    UPDATE actor SET last_update = DATETIME('now') WHERE actor_id = NEW.actor_id;
END;

CREATE TABLE category (
    category_id INTEGER PRIMARY KEY
    ,name TEXT NOT NULL
    ,last_update DATETIME DEFAULT (DATETIME('now')) NOT NULL
);

CREATE TRIGGER category_last_updated_after_update_trg AFTER UPDATE ON category BEGIN
    UPDATE category SET last_update = DATETIME('now') WHERE category_id = NEW.category_id;
END;

CREATE TABLE country (
    country_id INTEGER PRIMARY KEY
    ,country TEXT NOT NULL
    ,last_update DATETIME DEFAULT (DATETIME('now')) NOT NULL
);

CREATE TRIGGER country_last_updated_after_update_trg AFTER UPDATE ON country BEGIN
    UPDATE country SET last_update = DATETIME('now') WHERE country_id = NEW.country_id;
END;

CREATE TABLE city (
    city_id INTEGER PRIMARY KEY
    ,city TEXT NOT NULL
    ,country_id INT NOT NULL
    ,last_update DATETIME DEFAULT (DATETIME('now')) NOT NULL

    ,FOREIGN KEY (country_id) REFERENCES country (country_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX city_country_id_idx ON city (country_id);

CREATE TRIGGER city_last_updated_after_update_trg AFTER UPDATE ON city BEGIN
    UPDATE city SET last_update = DATETIME('now') WHERE city_id = NEW.city_id;
END;

CREATE TABLE address (
    address_id INTEGER PRIMARY KEY
    ,address TEXT NOT NULL
    ,address2 TEXT
    ,district TEXT NOT NULL
    ,city_id INT NOT NULL
    ,postal_code TEXT
    ,phone TEXT NOT NULL
    ,last_update DATETIME DEFAULT (DATETIME('now')) NOT NULL

    ,FOREIGN KEY (city_id) REFERENCES city (city_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX address_city_id_idx ON address (city_id);

CREATE TRIGGER address_last_updated_after_update_trg AFTER UPDATE ON address BEGIN
    UPDATE address SET last_update = DATETIME('now') WHERE address_id = NEW.address_id;
END;

CREATE TABLE language (
    language_id INTEGER PRIMARY KEY
    ,name TEXT NOT NULL
    ,last_update DATETIME DEFAULT (DATETIME('now')) NOT NULL
);

CREATE TRIGGER language_last_updated_after_update_trg AFTER UPDATE ON language BEGIN
    UPDATE language SET last_update = DATETIME('now') WHERE language_id = NEW.language_id;
END;

CREATE TABLE film (
    film_id INTEGER PRIMARY KEY
    ,title TEXT NOT NULL
    ,description TEXT
    ,release_year INT
    ,language_id INT NOT NULL
    ,original_language_id INT
    ,rental_duration INT DEFAULT 3 NOT NULL
    ,rental_rate DECIMAL(4,2) DEFAULT 4.99 NOT NULL
    ,length INT
    ,replacement_cost DECIMAL(5,2) DEFAULT 19.99 NOT NULL
    ,rating TEXT DEFAULT 'G'
    ,special_features JSON
    ,last_update DATETIME DEFAULT (DATETIME('now')) NOT NULL

    ,CONSTRAINT film_release_year_check CHECK (release_year >= 1901 AND release_year <= 2155)
    ,CONSTRAINT film_rating_check CHECK (rating IN ('G','PG','PG-13','R','NC-17'))
    ,FOREIGN KEY (language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (original_language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX film_title_idx ON film (title);

CREATE INDEX film_language_id_idx ON film (language_id);

CREATE INDEX film_original_language_id_idx ON film (original_language_id);

CREATE TRIGGER film_last_updated_after_update_trg AFTER UPDATE ON film BEGIN
    UPDATE film SET last_update = DATETIME('now') WHERE film_id = NEW.film_id;
END;

CREATE VIRTUAL TABLE film_text USING FTS5(title, description, content='film', content_rowid='film_id');

CREATE TRIGGER film_fts5_after_insert_trg AFTER INSERT ON film BEGIN
    INSERT INTO film_text (rowid, title, description) VALUES (NEW.film_id, NEW.title, NEW.description);
END;

CREATE TRIGGER film_fts5_after_delete_trg AFTER DELETE ON film BEGIN
    INSERT INTO film_text (film_text, rowid, title, description) VALUES ('delete', OLD.film_id, OLD.title, OLD.description);
END;

CREATE TRIGGER film_fts5_after_update_trg AFTER UPDATE ON film BEGIN
    INSERT INTO film_text (film_text, rowid, title, description) VALUES ('delete', OLD.film_id, OLD.title, OLD.description);
    INSERT INTO film_text (rowid, title, description) VALUES (NEW.film_id, NEW.title, NEW.description);
END;

CREATE TABLE film_actor (
    actor_id INT NOT NULL
    ,film_id INT NOT NULL
    ,last_update DATETIME DEFAULT (DATETIME('now')) NOT NULL

    ,FOREIGN KEY (actor_id) REFERENCES actor (actor_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX film_actor_actor_id_film_id_idx ON film_actor (actor_id, film_id);

CREATE INDEX film_actor_film_id_idx ON film_actor (film_id);

CREATE TRIGGER film_actor_last_updated_after_update_trg AFTER UPDATE ON film_actor BEGIN
    UPDATE film_actor SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;

CREATE TABLE film_category (
    film_id INT NOT NULL
    ,category_id INT NOT NULL
    ,last_update DATETIME DEFAULT (DATETIME('now')) NOT NULL

    ,FOREIGN KEY (category_id) REFERENCES category (category_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TRIGGER film_category_last_updated_after_update_trg AFTER UPDATE ON film_category BEGIN
    UPDATE film_category SET last_update = DATETIME('now') WHERE film_category_id = NEW.film_category_id;
END;

CREATE TABLE staff (
    staff_id INTEGER PRIMARY KEY
    ,first_name TEXT NOT NULL
    ,last_name TEXT NOT NULL
    ,address_id INT NOT NULL
    ,email TEXT
    ,store_id INT
    ,active BOOLEAN DEFAULT TRUE NOT NULL
    ,username TEXT NOT NULL
    ,password TEXT
    ,last_update DATETIME DEFAULT (DATETIME('now')) NOT NULL
    ,picture BLOB

    ,FOREIGN KEY (store_id) REFERENCES store (store_id)
    ,FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE TRIGGER staff_last_updated_after_update_trg AFTER UPDATE ON staff BEGIN
    UPDATE staff SET last_update = DATETIME('now') WHERE staff_id = NEW.staff_id;
END;

CREATE TABLE store (
    store_id INTEGER PRIMARY KEY
    ,manager_staff_id INT NOT NULL
    ,address_id INT NOT NULL
    ,last_update DATETIME DEFAULT (DATETIME('now')) NOT NULL

    ,FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (manager_staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX store_manager_staff_id_idx ON store (manager_staff_id);

CREATE TRIGGER store_last_updated_after_update_trg AFTER UPDATE ON store BEGIN
    UPDATE store SET last_update = DATETIME('now') WHERE store_id = NEW.store_id;
END;

CREATE TABLE customer (
    customer_id INTEGER PRIMARY KEY
    ,store_id INT NOT NULL
    ,first_name TEXT NOT NULL
    ,last_name TEXT NOT NULL
    ,email TEXT UNIQUE
    ,address_id INT NOT NULL
    ,active BOOLEAN DEFAULT TRUE NOT NULL
    ,create_date DATETIME DEFAULT (DATETIME('now')) NOT NULL
    ,last_update DATETIME DEFAULT (DATETIME('now'))

    ,CONSTRAINT customer_email_first_name_last_name_key UNIQUE (email, first_name, last_name)
    ,FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (store_id) REFERENCES store (store_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX customer_address_id_idx ON customer (address_id);

CREATE INDEX customer_store_id_idx ON customer (store_id);

CREATE INDEX customer_last_name_idx ON customer (last_name);

CREATE TRIGGER customer_last_updated_after_update_trg AFTER UPDATE ON customer BEGIN
    UPDATE customer SET last_update = DATETIME('now') WHERE customer_id = NEW.customer_id;
END;

CREATE TABLE inventory (
    inventory_id INTEGER PRIMARY KEY
    ,film_id INT NOT NULL
    ,store_id INT NOT NULL
    ,last_update DATETIME DEFAULT (DATETIME('now')) NOT NULL

    ,FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (store_id) REFERENCES store (store_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX inventory_store_id_film_id_idx ON inventory (store_id, film_id);

CREATE TRIGGER inventory_last_updated_after_update_trg AFTER UPDATE ON inventory BEGIN
    UPDATE inventory SET last_update = DATETIME('now') WHERE inventory_id = NEW.inventory_id;
END;

CREATE TABLE rental (
    rental_id INTEGER PRIMARY KEY
    ,rental_date DATETIME NOT NULL
    ,inventory_id INT NOT NULL
    ,customer_id INT NOT NULL
    ,return_date DATETIME
    ,staff_id INT NOT NULL
    ,last_update DATETIME DEFAULT (DATETIME('now')) NOT NULL

    ,FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (inventory_id) REFERENCES inventory (inventory_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX rental_rental_date_inventory_id_customer_id_idx ON rental (rental_date, inventory_id, customer_id);

CREATE INDEX rental_inventory_id_idx ON rental (inventory_id);

CREATE INDEX rental_customer_id_idx ON rental (customer_id);

CREATE INDEX rental_staff_id_idx ON rental (staff_id);

CREATE TRIGGER rental_last_updated_after_update_trg AFTER UPDATE ON rental BEGIN
    UPDATE rental SET last_update = DATETIME('now') WHERE rental_id = NEW.rental_id;
END;

CREATE TABLE payment (
    payment_id INTEGER PRIMARY KEY
    ,customer_id INT NOT NULL
    ,staff_id INT NOT NULL
    ,rental_id INT
    ,amount DECIMAL(5,2) NOT NULL
    ,payment_date DATETIME NOT NULL

    ,FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,FOREIGN KEY (rental_id) REFERENCES rental (rental_id) ON UPDATE CASCADE ON DELETE SET NULL
    ,FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX payment_customer_id_idx ON payment (customer_id);

CREATE INDEX payment_staff_id_idx ON payment (staff_id);

CREATE TABLE dummy_table (
    id1 INT
    ,id2 TEXT
    ,score INT
    ,color TEXT COLLATE NOCASE DEFAULT ('red')
    ,data JSON

    ,CONSTRAINT dummy_table_score_positive_check CHECK (score > 0)
    ,CONSTRAINT dummy_table_id1_id2_pkey PRIMARY KEY (id2, id1)
    ,CONSTRAINT dummy_table_score_color_key UNIQUE (score, color)
    ,CONSTRAINT dummy_table_score_id1_greater_than_check CHECK (score > id1)
);

-- CREATE INDEX dummy_table_score_color_data_idx ON dummy_table (score, (CAST(JSON_EXTRACT(data, '$.age') AS INT)), color) WHERE color = 'red';

CREATE INDEX dummy_table_score_color_data_idx ON dummy_table (score, (SUBSTR(color,1,2)), (color || ' abcd'), (CAST(JSON_EXTRACT(data, '$.age') AS INT))) WHERE color = 'red';

CREATE TABLE dummy_table_2 (
    id1 INT
    ,id2 TEXT

    ,FOREIGN KEY (id2, id1) REFERENCES dummy_table (id2, id1) ON UPDATE CASCADE ON DELETE RESTRICT
);
