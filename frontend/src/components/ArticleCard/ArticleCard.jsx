import React, { useState } from "react";
import { Link } from "react-router-dom";
import styles from "./ArticleCard.module.css";

const ArticleCard = ({ article, onLikeUpdate }) => {
  const [likesCount, setLikesCount] = useState(article.likes_count || 0);
  const [isLiked, setIsLiked] = useState(article.is_liked || false);
  const [liking, setLiking] = useState(false);

  // Форматируем дату
  const formattedDate = article.created_at
    ? new Date(article.created_at).toLocaleDateString("ru-RU", {
        year: "numeric",
        month: "long",
        day: "numeric",
      })
    : "";

  // Получаем первые 150 символов контента
  const getDescription = () => {
    if (!article.content) return "";
    const text = article.content.replace(/<[^>]*>/g, ""); // Убираем HTML теги
    if (text.length <= 150) return text;
    return text.substring(0, 150) + "...";
  };

  // Вычисляем примерное время чтения (средняя скорость 200 слов/мин, ~5 символов на слово)
  const getReadTime = () => {
    if (!article.content) return 0;
    const text = article.content.replace(/<[^>]*>/g, "");
    const words = text.split(/\s+/).filter((word) => word.length > 0).length;
    return Math.max(1, Math.ceil(words / 200));
  };

  const handleLike = async (e) => {
    e.preventDefault();
    e.stopPropagation();

    const token = localStorage.getItem("token");
    if (!token) {
      alert("Войдите, чтобы поставить лайк");
      return;
    }

    if (liking) return;

    try {
      setLiking(true);
      const response = await fetch(`/api/articles/${article.id}/like`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error("Ошибка при установке лайка");
      }

      const data = await response.json();
      setLikesCount(data.likes_count);
      setIsLiked(data.is_liked);

      // Если есть callback, вызываем его
      if (onLikeUpdate) {
        onLikeUpdate(data.likes_count, data.is_liked);
      }
    } catch (err) {
      console.error("Error liking article:", err);
      alert("Ошибка при установке лайка");
    } finally {
      setLiking(false);
    }
  };

  return (
    <div className={styles.card}>
      <Link to={`/articles/${article.id}`} className={styles.link}>
        <div className={styles.header}>
          <h3 className={styles.title}>{article.title || "Без названия"}</h3>
          <button
            onClick={handleLike}
            className={styles.likeButton}
            style={{
              background: "none",
              border: "none",
              cursor: "pointer",
              padding: 0,
              display: "flex",
              alignItems: "center",
              gap: "4px",
            }}
          >
            <span className={styles.likes} style={{ color: isLiked ? "#ef4444" : "#6b7280" }}>
              ❤️ {likesCount}
            </span>
          </button>
        </div>
        <p className={styles.description}>{getDescription()}</p>
        {article.author_username && (
          <div className={styles.author}>
            Автор:{" "}
            <Link
              to={`/profile/${article.author_username}`}
              style={{ color: "#030712", textDecoration: "none" }}
              onClick={(e) => e.stopPropagation()}
            >
              @{article.author_username}
            </Link>
          </div>
        )}
        {/* Отображение тегов */}
        {Array.isArray(article?.tags) && article.tags.length > 0 && (
          <div className={styles.tags}>
            {article.tags.map((tag, index) => {
              // Безопасная проверка тега
              if (!tag || typeof tag !== "string") return null;
              
              const tagName = tag.trim();
              if (!tagName) return null;

              return (
                <Link
                  key={`${tagName}-${index}`}
                  to={`/?search=%23${encodeURIComponent(tagName)}`}
                  className={styles.tag}
                  onClick={(e) => {
                    e.stopPropagation();
                  }}
                  style={{ textDecoration: "none" }}
                >
                  #{tagName}
                </Link>
              );
            })}
          </div>
        )}
        <div className={styles.meta}>
          {formattedDate} • {getReadTime()} min read
          {article.comments_count !== undefined && article.comments_count > 0 && (
            <> • {article.comments_count} комментариев</>
          )}
        </div>
      </Link>
    </div>
  );
};

export default ArticleCard;
