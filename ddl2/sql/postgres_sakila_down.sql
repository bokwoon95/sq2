DROP VIEW IF EXISTS
    public.actor_info
    ,public.customer_list
    ,public.film_list
    ,public.nicer_but_slower_film_list
    ,public.sales_by_film_category
    ,public.sales_by_store
    ,public.staff_list
CASCADE;

DROP MATERIALIZED VIEW IF EXISTS public.full_address CASCADE;

DROP TABLE IF EXISTS
    public.actor
    ,public.address
    ,public.category
    ,public.city
    ,public.country
    ,public.customer
    ,public.film
    ,public.film_actor
    ,public.film_actor_review
    ,public.film_category
    ,public.inventory
    ,public.language
    ,public.payment
    ,public.rental
    ,public.staff
    ,public.store
CASCADE;

DROP FUNCTION IF EXISTS public.last_update_trg() CASCADE;

DROP FUNCTION IF EXISTS public.refresh_full_address() CASCADE;

DROP EXTENSION IF EXISTS btree_gist, "uuid-ossp" CASCADE;
