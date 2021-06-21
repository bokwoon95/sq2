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
DROP TYPE IF EXISTS mpaa_rating CASCADE;
DROP DOMAIN IF EXISTS year CASCADE;
CREATE DOMAIN year AS INT CONSTRAINT year_check CHECK (VALUE >= 1901 AND VALUE <= 2155);
CREATE TYPE mpaa_rating AS ENUM ('G', 'PG', 'PG-13', 'R', 'NC-17');

CREATE TABLE public.actor (
    actor_id INT GENERATED ALWAYS AS IDENTITY
    ,first_name TEXT NOT NULL
    ,last_name TEXT NOT NULL
    ,full_name TEXT GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED
    ,full_name_reversed TEXT GENERATED ALWAYS AS (last_name || ' ' || first_name) STORED
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT actor_actor_id_pkey PRIMARY KEY (actor_id)
);

CREATE INDEX actor_last_name_idx ON public.actor (last_name);

CREATE TABLE public.category (
    category_id INT GENERATED BY DEFAULT AS IDENTITY
    ,name TEXT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT category_category_id_pkey PRIMARY KEY (category_id)
);

CREATE TABLE public.country (
    country_id INT GENERATED BY DEFAULT AS IDENTITY
    ,country TEXT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT country_country_id_pkey PRIMARY KEY (country_id)
);

CREATE TABLE public.city (
    city_id INT GENERATED BY DEFAULT AS IDENTITY
    ,city TEXT NOT NULL
    ,country_id INT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT city_city_id_pkey PRIMARY KEY (city_id)
);

CREATE INDEX city_country_id_idx ON public.city (country_id);

