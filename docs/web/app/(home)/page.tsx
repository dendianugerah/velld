"use client";

import {
  Code2,
  Database,
  PenTool,
  Rocket,
} from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { BackupStory } from "@/components/backup-story";
import { Dither } from "@/components/ui/dither";
import styles from "./page.module.css";

const repository = "https://github.com/dendianugerah/velld";
const databaseWords = ["databases.", "Postgres.", "MySQL.", "MongoDB."];
const activities = [
  { word: "building.", icon: Code2 },
  { word: "shipping.", icon: Rocket },
  { word: "creating.", icon: PenTool },
];

export default function HomePage() {
  const [databaseIndex, setDatabaseIndex] = useState(0);
  const [activityIndex, setActivityIndex] = useState(0);

  return (
    <div className={styles.home}>
      <a href="#main" className={styles.skipLink}>
        Skip to content
      </a>
      <header className={styles.header}>
        <nav className={styles.nav} aria-label="Main navigation">
          <Link href="/" className={styles.brand} aria-label="Velld home">
            velld
          </Link>
          <Link href="/docs" className={styles.docsLink}>
            Documentation
          </Link>
          <a
            href={repository}
            target="_blank"
            rel="noopener noreferrer"
            className={styles.githubLink}
          >
            GitHub
          </a>
        </nav>
      </header>

      <main id="main" tabIndex={-1} className={styles.main}>
        <Dither
          className={styles.pageDither}
          pattern="clouds"
          colors="oklch(0.5772 0.2324 260)"
          intensity={0.55}
          speed={0.25}
          dotSize={2}
          interactive
        />
        <section className={styles.hero} aria-labelledby="hero-title">
          <h1 id="hero-title" className={styles.title}>
            <span className={styles.titleLine}>
              Back up{" "}
              <button
                type="button"
                className={`${styles.wordChip} ${styles.databaseChip}`}
                onClick={() =>
                  setDatabaseIndex((index) => (index + 1) % databaseWords.length)
                }
                aria-label={`${databaseWords[databaseIndex]} Swap database`}
              >
                <span className={styles.chipIcon} aria-hidden="true">
                  <Database />
                </span>
                <span className={styles.wordStack} aria-hidden="true">
                  {databaseWords.map((word, index) => (
                    <span key={word} data-active={index === databaseIndex}>
                      {word}
                    </span>
                  ))}
                </span>
                <span className={styles.clickHint} aria-hidden="true">
                  <svg viewBox="0 0 42 30" fill="none">
                    <path d="M3 4c13 1 23 7 27 19m-8-4 9 5 3-10" />
                  </svg>
                  click me
                </span>
              </button>
            </span>
            <span className={styles.titleLine}>
              Get back to{" "}
              <button
                type="button"
                className={`${styles.wordChip} ${styles.buildingChip}`}
                onClick={() =>
                  setActivityIndex((index) => (index + 1) % activities.length)
                }
                aria-label={`${activities[activityIndex].word} Swap activity`}
              >
                <span className={styles.chipIcon} aria-hidden="true">
                  {activities.map(({ word, icon: Icon }, index) => (
                    <Icon key={word} data-active={index === activityIndex} />
                  ))}
                </span>
                <span className={styles.wordStack} aria-hidden="true">
                  {activities.map(({ word }, index) => (
                    <span key={word} data-active={index === activityIndex}>
                      {word}
                    </span>
                  ))}
                </span>
              </button>
            </span>
          </h1>
          <p className={styles.subtitle}>
            Backups, schedules, and restores. All in one place.
            <br />
            <span className={styles.highlight}>Open source. On your own server.</span>
          </p>
        </section>

        <BackupStory />
      </main>

    </div>
  );
}
