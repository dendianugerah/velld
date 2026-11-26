"use client";

import Script from "next/script";
import { useEffect, useRef, useState } from "react";
import { mount } from "./backup-figure";
import styles from "./backup-story.module.css";

// Story follows backup_service.go, backup_cmd.go, and backup_postgres.go.
// Compression is PostgreSQL-specific. S3 upload and local purge are optional;
// a completed dump does not itself guarantee upload or restore verification.
const stages = [
  {
    name: "Extract",
    title: "The database does the dumping.",
    body: "Velld runs the database’s own backup tool: pg_dump for PostgreSQL, mysqldump for MySQL and MariaDB, or mongodump for MongoDB. The tool reads the selected database and produces the dump.",
    note: "Multiple databases? Each is dumped separately, one after another.",
    detail: "database → native dump tool",
    caption: "The source stays. A separate dump takes shape.",
  },
  {
    name: "Write",
    title: "The dump becomes a local file.",
    body: "PostgreSQL’s SQL output streams through gzip into a .sql.gz file. MySQL and MariaDB write a .sql file; MongoDB writes a .archive file. Each lands in its connection’s folder on your server.",
    note: "Velld records the file’s size and completion time after the dump succeeds.",
    detail: ".sql.gz / .sql / .archive",
    caption: "A file on your server, in the database’s format.",
  },
  {
    name: "Store",
    title: "Local first. S3 if enabled.",
    body: "With S3 storage enabled, Velld uploads the finished file to your configured bucket. The local file stays unless you enable local purging, which runs only after a successful upload.",
    note: "An upload can fail independently of the dump. A completed backup is not a promise of an off-server copy.",
    detail: "local file → optional S3 copy",
    caption: "An optional second copy, in your own bucket.",
  },
];

export function BackupStory() {
  const [active, setActive] = useState(0);
  const [ready, setReady] = useState(false);
  const figure = useRef<HTMLDivElement>(null);
  const svg = useRef<SVGSVGElement>(null);
  const sections = useRef<(HTMLElement | null)[]>([]);
  const handle = useRef<ReturnType<typeof mount> | null>(null);

  useEffect(() => {
    // Find the text nearest the reading line, in either scroll direction.
    // On phones that line sits below the compact pinned illustration.
    const update = () => {
      const line = window.innerHeight * (window.innerWidth < 760 ? 0.7 : 0.5);
      let closest = Infinity;
      let index = 0;
      sections.current.forEach((section, i) => {
        if (!section) return;
        const bounds = section.getBoundingClientRect();
        const distance = Math.abs(bounds.top + bounds.height / 2 - line);
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

  useEffect(() => {
    if (!ready || !figure.current || !svg.current) return;
    handle.current = mount({
      stage: figure.current,
      svg: svg.current,
      read: { textContent: "rest" },
      interactive: false,
    }, 6);
    return () => { handle.current?.destroy(); handle.current = null; };
  }, [ready]);

  useEffect(() => { handle.current?.setStage(active); }, [active, ready]);

  return (
    <section className={styles.story} aria-labelledby="backup-story-title">
      <Script src="/hairline/kernel.js" strategy="afterInteractive" onReady={() => setReady(true)} />
      <header className={styles.intro}>
        <p className={styles.eyebrow}>Under the hood</p>
        <h2 id="backup-story-title">What goes into a backup.</h2>
        <p>From a running database to a file on your server.<br className={styles.desktopBreak} /> Here’s what happens to the data.</p>
      </header>
      <div className={styles.sequence}>
        <figure className={styles.pinned} aria-label="An isometric database releasing sheets into a local archive, then copying the file to a second store">
          <div className={styles.figureTop} aria-hidden="true">
            <span>The life of a backup</span>
            <span>0{active + 1} <span className={styles.separator}>/ 03</span></span>
          </div>
          <div ref={figure} className={styles.drawing} data-hairline="backup" data-theme="light" data-stage={active}>
            <svg ref={svg} viewBox="0 0 400 320" aria-hidden="true" />
            {!ready && <span className={styles.fallback}>Database → local backup → optional S3 copy</span>}
          </div>
          <figcaption className={styles.figureCaption} aria-hidden="true">{stages[active].caption}</figcaption>
          <ol className={styles.progress} aria-label="Backup stages">
            {stages.map((stage, i) => (
              <li key={stage.name} aria-current={active === i ? "step" : undefined}>
                <span className={styles.progressLine} />
                <span>0{i + 1} <span>{stage.name}</span></span>
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
              aria-labelledby={`backup-stage-${i}`}
            >
              <div className={styles.chapterContent}>
                <p className={styles.chapterNumber}>0{i + 1} <span>/ {stage.name}</span></p>
                <h3 id={`backup-stage-${i}`}>{stage.title}</h3>
                <p className={styles.body}>{stage.body}</p>
                <p className={styles.detail}>{stage.detail}</p>
                <p className={styles.note}>{stage.note}</p>
              </div>
            </article>
          ))}
        </div>
      </div>
      <footer className={styles.endnote}>Your database. Your backup files. Your infrastructure.</footer>
    </section>
  );
}
