DROP FUNCTION IF EXISTS inventory_in_stock;
DROP FUNCTION IF EXISTS film_in_stock;
DROP FUNCTION IF EXISTS film_not_in_stock;
DROP FUNCTION IF EXISTS get_customer_balance;
DROP FUNCTION IF EXISTS inventory_held_by_customer;
DROP FUNCTION IF EXISTS last_day;
DROP FUNCTION IF EXISTS rewards_report;

CREATE FUNCTION inventory_in_stock(p_inventory_id INT) RETURNS BOOLEAN AS $$ DECLARE
    v_rentals INT;
    v_out     INT;
BEGIN
    -- AN ITEM IS IN-STOCK IF THERE ARE EITHER NO ROWS IN THE rental TABLE
    -- FOR THE ITEM OR ALL ROWS HAVE return_date POPULATED
    SELECT COUNT(*) INTO v_rentals
    FROM rental
    WHERE inventory_id = p_inventory_id
    ;
    IF v_rentals = 0 THEN
      RETURN TRUE;
    END IF;
    SELECT
        COUNT(rental_id) INTO v_out
    FROM
        inventory
        LEFT JOIN rental USING(inventory_id)
    WHERE
        inventory.inventory_id = p_inventory_id
        AND rental.return_date IS NULL
    ;
    IF v_out > 0 THEN
      RETURN FALSE;
    ELSE
      RETURN TRUE;
    END IF;
END $$ LANGUAGE plpgsql;

CREATE FUNCTION film_in_stock(p_film_id INT, p_store_id INT, OUT p_film_count INT) RETURNS SETOF INT AS $$
    SELECT inventory_id
    FROM inventory
    WHERE film_id = p_film_id AND store_id = p_store_id AND inventory_in_stock(inventory_id);
$$ LANGUAGE sql;

CREATE FUNCTION film_not_in_stock(p_film_id INT, p_store_id INT, OUT p_film_count INT) RETURNS SETOF INT AS $$
    SELECT inventory_id
    FROM inventory
    WHERE film_id = p_film_id AND store_id = p_store_id AND NOT inventory_in_stock(inventory_id);
$$ LANGUAGE sql;

CREATE FUNCTION get_customer_balance(p_customer_id INT, p_effective_date TIMESTAMPTZ) RETURNS DECIMAL AS $$ DECLARE
    -- OK, WE NEED TO CALCULATE THE CURRENT BALANCE GIVEN A CUSTOMER_ID AND A DATE
    -- THAT WE WANT THE BALANCE TO BE EFFECTIVE FOR. THE BALANCE IS:
    --    1) RENTAL FEES FOR ALL PREVIOUS RENTALS
    --    2) ONE DOLLAR FOR EVERY DAY THE PREVIOUS RENTALS ARE OVERDUE
    --    3) IF A FILM IS MORE THAN RENTAL_DURATION * 2 OVERDUE, CHARGE THE REPLACEMENT_COST
    --    4) SUBTRACT ALL PAYMENTS MADE BEFORE THE DATE SPECIFIED
    v_rentfees DECIMAL(5,2); -- FEES PAID TO RENT THE VIDEOS INITIALLY
    v_overfees INT;          -- LATE FEES FOR PRIOR RENTALS
    v_payments DECIMAL(5,2); -- SUM OF PAYMENTS MADE PREVIOUSLY
