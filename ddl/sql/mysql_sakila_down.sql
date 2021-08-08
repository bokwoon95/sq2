DROP VIEW IF EXISTS
    db.actor_info
    ,db.customer_list
    ,db.film_list
    ,db.full_address
    ,db.nicer_but_slower_film_list
    ,db.sales_by_film_category
    ,db.sales_by_store
    ,db.staff_list
CASCADE;

DROP TABLE IF EXISTS
    db.actor
    ,db.address
    ,db.category
    ,db.city
    ,db.country
    ,db.customer
    ,db.film
    ,db.film_actor
    ,db.film_actor_review
    ,db.film_category
    ,db.film_text
    ,db.inventory
    ,db.language
    ,db.payment
    ,db.rental
    ,db.staff
    ,db.store
CASCADE;
