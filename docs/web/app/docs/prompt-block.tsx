'use client';

import { Check, ChevronDown, Copy } from 'lucide-react';
import { useEffect, useRef, useState, type ReactNode } from 'react';
import styles from './layout.module.css';

export function PromptBlock({ title, children }: { title: string; children: ReactNode }) {
  const prompt = useRef<HTMLPreElement>(null);
  const [status, setStatus] = useState<'idle' | 'copied' | 'failed'>('idle');

  useEffect(() => {
    if (status === 'idle') return;
    const timer = window.setTimeout(() => setStatus('idle'), 2500);
    return () => window.clearTimeout(timer);
  }, [status]);

  async function copy() {
    if (!prompt.current) return;
    try {
      await navigator.clipboard.writeText(prompt.current.textContent ?? '');
      setStatus('copied');
    } catch {
      setStatus('failed');
    }
  }

  return (
    <figure className={`not-prose ${styles.prompt}`} aria-label={title}>
      <div className={styles.promptHeader}>
        <p>Paste into a coding agent on your Docker host.</p>
        <button type="button" data-status={status} onClick={copy}>
          {status === 'copied' ? <Check size={14} aria-hidden="true" /> : <Copy size={14} aria-hidden="true" />}
          <span>{status === 'copied' ? 'Copied' : status === 'failed' ? 'Try again' : 'Copy prompt'}</span>
        </button>
        <span className="sr-only" role="status">
          {status === 'copied' ? 'Copied' : status === 'failed' ? 'Could not copy. Show the prompt to copy it manually.' : ''}
        </span>
      </div>
      <details className={styles.promptDetails}>
        <summary>
          <ChevronDown size={14} aria-hidden="true" />
          <span>Show prompt</span>
          <span>Hide prompt</span>
        </summary>
        <pre className={styles.promptBody} ref={prompt}>{children}</pre>
      </details>
    </figure>
  );
}
