package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UnreadRepository struct {
	DB *pgxpool.Pool
}

func (r *UnreadRepository) Increment(userID, roomID string) {
	query := `
		INSERT INTO unread_counts (user_id, room_id, count)
		VALUES ($1, $2, 1)
		ON CONFLICT (user_id, room_id) 
		DO UPDATE SET count = unread_counts.count + 1
		`
	r.DB.Exec(context.Background(), query, userID, roomID)
}

func (r *UnreadRepository) Reset(userID, roomID string) {
	query := `
		UPDATE unread_counts
		SET count = 0
		WHERE user_id = $1 AND room_id = $2
		`
	r.DB.Exec(context.Background(), query, userID, roomID)
}

func (r *UnreadRepository) Get(userID string) ([]map[string]interface{}, error) {
	rows, err := r.DB.Query(context.Background(),
		"SELECT room_id, count FROM unread_counts WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	for rows.Next() {
		var roomID string
		var count int
		rows.Scan(&roomID, &count)
		result = append(result, map[string]interface{}{
			"room_id": roomID,
			"count":   count,
		})
	}
	return result, nil
}
