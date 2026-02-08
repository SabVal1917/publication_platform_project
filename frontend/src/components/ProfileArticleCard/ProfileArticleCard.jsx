import React from "react";
import { Link } from "react-router-dom";
import styles from "./ProfileArticleCard.module.css";

const ProfileArticleCard = ({ article }) => {
  return (
    <Link to={`/articles/${article.id}`} className={styles.card}>
      <div className={styles.header}>
        <h3 className={styles.title}>{article.title}</h3>
        <span className={styles.likes}>{article.likes}▲</span>
      </div>
      <p className={styles.description}>{article.description}</p>
      <div className={styles.tags}>
        {article.tags.map((tag) => (
          <div key={tag} className={styles.tag}>
            <span>{tag}</span>
          </div>
        ))}
      </div>
      <div className={styles.meta}>
        {article.date} • {article.readTime} min read
      </div>
    </Link>
  );
};

export default ProfileArticleCard;
