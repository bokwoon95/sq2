DROP VIEW IF EXISTS
    actor_info
    ,customer_list
    ,film_list
    ,nicer_but_slower_film_list
    ,sales_by_film_category
    ,sales_by_store
    ,staff_list
CASCADE;

DROP MATERIALIZED VIEW IF EXISTS full_address CASCADE;

DROP TABLE IF EXISTS
    actor
    ,address
    ,category
    ,city
    ,country
    ,customer
    ,film
    ,film_actor
    ,film_actor_review
    ,film_category
    ,inventory
    ,language
    ,payment
    ,rental
    ,staff
    ,store
CASCADE;

DROP FUNCTION IF EXISTS last_update_trg() CASCADE;

DROP FUNCTION IF EXISTS refresh_full_address() CASCADE;

DROP EXTENSION IF EXISTS btree_gist, "uuid-ossp" CASCADE;
