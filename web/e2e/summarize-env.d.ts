declare module '*scripts/summarize.mjs' {
  export function summarizeEnvironment(env: Record<string, string>): Promise<Record<string, string>>;
}
