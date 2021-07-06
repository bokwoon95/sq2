DROP FUNCTION IF EXISTS last_update_trg;

CREATE FUNCTION last_update_trg() RETURNS trigger AS $$ BEGIN
    NEW.last_update = NOW();
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER actor_last_update_before_update_trg BEFORE UPDATE ON actor FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER category_last_update_before_update_trg BEFORE UPDATE ON category FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER country_last_update_before_update_trg BEFORE UPDATE ON country FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER city_last_update_before_update_trg BEFORE UPDATE ON city FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER address_last_update_before_update_trg BEFORE UPDATE ON address FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER language_last_update_before_update_trg BEFORE UPDATE ON language FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER film_last_update_before_update_trg BEFORE UPDATE ON film FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER film_fulltext_before_insert_update_trg BEFORE INSERT OR UPDATE ON film FOR EACH ROW EXECUTE PROCEDURE tsvector_update_trigger('fulltext', 'pg_catalog.english', 'title', 'description');

CREATE TRIGGER film_actor_last_update_before_update_trg BEFORE UPDATE ON film_actor FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER film_category_last_update_before_update_trg BEFORE UPDATE ON film_category FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER staff_last_update_before_update_trg BEFORE UPDATE ON staff FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER store_last_update_before_update_trg BEFORE UPDATE ON store FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER customer_last_update_before_update_trg BEFORE UPDATE ON customer FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER inventory_last_update_before_update_trg BEFORE UPDATE ON inventory FOR EACH ROW EXECUTE PROCEDURE last_update_trg();

CREATE TRIGGER rental_last_update_before_update_trg BEFORE UPDATE ON rental FOR EACH ROW EXECUTE PROCEDURE last_update_trg();
