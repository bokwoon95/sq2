DROP VIEW IF EXISTS
    actor_info
    ,customer_list
    ,film_list
    ,full_address
    ,nicer_but_slower_film_list
    ,sales_by_film_category
    ,sales_by_store
    ,staff_list
CASCADE;

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
    ,film_text
    ,inventory
    ,language
    ,payment
    ,rental
    ,staff
    ,store
CASCADE;
