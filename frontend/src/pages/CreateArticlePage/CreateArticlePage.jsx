import React, { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import formStyles from "../../styles/Forms.module.css";

const CreateArticlePage = () => {
  const {
    register,
    handleSubmit,
    formState: { errors },
    getValues,
  } = useForm();
  const navigate = useNavigate();
  const [availableTags, setAvailableTags] = useState([]);

  // Загружаем все статьи для получения уникальных тегов
  useEffect(() => {
    const fetchTags = async () => {
      try {
        const response = await fetch("/api/articles");
        if (!response.ok) {
          console.error("Failed to fetch articles for tags");
          return;
        }

        const articles = await response.json();
        if (!Array.isArray(articles)) {
          console.error("Articles is not an array");
          return;
        }

        const allTags = new Set();

        articles.forEach((article) => {
          // Безопасная проверка: tags может быть null, [] или массивом строк
          if (Array.isArray(article?.tags) && article.tags.length > 0) {
            article.tags.forEach((tag) => {
              // Проверяем, что тег - это непустая строка
              if (tag && typeof tag === "string") {
                const trimmedTag = tag.trim();
                if (trimmedTag.length > 0) {
                  allTags.add(trimmedTag.toLowerCase());
                }
              }
            });
          }
        });

        setAvailableTags(Array.from(allTags).sort());
      } catch (err) {
        console.error("Error fetching tags:", err);
        // Не показываем ошибку пользователю, просто не будет автодополнения
      }
    };

    fetchTags();
  }, []);

  // Функция для создания Slug (URL) из заголовка
  const generateSlug = (text) => {
    return (
      text
        .toString()
        .toLowerCase()
        .trim()
        .replace(/\s+/g, "-") // Заменяем пробелы на тире
        .replace(/[^\w\-]+/g, "") // Убираем спецсимволы
        .replace(/\-\-+/g, "-") + // Убираем двойные тире
      "-" +
      Date.now()
    ); // Добавляем время, чтобы slug был точно уникальным
  };

  const handleFormSubmit = async (status) => {
    const data = getValues();

    // 1. Достаем токен (без него бэкенд нас пошлет)
    const token = localStorage.getItem("token");
    if (!token) {
      alert("Вы должны войти в систему, чтобы публиковать статьи!");
      navigate("/login");
      return;
    }

    // 2. Готовим данные для Go
    // Обрабатываем теги: split, trim, фильтруем пустые
    let tagsArray = [];
    
    if (data.tags && typeof data.tags === "string") {
      tagsArray = data.tags
        .split(",")
        .map((tag) => tag.trim())
        .filter((tag) => tag.length > 0)
        .filter((tag) => typeof tag === "string"); // Дополнительная проверка типа
    }

    const articlePayload = {
      title: data.title,
      content: data.content,
      slug: generateSlug(data.title), // Генерируем slug
      tags: tagsArray, // Отправляем массив тегов
    };

    try {
      const response = await fetch("/api/articles", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(articlePayload),
      });

      if (response.ok) {
        alert("Статья успешно опубликована!");
        navigate("/");
      } else {
        const responseBody = await response.text();
        let errorMessage = responseBody.trim() || "Неизвестная ошибка";

        try {
          const errorData = JSON.parse(responseBody);
          errorMessage = errorData.message || errorMessage;
        } catch {
          // Backend может вернуть обычный текст через http.Error.
        }

        alert("Ошибка при публикации: " + errorMessage);
      }
    } catch (error) {
      console.error("Ошибка сети:", error);
      alert("Не удалось соединиться с сервером");
    }
  };

  return (
    <div className={formStyles.formContainer} style={{ maxWidth: "800px" }}>
      <h1 className={formStyles.title}>Создание новой статьи</h1>

      <form className={formStyles.form}>
        <div className={formStyles.field}>
          <label className={formStyles.label} htmlFor="title">
            Заголовок статьи
          </label>
          <input
            id="title"
            className={formStyles.input}
            {...register("title", { required: "Заголовок обязателен" })}
          />
          {errors.title && (
            <p className={formStyles.error}>{errors.title.message}</p>
          )}
        </div>

        <div className={formStyles.field}>
          <label className={formStyles.label} htmlFor="tags">
            Теги (через запятую)
          </label>
          <input
            id="tags"
            name="tags"
            className={formStyles.input}
            placeholder="Например: react, javascript"
            list="tags-list"
            {...register("tags")}
          />
          <datalist id="tags-list">
            {availableTags.map((tag) => (
              <option key={tag} value={tag} />
            ))}
          </datalist>
        </div>

        <div className={formStyles.field}>
          <label className={formStyles.label} htmlFor="content">
            Содержимое статьи
          </label>
          <textarea
            id="content"
            className={formStyles.textarea}
            rows="15"
            {...register("content", {
              required: "Содержимое статьи не может быть пустым",
            })}
          />
          {errors.content && (
            <p className={formStyles.error}>{errors.content.message}</p>
          )}
        </div>

        <div className={formStyles.buttonGroup}>
          {/* Кнопка черновика пока просто алерт */}
          <button
            type="button"
            className={`${formStyles.button} ${formStyles.secondaryButton}`}
            onClick={() => alert("Сохранение черновиков пока в разработке")}
          >
            Сохранить в черновики
          </button>

          <button
            type="button"
            className={formStyles.button}
            onClick={handleSubmit(() => handleFormSubmit("published"))}
          >
            Опубликовать
          </button>
        </div>
      </form>
    </div>
  );
};

export default CreateArticlePage;
