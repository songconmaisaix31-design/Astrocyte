export interface S1Server {
  apiURL: string;
  webURL: string;
  dataDir: string;
  temporary: string;
  /** Pass the original error being rethrown from a finally; cleanup failures remain on stderr and error.cleanupError. */
  close(options?: { preserveData?: boolean; primaryError?: unknown }): Promise<void>;
  restart(): Promise<void>;
  query(sql: string, ...parameters: (string | number)[]): Record<string, string | number | null>[];
}
export function startS1Server(options?: {
  browser?: boolean;
  /** Exact previously created helper directory and controller-approved owned root. No discovery. Reused stores default to preservation. */
  reuseOwnedTemporary?: { path: string; ownedRoot: string };
  env?: Record<string, string> | ((paths: { dataDir: string; temporary: string }) => Record<string, string>);
}): Promise<S1Server>;
