CREATE TRIGGER actor_last_updated_after_update_trg AFTER UPDATE ON actor BEGIN
    UPDATE actor SET last_update = DATETIME('now') WHERE actor_id = NEW.actor_id;
END;

CREATE TRIGGER category_last_updated_after_update_trg AFTER UPDATE ON category BEGIN
    UPDATE category SET last_update = DATETIME('now') WHERE category_id = NEW.category_id;
END;

CREATE TRIGGER country_last_updated_after_update_trg AFTER UPDATE ON country BEGIN
    UPDATE country SET last_update = DATETIME('now') WHERE country_id = NEW.country_id;
END;

CREATE TRIGGER city_last_updated_after_update_trg AFTER UPDATE ON city BEGIN
    UPDATE city SET last_update = DATETIME('now') WHERE city_id = NEW.city_id;
END;

CREATE TRIGGER address_last_updated_after_update_trg AFTER UPDATE ON address BEGIN
    UPDATE address SET last_update = DATETIME('now') WHERE address_id = NEW.address_id;
END;

CREATE TRIGGER language_last_updated_after_update_trg AFTER UPDATE ON language BEGIN
    UPDATE language SET last_update = DATETIME('now') WHERE language_id = NEW.language_id;
END;

CREATE TRIGGER film_last_updated_after_update_trg AFTER UPDATE ON film BEGIN
    UPDATE film SET last_update = DATETIME('now') WHERE film_id = NEW.film_id;
END;

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

CREATE TRIGGER film_actor_last_updated_after_update_trg AFTER UPDATE ON film_actor BEGIN
    UPDATE film_actor SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;

CREATE TRIGGER film_category_last_updated_after_update_trg AFTER UPDATE ON film_category BEGIN
    UPDATE film_category SET last_update = DATETIME('now') WHERE film_category_id = NEW.film_category_id;
END;

CREATE TRIGGER staff_last_updated_after_update_trg AFTER UPDATE ON staff BEGIN
    UPDATE staff SET last_update = DATETIME('now') WHERE staff_id = NEW.staff_id;
END;

CREATE TRIGGER store_last_updated_after_update_trg AFTER UPDATE ON store BEGIN
    UPDATE store SET last_update = DATETIME('now') WHERE store_id = NEW.store_id;
END;

CREATE TRIGGER customer_last_updated_after_update_trg AFTER UPDATE ON customer BEGIN
    UPDATE customer SET last_update = DATETIME('now') WHERE customer_id = NEW.customer_id;
END;

CREATE TRIGGER inventory_last_updated_after_update_trg AFTER UPDATE ON inventory BEGIN
    UPDATE inventory SET last_update = DATETIME('now') WHERE inventory_id = NEW.inventory_id;
END;

CREATE TRIGGER rental_last_updated_after_update_trg AFTER UPDATE ON rental BEGIN
    UPDATE rental SET last_update = DATETIME('now') WHERE rental_id = NEW.rental_id;
END;
