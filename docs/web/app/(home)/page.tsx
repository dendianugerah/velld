"use client";

import {
  ArrowRight,
  Check,
  Code2,
  Copy,
  Database,
  Github,
  LockKeyhole,
  ShieldCheck,
} from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import styles from "./page.module.css";

const repository = "https://github.com/dendianugerah/velld";
const cloneCommand = `git clone ${repository}.git`;

export default function HomePage() {
  const [copyState, setCopyState] = useState<"idle" | "copied" | "error">(
    "idle",
  );
  const resetCopy = useRef<ReturnType<typeof setTimeout> | null>(null);
  useEffect(
    () => () => {
      if (resetCopy.current) clearTimeout(resetCopy.current);
    },
    [],
  );

  async function copyCommand() {
    if (resetCopy.current) clearTimeout(resetCopy.current);
    try {
      await navigator.clipboard.writeText(cloneCommand);
      setCopyState("copied");
    } catch {
      setCopyState("error");
    }
    resetCopy.current = setTimeout(() => setCopyState("idle"), 2500);
  }

  return (
    <div className={styles.home}>
      <a href="#main" className={styles.skipLink}>
        Skip to content
      </a>
      <header className={styles.header}>
        <nav className={styles.nav} aria-label="Main navigation">
          <Link href="/" className={styles.brand} aria-label="Velld home">
            <Image src="/images/logo.png" alt="" width={29} height={29} />
            <span>
              velld<span className={styles.brandDot}>.</span>
            </span>
          </Link>
          <div className={styles.navLinks}>
            <a href="#installation">Install</a>
            <Link href="/docs">Docs</Link>
            <a
              href={repository}
              target="_blank"
              rel="noopener noreferrer"
              className={styles.navGithub}
            >
              <Github size={15} aria-hidden="true" /> GitHub{" "}
              <span aria-hidden="true">↗</span>
            </a>
          </div>
        </nav>
      </header>

      <main id="main">
        <section className={styles.hero} aria-labelledby="hero-title">
          <span className={styles.openSource}>
            <span className={styles.statusDot} /> Self-hosted. Open source.
          </span>
          <h1 id="hero-title" className={styles.title}>
            <span className={styles.titleLine}>
              Back up{" "}
              <span className={`${styles.wordChip} ${styles.databaseChip}`}>
                <span className={styles.chipIcon}>
                  <Database aria-hidden="true" />
                </span>
                databases.
              </span>
            </span>
            <span className={styles.titleLine}>
              Get back to{" "}
              <span className={`${styles.wordChip} ${styles.buildingChip}`}>
                <span className={styles.chipIcon}>
                  <Code2 aria-hidden="true" />
                </span>
                building.
              </span>
            </span>
          </h1>
          <p className={styles.subtitle}>
            Your databases, tucked in safely.
            <br className={styles.mobileBreak} /> Velld takes care of backups,
            <br className={styles.desktopBreak} /> schedules, and restores. All
            on your own server.
          </p>
          <div className={styles.compatibility}>
            <span>Works with</span>
            <ul aria-label="Supported databases" role="list">
              <li>
                <span
                  className={`${styles.databaseMark} ${styles.postgres}`}
                  aria-hidden="true"
                />
                PostgreSQL
              </li>
              <li>
                <span
                  className={`${styles.databaseMark} ${styles.mysql}`}
                  aria-hidden="true"
                />
                MySQL
              </li>
              <li>
                <span
                  className={`${styles.databaseMark} ${styles.mongo}`}
                  aria-hidden="true"
                />
                MongoDB
              </li>
            </ul>
          </div>
        </section>

        <section
          id="preview"
          className={styles.preview}
          aria-labelledby="preview-title"
        >
          <div className={styles.previewCaption}>
            <span>
              <span className={styles.statusDot} /> Your databases, in one place.
            </span>
          </div>
          <div className={styles.window}>
            <div className={styles.windowBar} aria-hidden="true">
              <div className={styles.windowDots}>
                <i />
                <i />
                <i />
              </div>
              <span>
                <LockKeyhole size={11} /> your-server / velld
              </span>
              <span className={styles.windowLabel}>CONNECTIONS</span>
            </div>
            <h2 id="preview-title" className={styles.srOnly}>
              One dashboard for your database backups
            </h2>
            <a
              href="/images/connections.png"
              target="_blank"
              rel="noopener noreferrer"
              className={styles.previewLink}
              aria-label="Open the Connections screenshot at full size in a new tab"
            >
              <Image
                src="/images/connections.png"
                alt="Velld connections view with PostgreSQL databases, on-demand backup actions, and scheduled backups"
                width={3024}
                height={1572}
                className={styles.dashboardImage}
                priority
                sizes="(max-width: 1100px) 94vw, 1060px"
              />
              <span className={styles.previewHint} aria-hidden="true">
                View full size ↗
              </span>
            </a>
          </div>
          <div className={styles.previewFootnote}>
            <ShieldCheck size={14} aria-hidden="true" /> Your infrastructure.
            Your credentials. Your data.
          </div>
        </section>

        <section
          id="installation"
          className={styles.setup}
          aria-labelledby="setup-title"
        >
          <div className={styles.setupIntro}>
            <h2 id="setup-title">
              Run it on{" "}
              <span
                className={`${styles.wordChip} ${styles.buildingChip} ${styles.setupTitleChip}`}
              >
                your server.
              </span>
            </h2>
            <p>You’ll need Docker and Docker Compose installed.</p>
          </div>
          <ol className={styles.steps} role="list">
            <li>
              <span className={styles.stepNumber} aria-hidden="true">
                1
              </span>
              <div>
                <h3>Clone it.</h3>
                <p>Get a copy of Velld on your server.</p>
                <div className={styles.command}>
                  <code>
                    git clone <wbr />
                    <span>https://github.com/</span>
                    <wbr />
                    <span>dendianugerah/</span>
                    <wbr />
                    <span>velld.git</span>
                  </code>
                  <button
                    type="button"
                    onClick={copyCommand}
                    aria-label={
                      copyState === "copied"
                        ? "Command copied"
                        : "Copy clone command"
                    }
                    aria-describedby={
                      copyState === "error" ? "clone-copy-status" : undefined
                    }
                    title={copyState === "copied" ? "Copied!" : "Copy clone command"}
                    data-copy-state={copyState}
                  >
                    {copyState === "copied" ? (
                      <Check size={15} />
                    ) : (
                      <Copy size={15} />
                    )}
                  </button>
                </div>
                <span
                  id="clone-copy-status"
                  className={styles.copyStatus}
                  role="status"
                  data-state={copyState}
                >
                  {copyState === "copied"
                    ? "Copied to clipboard!"
                    : copyState === "error"
                      ? "Couldn’t copy. Select the command to copy it manually."
                      : ""}
                </span>
              </div>
            </li>
            <li>
              <span className={styles.stepNumber} aria-hidden="true">
                2
              </span>
              <div>
                <h3>Make it yours.</h3>
                <p>
                  In the <code>velld</code> folder, copy <code>.env.example</code>{" "}
                  to <code>.env</code>, then set your admin credentials and
                  generate your secret keys.
                </p>
                <Link href="/docs/installation">
                  Configuration guide <ArrowRight size={14} aria-hidden="true" />
                </Link>
              </div>
            </li>
            <li>
              <span className={styles.stepNumber} aria-hidden="true">
                3
              </span>
              <div>
                <h3>Start it up.</h3>
                <p>
                  From the <code>velld</code> folder, start Velld. Then open{" "}
                  <code>localhost:3000</code> and connect your first database.
                </p>
                <code className={styles.launchCommand}>
                  docker compose up -d
                </code>
              </div>
            </li>
          </ol>
        </section>
      </main>
      <footer className={styles.footer}>
        <Link href="/" className={styles.brand}>
          <Image src="/images/logo.png" alt="" width={24} height={24} />
          <span>
            velld<span className={styles.brandDot}>.</span>
          </span>
        </Link>
        <div>
          <Link href="/docs">Documentation</Link>
          <a href={repository} target="_blank" rel="noopener noreferrer">
            GitHub <span aria-hidden="true">↗</span>
          </a>
          <a
            href={`${repository}/blob/main/LICENSE`}
            target="_blank"
            rel="noopener noreferrer"
          >
            MIT License
          </a>
        </div>
      </footer>
    </div>
  );
}
