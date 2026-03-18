package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepository struct {
	DB *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{DB: db}
}

func (r *AuthRepository) Register(username string, password string) error {
	return r.CreateUser(username, password)
}

func (r *AuthRepository) CreateUser(username, password string) error {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 14)

	_, err := r.DB.Exec(context.Background(),
		"INSERT INTO users (id, username, password) VALUES (gen_random_uuid(), $1, $2)",
		username,
		string(hash),
	)
	return err
}

func (r *AuthRepository) GetUser(username string) (string, string, error) {
	var id, hash string

	err := r.DB.QueryRow(context.Background(),
		"SELECT id, password FROM users WHERE username = $1",
		username,
	).Scan(&id, &hash)
	return id, hash, err
}
