CREATE TYPE pet_species AS ENUM ('CAT', 'DOG', 'FROG');
CREATE TYPE pet_status AS ENUM ('AVAILABLE', 'SOLD', 'REMOVED');

CREATE TABLE merchants (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email_hash      bytea NOT NULL,
    email_encrypted bytea NOT NULL,
    password_hash   text NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX merchants_email_hash_key ON merchants (email_hash);

CREATE TABLE stores (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id uuid NOT NULL REFERENCES merchants (id),
    name        text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX stores_merchant_id_key ON stores (merchant_id);

CREATE TABLE customers (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email_hash      bytea NOT NULL,
    email_encrypted bytea NOT NULL,
    password_hash   text NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX customers_email_hash_key ON customers (email_hash);

CREATE TABLE pets (
    id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id                uuid NOT NULL REFERENCES stores (id),
    name                    text NOT NULL,
    species                 pet_species NOT NULL,
    age_years               integer NOT NULL,
    description             text NOT NULL,
    breeder_name_encrypted  bytea NOT NULL,
    breeder_email_encrypted bytea NOT NULL,
    picture_object_key      text NOT NULL,
    status                  pet_status NOT NULL DEFAULT 'AVAILABLE',
    created_at              timestamptz NOT NULL DEFAULT now(),
    sold_at                 timestamptz,
    sold_by_customer_id     uuid REFERENCES customers (id),
    removed_at              timestamptz
);

CREATE INDEX pets_store_id_idx ON pets (store_id);

CREATE INDEX pets_available_idx ON pets (store_id, created_at, id)
    WHERE status = 'AVAILABLE';

CREATE INDEX pets_sold_idx ON pets (store_id, sold_at, id)
    WHERE status = 'SOLD';
