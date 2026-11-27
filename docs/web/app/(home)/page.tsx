"use client";

import {
  Code2,
  Database,
  PenTool,
  Rocket,
} from "lucide-react";
import Link from "next/link";
import Image from "next/image";
import { useState } from "react";
import { BackupStory } from "@/components/backup-story";
import { Dither } from "@/components/ui/dither";
import styles from "./page.module.css";

const repository = "https://github.com/dendianugerah/velld";
const databaseWords = ["databases.", "Postgres.", "MySQL.", "MongoDB."];
const databaseIcons = [null, "/images/postgresql.svg", "/images/mysql.svg", "/images/mongodb.svg"];
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
      <a href="#main" className="sr-only focus:not-sr-only focus:fixed focus:start-3 focus:top-3 focus:z-20 focus:rounded-lg focus:bg-white focus:p-3">
        Skip to content
      </a>
      <header className="mx-auto w-[min(92%,1280px)]">
        <nav className="flex min-h-[80px] items-center gap-8 text-[13px] font-medium max-[600px]:min-h-[72px] max-[600px]:gap-5 max-[600px]:text-xs" aria-label="Main navigation">
          <Link href="/" className="text-[29px] leading-none font-[750] tracking-[-1.8px] max-[600px]:text-[27px]" aria-label="Velld home">
            velld
          </Link>
          <Link href="/docs" className="ms-auto py-3 text-[#50565a] transition-colors duration-150 hover:text-[#161616]">
            Documentation
          </Link>
          <a
            href={repository}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-[7px] py-3 text-[#646b70] transition-colors duration-150 hover:text-[#161616]"
          >
            GitHub
          </a>
        </nav>
      </header>

      <main id="main" tabIndex={-1} className="relative isolate">
        <section className="relative isolate mx-auto flex min-h-[calc(100dvh-80px)] max-w-[1360px] flex-col justify-center px-6 pt-[clamp(88px,8vw,112px)] pb-[72px] text-center max-[900px]:pt-[88px] max-[900px]:pb-16 max-[759px]:min-h-0 max-[600px]:px-4 max-[600px]:pt-20 max-[600px]:pb-14" aria-labelledby="hero-title">
          <Dither
            className={styles.pageDither}
            pattern="clouds"
            colors="oklch(0.5772 0.2324 260)"
            intensity={0.55}
            speed={0.25}
            dotSize={2}
            interactive
          />
          <h1 id="hero-title" className="m-0 text-[clamp(42px,7.4vw,100px)] leading-[1.26] font-[650] tracking-[-0.06em] max-[600px]:text-[clamp(24px,8vw,48px)]">
            <span className="block">
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
                  {databaseIcons[databaseIndex] ? (
                    <Image src={databaseIcons[databaseIndex]!} alt="" width={24} height={24} className="absolute h-[57%] w-[57%] brightness-0 invert" />
                  ) : <Database />}
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
            <span className="mt-2 block">
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
          <p className="mx-auto mt-7 max-w-[640px] text-[clamp(17px,1.45vw,20px)] leading-[1.7] font-[450] tracking-[-0.015em] text-[#40484c] text-pretty max-[600px]:mt-6 max-[600px]:text-base">
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
