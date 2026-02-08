import React, { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import formStyles from "../../styles/Forms.module.css";

const RegisterPage = () => {
  const [formData, setFormData] = useState({
    username: "",
    email: "",
    password: "",
    confirmPassword: "",
  });
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");

    if (formData.password !== formData.confirmPassword) {
      setError("Пароли не совпадают");
      return;
    }

    try {
      const response = await fetch("/api/register", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          username: formData.username,
          email: formData.email,
          password_hash: formData.password, // Шлем как password_hash, надеясь что Go поймет
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        setError(data.message || "Ошибка при регистрации");
        return;
      }

      // Если всё ок, сохраняем токен
      localStorage.setItem("token", data.token);
      localStorage.setItem("user", JSON.stringify(data.user));

      // Отправляем событие для обновления Header
      window.dispatchEvent(new Event("localStorageChange"));

      navigate("/");
    } catch (err) {
      console.error("Register error:", err);
      setError("Ошибка соединения с сервером");
    }
  };

  return (
    <div className={formStyles.formContainer}>
      <h1 className={formStyles.title}>Регистрация</h1>

      <form className={formStyles.form} onSubmit={handleSubmit}>
        <div className={formStyles.field}>
          <label className={formStyles.label} htmlFor="username">
            Имя пользователя (Никнейм)
          </label>
          <input
            id="username"
            name="username"
            type="text"
            className={formStyles.input}
            value={formData.username}
            onChange={handleChange}
            required
            placeholder="CoolHacker2025"
          />
        </div>

        <div className={formStyles.field}>
          <label className={formStyles.label} htmlFor="email">
            Email
          </label>
          <input
            id="email"
            name="email"
            type="email"
            className={formStyles.input}
            value={formData.email}
            onChange={handleChange}
            required
            placeholder="example@mail.com"
          />
        </div>

        <div className={formStyles.field}>
          <label className={formStyles.label} htmlFor="password">
            Пароль
          </label>
          <input
            id="password"
            name="password"
            type="password"
            className={formStyles.input}
            value={formData.password}
            onChange={handleChange}
            required
            minLength={6}
            placeholder="Минимум 6 символов"
          />
        </div>

        <div className={formStyles.field}>
          <label className={formStyles.label} htmlFor="confirmPassword">
            Повторите пароль
          </label>
          <input
            id="confirmPassword"
            name="confirmPassword"
            type="password"
            className={formStyles.input}
            value={formData.confirmPassword}
            onChange={handleChange}
            required
            placeholder="Повторите пароль"
          />
        </div>

        {error && <p className={formStyles.error}>{error}</p>}

        <button type="submit" className={formStyles.button}>
          Зарегистрироваться
        </button>

        <p style={{ marginTop: "1rem", textAlign: "center", color: "#6b7280", fontSize: "14px" }}>
          Уже есть аккаунт?{" "}
          <Link to="/login" style={{ color: "#030712", textDecoration: "none" }}>
            Войти
          </Link>
        </p>
      </form>
    </div>
  );
};

export default RegisterPage;
