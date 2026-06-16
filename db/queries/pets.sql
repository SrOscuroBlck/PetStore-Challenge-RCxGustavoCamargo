-- name: CreatePet :exec
INSERT INTO pets (
    id, store_id, name, species, age_years, description,
    breeder_name_encrypted, breeder_email_encrypted, picture_object_key, status, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: GetPetByID :one
SELECT * FROM pets WHERE store_id = $1 AND id = $2;

-- name: GetPetByIDUnscoped :one
SELECT * FROM pets WHERE id = $1;

-- name: RemovePet :one
UPDATE pets
   SET status = 'REMOVED', removed_at = now()
 WHERE id = $1 AND store_id = $2 AND status = 'AVAILABLE'
RETURNING *;

-- name: ListAvailableByStore :many
SELECT * FROM pets
 WHERE store_id = sqlc.arg(store_id)
   AND status = 'AVAILABLE'
   AND (
     created_at > sqlc.arg(after_created_at)
     OR (created_at = sqlc.arg(after_created_at) AND id > sqlc.arg(after_id))
   )
 ORDER BY created_at, id
 LIMIT sqlc.arg(page_limit);

-- name: ListSoldByStore :many
SELECT * FROM pets
 WHERE store_id = sqlc.arg(store_id)
   AND status = 'SOLD'
   AND sold_at >= sqlc.arg(sold_from)
   AND sold_at <= sqlc.arg(sold_to)
   AND (
     sold_at > sqlc.arg(after_sold_at)
     OR (sold_at = sqlc.arg(after_sold_at) AND id > sqlc.arg(after_id))
   )
 ORDER BY sold_at, id
 LIMIT sqlc.arg(page_limit);

-- name: PurchasePet :one
UPDATE pets
   SET status = 'SOLD', sold_at = now(), sold_by_customer_id = $2
 WHERE id = $1 AND status = 'AVAILABLE'
RETURNING *;

-- name: LockPetsByIDs :many
SELECT * FROM pets
 WHERE id = ANY($1::uuid[])
 ORDER BY id
   FOR UPDATE;
