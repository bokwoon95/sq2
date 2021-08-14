CREATE FUNCTION film_not_in_stock(p_film_id INT, p_store_id INT) RETURNS INT READS SQL DATA BEGIN
    DECLARE v_count INT;

    SELECT COUNT(*)
    INTO v_count
    FROM inventory
    WHERE film_id = p_film_id AND store_id = p_store_id AND NOT inventory_in_stock(inventory_id)
    ;

    RETURN v_count;
END
