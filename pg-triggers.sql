DROP FUNCTION IF EXISTS last_updated_trg;

CREATE FUNCTION last_updated_trg() RETURNS trigger AS $$ BEGIN
    NEW.last_update = NOW();
    RETURN NEW;
END $$ LANGUAGE plpgsql;
