import React from "react";
import { Routes, Route } from "react-router-dom";
import Layout from "./components/Layout/Layout";
import HomePage from "./pages/HomePage/HomePage";
import LoginPage from "./pages/LoginPage/LoginPage";
import RegisterPage from "./pages/RegisterPage/RegisterPage";
import CreateArticlePage from "./pages/CreateArticlePage/CreateArticlePage";
import ProfilePage from "./pages/ProfilePage/ProfilePage";
import ArticlePage from "./pages/ArticlePage/ArticlePage";

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<HomePage />} />
        <Route path="login" element={<LoginPage />} />
        <Route path="register" element={<RegisterPage />} />
        <Route path="create" element={<CreateArticlePage />} />
        {/* /profile - редирект на профиль текущего пользователя, /profile/:username - просмотр профиля по нику */}
        <Route path="profile" element={<ProfilePage />} />
        <Route path="profile/:username" element={<ProfilePage />} />
        <Route path="articles/:id" element={<ArticlePage />} />
      </Route>
    </Routes>
  );
}

export default App;
