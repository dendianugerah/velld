import defaultMdxComponents from 'fumadocs-ui/mdx';
import type { MDXComponents } from 'mdx/types';
import type { ComponentProps } from 'react';
import { PromptBlock } from '@/app/docs/prompt-block';

function DocsCodeBlock(props: ComponentProps<'pre'>) {
  if (props.title !== 'Velld install prompt') {
    return <defaultMdxComponents.pre {...props} />;
  }

  return <PromptBlock title={props.title}>{props.children}</PromptBlock>;
}

// use this function to get MDX components, you will need it for rendering MDX
export function getMDXComponents(components?: MDXComponents): MDXComponents {
  return {
    ...defaultMdxComponents,
    pre: DocsCodeBlock,
    ...components,
  };
}
