package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash,omitempty"`
	AvatarURL    string    `json:"avatar_url"`
	Bio          string    `json:"bio"`
	CreatedAt    time.Time `json:"created_at"`
}
type Article struct {
	ID             int       `json:"id"`
	Title          string    `json:"title"`
	Slug           string    `json:"slug"`
	Content        string    `json:"content"`
	AuthorID       int       `json:"author_id"`
	LikesCount     int       `json:"likes_count"`
	CommentsCount  int       `json:"comments_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	AuthorUsername string    `json:"author_username,omitempty"`
	IsLiked        bool      `json:"is_liked"` 
	Tags           []string  `json:"tags"` 
}

type Comment struct {
	ID             int       `json:"id"`
	Content        string    `json:"content"`
	AuthorID       int       `json:"author_id"`
	ArticleID      int       `json:"article_id"`
	ParentID       *int      `json:"parent_id"`
	LikesCount     int       `json:"likes_count"`
	CreatedAt      time.Time `json:"created_at"`
	AuthorUsername string    `json:"author_username,omitempty"`
}