//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"backend/internal/domain/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestCanceledContextStopsQuery(t *testing.T) {
	db := integrationDB(t)
	repo := NewArticleRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.GetAll(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("GetAll() error = %v, want context.Canceled", err)
	}
}

func TestArticleQueriesReturnCommentsCount(t *testing.T) {
	db := integrationDB(t)
	userID, articleID := createFixture(t, db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	commentRepo := NewCommentRepository(db)
	comment := &models.Comment{Content: "visible counter", AuthorID: userID, ArticleID: articleID}
	if err := commentRepo.Create(ctx, comment); err != nil {
		t.Fatalf("Create() comment error = %v", err)
	}

	articleRepo := NewArticleRepository(db)
	article, err := articleRepo.GetByID(ctx, articleID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if article.CommentsCount != 1 {
		t.Fatalf("GetByID() CommentsCount = %d, want 1", article.CommentsCount)
	}

	articles, err := articleRepo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	for _, item := range articles {
		if item.ID == articleID {
			if item.CommentsCount != 1 {
				t.Fatalf("GetAll() CommentsCount = %d, want 1", item.CommentsCount)
			}
			return
		}
	}
	t.Fatalf("GetAll() did not return article %d", articleID)
}

func integrationDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open integration database: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("ping integration database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func createFixture(t *testing.T, db *sql.DB) (userID, articleID int) {
	t.Helper()
	suffix := time.Now().UnixNano()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.QueryRowContext(ctx,
		`INSERT INTO users (username, email, password_hash)
		 VALUES ($1, $2, 'integration-test')
		 RETURNING id`,
		fmt.Sprintf("integration-%d", suffix),
		fmt.Sprintf("integration-%d@example.com", suffix),
	).Scan(&userID)
	if err != nil {
		t.Fatalf("create integration user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})

	err = db.QueryRowContext(ctx,
		`INSERT INTO articles (title, slug, content, author_id)
		 VALUES ('Integration article', $1, 'content', $2)
		 RETURNING id`,
		fmt.Sprintf("integration-%d", suffix), userID,
	).Scan(&articleID)
	if err != nil {
		t.Fatalf("create integration article: %v", err)
	}
	return userID, articleID
}

func TestToggleLikeIsAtomic(t *testing.T) {
	db := integrationDB(t)
	userID, articleID := createFixture(t, db)
	repo := NewArticleRepository(db)

	const toggles = 20
	start := make(chan struct{})
	errorsCh := make(chan error, toggles)
	var ready sync.WaitGroup
	var done sync.WaitGroup
	ready.Add(toggles)
	done.Add(toggles)

	for range toggles {
		go func() {
			defer done.Done()
			ready.Done()
			<-start
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, _, err := repo.ToggleLike(ctx, articleID, userID)
			errorsCh <- err
		}()
	}
	ready.Wait()
	close(start)
	done.Wait()
	close(errorsCh)

	for err := range errorsCh {
		if err != nil {
			t.Errorf("concurrent ToggleLike() error = %v", err)
		}
	}

	var cachedCount, actualCount int
	err := db.QueryRow(
		`SELECT likes_count,
		        (SELECT count(*) FROM article_likes WHERE article_id = $1)
		 FROM articles WHERE id = $1`,
		articleID,
	).Scan(&cachedCount, &actualCount)
	if err != nil {
		t.Fatalf("read like counters: %v", err)
	}
	if cachedCount != 0 || actualCount != 0 {
		t.Fatalf("after %d toggles counts = cached:%d actual:%d, want 0:0", toggles, cachedCount, actualCount)
	}
}

func TestCreateCommentUpdatesCounterAtomically(t *testing.T) {
	db := integrationDB(t)
	userID, articleID := createFixture(t, db)
	repo := NewCommentRepository(db)

	const comments = 20
	errorsCh := make(chan error, comments)
	var wg sync.WaitGroup
	wg.Add(comments)
	for i := range comments {
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			comment := &models.Comment{
				Content:   fmt.Sprintf("comment-%d", i),
				AuthorID:  userID,
				ArticleID: articleID,
			}
			errorsCh <- repo.Create(ctx, comment)
		}()
	}
	wg.Wait()
	close(errorsCh)

	for err := range errorsCh {
		if err != nil {
			t.Errorf("concurrent Create() error = %v", err)
		}
	}

	var cachedCount, actualCount int
	err := db.QueryRow(
		`SELECT comments_count,
		        (SELECT count(*) FROM comments WHERE article_id = $1)
		 FROM articles WHERE id = $1`,
		articleID,
	).Scan(&cachedCount, &actualCount)
	if err != nil {
		t.Fatalf("read comment counters: %v", err)
	}
	if cachedCount != comments || actualCount != comments {
		t.Fatalf("comment counts = cached:%d actual:%d, want %d:%d", cachedCount, actualCount, comments, comments)
	}
}