BEGIN
    SELECT COALESCE(SUM(film.rental_rate), 0)
    INTO v_rentfees
    FROM film, inventory, rental
    WHERE
        film.film_id = inventory.film_id
        AND inventory.inventory_id = rental.inventory_id
        AND rental.rental_date <= p_effective_date
        AND rental.customer_id = p_customer_id
    ;
    SELECT COALESCE(SUM(CASE
        WHEN (rental.return_date - rental.rental_date) > (film.rental_duration * '1 day'::INTERVAL)
        THEN EXTRACT(DAY FROM (rental.return_date - rental.rental_date) - (film.rental_duration * '1 day'::INTERVAL))::INT
        ELSE 0
    END), 0)
    INTO v_overfees
    FROM rental, inventory, film
    WHERE
        film.film_id = inventory.film_id
        AND inventory.inventory_id = rental.inventory_id
        AND rental.rental_date <= p_effective_date
        AND rental.customer_id = p_customer_id
    ;
    SELECT COALESCE(SUM(payment.amount), 0)
    INTO v_payments
    FROM payment
    WHERE
        payment.payment_date <= p_effective_date
        AND payment.customer_id = p_customer_id
    ;
    RETURN v_rentfees + v_overfees - v_payments;
END $$ LANGUAGE plpgsql;

CREATE FUNCTION inventory_held_by_customer(p_inventory_id INT) RETURNS INT AS $$ DECLARE
    v_customer_id INT;
BEGIN
    SELECT customer_id
    INTO v_customer_id
    FROM rental
    WHERE
        return_date IS NULL
        AND inventory_id = p_inventory_id
    ;
    RETURN v_customer_id;
END $$ LANGUAGE plpgsql;

CREATE FUNCTION last_day(TIMESTAMPTZ) RETURNS date AS $$
    SELECT CASE
        WHEN EXTRACT(MONTH FROM $1) = 12
        THEN (((EXTRACT(YEAR FROM $1) + 1) || '-01-01')::DATE - INTERVAL '1 day')::DATE
        ELSE ((EXTRACT(YEAR FROM $1) || '-' || (EXTRACT(MONTH FROM $1) + 1) || '-01')::DATE - INTERVAL '1 day')::DATE
    END
$$ LANGUAGE sql IMMUTABLE STRICT;

CREATE FUNCTION rewards_report(min_monthly_purchases INT, min_dollar_amount_purchased DECIMAL) RETURNS SETOF customer AS $$ DECLARE
    last_month_start DATE;
    last_month_end DATE;
    rr RECORD;
    tmpSQL TEXT;
BEGIN
    /* Some sanity checks... */
    IF min_monthly_purchases = 0 THEN
        RAISE EXCEPTION 'Minimum monthly purchases parameter must be > 0';
    END IF;
    IF min_dollar_amount_purchased = 0.00 THEN
        RAISE EXCEPTION 'Minimum monthly dollar amount purchased parameter must be > $0.00';
    END IF;

    last_month_start := CURRENT_DATE - '3 month'::INTERVAL;
    last_month_start := to_date((extract(YEAR FROM last_month_start) || '-' || extract(MONTH FROM last_month_start) || '-01'),'YYYY-MM-DD');
    last_month_end := LAST_DAY(last_month_start);

    /*
    Create a temporary storage area for Customer IDs.
    */
    CREATE TEMPORARY TABLE tmpCustomer (customer_id INT NOT NULL PRIMARY KEY);

    /*
    Find all customers meeting the monthly purchase requirements
    */
    tmpSQL := 'INSERT INTO tmpCustomer (customer_id)
        SELECT p.customer_id
        FROM payment AS p
        WHERE DATE(p.payment_date) BETWEEN '||quote_literal(last_month_start) ||' AND '|| quote_literal(last_month_end) || '
        GROUP BY customer_id
        HAVING SUM(p.amount) > '|| min_dollar_amount_purchased || '
        AND COUNT(customer_id) > ' ||min_monthly_purchases ;
    EXECUTE tmpSQL;

    /*
    Output ALL customer information of matching rewardees.
    Customize output as needed.
    */
    FOR rr IN EXECUTE 'SELECT c.* FROM tmpCustomer AS t INNER JOIN customer AS c ON t.customer_id = c.customer_id' LOOP
        RETURN NEXT rr;
    END LOOP;

    /* Clean up */
    tmpSQL := 'DROP TABLE tmpCustomer';
    EXECUTE tmpSQL;
    RETURN;
END $$ LANGUAGE plpgsql SECURITY DEFINER;
