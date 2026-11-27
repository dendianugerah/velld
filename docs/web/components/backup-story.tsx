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
      { name: "SSH tunnel", text: "Connect securely through SSH when needed." },
      { name: "Protected credentials", text: "Your password stays out of command arguments." },
      { name: "Compatibility", text: "Backup tool compatibility is checked when needed." },
    ],
  },
  {
    name: "Dump",
    lede: "Your database’s own tool makes the backup file.",
    items: [
      { name: "Native tool", text: "Your database’s own tool creates the backup." },
      { name: "Backup processing", text: "The output is prepared as a backup file." },
      { name: "Stored locally", text: "Each database backup is saved on your server." },
    ],
  },
  {
    name: "Upload",
    lede: "If enabled, a copy goes to your S3 bucket.",
    items: [
      { name: "Your storage", text: "Send a copy to your S3-compatible storage." },
      { name: "Local backup", text: "An upload failure won’t remove your local backup." },
      { name: "Optional cleanup", text: "Remove the local file after a successful upload, if enabled." },
    ],
  },
  {
    name: "Verify",
    lede: "If enabled, Velld test-restores the file.",
    items: [
      { name: "Temporary container", text: "The backup is restored in an isolated container." },
      { name: "Restore check", text: "Velld checks the restored data." },
      { name: "Result", text: "See whether verification passed, failed, or was skipped." },
    ],
  },
];

export function BackupStory() {
  const [active, setActive] = useState(0);
  const [focus, setFocus] = useState<number | null>(null);
  const sections = useRef<(HTMLElement | null)[]>([]);
  const bars = useRef<(HTMLElement | null)[]>([]);

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
        const fill = Math.min(1, Math.max(0, (reading - bounds.top) / bounds.height));
        bars.current[i]?.style.setProperty("--fill", String(fill));
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


  const goTo = (i: number) => {
    const section = sections.current[i];
    if (!section) return;
    const smooth = !window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    const reading = window.innerHeight * (window.innerWidth < 760 ? 0.68 : 0.5);
    const bounds = section.getBoundingClientRect();
    window.scrollTo({
      top: window.scrollY + bounds.top + bounds.height / 2 - reading,
      behavior: smooth ? "smooth" : "auto",
    });
  };

  return (
    <section className={styles.story} aria-labelledby="backup-story-title">
      <header className={styles.intro}>
        <p className={styles.eyebrow}>Under the hood</p>
        <h2 id="backup-story-title">What happens when Velld backs up a database.</h2>
        <p className={styles.introLede}>One backup, step by step, straight from the source code.</p>
      </header>
      <div className={styles.sequence}>
        <figure className={styles.pinned} aria-label="Diagram of a Velld backup: the database, connection, backup tool, local storage, optional S3 storage and restore verification">
          <div className={styles.drawing}>
            <BackupFigure active={active} focus={focus} />
          </div>
          <ol className={styles.progress} aria-label="Backup steps">
            {stages.map((stage, i) => (
              <li key={stage.name} aria-current={active === i ? "step" : undefined} data-done={i < active}>
                <button type="button" onClick={() => goTo(i)}>
                  <span className={styles.progressLine} ref={(element) => { bars.current[i] = element; }} />
                  <span className={styles.progressNumber}>0{i + 1}</span>
                  {stage.name}
                </button>
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
                <p className={styles.chapterNumber}>0{i + 1} / 0{stages.length}</p>
                <h3 id={"backup-stage-" + i}>{stage.name}</h3>
                <p className={styles.lede}>{stage.lede}</p>
                <ul className={styles.items} onMouseLeave={() => setFocus(null)}>
                  {stage.items.map((item, j) => (
                    <li key={item.name} onMouseEnter={() => active === i && setFocus(j)}>
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
