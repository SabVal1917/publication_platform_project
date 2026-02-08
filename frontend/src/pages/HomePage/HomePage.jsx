import React, { useState, useEffect, useMemo } from "react";
import { Link, useSearchParams } from "react-router-dom";
import ArticleCard from "../../components/ArticleCard/ArticleCard";
import styles from "./HomePage.module.css";

const ITEMS_PER_PAGE = 10;

const HomePage = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const [articles, setArticles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  // Инициализируем поисковый запрос из URL, декодируя его
  const initialSearchParam = searchParams.get("search");
  const initialSearch = initialSearchParam ? decodeURIComponent(initialSearchParam) : "";
  const [searchQuery, setSearchQuery] = useState(initialSearch);
  const [sortBy, setSortBy] = useState("newest");

  useEffect(() => {
    const fetchArticles = async () => {
      try {
        setLoading(true);
        setError("");
        const response = await fetch("/api/articles");
        
        if (!response.ok) {
          throw new Error("Ошибка загрузки статей");
        }

        const data = await response.json();
        setArticles(data || []);
      } catch (err) {
        console.error("Error fetching articles:", err);
        setError(err.message || "Ошибка соединения с сервером");
      } finally {
        setLoading(false);
      }
    };

    fetchArticles();
  }, []);

  // Фильтрация и сортировка статей
  const filteredAndSortedArticles = useMemo(() => {
    let filtered = [...articles];

    // Фильтрация по поисковому запросу
    if (searchQuery && searchQuery.trim()) {
      const query = searchQuery.trim();
      
      if (query.startsWith("#")) {
        // Поиск по тегам - точное совпадение (case-insensitive)
        const tagName = query.substring(1).toLowerCase().trim();
        
        if (tagName.length > 0) {
          filtered = filtered.filter((article) => {
            // Безопасная проверка: tags может быть null, [] или массивом
            if (!Array.isArray(article?.tags) || article.tags.length === 0) {
              return false;
            }
            
            // Ищем точное совпадение тега (case-insensitive)
            return article.tags.some((tag) => {
              if (!tag || typeof tag !== "string") return false;
              return tag.trim().toLowerCase() === tagName;
            });
          });
        }
      } else {
        // Поиск по заголовку или автору (case-insensitive)
        const lowerQuery = query.toLowerCase();
        filtered = filtered.filter((article) => {
          const title = article?.title || "";
          const author = article?.author_username || "";
          
          const titleMatch = title.toLowerCase().includes(lowerQuery);
          const authorMatch = author.toLowerCase().includes(lowerQuery);
          
          return titleMatch || authorMatch;
        });
      }
    }

    // Сортировка
    switch (sortBy) {
      case "newest":
        filtered.sort((a, b) => {
          const dateA = new Date(a.created_at);
          const dateB = new Date(b.created_at);
          return dateB - dateA;
        });
        break;
      case "oldest":
        filtered.sort((a, b) => {
          const dateA = new Date(a.created_at);
          const dateB = new Date(b.created_at);
          return dateA - dateB;
        });
        break;
      case "popular":
        filtered.sort((a, b) => {
          const scoreA = (a.likes_count || 0) + (a.comments_count || 0);
          const scoreB = (b.likes_count || 0) + (b.comments_count || 0);
          return scoreB - scoreA;
        });
        break;
      case "discussed":
        filtered.sort((a, b) => {
          const commentsA = a.comments_count || 0;
          const commentsB = b.comments_count || 0;
          return commentsB - commentsA;
        });
        break;
      default:
        break;
    }

    return filtered;
  }, [articles, searchQuery, sortBy]);

  // Обновляем URL при изменении поискового запроса
  useEffect(() => {
    if (searchQuery) {
      setSearchParams({ search: searchQuery });
    } else {
      setSearchParams({});
    }
  }, [searchQuery, setSearchParams]);

  const handleSearchChange = (e) => {
    setSearchQuery(e.target.value);
    setCurrentPage(1); // Сбрасываем на первую страницу при новом поиске
  };

  const handleSearchClear = () => {
    setSearchQuery("");
    setCurrentPage(1);
    setSearchParams({});
  };

  // Вычисляем пагинацию для отфильтрованных статей
  const totalPages = Math.ceil(filteredAndSortedArticles.length / ITEMS_PER_PAGE);
  const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
  const endIndex = startIndex + ITEMS_PER_PAGE;
  const currentArticles = filteredAndSortedArticles.slice(startIndex, endIndex);

  const handlePageChange = (page) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  return (
    <>
      <section className={styles.heroSection}>
        <div className={styles.heroContainer}>
          <h1 className={styles.heading1}>
            Discover concise programming articles
          </h1>
          <p className={styles.subheading}>
            Curated, minimal and focused — quickly find what matters.
          </p>
          <div className={styles.inputWrapper} style={{ position: "relative" }}>
            <input
              type="text"
              className={styles.searchInput}
              placeholder="Search articles, topics, authors... (use #tag for tags)"
              value={searchQuery}
              onChange={handleSearchChange}
              style={{ paddingRight: searchQuery ? "32px" : "16px" }}
            />
            {searchQuery && (
              <button
                onClick={handleSearchClear}
                style={{
                  position: "absolute",
                  right: "8px",
                  top: "50%",
                  transform: "translateY(-50%)",
                  background: "none",
                  border: "none",
                  cursor: "pointer",
                  padding: "4px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  color: "#6b7280",
                  fontSize: "16px",
                  lineHeight: "1",
                }}
                aria-label="Очистить поиск"
              >
                ×
              </button>
            )}
          </div>
        </div>
      </section>

      <div className={styles.articlesSectionWrapper}>
        <section className={styles.articlesSection}>
          {!loading && !error && filteredAndSortedArticles.length > 0 && (
            <div style={{ 
              display: "flex", 
              justifyContent: "space-between", 
              alignItems: "center",
              marginBottom: "24px",
              gap: "16px",
              flexWrap: "wrap"
            }}>
              <div style={{ fontSize: "14px", color: "#6b7280" }}>
                Найдено статей: {filteredAndSortedArticles.length}
              </div>
              <select
                value={sortBy}
                onChange={(e) => {
                  setSortBy(e.target.value);
                  setCurrentPage(1);
                }}
                style={{
                  padding: "8px 12px",
                  borderRadius: "10px",
                  border: "1px solid #e5e7eb",
                  fontSize: "14px",
                  color: "#030712",
                  background: "#ffffff",
                  cursor: "pointer",
                }}
              >
                <option value="newest">Новые</option>
                <option value="oldest">Старые</option>
                <option value="popular">Популярные</option>
                <option value="discussed">Обсуждаемые</option>
              </select>
            </div>
          )}

          {loading ? (
            <div style={{ textAlign: "center", padding: "40px", color: "#6b7280" }}>
              Загрузка...
            </div>
          ) : error ? (
            <div style={{ textAlign: "center", padding: "40px", color: "#ef4444" }}>
              {error}
            </div>
          ) : filteredAndSortedArticles.length === 0 ? (
            <div style={{ textAlign: "center", padding: "40px" }}>
              <p style={{ color: "#6b7280", fontSize: "16px", marginBottom: "20px" }}>
                {searchQuery ? "Ничего не найдено по вашему запросу" : "Статей пока нет. Будьте первым!"}
              </p>
              <Link
                to="/create"
                style={{
                  display: "inline-block",
                  padding: "12px 20px",
                  borderRadius: "10px",
                  border: "1px solid #030712",
                  background: "#030712",
                  color: "#ffffff",
                  textDecoration: "none",
                  fontSize: "14px",
                  fontWeight: "500",
                  transition: "opacity 0.2s",
                }}
                onMouseEnter={(e) => (e.target.style.opacity = "0.9")}
                onMouseLeave={(e) => (e.target.style.opacity = "1")}
              >
                Создать статью
              </Link>
            </div>
          ) : (
            <>
              <div className={styles.articlesGrid}>
                {currentArticles.map((article) => (
                  <ArticleCard key={article.id} article={article} />
                ))}
              </div>

              {totalPages > 1 && (
                <nav className={styles.paginationNav}>
                  <button
                    className={styles.paginationButton}
                    onClick={() => handlePageChange(currentPage - 1)}
                    disabled={currentPage === 1}
                  >
                    Назад
                  </button>
                  
                  {Array.from({ length: totalPages }, (_, i) => i + 1).map((page) => (
                    <button
                      key={page}
                      className={`${styles.paginationButton} ${
                        currentPage === page ? styles.active : ""
                      }`}
                      onClick={() => handlePageChange(page)}
                    >
                      {page}
                    </button>
                  ))}
                  
                  <button
                    className={styles.paginationButton}
                    onClick={() => handlePageChange(currentPage + 1)}
                    disabled={currentPage === totalPages}
                  >
                    Вперед
                  </button>
                </nav>
              )}
            </>
          )}
        </section>
      </div>
    </>
  );
};

export default HomePage;
