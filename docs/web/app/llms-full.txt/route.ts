import { llmsFull, llmsText } from '@/lib/llms';

export const dynamic = 'force-static';

export async function GET() {
  return llmsText(await llmsFull());
}
