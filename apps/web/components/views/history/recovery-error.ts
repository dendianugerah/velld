export function recoveryError(error: unknown, fallback: string) {
  if (error instanceof Error) {
    try { return (JSON.parse(error.message) as {message?: string}).message || fallback; } catch { return fallback; }
  }
  return fallback;
}
