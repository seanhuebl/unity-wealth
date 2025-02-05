-- name: CreateTransaction :exec
INSERT INTO transactions (
        id,
        user_id,
        transaction_date,
        merchant,
        amount_cents,
        detailed_category_id
    )
VALUES (?1, ?2, ?3, ?4, ?5, ?6);
-- name: GetDetailedCategoryId :one
SELECT id
FROM detailed_categories
WHERE name = ?1;
-- name: UpdateTransactionByID :one
UPDATE transactions
SET transaction_date = ?1,
    merchant = ?2,
    amount_cents = ?3,
    detailed_category_id = ?4,
    updated_at = ?5
WHERE id = ?6
RETURNING id,
    transaction_date,
    merchant,
    amount_cents,
    detailed_category_id;
-- name: GetCategoryNames :one
SELECT pc.name, dc.name
FROM primary_categories AS pc
INNER JOIN detailed_categories AS dc
ON pc.id = dc.primary_category_id
WHERE