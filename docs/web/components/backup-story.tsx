"use client";

import { useEffect, useRef, useState } from "react";
import { BackupFigure } from "./backup-figure";
import styles from "./backup-story.module.css";

// Follows apps/api/internal/backup: backup_cmd.go, backup_postgres.go,
// backup_service.go, s3_storage.go and restore_check.go.
const stages = [
  {
    name: "Connect",
    lede: "Directly or through an SSH tunnel.",
  },
  {
    name: "Dump",
    lede: "Native database tools write a compressed backup to your server.",
  },
  {
    name: "Upload",
    lede: "Optionally copy it to S3-compatible storage or Cloudflare R2.",
  },
  {
    name: "Verify",
    lede: "With Docker access, restore into a temporary database and check the data.",
  },
];

export function BackupStory() {
  const [active, setActive] = useState(0);
  const sections = useRef<(HTMLElement | null)[]>([]);
  const bars = useRef<(HTMLElement | null)[]>([]);
  const pinned = useRef<HTMLElement>(null);

  const readingLine = () => {
    if (window.innerWidth >= 760 || window.innerHeight <= 520) return window.innerHeight * 0.5;
    const height = pinned.current?.offsetHeight ?? 0;
    return height + (window.innerHeight - height) / 2;
  };

  useEffect(() => {
    // The chapter nearest the reading line drives the figure, in either direction.
    // On phones that line sits below the pinned figure.
    const update = () => {
      const reading = readingLine();
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
    const reading = readingLine();
    const bounds = section.getBoundingClientRect();
    window.scrollTo({
      top: window.scrollY + bounds.top + bounds.height / 2 - reading,
      behavior: smooth ? "smooth" : "auto",
    });
  };

  return (
    <section className={styles.story} aria-label="What happens when Velld backs up a database">
      <header className={`${styles.intro} ${styles.desktopIntro}`}>
        <p className={styles.eyebrow}>PostgreSQL · MySQL · MongoDB</p>
        <h2 id="backup-story-title">From database to verified backup.</h2>
      </header>
      <div className={styles.sequence}>
        <figure ref={pinned} className={styles.pinned} aria-label="PostgreSQL, MySQL and MongoDB backup flow: connection, native tool, local storage, optional S3 storage and restore verification">
          <div className={`${styles.intro} ${styles.mobileIntro}`}>
            <p className={styles.eyebrow}>PostgreSQL · MySQL · MongoDB</p>
            <h2 className={styles.mobileTitle}>From database to verified backup.</h2>
          </div>
          <div className={styles.drawing}>
            <BackupFigure active={active} />
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
              </div>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
