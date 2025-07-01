-- +goose Up
/* Shared helper functions & extensions */
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE EXTENSION IF NOT EXISTS citext;

/* ---------- timestamp maintenance helpers ---------- */
CREATE OR REPLACE FUNCTION trg_set_timestamp () RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at := NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION trg_touch_last_used_at () RETURNS TRIGGER AS $$
BEGIN
  NEW.last_used_at := NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- +goose Down
DROP FUNCTION IF EXISTS trg_set_timestamp ();

DROP FUNCTION IF EXISTS trg_touch_last_used_at ();

DROP EXTENSION IF EXISTS citext;

DROP EXTENSION IF EXISTS pgcrypto;