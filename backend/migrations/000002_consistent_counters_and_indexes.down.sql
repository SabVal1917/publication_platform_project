DROP INDEX IF EXISTS article_likes_article_id_idx;
DROP INDEX IF EXISTS comments_parent_id_idx;
DROP INDEX IF EXISTS comments_article_id_idx;
DROP INDEX IF EXISTS articles_author_id_idx;

ALTER TABLE articles
    DROP CONSTRAINT IF EXISTS articles_comments_count_nonnegative,
    DROP CONSTRAINT IF EXISTS articles_likes_count_nonnegative,
    ALTER COLUMN comments_count DROP NOT NULL,
    ALTER COLUMN likes_count DROP NOT NULL;
