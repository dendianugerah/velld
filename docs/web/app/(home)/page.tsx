"use client";

import {
  ArrowRight,
  Check,
  Code2,
  Copy,
  Database,
  Github,
  LockKeyhole,
} from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import styles from "./page.module.css";

const repository = "https://github.com/dendianugerah/velld";
const cloneCommand = `git clone ${repository}.git`;
const agentPrompt = `Help me set up and run Velld on my own machine or server using Docker Compose. Use the repository README and Compose files as the source of truth. Check that Docker and the Compose plugin are available.

1. Run ${cloneCommand}, then cd velld. If the repository is already checked out, use that checkout instead.
2. If .env does not exist, copy .env.example to .env. Generate JWT_SECRET and ENCRYPTION_KEY separately with openssl rand -hex 32 and put the literal outputs in .env, not shell expressions. ENCRYPTION_KEY must be 64 hex characters. Set ADMIN_USERNAME_CREDENTIAL and a strong ADMIN_PASSWORD_CREDENTIAL for the initial admin account. Do not print or commit secrets, and keep existing keys when updating.
3. Set NEXT_PUBLIC_API_URL to the API address reachable from my browser: http://localhost:8080 for a local install, or the server's reachable address on port 8080 (or its HTTPS reverse-proxy URL) for a remote install.
4. Run docker compose up -d from the repository root and check docker compose ps. If a service fails, inspect docker compose logs. Open http://localhost:3000 locally, or the server's web address on port 3000 (or its HTTPS URL), and sign in with the configured admin credentials.
5. To update a source install, pull the latest changes for the installed branch and run docker compose up -d --build. Keep .env and the api_data and backup_data volumes intact.

Tell me the UI URL and service status without exposing secrets.`;

function useClipboard(value: string) {
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

  async function copy() {
    if (resetCopy.current) clearTimeout(resetCopy.current);
    try {
      await navigator.clipboard.writeText(value);
      setCopyState("copied");
    } catch {
      setCopyState("error");
    }
    resetCopy.current = setTimeout(() => setCopyState("idle"), 2500);
  }

  return { copyState, copy };
}

export default function HomePage() {
  const { copyState, copy: copyCommand } = useClipboard(cloneCommand);
  const { copyState: promptCopyState, copy: copyPrompt } =
    useClipboard(agentPrompt);

  return (
    <div className={styles.home}>
      <a href="#main" className={styles.skipLink}>
        Skip to content
      </a>
      <header className={styles.header}>
        <nav className={styles.nav} aria-label="Main navigation">
          <Link href="/" className={styles.brand} aria-label="Velld home">
            <span>velld</span>
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
        </section>

        <section
          id="installation"
          className={styles.setup}
          aria-labelledby="setup-title"
        >
          <div className={styles.installHeader}>
            <h2 id="setup-title">Install Velld</h2>
            <Link href="/docs/installation">
              Installation guide <ArrowRight size={14} aria-hidden="true" />
            </Link>
          </div>
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
                copyState === "copied" ? "Command copied" : "Copy clone command"
              }
              aria-describedby={
                copyState === "error" ? "clone-copy-status" : undefined
              }
              title={copyState === "copied" ? "Copied!" : "Copy clone command"}
              data-copy-state={copyState}
            >
              <Copy
                size={15}
                className={styles.copyGlyph}
                data-visible={copyState !== "copied"}
                aria-hidden="true"
              />
              <Check
                size={15}
                className={styles.copyGlyph}
                data-visible={copyState === "copied"}
                aria-hidden="true"
              />
            </button>
          </div>
          <span
            id="clone-copy-status"
            className={styles.copyStatus}
            role="status"
            data-state={copyState}
          >
            {copyState === "copied"
              ? "Copied to clipboard"
              : copyState === "error"
                ? "Couldn’t copy. Select the command to copy it manually."
                : ""}
          </span>
          <div className={styles.promptCopy}>
            <button type="button" onClick={copyPrompt}>
              Copy agent prompt
            </button>
            <span role="status" data-state={promptCopyState}>
              {promptCopyState === "copied"
                ? "Copied"
                : promptCopyState === "error"
                  ? "Couldn’t copy the prompt."
                  : ""}
            </span>
          </div>
        </section>
      </main>
      <footer className={styles.footer}>
        <Link href="/" className={styles.brand}>
          <span>velld</span>
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
