DROP TABLE IF EXISTS public.dummy_table_2 CASCADE;
DROP TABLE IF EXISTS public.dummy_table CASCADE;
DROP TABLE IF EXISTS public.payment CASCADE;
DROP TABLE IF EXISTS public.rental CASCADE;
DROP TABLE IF EXISTS public.inventory CASCADE;
DROP TABLE IF EXISTS public.customer CASCADE;
DROP TABLE IF EXISTS public.store CASCADE;
DROP TABLE IF EXISTS public.staff CASCADE;
DROP TABLE IF EXISTS public.film_category CASCADE;
DROP TABLE IF EXISTS public.film_actor CASCADE;
DROP TABLE IF EXISTS public.film CASCADE;
DROP TABLE IF EXISTS public.language CASCADE;
DROP TABLE IF EXISTS public.address CASCADE;
DROP TABLE IF EXISTS public.city CASCADE;
DROP TABLE IF EXISTS public.country CASCADE;
DROP TABLE IF EXISTS public.category CASCADE;
DROP TABLE IF EXISTS public.actor CASCADE;

CREATE TABLE IF NOT EXISTS public.actor (
    actor_id INT GENERATED BY DEFAULT AS IDENTITY
    ,first_name TEXT NOT NULL
    ,last_name TEXT NOT NULL
    ,full_name TEXT GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED
    ,full_name_reversed TEXT GENERATED ALWAYS AS (last_name || ' ' || first_name) STORED
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT actor_actor_id_pkey PRIMARY KEY (actor_id)
);

CREATE INDEX IF NOT EXISTS actor_last_name_idx ON public.actor (last_name);

CREATE TABLE IF NOT EXISTS public.category (
    category_id INT GENERATED BY DEFAULT AS IDENTITY
    ,name TEXT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT category_category_id_pkey PRIMARY KEY (category_id)
);

CREATE TABLE IF NOT EXISTS public.country (
    country_id INT GENERATED BY DEFAULT AS IDENTITY
    ,country TEXT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT country_country_id_pkey PRIMARY KEY (country_id)
);

CREATE TABLE IF NOT EXISTS public.city (
    city_id INT GENERATED BY DEFAULT AS IDENTITY
    ,city TEXT NOT NULL
    ,country_id INT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT city_city_id_pkey PRIMARY KEY (city_id)
);

CREATE INDEX IF NOT EXISTS city_country_id_idx ON public.city (country_id);

CREATE TABLE IF NOT EXISTS public.address (
    address_id INT GENERATED BY DEFAULT AS IDENTITY
    ,address TEXT NOT NULL
    ,address2 TEXT
    ,district TEXT NOT NULL
    ,city_id INT NOT NULL
    ,postal_code TEXT
    ,phone TEXT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT address_address_id_pkey PRIMARY KEY (address_id)
);

CREATE INDEX IF NOT EXISTS address_city_id_idx ON public.address (city_id);

CREATE TABLE IF NOT EXISTS public.language (
    language_id INT GENERATED BY DEFAULT AS IDENTITY
    ,name TEXT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT language_language_id_pkey PRIMARY KEY (language_id)
);

CREATE TABLE IF NOT EXISTS public.film (
    film_id INT GENERATED BY DEFAULT AS IDENTITY
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
    ,special_features TEXT[]
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()
    ,fulltext TSVECTOR

    ,CONSTRAINT film_film_id_pkey PRIMARY KEY (film_id)
    ,CONSTRAINT film_release_year_check CHECK (release_year >= 1901 AND release_year <= 2155)
    ,CONSTRAINT film_rating_check CHECK (rating IN ('G','PG','PG-13','R','NC-17'))
);

CREATE INDEX IF NOT EXISTS film_title_idx ON public.film (title);

CREATE INDEX IF NOT EXISTS film_language_id_idx ON public.film (language_id);

CREATE INDEX IF NOT EXISTS film_original_language_id_idx ON public.film (original_language_id);

CREATE INDEX IF NOT EXISTS film_fulltext_idx ON public.film USING GIST (fulltext);

CREATE TABLE IF NOT EXISTS public.film_text (
    film_id INT
    ,title TEXT
    ,description TEXT
);

