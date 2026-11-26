"use client";

import { useEffect, useRef, useState } from "react";
import { BackupFigure } from "./backup-figure";
import styles from "./backup-story.module.css";

// Follows apps/api/internal/backup: backup_cmd.go, backup_postgres.go,
// backup_service.go, s3_storage.go and restore_check.go.
const stages = [
  {
    name: "Connect",
    lede: "Velld reaches the database without exposing the password.",
    items: [
      { name: "SSH tunnel", text: "Optional. Connects through a local port." },
      { name: "Password via env", text: "PGPASSWORD, MYSQL_PWD or a temp Mongo config. Never in args." },
      { name: "Version check", text: "pg_dump must be at least the server’s version." },
    ],
  },
  {
    name: "Dump",
    lede: "Your database’s own tool makes the backup file.",
    items: [
      { name: "Native tool", text: "pg_dump, mysqldump or mongodump, one database at a time." },
      { name: "gzip while streaming", text: "PostgreSQL only. A failed dump leaves no partial file." },
      { name: "One file per database", text: "backups/<connection>/<db>_<time>_<id>.sql.gz" },
    ],
  },
  {
    name: "Upload",
    lede: "If enabled, a copy goes to your S3 bucket.",
    items: [
      { name: "Object key", text: "<prefix>/<connection>/<file>, on any S3-compatible storage." },
      { name: "Failure is a warning", text: "The local backup stays completed." },
      { name: "Local purge", text: "Optional, only after a successful upload." },
    ],
  },
  {
    name: "Verify",
    lede: "If enabled, Velld test-restores the file.",
    items: [
      { name: "Throwaway container", text: "Same engine, in Docker, removed afterwards." },
      { name: "Tables and rows", text: "Checked against what the dump contains." },
      { name: "Result", text: "Passed, failed, or skipped without Docker." },
    ],
  },
];

export function BackupStory() {
  const [active, setActive] = useState(0);
  const sections = useRef<(HTMLElement | null)[]>([]);

  useEffect(() => {
    // The chapter nearest the reading line drives the figure, in either direction.
    // On phones that line sits below the pinned figure.
    const update = () => {
      const reading = window.innerHeight * (window.innerWidth < 760 ? 0.68 : 0.5);
      let closest = Infinity;
      let index = 0;
      sections.current.forEach((section, i) => {
        if (!section) return;
        const bounds = section.getBoundingClientRect();
        const distance = Math.abs(bounds.top + bounds.height / 2 - reading);
        if (distance < closest) { closest = distance; index = i; }
      });
      setActive(index);
    };
    update();
    window.addEventListener("scroll", update, { passive: true });
    window.addEventListener("resize", update);
    return () => {
      window.removeEventListener("scroll", update);
      window.removeEventListener("resize", update);
    };
  }, []);

  return (
    <section className={styles.story} aria-labelledby="backup-story-title">
      <header className={styles.intro}>
        <p className={styles.eyebrow}>Under the hood</p>
        <h2 id="backup-story-title">What happens when Velld backs up a database.</h2>
      </header>
      <div className={styles.sequence}>
        <figure className={styles.pinned} aria-label="Diagram of a Velld backup: the database host, an SSH tunnel, the dump tool, gzip, the backups folder, an S3 bucket and a throwaway restore container">
          <div className={styles.drawing}>
            <BackupFigure active={active} />
          </div>
          <ol className={styles.progress} aria-label="Backup steps">
            {stages.map((stage, i) => (
              <li key={stage.name} aria-current={active === i ? "step" : undefined}>
                <span className={styles.progressLine} />
                0{i + 1}&nbsp;&nbsp;{stage.name}
              </li>
            ))}
          </ol>
        </figure>
        <div className={styles.chapters}>
          {stages.map((stage, i) => (
            <article
              key={stage.name}
              ref={(element) => { sections.current[i] = element; }}
              className={styles.chapter}
              data-active={active === i}
              aria-labelledby={"backup-stage-" + i}
            >
              <div className={styles.chapterContent}>
                <p className={styles.chapterNumber}>0{i + 1}</p>
                <h3 id={"backup-stage-" + i}>{stage.name}</h3>
                <p className={styles.lede}>{stage.lede}</p>
                <ul className={styles.items}>
                  {stage.items.map((item, j) => (
                    <li key={item.name}>
                      <span className={styles.letter} aria-hidden="true">{"ABC"[j]}</span>
                      <div>
                        <strong>{item.name}</strong>
                        <p>{item.text}</p>
                      </div>
                    </li>
                  ))}
                </ul>
              </div>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}

