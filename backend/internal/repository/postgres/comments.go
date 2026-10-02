package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"backend/internal/domain/models"
)

type CommentRepository struct {
	DB *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{DB: db}
}

func (r *CommentRepository) GetByArticleID(ctx context.Context, articleID int) ([]models.Comment, error) {
	query := `
		SELECT c.id, c.content, c.author_id, c.article_id, c.parent_id, c.likes_count, c.created_at, u.username
		FROM comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.article_id = $1 
		ORDER BY c.created_at ASC
	`

	rows, err := r.DB.QueryContext(ctx, query, articleID)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *CommentRepository) Create(ctx context.Context, c *models.Comment) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin comment transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		INSERT INTO comments (content, author_id, article_id, parent_id) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, created_at
	`
	if err := tx.QueryRowContext(ctx, query, c.Content, c.AuthorID, c.ArticleID, c.ParentID).Scan(&c.ID, &c.CreatedAt); err != nil {
		return fmt.Errorf("insert comment: %w", err)
	}

	result, err := tx.ExecContext(ctx,
		`UPDATE articles
		 SET comments_count = comments_count + 1
		 WHERE id = $1`,
		c.ArticleID,
	)
	if err != nil {
		return fmt.Errorf("increment article comments count: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated article rows: %w", err)
	}
	if affected != 1 {
		return ErrArticleNotFound
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit comment transaction: %w", err)
	}
	return nil
}
