package postgres

import (
	"database/sql"
	"backend/internal/domain/models"
	"github.com/lib/pq"
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
			a.id, a.title, a.slug, a.content, a.likes_count, a.created_at, u.username,
			COALESCE(array_agg(t.name) FILTER (WHERE t.name IS NOT NULL), '{}') as tags
		FROM articles a
		JOIN users u ON a.author_id = u.id
		LEFT JOIN article_tags at ON a.id = at.article_id
		LEFT JOIN tags t ON at.tag_id = t.id
		WHERE a.id = $1
		GROUP BY a.id, u.username
	`
	
	var a models.Article
	err := r.DB.QueryRow(query, id).Scan(&a.ID, &a.Title, &a.Slug, &a.Content, &a.LikesCount, &a.CreatedAt, &a.AuthorUsername, pq.Array(&a.Tags))
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
		if err != nil { return 0, false, err }
		_, err = tx.Exec(`UPDATE articles SET likes_count = likes_count - 1 WHERE id=$1`, articleID)
	} else {
		_, err = tx.Exec(`INSERT INTO article_likes (user_id, article_id) VALUES ($1, $2)`, userID, articleID)
		if err != nil { return 0, false, err }
		_, err = tx.Exec(`UPDATE articles SET likes_count = likes_count + 1 WHERE id=$1`, articleID)
	}
	if err != nil { return 0, false, err }

	var newCount int
	err = tx.QueryRow(`SELECT likes_count FROM articles WHERE id=$1`, articleID).Scan(&newCount)
	if err != nil { return 0, false, err }

	return newCount, !exists, tx.Commit()
}