package database

const paymentsSchema = `
	CREATE EXTENSION IF NOT EXISTS pgcrypto;
	CREATE TABLE IF NOT EXISTS payments (
		payment_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id              UUID NOT NULL,
		booking_id           UUID NOT NULL,
		amount               BIGINT NOT NULL,
		payment_completed_at TIMESTAMPTZ,
		(user_id, booking_id) UNIQUE
	);
`
