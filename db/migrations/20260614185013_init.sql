-- Create enum type "pet_species"
CREATE TYPE "pet_species" AS ENUM ('CAT', 'DOG', 'FROG');
-- Create enum type "pet_status"
CREATE TYPE "pet_status" AS ENUM ('AVAILABLE', 'SOLD', 'REMOVED');
-- Create "customers" table
CREATE TABLE "customers" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "email_hash" bytea NOT NULL, "email_encrypted" bytea NOT NULL, "password_hash" text NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"));
-- Create index "customers_email_hash_key" to table: "customers"
CREATE UNIQUE INDEX "customers_email_hash_key" ON "customers" ("email_hash");
-- Create "merchants" table
CREATE TABLE "merchants" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "email_hash" bytea NOT NULL, "email_encrypted" bytea NOT NULL, "password_hash" text NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"));
-- Create index "merchants_email_hash_key" to table: "merchants"
CREATE UNIQUE INDEX "merchants_email_hash_key" ON "merchants" ("email_hash");
-- Create "stores" table
CREATE TABLE "stores" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "merchant_id" uuid NOT NULL, "name" text NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"), CONSTRAINT "stores_merchant_id_fkey" FOREIGN KEY ("merchant_id") REFERENCES "merchants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "stores_merchant_id_key" to table: "stores"
CREATE UNIQUE INDEX "stores_merchant_id_key" ON "stores" ("merchant_id");
-- Create "pets" table
CREATE TABLE "pets" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "store_id" uuid NOT NULL, "name" text NOT NULL, "species" "pet_species" NOT NULL, "age_years" integer NOT NULL, "description" text NOT NULL, "breeder_name_encrypted" bytea NOT NULL, "breeder_email_encrypted" bytea NOT NULL, "picture_object_key" text NOT NULL, "status" "pet_status" NOT NULL DEFAULT 'AVAILABLE', "created_at" timestamptz NOT NULL DEFAULT now(), "sold_at" timestamptz NULL, "sold_by_customer_id" uuid NULL, "removed_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "pets_sold_by_customer_id_fkey" FOREIGN KEY ("sold_by_customer_id") REFERENCES "customers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "pets_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "stores" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "pets_available_idx" to table: "pets"
CREATE INDEX "pets_available_idx" ON "pets" ("store_id", "created_at", "id") WHERE (status = 'AVAILABLE'::pet_status);
-- Create index "pets_sold_idx" to table: "pets"
CREATE INDEX "pets_sold_idx" ON "pets" ("store_id", "sold_at", "id") WHERE (status = 'SOLD'::pet_status);
-- Create index "pets_store_id_idx" to table: "pets"
CREATE INDEX "pets_store_id_idx" ON "pets" ("store_id");
