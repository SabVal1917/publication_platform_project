package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"backend/internal/domain/models"

	"github.com/lib/pq"
)

var (
	ErrArticleNotFound  = errors.New("article not found")
	ErrArticleForbidden = errors.New("article operation forbidden")
)

type ArticleRepository struct {
	DB *sql.DB
}

func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{DB: db}
}

func (r *ArticleRepository) GetAll() ([]models.Article, error) {
	query := `
		SELECT 
			a.id, a.title, a.slug, a.content, a.likes_count, a.created_at, u.username,
			COALESCE(array_agg(t.name) FILTER (WHERE t.name IS NOT NULL), '{}') as tags
		FROM articles a
		JOIN users u ON a.author_id = u.id
		LEFT JOIN article_tags at ON a.id = at.article_id
		LEFT JOIN tags t ON at.tag_id = t.id
		GROUP BY a.id, u.username
		ORDER BY a.created_at DESC
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var a models.Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Slug, &a.Content, &a.LikesCount, &a.CreatedAt, &a.AuthorUsername, pq.Array(&a.Tags)); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, nil
}

func (r *ArticleRepository) GetByID(id int) (*models.Article, error) {
	query := `
		SELECT 
			a.id, a.title, a.slug, a.content, a.author_id, a.likes_count, a.created_at, u.username,
			COALESCE(array_agg(t.name) FILTER (WHERE t.name IS NOT NULL), '{}') as tags
		FROM articles a
		JOIN users u ON a.author_id = u.id
		LEFT JOIN article_tags at ON a.id = at.article_id
		LEFT JOIN tags t ON at.tag_id = t.id
		WHERE a.id = $1
		GROUP BY a.id, u.username
	`

	var a models.Article
	err := r.DB.QueryRow(query, id).Scan(&a.ID, &a.Title, &a.Slug, &a.Content, &a.AuthorID, &a.LikesCount, &a.CreatedAt, &a.AuthorUsername, pq.Array(&a.Tags))
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *ArticleRepository) Create(a *models.Article) (int, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `INSERT INTO articles (title, slug, content, author_id) VALUES ($1, $2, $3, $4) RETURNING id`
	var articleID int
	err = tx.QueryRow(query, a.Title, a.Slug, a.Content, a.AuthorID).Scan(&articleID)
	if err != nil {
		return 0, err
	}

	if len(a.Tags) > 0 {
		for _, tagName := range a.Tags {
			var tagID int
			err = tx.QueryRow(`SELECT id FROM tags WHERE name = $1`, tagName).Scan(&tagID)
			if err == sql.ErrNoRows {
				err = tx.QueryRow(`INSERT INTO tags (name) VALUES ($1) RETURNING id`, tagName).Scan(&tagID)
			}
			if err != nil {
				return 0, err
			}
			_, err = tx.Exec(`INSERT INTO article_tags (article_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, articleID, tagID)
			if err != nil {
				return 0, err
			}
		}
	}

	return articleID, tx.Commit()
}

func (r *ArticleRepository) ToggleLike(articleID, userID int) (int, bool, error) {
	var exists bool
	queryCheck := `SELECT EXISTS(SELECT 1 FROM article_likes WHERE user_id=$1 AND article_id=$2)`
	err := r.DB.QueryRow(queryCheck, userID, articleID).Scan(&exists)
	if err != nil {
		return 0, false, err
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback()

	if exists {
		_, err = tx.Exec(`DELETE FROM article_likes WHERE user_id=$1 AND article_id=$2`, userID, articleID)
		if err != nil {
			return 0, false, err
		}
		_, err = tx.Exec(`UPDATE articles SET likes_count = likes_count - 1 WHERE id=$1`, articleID)
	} else {
		_, err = tx.Exec(`INSERT INTO article_likes (user_id, article_id) VALUES ($1, $2)`, userID, articleID)
		if err != nil {
			return 0, false, err
		}
		_, err = tx.Exec(`UPDATE articles SET likes_count = likes_count + 1 WHERE id=$1`, articleID)
	}
	if err != nil {
		return 0, false, err
	}

	var newCount int
	err = tx.QueryRow(`SELECT likes_count FROM articles WHERE id=$1`, articleID).Scan(&newCount)
	if err != nil {
		return 0, false, err
	}

	return newCount, !exists, tx.Commit()
}

func (r *ArticleRepository) Update(ctx context.Context, articleID, userID int, input models.UpdateArticleRequest) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin article update transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var authorID int
	err = tx.QueryRowContext(ctx,
		`SELECT author_id
		FROM articles
		WHERE id = $1
		FOR UPDATE`,
		articleID).Scan(&authorID)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return ErrArticleNotFound
	case err != nil:
		return fmt.Errorf("select article author: %w", err)
	case authorID != userID:
		return ErrArticleForbidden
	}
	result, err := tx.ExecContext(
		ctx,
		`UPDATE articles
		SET title = $1,
			content = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $3`,
		input.Title, input.Content, articleID,
	)
	if err != nil {
		return fmt.Errorf("update article: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get updated article rows: %w", err)
	}

	if affected != 1 {
		return fmt.Errorf(
			"update article %d: expected 1 affected row, got %d",
			articleID,
			affected,
		)
	}

	_, err = tx.ExecContext(
		ctx,
		`DELETE FROM article_tags WHERE article_id = $1`,
		articleID,
	)
	if err != nil {
		return fmt.Errorf("delete article tags: %w", err)
	}

	for _, tagName := range input.Tags {
		var tagID int

		err = tx.QueryRowContext(
			ctx,
			`INSERT INTO tags (name)
         VALUES ($1)
         ON CONFLICT (name)
         DO UPDATE SET name = EXCLUDED.name
         RETURNING id`,
			tagName,
		).Scan(&tagID)
		if err != nil {
			return fmt.Errorf("upsert tag %q: %w", tagName, err)
		}
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO article_tags (article_id, tag_id)
         VALUES ($1, $2)
         ON CONFLICT DO NOTHING`,
			articleID,
			tagID,
		)
		if err != nil {
			return fmt.Errorf(
				"link tag %d to article %d: %w",
				tagID,
				articleID,
				err,
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit article update: %w", err)
	}

	return nil
}

func (r *ArticleRepository) Delete(ctx context.Context, articleID, userID int) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin article delete transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var authorID int
	err = tx.QueryRowContext(
		ctx,
		`SELECT author_id
		 FROM articles
		 WHERE id = $1
		 FOR UPDATE`,
		articleID,
	).Scan(&authorID)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return ErrArticleNotFound
	case err != nil:
		return fmt.Errorf("select article author for delete: %w", err)
	case authorID != userID:
		return ErrArticleForbidden
	}

	result, err := tx.ExecContext(
		ctx,
		`DELETE FROM articles WHERE id = $1`,
		articleID,
	)
	if err != nil {
		return fmt.Errorf("delete article: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted article rows: %w", err)
	}
	if affected != 1 {
		return fmt.Errorf(
			"delete article %d: expected 1 affected row, got %d",
			articleID,
			affected,
		)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit article delete: %w", err)
	}

	return nil
}
