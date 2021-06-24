DELIMITER ;;

-- ddl: film_after_insert_trg
CREATE TRIGGER film_after_insert_trg AFTER INSERT ON film FOR EACH ROW BEGIN
    INSERT INTO film_text
        (film_id, title, description)
    VALUES
        (NEW.film_id, NEW.title, NEW.description)
    ;
END;;

-- ddl: film_after_update_trg
CREATE TRIGGER film_after_update_trg AFTER UPDATE ON film FOR EACH ROW BEGIN
    IF OLD.title <> NEW.title OR OLD.description <> NEW.description THEN
        UPDATE
            film_text
        SET
            title = NEW.title
            ,description = NEW.description
            ,film_id = NEW.film_id
        WHERE
            film_id = OLD.film_id
        ;
    END IF;
END;;

-- ddl: film_after_delete_trg
CREATE TRIGGER film_after_delete_trg AFTER DELETE ON film FOR EACH ROW BEGIN
    DELETE FROM film_text WHERE film_id = OLD.film_id;
END;;

DELIMITER ;
