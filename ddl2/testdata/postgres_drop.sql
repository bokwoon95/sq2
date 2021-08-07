DROP VIEW IF EXISTS
    public.actor_info
    ,public.customer_list
    ,public.film_list
    ,public.nicer_but_slower_film_list
    ,public.sales_by_film_category
    ,public.sales_by_store
CASCADE;

DROP MATERIALIZED VIEW public.full_address CASCADE;

DROP TABLE IF EXISTS
    public.actor
    ,public.category
    ,public.country
    ,public.city
    ,public.address
    ,public.language
    ,public.film
    ,public.film_actor
    ,public.film_actor_review
    ,public.film_category
    ,public.staff
    ,public.store
    ,public.customer
    ,public.inventory
    ,public.rental
    ,public.payment
CASCADE;

DROP FUNCTION IF EXISTS last_update_trg() CASCADE;

DROP FUNCTION IF EXISTS refresh_full_address() CASCADE;
