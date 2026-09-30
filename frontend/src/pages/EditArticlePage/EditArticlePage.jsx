import React, { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate, useParams } from "react-router-dom";
import formStyles from "../../styles/Forms.module.css";

const readErrorMessage = async (response, fallback) => {
  const responseBody = await response.text();
  if (!responseBody.trim()) return fallback;

  try {
    const data = JSON.parse(responseBody);
    return data.message || fallback;
  } catch {
    return responseBody.trim();
  }
};

const readStoredUser = () => {
  try {
    return JSON.parse(localStorage.getItem("user"));
  } catch {
    return null;
  }
};

const EditArticlePage = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState("");
  const [submitError, setSubmitError] = useState("");
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm({
    defaultValues: { title: "", content: "", tags: "" },
  });

  useEffect(() => {
    const token = localStorage.getItem("token");
    const currentUser = readStoredUser();
    if (!token || !currentUser?.id) {
      navigate("/login", { replace: true });
      return undefined;
    }

    const controller = new AbortController();

    const loadArticle = async () => {
      try {
        setLoading(true);
        setLoadError("");

        const response = await fetch(`/api/articles/${id}`, {
          signal: controller.signal,
        });
        if (!response.ok) {
          setLoadError(
            response.status === 404
              ? "Статья не найдена"
              : "Не удалось загрузить статью",
          );
          return;
        }

        const article = await response.json();
        if (Number(article.author_id) !== Number(currentUser.id)) {
          setLoadError("Редактировать статью может только её автор");
          return;
        }

        reset({
          title: article.title || "",
          content: article.content || "",
          tags: Array.isArray(article.tags) ? article.tags.join(", ") : "",
        });
      } catch (error) {
        if (error.name !== "AbortError") {
          setLoadError("Ошибка соединения с сервером");
        }
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    };

    loadArticle();
    return () => controller.abort();
  }, [id, navigate, reset]);

  const onSubmit = async (formData) => {
    const token = localStorage.getItem("token");
    if (!token) {
      navigate("/login", { replace: true });
      return;
    }

    setSubmitError("");
    const tags = formData.tags
      .split(",")
      .map((tag) => tag.trim())
      .filter(Boolean);

    try {
      const response = await fetch(`/api/articles/${id}`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          title: formData.title,
          content: formData.content,
          tags,
        }),
      });

      if (response.ok) {
        navigate(`/articles/${id}`, { replace: true });
        return;
      }

      if (response.status === 401) {
        navigate("/login", { replace: true });
        return;
      }

      setSubmitError(await readErrorMessage(response, "Не удалось обновить статью"));
    } catch {
      setSubmitError("Ошибка соединения с сервером");
    }
  };

  if (loading) {
    return (
      <div className={formStyles.formContainer}>
        <p>Загрузка статьи...</p>
      </div>
    );
  }

  if (loadError) {
    return (
      <div className={formStyles.formContainer}>
        <h1 className={formStyles.title}>Редактирование статьи</h1>
        <p className={formStyles.error}>{loadError}</p>
        <Link to={`/articles/${id}`}>Вернуться к статье</Link>
      </div>
    );
  }

  return (
    <div className={formStyles.formContainer} style={{ maxWidth: "800px" }}>
      <h1 className={formStyles.title}>Редактирование статьи</h1>

      <form className={formStyles.form} onSubmit={handleSubmit(onSubmit)}>
        <div className={formStyles.field}>
          <label className={formStyles.label} htmlFor="title">
            Заголовок статьи
          </label>
          <input
            id="title"
            className={formStyles.input}
            {...register("title", {
              required: "Заголовок обязателен",
              maxLength: {
                value: 200,
                message: "Заголовок не должен превышать 200 символов",
              },
            })}
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
            className={formStyles.input}
            placeholder="Например: go, sre"
            {...register("tags")}
          />
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

        {submitError && (
          <p className={formStyles.error} role="alert">
            {submitError}
          </p>
        )}

        <div className={formStyles.buttonGroup}>
          <button
            type="button"
            className={`${formStyles.button} ${formStyles.secondaryButton}`}
            onClick={() => navigate(`/articles/${id}`)}
            disabled={isSubmitting}
          >
            Отмена
          </button>
          <button
            type="submit"
            className={formStyles.button}
            disabled={isSubmitting}
          >
            {isSubmitting ? "Сохранение..." : "Сохранить"}
          </button>
        </div>
      </form>
    </div>
  );
};

export default EditArticlePage;
