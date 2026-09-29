import React, { useEffect, useState, useCallback } from "react";
import { useParams, Link, useNavigate } from "react-router-dom";
import styles from "./ArticlePage.module.css";

// Компонент формы ответа - вынесен наружу для стабильности
const ReplyForm = ({ commentId, onSubmit, onCancel, value, onChange }) => {
  const textareaId = `reply-input-${commentId}`;
  const textareaName = `reply-${commentId}`;

  return (
    <div style={{ marginTop: "8px", marginBottom: "12px" }}>
      <textarea
        id={textareaId}
        name={textareaName}
        value={value || ""}
        onChange={(e) => onChange(e.target.value)}
        rows={3}
        placeholder="Напишите ответ..."
        style={{
          width: "100%",
          padding: "8px 12px",
          borderRadius: "8px",
          border: "1px solid #e5e7eb",
          fontSize: "14px",
          fontFamily: "inherit",
          resize: "vertical",
          marginBottom: "8px",
        }}
      />
      <div style={{ display: "flex", gap: "8px" }}>
        <button
          onClick={() => onSubmit()}
          disabled={!value?.trim()}
          style={{
            padding: "6px 12px",
            borderRadius: "8px",
            border: "1px solid #030712",
            background: "#030712",
            color: "#ffffff",
            fontSize: "12px",
            fontWeight: 500,
            cursor: "pointer",
            opacity: !value?.trim() ? 0.5 : 1,
          }}
        >
          Отправить
        </button>
        <button
          onClick={onCancel}
          style={{
            padding: "6px 12px",
            borderRadius: "8px",
            border: "1px solid #e5e7eb",
            background: "#ffffff",
            color: "#6b7280",
            fontSize: "12px",
            fontWeight: 500,
            cursor: "pointer",
          }}
        >
          Отмена
        </button>
      </div>
    </div>
  );
};

// Компонент комментария - вынесен наружу для стабильности
const CommentItem = ({ comment, depth = 0, replyingTo, onReplyClick, replyText, onReplyTextChange, onReplySubmit, articleId }) => {
  const commentDate =
    comment.created_at &&
    new Date(comment.created_at).toLocaleDateString("ru-RU", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });

  const authorIdentifier = comment.author_username || (comment.author_id ? `user_${comment.author_id}` : "Аноним");
  const isReplying = replyingTo === comment.id;

  return (
    <li
      style={{
        padding: "12px 0",
        borderBottom: depth === 0 ? "1px solid #e5e7eb" : "none",
        paddingLeft: depth > 0 ? `${depth * 24}px` : "0",
        marginLeft: depth > 0 ? "16px" : "0",
        borderLeft: depth > 0 ? "2px solid #e5e7eb" : "none",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          marginBottom: "4px",
        }}
      >
        <span style={{ fontSize: "14px", fontWeight: 500 }}>
          {authorIdentifier ? (
            <Link
              to={`/profile/${authorIdentifier}`}
              style={{
                color: "#030712",
                textDecoration: "none",
              }}
            >
              @{authorIdentifier}
            </Link>
          ) : (
            "Аноним"
          )}
        </span>
        {commentDate && (
          <span style={{ fontSize: "12px", color: "#9ca3af" }}>
            {commentDate}
          </span>
        )}
      </div>
      <p
        style={{
          margin: "4px 0 8px 0",
          fontSize: "14px",
          color: "#374151",
        }}
      >
        {comment.content}
      </p>
      <div style={{ display: "flex", gap: "8px", marginBottom: "8px" }}>
        <button
          onClick={() => onReplyClick(isReplying ? null : comment.id)}
          style={{
            background: "none",
            border: "none",
            color: "#6b7280",
            fontSize: "12px",
            cursor: "pointer",
            padding: "4px 8px",
          }}
        >
          {isReplying ? "Отмена" : "Ответить"}
        </button>
      </div>
      {isReplying && (
        <ReplyForm
          commentId={comment.id}
          value={replyText[comment.id] || ""}
          onChange={(value) => onReplyTextChange(comment.id, value)}
          onSubmit={() => onReplySubmit(comment.id)}
          onCancel={() => onReplyClick(null)}
        />
      )}
      {comment.replies && comment.replies.length > 0 && (
        <ul style={{ listStyle: "none", padding: 0, marginTop: "8px" }}>
          {comment.replies.map((reply) => (
            <CommentItem
              key={reply.id}
              comment={reply}
              depth={depth + 1}
              replyingTo={replyingTo}
              onReplyClick={onReplyClick}
              replyText={replyText}
              onReplyTextChange={onReplyTextChange}
              onReplySubmit={onReplySubmit}
              articleId={articleId}
            />
          ))}
        </ul>
      )}
    </li>
  );
};

