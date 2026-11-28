import { source } from '@/lib/source';
import { DocsBody } from 'fumadocs-ui/page';
import { ArrowLeft, ArrowRight, ChevronDown } from 'lucide-react';
import Link from 'next/link';
import { notFound } from 'next/navigation';
import { createRelativeLink } from 'fumadocs-ui/mdx';
import { getMDXComponents } from '@/mdx-components';
import styles from '../layout.module.css';

const setupPages = new Set(['/docs', '/docs/installation', '/docs/databases']);
const sidebarLinks = source.pageTree.children.filter(node => node.type === 'page');
const guideGroups = [
  { title: 'Get started', pages: sidebarLinks.filter(node => setupPages.has(node.url)) },
  { title: 'Operate', pages: sidebarLinks.filter(node => !setupPages.has(node.url)) },
];

export default async function Page(props: {
  params: Promise<{ slug?: string[] }>;
}) {
  const params = await props.params;
  const page = source.getPage(params.slug);
  if (!page) notFound();

  const MDXContent = page.data.body;
  const sections = page.data.toc.filter(item => item.depth === 2);
  const current = sidebarLinks.findIndex(item => item.url === page.url);
  const previous = sidebarLinks[current - 1];
  const next = current === -1 ? undefined : sidebarLinks[current + 1];
  const guideNav = (
    <nav aria-label="Documentation guides" className={styles.guideNav}>
      {guideGroups.map(group => (
        <div key={group.title}>
          <p className={styles.navLabel}>{group.title}</p>
          <ul>
            {group.pages.map(item => (
              <li key={item.url}>
                <Link href={item.url} aria-current={item.url === page.url ? 'page' : undefined}>{item.name}</Link>
              </li>
            ))}
          </ul>
        </div>
      ))}
    </nav>
  );

  return (
    <div className={styles.frame}>
      <aside className={styles.rail}>{guideNav}</aside>
      <details className={styles.mobileNav} key={page.url}>
        <summary>Browse docs <ChevronDown size={16} aria-hidden="true" /></summary>
        {guideNav}
      </details>
      <article className={styles.article}>
        <h1>{page.data.title}</h1>
        <DocsBody className={styles.body}>
          <MDXContent components={getMDXComponents({ a: createRelativeLink(source, page) })} />
        </DocsBody>
        {(previous || next) && (
          <nav className={styles.pagination} aria-label="Previous and next guides">
            {previous && (
              <Link href={previous.url} rel="prev">
                <span>Previous</span>
                <strong><ArrowLeft size={14} aria-hidden="true" />{previous.name}</strong>
              </Link>
            )}
            {next && (
              <Link href={next.url} rel="next">
                <span>Next</span>
                <strong>{next.name}<ArrowRight size={14} aria-hidden="true" /></strong>
              </Link>
            )}
          </nav>
        )}
      </article>
      {sections.length > 0 && (
        <nav className={styles.toc} aria-label="On this page">
          <p className={styles.navLabel}>On this page</p>
          <ul>
            {sections.map(item => (
              <li key={item.url}><a href={item.url}>{item.title}</a></li>
            ))}
          </ul>
        </nav>
      )}
    </div>
  );
}

export async function generateStaticParams() {
  return source.generateParams();
}

export async function generateMetadata(props: {
  params: Promise<{ slug?: string[] }>;
}) {
  const params = await props.params;
  const page = source.getPage(params.slug);
  if (!page) notFound();

  return {
    title: page.data.title,
    description: page.data.description,
  };
}
