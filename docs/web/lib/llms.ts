import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import navigation from '@/content/docs/meta.json';
import { source } from '@/lib/source';

const site = 'https://velld.vercel.app';
const intro = '# Velld\n\n> Self-hosted database backups, restore verification, and recovery.\n';
const pages = navigation.pages.map((slug) => {
  const page = source.getPage(slug === 'index' ? [] : [slug]);
  if (!page) throw new Error(`Missing documentation page: ${slug}`);
  return page;
});

export function llmsIndex() {
  const links = pages.map(({ url, data }) =>
    `- [${data.title}](${site}${url}): ${data.description}`,
  );
  return [intro, '## Documentation\n', ...links, '', '## Optional\n',
    `- [Full documentation](${site}/llms-full.txt): All guides in one Markdown file.`, '',
  ].join('\n');
}

export async function llmsFull() {
  const sections = await Promise.all(pages.map(async ({ file, url, data }) => {
    const markdown = await readFile(join(process.cwd(), 'content/docs', file.path), 'utf8');
    const body = markdown.replace(/^---\r?\n[\s\S]*?\r?\n---\r?\n/, '').trim();
    return `# ${data.title}\n\nSource: ${site}${url}\n\n${body}`;
  }));
  return [intro.trimEnd(), ...sections].join('\n\n---\n\n') + '\n';
}

export function llmsText(body: string) {
  return new Response(body, { headers: { 'Content-Type': 'text/plain; charset=utf-8' } });
}