CREATE TABLE public.address (
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

CREATE INDEX address_city_id_idx ON public.address (city_id);

CREATE TABLE public.language (
    language_id INT GENERATED BY DEFAULT AS IDENTITY
    ,name TEXT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT language_language_id_pkey PRIMARY KEY (language_id)
);

CREATE TABLE public.film (
    film_id INT GENERATED BY DEFAULT AS IDENTITY
    ,title TEXT NOT NULL
    ,description TEXT
    ,release_year year
    ,language_id INT NOT NULL
    ,original_language_id INT
    ,rental_duration INT NOT NULL DEFAULT 3
    ,rental_rate DECIMAL(4,2) NOT NULL DEFAULT 4.99
    ,length INT
    ,replacement_cost DECIMAL(5,2) NOT NULL DEFAULT 19.99
    ,rating mpaa_rating DEFAULT 'G'::mpaa_rating
    ,special_features TEXT[]
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()
    ,fulltext TSVECTOR NOT NULL

    ,CONSTRAINT film_film_id_pkey PRIMARY KEY (film_id)
);

CREATE INDEX film_title_idx ON public.film (title);

CREATE INDEX film_language_id_idx ON public.film (language_id);

CREATE INDEX film_original_language_id_idx ON public.film (original_language_id);

CREATE TABLE public.film_actor (
    actor_id INT NOT NULL
    ,film_id INT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX film_actor_actor_id_film_id_idx ON public.film_actor (actor_id, film_id);

CREATE INDEX film_actor_film_id_idx ON public.film_actor (film_id);

CREATE TABLE public.film_category (
    film_id INT NOT NULL
    ,category_id INT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE public.staff (
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

CREATE TABLE public.store (
    store_id INT GENERATED BY DEFAULT AS IDENTITY
    ,manager_staff_id INT NOT NULL
    ,address_id INT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT store_store_id_pkey PRIMARY KEY (store_id)
);

CREATE UNIQUE INDEX store_manager_staff_id_idx ON public.store (manager_staff_id);

CREATE TABLE public.customer (
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

CREATE INDEX customer_store_id_idx ON public.customer (store_id);

CREATE INDEX customer_last_name_idx ON public.customer (last_name);

CREATE INDEX customer_address_id_idx ON public.customer (address_id);

CREATE TABLE public.inventory (
    inventory_id INT GENERATED BY DEFAULT AS IDENTITY
    ,film_id INT NOT NULL
    ,store_id INT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT inventory_inventory_id_pkey PRIMARY KEY (inventory_id)
);

CREATE INDEX inventory_store_id_film_id_idx ON public.inventory (store_id, film_id);

CREATE TABLE public.rental (
    rental_id INT GENERATED BY DEFAULT AS IDENTITY
    ,rental_date TIMESTAMPTZ NOT NULL
    ,inventory_id INT NOT NULL
    ,customer_id INT NOT NULL
    ,return_date TIMESTAMPTZ
    ,staff_id INT NOT NULL
    ,last_update TIMESTAMPTZ NOT NULL DEFAULT NOW()

    ,CONSTRAINT rental_rental_id_pkey PRIMARY KEY (rental_id)
);

CREATE UNIQUE INDEX rental_rental_date_inventory_id_customer_id_idx ON public.rental (rental_date, inventory_id, customer_id);

CREATE INDEX rental_inventory_id_idx ON public.rental (inventory_id);

CREATE INDEX rental_customer_id_idx ON public.rental (customer_id);

CREATE INDEX rental_staff_id_idx ON public.rental (staff_id);

CREATE TABLE public.payment (
    payment_id INT GENERATED BY DEFAULT AS IDENTITY
    ,customer_id INT NOT NULL
    ,staff_id INT NOT NULL
    ,rental_id INT
    ,amount DECIMAL(5,2) NOT NULL
    ,payment_date TIMESTAMPTZ NOT NULL

    ,CONSTRAINT payment_payment_id_pkey PRIMARY KEY (payment_id)
);

CREATE INDEX payment_customer_id_idx ON public.payment (customer_id);

CREATE INDEX payment_staff_id_idx ON public.payment (staff_id);

CREATE TABLE public.dummy_table (
    id1 INT
    ,id2 TEXT
    ,score INT
    ,color TEXT DEFAULT 'red' COLLATE "C"
    ,data JSONB

    ,CONSTRAINT dummy_table_id1_id2_pkey PRIMARY KEY (id1, id2)
    ,CONSTRAINT dummy_table_score_color_key UNIQUE (score, color)
    ,CONSTRAINT dummy_table_score_positive_check CHECK (score > 0)
    ,CONSTRAINT dummy_table_score_id1_greater_than_check CHECK (score > id1)
);

CREATE INDEX dummy_table_complex_expr_idx ON public.dummy_table (score, (SUBSTR(color, 1, 2)), (color || ' abcd'), ((data->>'age')::INT)) WHERE color = 'red';

CREATE TABLE public.dummy_table_2 (
    id1 INT
    ,id2 TEXT
);

ALTER TABLE public.city ADD CONSTRAINT city_country_id_fkey FOREIGN KEY (country_id) REFERENCES country ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.address ADD CONSTRAINT address_city_id_fkey FOREIGN KEY (city_id) REFERENCES city (city_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.film ADD CONSTRAINT film_language_id_fkey FOREIGN KEY (language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.film ADD CONSTRAINT film_original_language_id_fkey FOREIGN KEY (original_language_id) REFERENCES language (language_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.film_actor ADD CONSTRAINT film_actor_actor_id_fkey FOREIGN KEY (actor_id) REFERENCES actor (actor_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.film_actor ADD CONSTRAINT film_actor_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.film_category ADD CONSTRAINT film_category_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.film_category ADD CONSTRAINT film_category_category_id_fkey FOREIGN KEY (category_id) REFERENCES category (category_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.staff ADD CONSTRAINT staff_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.staff ADD CONSTRAINT staff_store_id_fkey FOREIGN KEY (store_id) REFERENCES store (store_id);

ALTER TABLE public.store ADD CONSTRAINT store_manager_staff_id_fkey FOREIGN KEY (manager_staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.store ADD CONSTRAINT store_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.customer ADD CONSTRAINT customer_address_id_fkey FOREIGN KEY (address_id) REFERENCES address (address_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.inventory ADD CONSTRAINT inventory_film_id_fkey FOREIGN KEY (film_id) REFERENCES film (film_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.inventory ADD CONSTRAINT inventory_store_id_fkey FOREIGN KEY (store_id) REFERENCES store (store_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.rental ADD CONSTRAINT rental_inventory_id_fkey FOREIGN KEY (inventory_id) REFERENCES inventory (inventory_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.rental ADD CONSTRAINT rental_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.rental ADD CONSTRAINT rental_staff_id_fkey FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.payment ADD CONSTRAINT payment_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES customer (customer_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.payment ADD CONSTRAINT payment_staff_id_fkey FOREIGN KEY (staff_id) REFERENCES staff (staff_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.payment ADD CONSTRAINT payment_rental_id_fkey FOREIGN KEY (rental_id) REFERENCES rental (rental_id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.dummy_table_2 ADD CONSTRAINT dummy_table_2_id1_id2_fkey FOREIGN KEY (id1, id2) REFERENCES dummy_table (id1, id2) ON UPDATE CASCADE ON DELETE RESTRICT;
