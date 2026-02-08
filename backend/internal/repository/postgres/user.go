package postgres

import (
	"backend/internal/domain/models"
	"database/sql"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Create(u *models.User) error {
	query := `INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id`
	return r.DB.QueryRow(query, u.Username, u.Email, u.PasswordHash).Scan(&u.ID)
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	u := &models.User{}
	query := `
		SELECT id, username, email, password_hash, 
		       COALESCE(avatar_url, ''), 
		       COALESCE(bio, '') 
		FROM users 
		WHERE email = $1
	`
	
	err := r.DB.QueryRow(query, email).Scan(
		&u.ID, 
		&u.Username, 
		&u.Email, 
		&u.PasswordHash, 
		&u.AvatarURL, 
		&u.Bio,
	)
	
	if err != nil {
		return nil, err
	}
	return u, nil
}