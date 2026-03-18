package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Message struct {
	ID        string
	RoomID    string
	SenderID  string
	Content   string
	FileURL   string
	Type      string
	Status    string
	CreatedAt time.Time
}

type MessageRepository struct {
	DB *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{DB: db}
}

func (r *MessageRepository) SaveMessage(
	roomID string,
	senderID string,
	content string,
	fileURL string,
	msgtype string,
) error {
	query := `
		INSERT INTO messages (id, room_id, sender_id, content, file_url, type, status)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, 'sent')
	`
	_, err := r.DB.Exec(context.Background(),
		query,
		roomID,
		senderID,
		content,
		fileURL,
		msgtype,
	)

	return err
}

func (r *MessageRepository) GetMessage(roomID string, messageID string, page int, limit int) ([]Message, error) {
	offset := (page - 1) * limit

	query := `
		SELECT id, room_id, sender_id,content, file_url, type, status, created_at
		FROM messages
		WHERE room_id=$1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
		`

	rows, err := r.DB.Query(context.Background(), query, roomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message

	for rows.Next() {
		var m Message
		err := rows.Scan(
			&m.ID,
			&m.RoomID,
			&m.SenderID,
			&m.Content,
			&m.FileURL,
			&m.Type,
			&m.Status,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (r *MessageRepository) UpdateStatus(messageID string, status string) error {
	query := `
		UPDATE messages
		SET status = $1
		WHERE id = $2
		`

	_, err := r.DB.Exec(context.Background(),
		query,
		status,
		messageID,
	)
	return err
}

func (r *MessageRepository) CanUserAccessRoom(userID, roomID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM room_members 
			WHERE user_id::text = $1 AND room_id::text = $2
		) OR EXISTS (
			SELECT 1 FROM rooms 
			WHERE id::text = $2 AND created_by::text = $1
		)
	`

	var hasAccess bool
	err := r.DB.QueryRow(context.Background(), query, userID, roomID).Scan(&hasAccess)
	if err != nil {
		return false, err
	}

	return hasAccess, nil
}

func (r *MessageRepository) EnsureUserRoomAccess(userID, roomRef string) (string, error) {
	roomID, err := r.ResolveRoomID(roomRef)
	if err != nil {
		return "", err
	}

	if roomID == "" {
		createRoomQuery := `
			INSERT INTO rooms (id, name, type, created_by)
			VALUES (gen_random_uuid()::text, $1, 'group', $2)
			RETURNING id::text
		`

		err = r.DB.QueryRow(context.Background(), createRoomQuery, roomRef, userID).Scan(&roomID)
		if err != nil {
			return "", err
		}
	}

	joinQuery := `
		INSERT INTO room_members (user_id, room_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, room_id) DO NOTHING
	`

	_, err = r.DB.Exec(context.Background(), joinQuery, userID, roomID)
	if err != nil {
		return "", err
	}

	return roomID, nil
}

func (r *MessageRepository) ResolveRoomID(roomRef string) (string, error) {
	query := `
		SELECT id::text
		FROM rooms
		WHERE id::text = $1 OR name = $1
		LIMIT 1
	`

	var roomID string
	err := r.DB.QueryRow(context.Background(), query, roomRef).Scan(&roomID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	return roomID, nil
}
