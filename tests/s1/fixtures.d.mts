import type { ImportInput } from './api.mjs';
export const paperText: string;
export const changedPaperText: string;
export const videoTranscript: string;
export function paperImport(locator?: string, text?: string): ImportInput & { title: string; export_text: string };
export function videoImport(locator?: string, options?: { summaryOnly?: boolean }): ImportInput & { export_text: string };
