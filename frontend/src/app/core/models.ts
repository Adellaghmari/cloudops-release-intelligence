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
  scenario?: string;
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
  migration_reversible?: boolean;
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
  dependencies: { name: string; status: string }[];
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
  edges: { from: string; to: string }[];
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

export interface RollbackSignal {
  id: string;
  ok: boolean;
  not_applicable: boolean;
  detail: string;
  missing: boolean;
}

export interface RollbackResponse {
  release_id: string;
  status: string;
  signals: RollbackSignal[];
  missing: string[];
  model_version: string;
  assessed_at: string;
  disclaimer: string;
}

export interface PolicyRule {
  id: string;
  result: string;
  message: string;
  input_excerpt: string;
  skipped: boolean;
}

export interface PolicyResponse {
  release_id: string;
  policy_version: string;
  result: string;
  phase: string;
  rules: PolicyRule[];
  evaluated_at: string;
}

export interface TimelineEntry {
  event_id: string;
  event_type: string;
  occurred_at: string;
  producer: string;
  summary: string;
  release_id?: string;
  service_id?: string;
}

export interface TimelineResponse {
  release_id: string;
  entries: TimelineEntry[];
}

export interface ReplayField {
  path: string;
  kind: string;
  a: string;
  b: string;
}

export interface ReplayResponse {
  a: string;
  b: string;
  fields: ReplayField[];
}

export interface ApiErrorBody {
  error: {
    code: string;
    message: string;
    request_id: string;
  };
}

export interface OperationalEvidence {
  id: string;
  kind: string;
  git_sha?: string;
  branch?: string;
  workflow_name?: string;
  workflow_run_id?: string;
  workflow_result?: string;
  test_result?: string;
  build_duration_ms?: number;
  image_digest?: string;
  artifact_id?: string;
  security_scan_result?: string;
  deployment_timestamp?: string;
  recorded_at: string;
  source: DataSource;
  label: string;
}

export interface SystemStatusResponse {
  service: string;
  version: string;
  git_sha?: string;
  store: string;
  live_project_data: OperationalEvidence[];
  synthetic_demo_note: string;
}
