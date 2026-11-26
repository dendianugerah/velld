"use client";

import { useEffect, useRef, useState } from "react";
import { BackupFigure } from "./backup-figure";
import styles from "./backup-story.module.css";

// Follows apps/api/internal/backup: backup_cmd.go, backup_postgres.go,
// backup_service.go, s3_storage.go and restore_check.go.
const stages = [
  {
    name: "Connect",
    lede: "Before any data moves, Velld opens a route to the database and hands the dump tool its password.",
    items: [
      { name: "SSH tunnel", text: "Optional. The dump tool connects to a local port that forwards to the database host." },
      { name: "Password in the environment", text: "Passed as PGPASSWORD or MYSQL_PWD, or in a temporary MongoDB config file. Never on the command line." },
      { name: "Version check", text: "For PostgreSQL, pg_dump must match or exceed the server’s major version, or the backup stops." },
    ],
  },
  {
    name: "Dump",
    lede: "The database’s own tool reads the data. Velld decides where the output goes.",
    items: [
      { name: "Native tool", text: "pg_dump, mysqldump or mongodump runs as a separate process, one database at a time." },
      { name: "Streamed through gzip", text: "PostgreSQL output is compressed on its way to disk. If the dump fails, the partial file is deleted." },
      { name: "One file per database", text: "backups/<connection>/<db>_<time>_<id> as .sql.gz, .sql or .archive. Its size is recorded when it finishes." },
    ],
  },
  {
    name: "Upload",
    lede: "With S3 storage turned on, the finished file is copied to your bucket.",
    items: [
      { name: "Object key", text: "<prefix>/<connection>/<file>, on AWS S3 or any S3-compatible endpoint." },
      { name: "Failure is a warning", text: "A failed upload is logged. The local backup still counts as completed." },
      { name: "Local purge", text: "Optional. The local file is deleted only after the upload succeeds." },
    ],
  },
  {
    name: "Verify",
    lede: "In the background, Velld restores the file somewhere safe to prove it works. If the local copy was purged, it is fetched from S3 first.",
    items: [
      { name: "Throwaway container", text: "A fresh database of the same engine, started through Docker and removed afterwards." },
      { name: "Restore and count", text: "The file is restored, then the tables and rows are checked against what the dump contains." },
      { name: "Recorded result", text: "Passed, failed, or skipped when Docker isn’t reachable. A failure sends a notification." },
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
        <div>
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

