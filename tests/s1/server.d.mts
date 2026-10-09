export interface S1Server {
  apiURL: string;
  webURL: string;
  dataDir: string;
  temporary: string;
  close(): Promise<void>;
  restart(): Promise<void>;
  query(sql: string, ...parameters: (string | number)[]): Record<string, string | number | null>[];
}
export function startS1Server(options?: {
  browser?: boolean;
  env?: Record<string, string> | ((paths: { dataDir: string; temporary: string }) => Record<string, string>);
}): Promise<S1Server>;
