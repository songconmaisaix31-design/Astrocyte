import type { components } from '../../web/src/api/schema';

export type PublicPaperImport = Omit<components['schemas']['ImportMaterialRequestV1'], 'schema_version' | 'request_id' | 'expected_version'> & {
  kind: 'paper';
  adapter: 'manual';
  export_text: string;
  title: string;
  source_spans: string[];
  collection_reason: null;
};

export interface PublicPaperInput {
  source: {
    source_url: string;
    source_key: string;
    title: string;
    observed_version: string;
    text: string;
    content_state: 'readable_fulltext';
  };
  text: string;
  import: PublicPaperImport;
}

export interface PublicPaperSnapshot {
  schema_version: 1;
  source_url: string;
  host_family: string;
  source_key: string;
  title: string;
  abstract: string;
  authors: string[];
  doi: string;
  arxiv_id: string;
  observed_version: string;
  pdf_urls: string[];
  content_state: 'readable_fulltext' | 'abstract_only';
  text: string;
  warning: string;
  truncated: false;
  provenance: { processor: string; version: string; mode: string; source: string };
}

export function loadPublicPaperSnapshots(absoluteDirectory: string): Promise<{
  original: PublicPaperSnapshot;
  revised: PublicPaperSnapshot;
  secondary: PublicPaperSnapshot;
  abstract: PublicPaperSnapshot;
}>;

export function loadPublicPaperInputs(absoluteDirectory: string): Promise<{
  original: PublicPaperInput;
  revised: PublicPaperInput;
  secondary: PublicPaperInput;
}>;
