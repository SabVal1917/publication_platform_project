import React, { useState, useEffect } from "react";
import { Link, NavLink, useNavigate } from "react-router-dom";
import styles from "./Header.module.css";

const Header = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState(null);
  const navigate = useNavigate();

  // Проверка авторизации при монтировании и при изменении localStorage
  useEffect(() => {
    const checkAuth = () => {
      const token = localStorage.getItem("token");
      const userData = localStorage.getItem("user");
      
      if (token && userData) {
        try {
          const parsedUser = JSON.parse(userData);
          setIsAuthenticated(true);
          setUser(parsedUser);
        } catch (err) {
          console.error("Error parsing user data:", err);
          setIsAuthenticated(false);
          setUser(null);
        }
      } else {
        setIsAuthenticated(false);
        setUser(null);
      }
    };

    // Проверяем при монтировании
    checkAuth();

    // Слушаем изменения в localStorage (для обновления при входе/выходе в других вкладках)
    const handleStorageChange = (e) => {
      if (e.key === "token" || e.key === "user") {
        checkAuth();
      }
    };

    window.addEventListener("storage", handleStorageChange);
    
    // Также слушаем события на этой же вкладке через кастомное событие
    const handleCustomStorageChange = () => {
      checkAuth();
    };

    window.addEventListener("localStorageChange", handleCustomStorageChange);

    return () => {
      window.removeEventListener("storage", handleStorageChange);
      window.removeEventListener("localStorageChange", handleCustomStorageChange);
    };
  }, []);

  const handleLogout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user");
    setIsAuthenticated(false);
    setUser(null);
    
    // Отправляем событие для обновления других компонентов
    window.dispatchEvent(new Event("localStorageChange"));
    
    navigate("/");
  };

  return (
    <header className={styles.header}>
      <div className={styles.container}>
        {/* Логотип - это обычный Link */}
        <Link to="/" className={styles.logoLink}>
          <div className={styles.gradient}></div>
          <div className={styles.logoTextContainer}>
            <span className={styles.logoText}>Programming Article Hub</span>
          </div>
        </Link>

        {/* Навигация */}
        <nav className={styles.nav}>
          <NavLink
            to="/"
            className={({ isActive }) =>
              isActive ? styles.navLinkActive : styles.navLink
            }
          >
            Home
          </NavLink>

          {isAuthenticated ? (
            <>
              <NavLink
                to="/create"
                className={({ isActive }) =>
                  isActive ? styles.navLinkActive : styles.navLink
                }
              >
                Create
              </NavLink>
              <NavLink
                to="/profile"
                className={({ isActive }) =>
                  isActive ? styles.navLinkActive : styles.navLink
                }
              >
                {user?.username || "Profile"}
              </NavLink>
              <button onClick={handleLogout} className={styles.logoutButton}>
                Выйти
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className={styles.navLink}>
                Вход
              </Link>
              <Link to="/register" className={styles.navLink}>
                Регистрация
              </Link>
            </>
          )}
        </nav>
      </div>
    </header>
  );
};

export default Header;
