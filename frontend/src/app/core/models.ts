export type DataSource = 'live' | 'synthetic';

export interface ServiceRecord {
  id: string;
  name: string;
  description?: string;
  criticality: string;
  source: DataSource;
  created_at: string;
  updated_at: string;
}

export interface DependencyRecord {
  from: string;
  to: string;
  kind: string;
}

export interface ReleaseRecord {
  id: string;
  service_id: string;
  version: string;
  git_sha: string;
  environment: string;
  status: string;
  source: DataSource;
  created_at: string;
}

export interface CommitRecord {
  sha: string;
  message: string;
  author: string;
  files_changed: number;
  lines_added: number;
  lines_deleted: number;
  migration_present: boolean;
  config_change_present: boolean;
  committed_at: string;
}

export interface DeploymentRecord {
  id: string;
  status: string;
  environment: string;
  target: string;
  image_digest?: string;
  artifact_uri?: string;
  started_at: string;
  completed_at?: string;
}

export interface CIRunRecord {
  id: string;
  workflow_name: string;
  status: string;
  failed_tests: number;
  started_at: string;
  completed_at?: string;
}

export interface ServiceListResponse {
  services: ServiceRecord[];
}

export interface ServiceDetailResponse {
  service: ServiceRecord;
  depends_on: DependencyRecord[];
  depended_by: DependencyRecord[];
}

export interface ReleaseListResponse {
  releases: ReleaseRecord[];
}

export interface ReleaseDetailResponse {
  release: ReleaseRecord;
  service: ServiceRecord;
  commit?: CommitRecord;
  deployment?: DeploymentRecord;
  ci_run?: CIRunRecord;
}

export interface HealthResponse {
  status: string;
  service: string;
  version: string;
}

export interface ReadyResponse {
  status: string;
  service: string;
  version: string;
  dependencies: Array<{ name: string; status: string }>;
}

export interface RiskFactor {
  code: string;
  label: string;
  points: number;
  rationale: string;
  input: string;
  omitted: boolean;
}

export interface RiskResponse {
  release_id: string;
  score: number;
  score_raw: number;
  category: string;
  model_version: string;
  assessed_at: string;
  disclaimer: string;
  factors: RiskFactor[];
}

export interface HealthMetric {
  name: string;
  baseline: number;
  post: number;
  abs_delta: number;
  pct_delta?: number;
  threshold: string;
  verdict: string;
  available: boolean;
  reason: string;
}

export interface ImpactNode {
  id: string;
  role: string;
  depth: number;
}

export interface ImpactResponse {
  changed_service_id: string;
  direct_dependents: string[];
  transitive_dependents: string[];
  upstream_dependencies: string[];
  critical_in_radius: string[];
  nodes: ImpactNode[];
  edges: Array<{ from: string; to: string }>;
  max_dependent_depth: number;
  cycles: string[][];
  unknown: boolean;
  empty: boolean;
  algorithm: string;
  disclaimer: string;
}

export interface HealthCompareResponse {
  release_id: string;
  overall: string;
  correlation: string;
  reasons: string[];
  metrics: HealthMetric[];
  baseline_from: string;
  baseline_to: string;
  post_from: string;
  post_to: string;
  model_version: string;
  compared_at: string;
  disclaimer: string;
}

export interface ApiErrorBody {
  error: {
    code: string;
    message: string;
    request_id: string;
  };
}
