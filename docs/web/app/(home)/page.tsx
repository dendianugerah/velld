import Image from "next/image";
import Link from "next/link";
import styles from "./page.module.css";

const repository = "https://github.com/dendianugerah/velld";

const capabilities = [
  {
    number: "01",
    title: "Connect your databases",
    description:
      "PostgreSQL, MySQL, MongoDB, and Redis. Manage connections from one interface.",
  },
  {
    number: "02",
    title: "Set a backup schedule",
    description:
      "Run a backup when you need one, or automate it with a cron schedule and retention policy.",
  },
  {
    number: "03",
    title: "Keep a way back",
    description:
      "Store backups locally or in S3-compatible storage. Download a file or restore to an empty database.",
  },
];

export default function HomePage() {
  return (
    <div className={styles.home}>
      <a href="#main" className={styles.skipLink}>
        Skip to content
      </a>
      <header className={styles.header}>
        <Link href="/" className={styles.brand} aria-label="Velld home">
          <Image src="/images/logo.png" alt="" width={28} height={28} />
          <span>velld.</span>
        </Link>
        <nav aria-label="Main navigation" className={styles.navigation}>
          <Link href="/docs">Documentation</Link>
          <a href={repository} target="_blank" rel="noopener noreferrer">
            GitHub <span aria-hidden="true">↗</span>
          </a>
        </nav>
      </header>

      <main id="main" className={styles.main}>
        <section className={styles.intro} aria-labelledby="page-title">
          <p className={styles.eyebrow}>OPEN SOURCE / SELF-HOSTED</p>
          <h1 id="page-title">
            Database backups.
            <br />
            <span>On your server.</span>
          </h1>
          <p className={styles.description}>
            Connect your databases, schedule backups, and restore when you need
            to. Velld runs on your own infrastructure, with storage you control.
          </p>
          <div className={styles.actions}>
            <Link href="/docs/installation" className={styles.primaryLink}>
              Get started <span aria-hidden="true">→</span>
            </Link>
            <a
              href={repository}
              target="_blank"
              rel="noopener noreferrer"
              className={styles.sourceLink}
            >
              Browse the source <span aria-hidden="true">↗</span>
            </a>
          </div>
          <ul className={styles.databases} aria-label="Supported databases">
            <li>PostgreSQL</li>
            <li>MySQL</li>
            <li>MongoDB</li>
            <li>Redis</li>
          </ul>
        </section>

        <figure className={styles.preview}>
          <div className={styles.previewImage}>
            <Image
              src="/images/connections.png"
              alt="Velld connections interface: PostgreSQL connections with Backup Now and Schedule actions"
              width={3024}
              height={1572}
              priority
              sizes="(max-width: 1168px) 94vw, 1120px"
            />
          </div>
          <figcaption>
            <span>The connections view</span>
            <span>Back up now, or leave it on a schedule.</span>
          </figcaption>
        </figure>

        <section
          className={styles.capabilities}
          aria-labelledby="capabilities-title"
        >
          <div className={styles.sectionHeading}>
            <h2 id="capabilities-title">The essentials, in one place.</h2>
            <p>A small tool for a job worth getting right.</p>
          </div>
          <div className={styles.capabilityGrid}>
            {capabilities.map(({ number, title, description }) => (
              <article key={number}>
                <span className={styles.number} aria-hidden="true">
                  {number}
                </span>
                <h3>{title}</h3>
                <p>{description}</p>
              </article>
            ))}
          </div>
        </section>

        <aside className={styles.projectNote} aria-label="About the project">
          <p>Self-hosted, not a hosted service.</p>
          <p>
            Deploy with Docker Compose. Keep your credentials, backups, and
            configuration on infrastructure you control. Free to use and modify
            under the MIT license.
          </p>
          <Link href="/docs/installation">
            Read the installation guide <span aria-hidden="true">→</span>
          </Link>
        </aside>
      </main>

      <footer className={styles.footer}>
        <span>
          velld <span className={styles.footerDivider}>/</span> An open-source
          project.
        </span>
        <nav aria-label="Project links">
          <a
            href={`${repository}/blob/main/LICENSE`}
            target="_blank"
            rel="noopener noreferrer"
          >
            MIT License
          </a>
          <a
            href={`${repository}/issues`}
            target="_blank"
            rel="noopener noreferrer"
          >
            Issues
          </a>
          <a href={repository} target="_blank" rel="noopener noreferrer">
            Source <span aria-hidden="true">↗</span>
          </a>
        </nav>
      </footer>
    </div>
  );
}
