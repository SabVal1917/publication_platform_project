import React from "react";
import styles from "./Footer.module.css";

const Footer = () => {
  return (
    <footer className={styles.footerMargin}>
      <div className={styles.footer}>
        <div className={styles.container}>
          <span>© 2025 Programming Hub. All rights reserved.</span>
          <span>Created as a demo project.</span>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
