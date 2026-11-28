import { Github } from 'lucide-react';
import { LargeSearchToggle } from 'fumadocs-ui/components/layout/search-toggle';
import Link from 'next/link';
import type { ReactNode } from 'react';
import styles from './layout.module.css';

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <div className={styles.site}>
      <header className={styles.topbar}>
        <nav className={styles.topbarInner} aria-label="Velld">
          <Link href="/" className={styles.wordmark} aria-label="Velld home">velld</Link>
          <Link href="/docs" className={styles.docsLabel}>Docs</Link>
          <div className={styles.topbarActions}>
            <LargeSearchToggle hideIfDisabled aria-label="Open Search" className={styles.search} />
            <a className={styles.iconLink} href="https://github.com/dendianugerah/velld" aria-label="Velld on GitHub">
              <Github size={16} aria-hidden="true" />
            </a>
          </div>
        </nav>
      </header>
      <main className={styles.main}>{children}</main>
    </div>
  );
}
