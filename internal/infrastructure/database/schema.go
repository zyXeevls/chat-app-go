package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func EnsureSchema(db *pgxpool.Pool) error {
	query := `
	CREATE EXTENSION IF NOT EXISTS pgcrypto;

	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
		username VARCHAR(50) NOT NULL UNIQUE,
		password TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS rooms (
		id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
		name VARCHAR(100) NOT NULL,
		type VARCHAR(20) NOT NULL DEFAULT 'group',
		created_by TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS room_members (
		user_id TEXT NOT NULL,
		room_id TEXT NOT NULL,
		joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
		PRIMARY KEY (user_id, room_id)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
		room_id TEXT NOT NULL,
		sender_id TEXT NOT NULL,
		content TEXT,
		file_url TEXT,
		type VARCHAR(20) NOT NULL DEFAULT 'text',
		status VARCHAR(20) NOT NULL DEFAULT 'sent',
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS unread_counts (
		user_id TEXT NOT NULL,
		room_id TEXT NOT NULL,
		count INT NOT NULL DEFAULT 0,
		PRIMARY KEY (user_id, room_id)
	);

	CREATE INDEX IF NOT EXISTS idx_messages_room_created_at
	ON messages(room_id, created_at);
	`

	_, err := db.Exec(context.Background(), query)
	return err
}
