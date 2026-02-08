import React, { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import formStyles from "../../styles/Forms.module.css";

const LoginPage = () => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");

    try {
      const response = await fetch("/api/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          email: email,
          password: password,
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        setError(data.message || "Неверная почта или пароль");
        return;
      }

      localStorage.setItem("token", data.token);
      localStorage.setItem("user", JSON.stringify(data.user));

      // Отправляем событие для обновления Header
      window.dispatchEvent(new Event("localStorageChange"));

      navigate("/");
    } catch (err) {
      console.error("Login error:", err);
      setError("Ошибка соединения с сервером");
    }
  };

  return (
    <div className={formStyles.formContainer}>
      <h1 className={formStyles.title}>Вход в аккаунт</h1>

      <form className={formStyles.form} onSubmit={handleSubmit}>
        <div className={formStyles.field}>
          <label className={formStyles.label} htmlFor="email">
            Email
          </label>
          <input
            id="email"
            type="email"
            className={formStyles.input}
            value={email}
            onChange={(e) => setEmail(e.target.value)}
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
            type="password"
            className={formStyles.input}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            placeholder="Ваш пароль"
          />
        </div>

        {error && <p className={formStyles.error}>{error}</p>}

        <button type="submit" className={formStyles.button}>
          Войти
        </button>

        <p style={{ marginTop: "1rem", textAlign: "center", color: "#6b7280", fontSize: "14px" }}>
          Нет аккаунта?{" "}
          <Link to="/register" style={{ color: "#030712", textDecoration: "none" }}>
            Зарегистрироваться
          </Link>
        </p>
      </form>
    </div>
  );
};

export default LoginPage;
