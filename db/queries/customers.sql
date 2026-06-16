-- name: CreateCustomer :exec
INSERT INTO customers (id, email_hash, email_encrypted, password_hash, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetCustomerByEmailHash :one
SELECT * FROM customers WHERE email_hash = $1;

-- name: GetCustomerByID :one
SELECT * FROM customers WHERE id = $1;
