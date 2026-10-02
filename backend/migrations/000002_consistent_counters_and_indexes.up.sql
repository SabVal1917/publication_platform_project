UPDATE articles AS a
SET likes_count = (
        SELECT count(*) FROM article_likes AS al WHERE al.article_id = a.id
    ),
    comments_count = (
        SELECT count(*) FROM comments AS c WHERE c.article_id = a.id
    );

ALTER TABLE articles
    ALTER COLUMN likes_count SET NOT NULL,
    ALTER COLUMN comments_count SET NOT NULL,
    ADD CONSTRAINT articles_likes_count_nonnegative CHECK (likes_count >= 0),
    ADD CONSTRAINT articles_comments_count_nonnegative CHECK (comments_count >= 0);

CREATE INDEX articles_author_id_idx ON articles (author_id);
CREATE INDEX comments_article_id_idx ON comments (article_id);
CREATE INDEX comments_parent_id_idx ON comments (parent_id);
CREATE INDEX article_likes_article_id_idx ON article_likes (article_id);
