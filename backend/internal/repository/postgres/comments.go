package postgres

import (
	"backend/internal/domain/models"
	"database/sql"
)

type CommentRepository struct {
	DB *sql.DB
}
func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{DB: db}
}

func (r *CommentRepository) GetByArticleID(articleID int) ([]models.Comment, error) {
	query := `
		SELECT c.id, c.content, c.author_id, c.article_id, c.parent_id, c.likes_count, c.created_at, u.username
		FROM comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.article_id = $1 
		ORDER BY c.created_at ASC
	`
	
	rows, err := r.DB.Query(query, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(&c.ID, &c.Content, &c.AuthorID, &c.ArticleID, &c.ParentID, &c.LikesCount, &c.CreatedAt, &c.AuthorUsername); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (r *CommentRepository) Create(c *models.Comment) error {
	query := `
		INSERT INTO comments (content, author_id, article_id, parent_id) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, created_at
	`
	return r.DB.QueryRow(query, c.Content, c.AuthorID, c.ArticleID, c.ParentID).Scan(&c.ID, &c.CreatedAt)
}