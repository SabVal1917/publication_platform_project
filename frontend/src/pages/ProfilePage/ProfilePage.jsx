import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import ArticleCard from "../../components/ArticleCard/ArticleCard";
import styles from "./ProfilePage.module.css";

const ProfilePage = () => {
  const { username: routeUsername } = useParams();
  const [user, setUser] = useState(null); // текущий залогиненный пользователь
  const [profileUsername, setProfileUsername] = useState(null); // чей профиль смотрим
  const [isOwnProfile, setIsOwnProfile] = useState(false);
  const [userArticles, setUserArticles] = useState([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    const userData = localStorage.getItem("user");
    const token = localStorage.getItem("token");

    if (!token || !userData) {
      navigate("/login");
      return;
    }

    try {
      const parsedUser = JSON.parse(userData);
      setUser(parsedUser);

       // Если username не передан в URL, редиректим на /profile/:username текущего пользователя
      if (!routeUsername && parsedUser.username) {
        navigate(`/profile/${parsedUser.username}`, { replace: true });
        return;
      }

      const targetUsername = routeUsername || parsedUser.username;
      setProfileUsername(targetUsername);
      setIsOwnProfile(targetUsername === parsedUser.username);
    } catch (err) {
      console.error("Error parsing user data:", err);
      navigate("/login");
    }
  }, [navigate, routeUsername]);

  useEffect(() => {
    const fetchUserArticles = async () => {
      if (!profileUsername) return;

      try {
        setLoading(true);
        const response = await fetch("/api/articles");

        if (!response.ok) {
          throw new Error("Ошибка загрузки статей");
        }

        const allArticles = await response.json();
        
        // Фильтруем статьи по username профиля
        const filtered = (allArticles || []).filter(
          (article) => article.author_username === profileUsername
        );

        // Сортируем по дате (от новых к старым)
        const sorted = filtered.sort((a, b) => {
          const dateA = new Date(a.created_at);
          const dateB = new Date(b.created_at);
          return dateB - dateA;
        });

        setUserArticles(sorted);
      } catch (err) {
        console.error("Error fetching user articles:", err);
      } finally {
        setLoading(false);
      }
    };

    fetchUserArticles();
  }, [profileUsername]);

  const handleLogout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user");
    window.dispatchEvent(new Event("localStorageChange"));
    navigate("/");
  };

  if (!user || !profileUsername) {
    return null; // Показываем ничего во время проверки/редиректа
  }

  return (
    <div className={styles.profileWrapper}>
      <div className={styles.profileContainer}>
        <header className={styles.profileHeader}>
          {/* Заглушка для аватара - можно будет добавить позже */}
          <div
            className={styles.avatar}
            style={{
              backgroundColor: "#e5e7eb",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              color: "#6b7280",
              fontSize: "36px",
            }}
          >
            {profileUsername?.[0]?.toUpperCase() || "U"}
          </div>

          <div className={styles.infoBlock}>
            <div className={styles.mainInfo}>
              <div className={styles.nameAndHandle}>
                <h1 className={styles.name}>
                  Профиль пользователя: @{profileUsername || "пользователь"}
                </h1>
                <p className={styles.username}>@{profileUsername}</p>
              </div>
              <div className={styles.actionButtons}>
                {isOwnProfile && (
                  <button
                    onClick={handleLogout}
                    className={styles.logoutButton}
                  >
                    Выйти из аккаунта
                  </button>
                )}
              </div>
            </div>

            <div className={styles.details}>
              {isOwnProfile ? (
                <span>📧 Email: {user.email || "Email не указан"}</span>
              ) : (
                <span>📧 Email скрыт</span>
              )}
            </div>
          </div>
        </header>

        <div className={styles.contentTabs}>
          <nav className={styles.tabNav}>
            <button className={`${styles.tabButton} ${styles.active}`}>
              Мои статьи
            </button>
          </nav>
        </div>

        <div className={styles.articlesGrid}>
          {loading ? (
            <p style={{ color: "#6b7280", fontSize: "14px", gridColumn: "1 / -1", textAlign: "center" }}>
              Загрузка...
            </p>
          ) : userArticles.length === 0 ? (
            <p style={{ color: "#6b7280", fontSize: "14px", gridColumn: "1 / -1", textAlign: "center" }}>
              Вы еще не опубликовали ни одной статьи
            </p>
          ) : (
            userArticles.map((article) => (
              <ArticleCard key={article.id} article={article} />
            ))
          )}
        </div>
      </div>
    </div>
  );
};

export default ProfilePage;
