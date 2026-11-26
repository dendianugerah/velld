"use client";

import { useEffect, useRef, useState } from "react";
import { BackupFigure, parts, partsIn } from "./backup-figure";
import styles from "./backup-story.module.css";

// Follows apps/api/internal/backup: backup_cmd.go, backup_postgres.go,
// backup_service.go, s3_storage.go and restore_check.go.
const steps = [
  {
    label: "Connect",
    line: "Velld reaches your database without exposing the password.",
    facts: ["SSH tunnel, if you set one", "Password via env, never in args", "pg_dump version checked first"],
  },
  {
    label: "Dump",
    line: "Your database’s own tool makes the backup.",
    facts: ["pg_dump, mysqldump or mongodump", "Gzipped while it streams", "One file per database"],
  },
  {
    label: "Upload",
    line: "A copy goes to your S3 bucket, if enabled.",
    facts: ["Any S3-compatible storage", "Failed upload: local file kept", "Purge only after success"],
  },
  {
    label: "Verify",
    line: "Velld test-restores it in a throwaway database.",
    facts: ["Optional, runs in Docker", "Tables and rows checked", "Passed, failed or skipped"],
  },
];

function Card({ active, still = false }: { active: number; still?: boolean }) {
  return (
    <div className={styles.card}>
      <div className={styles.drawing}>
        <BackupFigure active={active} labels={!still} />
      </div>
      <ul className={styles.legend} aria-label="Parts in this step">
        {partsIn(active).map((key) => <li key={key}>{parts[key].name}</li>)}
      </ul>
    </div>
  );
}

export function BackupStory() {
  const [active, setActive] = useState(0);
  const sections = useRef<(HTMLElement | null)[]>([]);

  useEffect(() => {
    const update = () => {
      const reading = window.innerHeight * 0.5;
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
        <h2 id="backup-story-title">What actually happens in a backup.</h2>
        <p>
          No custom format, no black box. Velld runs{" "}
          <span className={styles.highlight}>your database’s own dump tool</span>{" "}
          and takes care of the file it produces.
        </p>
      </header>
      <div className={styles.sequence}>
        <figure className={styles.pinned} aria-label="Diagram of a Velld backup: database, SSH tunnel, dump tool, gzip, backups folder, S3 bucket and a throwaway restore database">
          <Card active={active} />
          <ol className={styles.progress} aria-label="Backup steps">
            {steps.map((step, i) => (
              <li key={step.label} aria-current={active === i ? "step" : undefined}>{step.label}</li>
            ))}
          </ol>
        </figure>
        <div>
          {steps.map((step, i) => (
            <article
              key={step.label}
              ref={(element) => { sections.current[i] = element; }}
              className={styles.step}
              data-active={active === i}
              aria-labelledby={"backup-step-" + i}
            >
              <div className={styles.inlineCard}><Card active={i} still /></div>
              <p className={styles.stepLabel}>0{i + 1}</p>
              <h3 id={"backup-step-" + i}>{step.label}</h3>
              <p className={styles.body}>{step.line}</p>
              <ul className={styles.facts}>
                {step.facts.map((fact) => <li key={fact}>{fact}</li>)}
              </ul>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}

