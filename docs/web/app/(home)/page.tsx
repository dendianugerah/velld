"use client";

import { ArrowUpRight, Code2, Maximize2, X } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { useRef, useState } from "react";
import styles from "./page.module.css";

const repository = "https://github.com/dendianugerah/velld";
const databases = [
  { name: "PostgreSQL", image: "/images/postgresql.svg" },
  { name: "MySQL", image: "/images/mysql.svg" },
  { name: "MongoDB", image: "/images/mongodb.svg" },
];
const screens = [
  {
    name: "Connections",
    image: "/images/connections.png",
    description: "Your databases, together in one place.",
    alt: "Velld Connections: database connections with on-demand backup actions and scheduled backups",
  },
  {
    name: "Dashboard",
    image: "/images/dashboard.png",
    description: "A little less checking. A little more peace of mind.",
    alt: "Velld Dashboard: an overview of database connections, backup activity, and storage",
  },
  {
    name: "History",
    image: "/images/history.png",
    description: "Every backup, with its history close by.",
    alt: "Velld History: completed database backups with status, size, and restore actions",
  },
];

export default function HomePage() {
  const [databaseIndex, setDatabaseIndex] = useState(0);
  const [screenIndex, setScreenIndex] = useState(0);
  const dialog = useRef<HTMLDialogElement>(null);
  const database = databases[databaseIndex];
  const screen = screens[screenIndex];

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
            GitHub <ArrowUpRight size={16} aria-hidden="true" />
          </a>
        </nav>
      </header>

      <main id="main" tabIndex={-1}>
        <section className={styles.hero} aria-labelledby="hero-title">
          <h1 id="hero-title" className={styles.title}>
            <span className={styles.titleLine}>
              Back up{" "}
              <button
                type="button"
                className={styles.databaseTile}
                onClick={() =>
                  setDatabaseIndex((index) => (index + 1) % databases.length)
                }
                aria-label={`Showing ${database.name}. Show next database`}
                title="Click to meet your databases"
              >
                {databases.map((item, index) => (
                  <Image
                    key={item.name}
                    src={item.image}
                    alt=""
                    width={54}
                    height={54}
                    className={styles.databaseIcon}
                    data-active={index === databaseIndex}
                    aria-hidden="true"
                    priority
                  />
                ))}
              </button>{" "}
              databases.
            </span>
            <span className={styles.titleLine}>
              Get back to{" "}
              <span className={styles.codeTile} aria-hidden="true">
                <Code2 strokeWidth={1.6} />
              </span>{" "}
              building.
            </span>
          </h1>
          <p className={styles.subtitle}>
            Backups, schedules, and restores. All in one place.
            <br />
            Open source. On your own server.
          </p>
          <ul className={styles.compatibility} aria-label="Supported databases">
            {databases.map((item, index) => (
              <li key={item.name}>
                <button
                  type="button"
                  aria-pressed={index === databaseIndex}
                  onClick={() => setDatabaseIndex(index)}
                >
                  <Image src={item.image} alt="" width={18} height={18} />
                  {item.name}
                </button>
              </li>
            ))}
          </ul>
        </section>

        <section className={styles.preview} aria-labelledby="preview-title">
          <div className={styles.previewToolbar}>
            <h2 id="preview-title">A look inside</h2>
            <div
              className={styles.screenPicker}
              role="group"
              aria-label="Choose a screenshot"
            >
              {screens.map((item, index) => (
                <button
                  key={item.name}
                  type="button"
                  onClick={() => setScreenIndex(index)}
                  aria-pressed={screenIndex === index}
                  aria-controls="screen-preview"
                >
                  {item.name}
                </button>
              ))}
            </div>
            <button
              type="button"
              className={styles.expandButton}
              aria-label={`Expand ${screen.name} screenshot`}
              onClick={() => dialog.current?.showModal()}
            >
              <Maximize2 size={17} aria-hidden="true" />
            </button>
          </div>
          <figure id="screen-preview" className={styles.figure}>
            <button
              type="button"
              className={styles.previewButton}
              onClick={() => dialog.current?.showModal()}
              aria-label={`View ${screen.name} screenshot at full size`}
            >
              {screens.map((item, index) => (
                <Image
                  key={item.name}
                  src={item.image}
                  alt={item.alt}
                  width={3024}
                  height={1572}
                  className={styles.dashboardImage}
                  hidden={screenIndex !== index}
                  priority={index === 0}
                  sizes="(max-width: 700px) 94vw, (max-width: 1440px) 86vw, 1240px"
                />
              ))}
              <span className={styles.previewHint} aria-hidden="true">
                Take a closer look <ArrowUpRight size={14} />
              </span>
            </button>
            <figcaption className={styles.caption} aria-live="polite">
              {screen.description}
            </figcaption>
          </figure>
        </section>
      </main>

      <footer className={styles.footer}>
        <div className={styles.footerIdentity}>
          <Link href="/" className={styles.brand}>
            velld
          </Link>
          <span>Self-hosted. Open source. Yours.</span>
        </div>
        <nav aria-label="Footer navigation">
          <Link href="/docs">Docs</Link>
          <a href={repository} target="_blank" rel="noopener noreferrer">
            GitHub <ArrowUpRight size={13} aria-hidden="true" />
          </a>
          <a
            href={`${repository}/blob/main/LICENSE`}
            target="_blank"
            rel="noopener noreferrer"
          >
            MIT License
          </a>
        </nav>
      </footer>

      <dialog
        ref={dialog}
        className={styles.lightbox}
        aria-labelledby="lightbox-title"
        onClick={(event) => {
          if (event.target === event.currentTarget) dialog.current?.close();
        }}
      >
        <div className={styles.lightboxHeader}>
          <h2 id="lightbox-title">{screen.name}</h2>
          <span>Preview</span>
          <button
            type="button"
            onClick={() => dialog.current?.close()}
            aria-label="Close screenshot"
          >
            <X size={20} aria-hidden="true" />
          </button>
        </div>
        <div className={styles.lightboxImage}>
          <Image
            src={screen.image}
            alt={screen.alt}
            width={3024}
            height={1572}
            sizes="96vw"
          />
        </div>
      </dialog>
    </div>
  );
}
