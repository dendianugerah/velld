import { llmsIndex, llmsText } from '@/lib/llms';

export const dynamic = 'force-static';

export function GET() {
  return llmsText(llmsIndex());
}
