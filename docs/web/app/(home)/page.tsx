"use client";

import {
  ArrowDown,
  ArrowRight,
  Check,
  ChevronRight,
  Clock3,
  Code2,
  Copy,
  Database,
  Github,
  HardDrive,
  Heart,
  LockKeyhole,
  RotateCcw,
  Server,
  ShieldCheck,
  Sparkles,
  Terminal,
} from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { ThemeToggle } from "fumadocs-ui/components/layout/theme-toggle";
import styles from "./page.module.css";

const repository = "https://github.com/dendianugerah/velld";
const cloneCommand = `git clone ${repository}.git`;
const features = [
  {
    icon: Clock3,
    color: "purple",
    label: "SET IT. FORGET IT.",
    title: "A little routine goes a long way.",
    description:
      "Hourly, daily, or your own cron schedule. Let your backups happen while you do literally anything else.",
    detail: "0 2 * * *",
    caption: "Every day, at 2:00 AM",
  },
  {
    icon: HardDrive,
    color: "blue",
    label: "A HOME FOR YOUR DATA",
    title: "Your storage. Your rules.",
    description:
      "Keep backups on your server or send them to S3-compatible storage. You decide where your data lives.",
    detail: "Local storage + S3",
    caption: "A place you already trust",
  },
  {
    icon: RotateCcw,
    color: "green",
    label: "FOR THE JUST-IN-CASES",
    title: "A way back when you need it.",
    description:
      "Browse your backup history, download a snapshot, and restore to an empty database when it’s time to recover.",
    detail: "A fresh start, not a redo.",
    caption: "Keep building with confidence",
  },
];

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
            <a href="#features" className={styles.featuresLink}>
              Why Velld?
            </a>
            <Link href="/docs">Docs</Link>
            <ThemeToggle className={styles.themeToggle} />
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
          <div
            className={`${styles.doodle} ${styles.doodleLeft}`}
            aria-hidden="true"
          >
            <span className={styles.doodleIcon}>
              <Database size={27} />
            </span>
            <span>safe &amp; sound</span>
            <svg viewBox="0 0 92 70" fill="none">
              <path d="M10 5C0 56 76 14 72 56m-9-9 9 11 10-12" />
            </svg>
          </div>
          <div
            className={`${styles.doodle} ${styles.doodleRight}`}
            aria-hidden="true"
          >
            <Sparkles size={26} />
            <span>
              one less thing
              <br />
              to worry about
            </span>
          </div>
          <a
            className={styles.openSource}
            href={repository}
            target="_blank"
            rel="noopener noreferrer"
          >
            <span className={styles.statusDot} /> Self-hosted. Open source. A
            little peace of mind.
            <ChevronRight size={13} aria-hidden="true" />
          </a>
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
          <div className={styles.heroActions}>
            <Link href="/docs/installation" className={styles.primaryButton}>
              Get started <ArrowRight size={16} aria-hidden="true" />
            </Link>
            <a href="#preview" className={styles.secondaryButton}>
              Take a peek <ArrowDown size={15} aria-hidden="true" />
            </a>
          </div>
          <div className={styles.compatibility}>
            <span>Happy together with</span>
            <ul aria-label="Supported databases">
              <li>
                <span className={styles.postgres}>
                  <Database size={14} aria-hidden="true" />
                </span>
                PostgreSQL
              </li>
              <li>
                <span className={styles.mysql}>
                  <Database size={14} aria-hidden="true" />
                </span>
                MySQL
              </li>
              <li>
                <span className={styles.mongo}>
                  <Database size={14} aria-hidden="true" />
                </span>
                MongoDB
              </li>
              <li>
                <span className={styles.redis}>
                  <Database size={14} aria-hidden="true" />
                </span>
                Redis
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
              <span className={styles.statusDot} /> A little order. A lot less
              worry.
            </span>
            <span className={styles.handwritten}>
              meet your backup buddy <span aria-hidden="true">↙</span>
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
              <span className={styles.windowLabel}>
                YOUR LITTLE CONTROL ROOM
              </span>
            </div>
            <h2 id="preview-title" className={styles.srOnly}>
              One dashboard for your database backups
            </h2>
            <Image
              src="/images/connections.png"
              alt="Velld connections view with PostgreSQL databases, on-demand backup actions, and scheduled backups"
              width={3024}
              height={1572}
              className={styles.dashboardImage}
              priority
              sizes="(max-width: 1100px) 94vw, 1060px"
            />
          </div>
          <div className={styles.previewFootnote}>
            <ShieldCheck size={14} aria-hidden="true" /> Your infrastructure.
            Your credentials. Your data.
          </div>
        </section>

        <section
          id="features"
          className={styles.features}
          aria-labelledby="features-title"
        >
          <div className={styles.sectionHeading}>
            <span className={styles.eyebrow}>SMALL TOOL. BIG EXHALE.</span>
            <h2 id="features-title">
              Good backups. <span>Better sleep.</span>
            </h2>
            <p>
              Not another complicated platform.
              <br />
              Just the things you need to keep your databases safe.
            </p>
          </div>
          <div className={styles.featureGrid}>
            {features.map(
              ({
                icon: Icon,
                color,
                label,
                title,
                description,
                detail,
                caption,
              }) => (
                <article
                  className={`${styles.featureCard} ${styles[color]}`}
                  key={title}
                >
                  <span className={styles.featureIcon}>
                    <Icon size={22} aria-hidden="true" />
                  </span>
                  <span className={styles.featureLabel}>{label}</span>
                  <h3>{title}</h3>
                  <p>{description}</p>
                  <div className={styles.featureDetail}>
                    <span>{detail}</span>
                    <small>{caption}</small>
                  </div>
                </article>
              ),
            )}
          </div>
        </section>

        <section className={styles.setup} aria-labelledby="setup-title">
          <div className={styles.setupIntro}>
            <span className={styles.eyebrow}>MAKE YOURSELF AT HOME</span>
            <h2 id="setup-title">
              A few small steps.
              <br />
              <span>A safer database.</span>
            </h2>
            <p>
              No account with us. No subscription.
              <br />
              Just Velld, running on a server you control.
            </p>
            <Link href="/docs/installation" className={styles.textLink}>
              Let’s get you set up <ArrowRight size={15} aria-hidden="true" />
            </Link>
          </div>
          <div className={styles.setupCard}>
            <div className={styles.setupCardTitle}>
              <Terminal size={16} aria-hidden="true" /> A little terminal time{" "}
              <span>Docker Compose</span>
            </div>
            <ol className={styles.steps}>
              <li>
                <span className={styles.stepNumber}>1</span>
                <div>
                  <h3>Bring Velld home</h3>
                  <div className={styles.command}>
                    <code>{cloneCommand}</code>
                    <button
                      type="button"
                      onClick={copyCommand}
                      aria-label={
                        copyState === "copied"
                          ? "Command copied"
                          : "Copy clone command"
                      }
                      title="Copy clone command"
                    >
                      {copyState === "copied" ? (
                        <Check size={15} />
                      ) : (
                        <Copy size={15} />
                      )}
                    </button>
                  </div>
                  <span className={styles.copyStatus} role="status">
                    {copyState === "copied"
                      ? "Copied to clipboard!"
                      : copyState === "error"
                        ? "Couldn’t copy. Select the command to copy it manually."
                        : ""}
                  </span>
                </div>
              </li>
              <li>
                <span className={styles.stepNumber}>2</span>
                <div>
                  <h3>Give it the keys</h3>
                  <p>
                    In the cloned folder, copy <code>.env.example</code> to{" "}
                    <code>.env</code>, then set your admin credentials and
                    generate your secret keys.
                  </p>
                  <Link href="/docs/installation">
                    Follow the configuration guide{" "}
                    <ChevronRight size={12} aria-hidden="true" />
                  </Link>
                </div>
              </li>
              <li>
                <span className={styles.stepNumber}>3</span>
                <div>
                  <h3>You’re ready to roll</h3>
                  <code className={styles.launchCommand}>
                    docker compose up -d
                  </code>
                  <p>
                    From the cloned folder, start Velld. Then open{" "}
                    <code>localhost:3000</code> and connect your first database.
                  </p>
                </div>
              </li>
            </ol>
          </div>
        </section>

        <section className={styles.lastCall} aria-labelledby="last-call-title">
          <span className={styles.lastCallIcon}>
            <Heart size={25} aria-hidden="true" />
          </span>
          <span className={styles.eyebrow}>
            BUILT IN THE OPEN. MADE FOR YOUR SERVER.
          </span>
          <h2 id="last-call-title">
            Your data deserves
            <br />a happy little home.
          </h2>
          <p>Free, self-hosted, and open source. Always in your hands.</p>
          <div className={styles.heroActions}>
            <Link href="/docs/installation" className={styles.primaryButton}>
              Make your first backup <ArrowRight size={16} aria-hidden="true" />
            </Link>
            <a
              href={repository}
              target="_blank"
              rel="noopener noreferrer"
              className={styles.secondaryButton}
            >
              <Github size={16} aria-hidden="true" /> Explore the code
            </a>
          </div>
          <div className={styles.trustNotes}>
            <span>
              <Server size={13} aria-hidden="true" /> Self-hosted
            </span>
            <span>
              <Code2 size={13} aria-hidden="true" /> MIT licensed
            </span>
            <span>
              <Heart size={13} aria-hidden="true" /> No subscriptions
            </span>
          </div>
        </section>
      </main>
      <footer className={styles.footer}>
        <Link href="/" className={styles.brand}>
          <Image src="/images/logo.png" alt="" width={24} height={24} />
          <span>
            velld<span className={styles.brandDot}>.</span>
          </span>
        </Link>
        <span>A little care for your databases.</span>
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
