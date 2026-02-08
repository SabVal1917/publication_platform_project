import React from "react";
import { Outlet } from "react-router-dom";
import Header from "../Header/Header";
import Footer from "../Footer/Footer";
import styles from "./Layout.module.css";

const Layout = () => {
  return (
    <div className={styles.pageWrapper}>
      <Header />

      <div className={styles.headerSeparator}></div>

      <main className={styles.mainContent}>
        <Outlet />
      </main>

      <Footer />
    </div>
  );
};

export default Layout;