const ArticlePage = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  const [article, setArticle] = useState(null);
  const [comments, setComments] = useState([]);
  const [loadingArticle, setLoadingArticle] = useState(true);
  const [loadingComments, setLoadingComments] = useState(true);
  const [errorArticle, setErrorArticle] = useState("");
  const [errorComments, setErrorComments] = useState("");
  const [newComment, setNewComment] = useState("");
  const [postingComment, setPostingComment] = useState(false);
  const [postError, setPostError] = useState("");
  const [likesCount, setLikesCount] = useState(0);
  const [isLiked, setIsLiked] = useState(false);
  const [liking, setLiking] = useState(false);
  const [replyingTo, setReplyingTo] = useState(null);
  const [replyText, setReplyText] = useState({});

  // 1. Сначала определяем ВСЕ хуки (useCallback)
  const handleReply = useCallback(async (parentId) => {
    const text = replyText[parentId] || "";
    if (!text.trim()) return;

    const token = localStorage.getItem("token");
    if (!token) {
      alert("Нужно войти, чтобы ответить на комментарий");
      navigate("/login");
      return;
    }

    try {
      const response = await fetch(`/api/articles/${id}/comments`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ content: text, parent_id: parentId }),
      });

      if (!response.ok) {
        throw new Error("Ошибка при отправке ответа");
      }

      const created = await response.json();
      // Обновляем список локально или перезагружаем
      // Для простоты перезагружаем список:
      const commentsResponse = await fetch(`/api/articles/${id}/comments`);
      if (commentsResponse.ok) {
        const commentsData = await commentsResponse.json();
        setComments(commentsData || []);
      }
      
      setReplyText((prev) => ({ ...prev, [parentId]: "" }));
      setReplyingTo(null);

    } catch (err) {
      console.error("Error posting reply:", err);
      alert("Ошибка при отправке ответа");
    }
  }, [id, replyText, navigate]);

  const handleReplyTextChange = useCallback((commentId, value) => {
    setReplyText((prev) => ({
      ...prev,
      [commentId]: value,
    }));
  }, []);

  const handleReplyClick = useCallback((commentId) => {
    setReplyingTo(commentId);
  }, []);

  // 2. Потом useEffect и прочая логика
  useEffect(() => {
    const fetchArticle = async () => {
      try {
        setLoadingArticle(true);
        setErrorArticle("");

        const response = await fetch(`/api/articles/${id}`);
        if (!response.ok) {
          if (response.status === 404) {
            setErrorArticle("Статья не найдена");
          } else {
            setErrorArticle("Ошибка загрузки статьи");
          }
          return;
        }

        const data = await response.json();
        setArticle(data);
        setLikesCount(data.likes_count || 0);
        setIsLiked(data.is_liked || false);
      } catch (err) {
        console.error("Error fetching article:", err);
        setErrorArticle("Ошибка соединения с сервером");
      } finally {
        setLoadingArticle(false);
      }
    };

    const fetchComments = async () => {
      try {
        setLoadingComments(true);
        setErrorComments("");

        const response = await fetch(`/api/articles/${id}/comments`);
        if (!response.ok) {
          setErrorComments("Ошибка загрузки комментариев");
          return;
        }

        const data = await response.json();
        setComments(data || []);
      } catch (err) {
        console.error("Error fetching comments:", err);
        setErrorComments("Ошибка соединения с сервером");
      } finally {
        setLoadingComments(false);
      }
    };

    if (id) {
      fetchArticle();
      fetchComments();
    }
  }, [id]);

  const formattedDate =
    article?.created_at &&
    new Date(article.created_at).toLocaleDateString("ru-RU", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });

  const buildCommentTree = (flatComments) => {
    const commentMap = new Map();
    const rootComments = [];

    flatComments.forEach((comment) => {
      commentMap.set(comment.id, { ...comment, replies: [] });
    });

    flatComments.forEach((comment) => {
      const commentNode = commentMap.get(comment.id);
      if (comment.parent_id === null || comment.parent_id === undefined) {
        rootComments.push(commentNode);
      } else {
        const parent = commentMap.get(comment.parent_id);
        if (parent) {
          parent.replies.push(commentNode);
        } else {
          rootComments.push(commentNode);
        }
      }
    });

    return rootComments;
  };

  const commentTree = buildCommentTree(comments);

  const handleLike = async () => {
    const token = localStorage.getItem("token");
    if (!token) {
      alert("Войдите, чтобы поставить лайк");
      return;
    }
    if (liking) return;
    try {
      setLiking(true);
      const response = await fetch(`/api/articles/${id}/like`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
      });
      if (!response.ok) throw new Error("Err");
      const data = await response.json();
      setLikesCount(data.likes_count);
      setIsLiked(data.is_liked);
      setArticle((prev) => ({ ...prev, likes_count: data.likes_count, is_liked: data.is_liked }));
    } catch (err) {
      alert("Ошибка при лайке");
    } finally {
      setLiking(false);
    }
  };

  const handleSubmitComment = async (e) => {
    e.preventDefault();
    if (!newComment.trim()) return;

    const token = localStorage.getItem("token");
    if (!token) {
      setPostError("Нужно войти");
      navigate("/login");
      return;
    }

    try {
      setPostingComment(true);
      setPostError("");
      const response = await fetch(`/api/articles/${id}/comments`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ content: newComment }),
      });
      if (!response.ok) {
        setPostError("Ошибка");
        return;
      }
      const created = await response.json();
      setComments((prev) => [...prev, created]);
      setNewComment("");
      
      const commentsResponse = await fetch(`/api/articles/${id}/comments`);
      if (commentsResponse.ok) {
         const data = await commentsResponse.json();
         setComments(data || []);
      }
    } catch (err) {
      setPostError("Ошибка сети");
    } finally {
      setPostingComment(false);
    }
  };

  // 3. И только теперь делаем return
  if (loadingArticle) {
    return (
      <div className={styles.articleWrapper}>
        <div className={styles.articleContainer}>
          <p style={{ color: "#6b7280" }}>Загрузка статьи...</p>
        </div>
      </div>
    );
  }

  if (errorArticle || !article) {
    return (
      <div className={styles.notFound}>
        <h1>404</h1>
        <p>{errorArticle || "Статья не найдена!"}</p>
        <Link to="/" className={styles.homeLink}>
          Вернуться на главную
        </Link>
      </div>
    );
  }

  return (
    <div className={styles.articleWrapper}>
      <div className={styles.articleContainer}>
        <header className={styles.articleHeader}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
            <h1 className={styles.title}>{article.title}</h1>
            <button
              onClick={handleLike}
              disabled={liking}
              style={{
                background: "none",
                border: "1px solid #e5e7eb",
                borderRadius: "10px",
                padding: "8px 16px",
                cursor: "pointer",
                display: "flex",
                alignItems: "center",
                gap: "6px",
                color: isLiked ? "#ef4444" : "#6b7280",
                fontSize: "14px",
                fontWeight: 500,
              }}
            >
              ❤️ {likesCount}
            </button>
          </div>
          <div>
            {article.author_username && (
              <div style={{ fontSize: "14px", color: "#6b7280" }}>
                Автор:{" "}
                <Link
                  to={`/profile/${article.author_username}`}
                  style={{ color: "#030712", textDecoration: "none" }}
                >
                  @{article.author_username}
                </Link>
              </div>
            )}
            {formattedDate && (
              <div style={{ fontSize: "12px", color: "#9ca3af" }}>
                Опубликовано: {formattedDate}
              </div>
            )}
          </div>
        </header>

        <article className={styles.articleBody}>{article.content}</article>

        <section style={{ marginTop: "32px" }}>
          <h2 style={{ fontSize: "20px", marginBottom: "16px" }}>Комментарии</h2>

          {loadingComments ? (
            <p style={{ color: "#6b7280" }}>Загрузка комментариев...</p>
          ) : errorComments ? (
            <p style={{ color: "#ef4444" }}>{errorComments}</p>
          ) : commentTree.length === 0 ? (
            <p style={{ color: "#6b7280" }}>
              Пока нет комментариев. Будьте первым!
            </p>
          ) : (
            <ul style={{ listStyle: "none", padding: 0, marginBottom: "24px" }}>
              {commentTree.map((comment) => (
                <CommentItem
                  key={comment.id}
                  comment={comment}
                  depth={0}
                  replyingTo={replyingTo}
                  onReplyClick={handleReplyClick}
                  replyText={replyText}
                  onReplyTextChange={handleReplyTextChange}
                  onReplySubmit={handleReply}
                  articleId={id}
                />
              ))}
            </ul>
          )}

          <form onSubmit={handleSubmitComment}>
            <label
              htmlFor="comment"
              style={{
                display: "block",
                marginBottom: "8px",
                fontSize: "14px",
                color: "#6b7280",
              }}
            >
              Добавить комментарий
            </label>
            <textarea
              id="comment"
              name="comment"
              value={newComment}
              onChange={(e) => setNewComment(e.target.value)}
              rows={4}
              style={{
                width: "100%",
                padding: "10px 16px",
                borderRadius: "10px",
                border: "1px solid #e5e7eb",
                fontSize: "14px",
                fontFamily: "inherit",
                resize: "vertical",
                marginBottom: "8px",
              }}
            />
            {postError && (
              <p style={{ color: "#ef4444", fontSize: "12px" }}>{postError}</p>
            )}
            <button
              type="submit"
              disabled={postingComment}
              style={{
                padding: "10px 18px",
                borderRadius: "10px",
                border: "1px solid #030712",
                background: "#030712",
                color: "#ffffff",
                fontSize: "14px",
                fontWeight: 500,
                cursor: "pointer",
                opacity: postingComment ? 0.7 : 1,
              }}
            >
              {postingComment ? "Отправка..." : "Отправить комментарий"}
            </button>
          </form>
        </section>
      </div>
    </div>
  );
};

export default ArticlePage;