CREATE TABLE IF NOT EXISTS public.film_actor (
    actor_id INT NOT NULL
    ,film_id INT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS film_actor_actor_id_film_id_idx ON public.film_actor (actor_id, film_id);

CREATE INDEX IF NOT EXISTS film_actor_film_id_idx ON public.film_actor (film_id);

CREATE TABLE IF NOT EXISTS public.film_category (
    film_id INT NOT NULL
    ,category_id INT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS public.staff (
    staff_id INT GENERATED BY DEFAULT AS IDENTITY
    ,first_name TEXT NOT NULL
    ,last_name TEXT NOT NULL
    ,address_id INT NOT NULL
    ,email TEXT
    ,store_id INT
    ,active BOOLEAN NOT NULL DEFAULT TRUE
    ,username TEXT NOT NULL
    ,password TEXT
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()
    ,picture BYTEA

    ,CONSTRAINT staff_staff_id_pkey PRIMARY KEY (staff_id)
);

CREATE TABLE IF NOT EXISTS public.store (
    store_id INT GENERATED BY DEFAULT AS IDENTITY
    ,manager_staff_id INT NOT NULL
    ,address_id INT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT store_store_id_pkey PRIMARY KEY (store_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS store_manager_staff_id_idx ON public.store (manager_staff_id);

CREATE TABLE IF NOT EXISTS public.customer (
    customer_id INT GENERATED BY DEFAULT AS IDENTITY
    ,store_id INT NOT NULL
    ,first_name TEXT NOT NULL
    ,last_name TEXT NOT NULL
    ,email TEXT
    ,address_id INT NOT NULL
    ,active BOOLEAN NOT NULL DEFAULT TRUE
    ,data JSONB
    ,create_date TIMESTAMPTZ NOT NULL DEFAULT NOW()
    ,last_update TIMESTAMPTZ DEFAULT NOW()

    ,CONSTRAINT customer_email_first_name_last_name_key UNIQUE (email, first_name, last_name)
    ,CONSTRAINT customer_customer_id_pkey PRIMARY KEY (customer_id)
    ,CONSTRAINT customer_email_key UNIQUE (email)
);

CREATE INDEX IF NOT EXISTS customer_store_id_idx ON public.customer (store_id);

CREATE INDEX IF NOT EXISTS customer_last_name_idx ON public.customer (last_name);

CREATE INDEX IF NOT EXISTS customer_address_id_idx ON public.customer (address_id);

CREATE TABLE IF NOT EXISTS public.inventory (
    inventory_id INT GENERATED BY DEFAULT AS IDENTITY
    ,film_id INT NOT NULL
    ,store_id INT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT inventory_inventory_id_pkey PRIMARY KEY (inventory_id)
);

CREATE INDEX IF NOT EXISTS inventory_store_id_film_id_idx ON public.inventory (store_id, film_id);

CREATE TABLE IF NOT EXISTS public.rental (
    rental_id INT GENERATED BY DEFAULT AS IDENTITY
    ,rental_date TIMESTAMPTZ NOT NULL
    ,inventory_id INT NOT NULL
    ,customer_id INT NOT NULL
    ,return_date TIMESTAMPTZ
    ,staff_id INT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT rental_rental_id_pkey PRIMARY KEY (rental_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS rental_rental_date_inventory_id_customer_id_idx ON public.rental (rental_date, inventory_id, customer_id);

CREATE INDEX IF NOT EXISTS rental_inventory_id_idx ON public.rental (inventory_id);

CREATE INDEX IF NOT EXISTS rental_customer_id_idx ON public.rental (customer_id);

CREATE INDEX IF NOT EXISTS rental_staff_id_idx ON public.rental (staff_id);

CREATE TABLE IF NOT EXISTS public.payment (
    payment_id INT GENERATED BY DEFAULT AS IDENTITY
    ,customer_id INT NOT NULL
    ,staff_id INT NOT NULL
    ,rental_id INT
    ,amount DECIMAL(5,2) NOT NULL
    ,payment_date TIMESTAMPTZ NOT NULL

    ,CONSTRAINT payment_payment_id_pkey PRIMARY KEY (payment_id)
);

CREATE INDEX IF NOT EXISTS payment_customer_id_idx ON public.payment (customer_id);

CREATE INDEX IF NOT EXISTS payment_staff_id_idx ON public.payment (staff_id);

CREATE TABLE IF NOT EXISTS public.dummy_table (
    id1 INT GENERATED ALWAYS AS IDENTITY
    ,id2 TEXT
    ,score INT
    ,color TEXT DEFAULT 'red' COLLATE "C"
    ,data JSONB

    ,CONSTRAINT dummy_table_id1_id2_pkey PRIMARY KEY (id1, id2)
    ,CONSTRAINT dummy_table_score_color_key UNIQUE (score, color)
    ,CONSTRAINT dummy_table_score_positive_check CHECK (score > 0)
    ,CONSTRAINT dummy_table_score_id1_greater_than_check CHECK (score > id1)
);

CREATE INDEX IF NOT EXISTS dummy_table_complex_expr_idx ON public.dummy_table (score, (SUBSTR(color, 1, 2)), (color || ' abcd'), ((data->>'age')::INT)) WHERE color = 'red';

CREATE INDEX IF NOT EXISTS dummy_table_id2_idx ON public.dummy_table (id2 COLLATE "C");

CREATE INDEX IF NOT EXISTS dummy_table_color_idx ON public.dummy_table (color text_pattern_ops);

CREATE TABLE IF NOT EXISTS public.dummy_table_2 (
    id1 INT
    ,id2 TEXT
);

CREATE OR REPLACE VIEW public.actor_info AS
SELECT
    a.actor_id
    ,a.first_name
    ,a.last_name
    ,jsonb_object_agg(c.name, (
        SELECT
            jsonb_agg(f.title)
        FROM
            public.film AS f
            JOIN public.film_category AS fc ON fc.film_id = f.film_id
            JOIN public.film_actor AS fa ON fa.film_id = f.film_id
        WHERE
            fc.category_id = c.category_id
            AND fa.actor_id = a.actor_id
        GROUP BY
            fa.actor_id
    )) AS film_info
FROM
    public.actor AS a
    LEFT JOIN public.film_actor AS fa ON fa.actor_id = a.actor_id
    LEFT JOIN public.film_category AS fc ON fc.film_id = fa.film_id
    LEFT JOIN public.category AS c ON c.category_id = fc.category_id
GROUP BY
    a.actor_id
    ,a.first_name
    ,a.last_name
;

CREATE OR REPLACE VIEW public.customer_list AS
SELECT
    cu.customer_id AS id
    ,cu.first_name || ' ' || cu.last_name AS name
    ,a.address, a.postal_code AS "zip code"
    ,a.phone
    ,city.city
    ,country.country
    ,CASE
        WHEN cu.active THEN 'active'
        ELSE ''
    END AS notes
    ,cu.store_id AS sid
FROM
    public.customer AS cu
    JOIN public.address AS a ON a.address_id = cu.address_id
    JOIN public.city ON city.city_id = a.city_id
    JOIN public.country ON country.country_id = city.country_id
;

CREATE OR REPLACE VIEW public.film_list AS
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
    public.category
    LEFT JOIN public.film_category ON film_category.category_id = category.category_id
    LEFT JOIN public.film ON film.film_id = film_category.film_id
    JOIN public.film_actor ON film_actor.film_id = film.film_id
    JOIN public.actor ON actor.actor_id = film_actor.actor_id
GROUP BY
    film.film_id
    ,film.title
    ,film.description
    ,category.name
    ,film.rental_rate
    ,film.length
    ,film.rating
;

CREATE OR REPLACE VIEW public.nicer_but_slower_film_list AS
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
    public.category
    LEFT JOIN public.film_category ON film_category.category_id = category.category_id
    LEFT JOIN public.film ON film.film_id = film_category.film_id
    JOIN public.film_actor ON film_actor.film_id = film.film_id
    JOIN public.actor ON actor.actor_id = film_actor.actor_id
GROUP BY
    film.film_id
    ,film.title
    ,film.description
    ,category.name
    ,film.rental_rate
    ,film.length
    ,film.rating
;

CREATE OR REPLACE VIEW public.sales_by_film_category AS
SELECT
    c.name AS category
    ,SUM(p.amount) AS total_sales
FROM
    public.payment AS p
    JOIN public.rental AS r ON r.rental_id = p.rental_id
    JOIN public.inventory AS i ON i.inventory_id = r.inventory_id
    JOIN public.film AS f ON f.film_id = i.film_id
    JOIN public.film_category AS fc ON fc.film_id = f.film_id
    JOIN public.category AS c ON c.category_id = fc.category_id
GROUP BY
    c.name
ORDER BY
    SUM(p.amount) DESC
;

CREATE OR REPLACE VIEW public.sales_by_store AS
SELECT
    ci.city || ',' || co.country AS store
    ,m.first_name || ' ' || m.last_name AS manager
    ,SUM(p.amount) AS total_sales
FROM
    public.payment AS p
    JOIN public.rental AS r ON r.rental_id = p.rental_id
    JOIN public.inventory AS i ON i.inventory_id = r.inventory_id
    JOIN public.store AS s ON s.store_id = i.store_id
    JOIN public.address AS a ON a.address_id = s.address_id
    JOIN public.city AS ci ON ci.city_id = a.city_id
    JOIN public.country AS co ON co.country_id = ci.country_id
    JOIN public.staff AS m ON m.staff_id = s.manager_staff_id
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

CREATE OR REPLACE VIEW public.staff_list AS SELECT s.staff_id AS id, s.first_name || ' ' || s.last_name AS name, a.address, a.postal_code AS "zip code", a.phone, ci.city, co.country, s.store_id AS sid FROM public.staff AS s JOIN public.address AS a ON a.address_id = s.address_id JOIN public.city AS ci ON ci.city_id = a.city_id JOIN public.country AS co ON co.country_id = ci.country_id;

CREATE OR REPLACE FUNCTION last_update_trg() RETURNS trigger AS $$ BEGIN
    NEW.last_update = NOW();
    RETURN NEW;
END; $$ LANGUAGE plpgsql;

CREATE TRIGGER actor_last_update_before_update_trg BEFORE UPDATE ON public.actor
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER category_last_update_before_update_trg BEFORE UPDATE ON public.category
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER country_last_update_before_update_trg BEFORE UPDATE ON public.country
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER city_last_update_before_update_trg BEFORE UPDATE ON public.city
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER city_last_update_before_update_trg BEFORE UPDATE ON public.address
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER language_last_update_before_update_trg BEFORE UPDATE ON public.language
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER film_last_update_before_update_trg BEFORE UPDATE ON public.film
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER film_fulltext_before_insert_update_trg BEFORE INSERT OR UPDATE ON public.film
FOR EACH ROW EXECUTE PROCEDURE tsvector_update_trigger(fulltext, 'pg_catalog.english', title, description);

CREATE TRIGGER film_actor_last_update_before_update_trg BEFORE UPDATE ON public.film_actor
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER film_category_last_update_before_update_trg BEFORE UPDATE ON public.film_category
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER staff_last_update_before_update_trg BEFORE UPDATE ON public.staff
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER store_last_update_before_update_trg BEFORE UPDATE ON public.store
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER customer_last_update_before_update_trg BEFORE UPDATE ON public.customer
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER inventory_last_update_before_update_trg BEFORE UPDATE ON public.inventory
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER rental_last_update_before_update_trg BEFORE UPDATE ON public.rental
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

ALTER TABLE IF EXISTS public.city
    ADD CONSTRAINT city_country_id_fkey FOREIGN KEY (country_id) REFERENCES country ON UPDATE CASCADE ON DELETE RESTRICT
;

ALTER TABLE IF EXISTS public.address
    ADD CONSTRAINT address_city_id_fkey FOREIGN KEY (city_id) REFERENCES city (city_id) ON UPDATE CASCADE ON DELETE RESTRICT
;

ALTER TABLE IF EXISTS public.film
    ADD CONSTRAINT film_language_id_fkey FOREIGN KEY (language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT film_original_language_id_fkey FOREIGN KEY (original_language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT
;

ALTER TABLE IF EXISTS public.film_actor
    ADD CONSTRAINT film_actor_actor_id_fkey FOREIGN KEY (actor_id) REFERENCES actor (actor_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT film_actor_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT
;

ALTER TABLE IF EXISTS public.film_category
    ADD CONSTRAINT film_category_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT film_category_category_id_fkey FOREIGN KEY (category_id) REFERENCES category (category_id) ON UPDATE CASCADE ON DELETE RESTRICT
;

ALTER TABLE IF EXISTS public.staff
    ADD CONSTRAINT staff_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT staff_store_id_fkey FOREIGN KEY (store_id) REFERENCES store (store_id)
;

ALTER TABLE IF EXISTS public.store
    ADD CONSTRAINT store_manager_staff_id_fkey FOREIGN KEY (manager_staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT store_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT
;

ALTER TABLE IF EXISTS public.customer
    ADD CONSTRAINT customer_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT
;

ALTER TABLE IF EXISTS public.inventory
    ADD CONSTRAINT inventory_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT inventory_store_id_fkey FOREIGN KEY (store_id) REFERENCES store (store_id) ON UPDATE CASCADE ON DELETE RESTRICT
;

ALTER TABLE IF EXISTS public.rental
    ADD CONSTRAINT rental_inventory_id_fkey FOREIGN KEY (inventory_id) REFERENCES inventory (inventory_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT rental_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT rental_staff_id_fkey FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT
;

ALTER TABLE IF EXISTS public.payment
    ADD CONSTRAINT payment_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT payment_staff_id_fkey FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT
    ,ADD CONSTRAINT payment_rental_id_fkey FOREIGN KEY (rental_id) REFERENCES rental (rental_id) ON UPDATE CASCADE ON DELETE RESTRICT
;

ALTER TABLE IF EXISTS public.dummy_table_2
    ADD CONSTRAINT dummy_table_2_id1_id2_fkey FOREIGN KEY (id1, id2) REFERENCES dummy_table (id1, id2) ON UPDATE CASCADE ON DELETE RESTRICT
;
