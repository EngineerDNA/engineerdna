export interface Event {
  id: string;
  type: string;
  source: string;
  source_id: string;
  timestamp: string;
  actor: string;
  data: Record<string, unknown>;
  anonymized: boolean;
}

export interface ConfigField {
  name: string;
  type: 'string' | 'password' | 'boolean' | 'select';
  required: boolean;
  description: string;
  secret: boolean;
  default?: string;
  options?: string[];
}

export interface Plugin {
  name: string;
  type: 'source' | 'destination' | 'processor';
  enabled: boolean;
  last_sync?: string;
  health: 'healthy' | 'error' | 'unknown';
  config_fields?: ConfigField[];
}

export interface AnonymizationPolicy {
  plugin_name: string;
  enabled: boolean;
  strategy: 'sequential' | 'uuid' | 'hash';
  fields: string[];
}

export interface AnonymizationMapping {
  real_name: string;
  anonymized_id: string;
  plugins_using: string[];
}

export interface Insight {
  id: string;
  plugin_name: string;
  generated_at: string;
  severity: 'info' | 'warning' | 'error';
  title: string;
  description: string;
  recommendation?: string;
  metrics: Record<string, unknown>;
  dismissed: boolean;
}

export interface AuditLogEntry {
  id: string;
  timestamp: string;
  plugin_name: string;
  action: string;
  event_count: number;
  anonymized: boolean;
  destination?: string;
}

export interface ThroughputData {
  period: { start: string; end: string };
  data: Array<{
    date: string;
    pull_requests: number;
    issues: number;
  }>;
}

export interface TestResult {
  healthy: boolean;
  message: string;
}

export interface SystemInfo {
  database: {
    path: string;
    size_bytes: number;
    event_count: number;
  };
  encryption: {
    master_key_source: string;
  };
  plugins: {
    count: number;
  };
  version: string;
}

export interface Settings {
  id: string;
  sync_schedule: string;
  anonymization_strategy: string;
  port: number;
  created_at: string;
  updated_at: string;
}

export interface Engineer {
  id: string;
  name: string;
  email?: string;
  manager?: string;
  role_id?: string;
  identifiers: Record<string, string>;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface UnresolvedIdentity {
  id: string;
  source: string;
  identifier: string;
  first_seen: string;
  event_count: number;
  ignored: boolean;
}

export interface MatchSuggestion {
  unresolved_id: string;
  engineer_id: string;
  engineer_name: string;
  confidence: number;
  reason: string;
}

export interface EngineerActivity {
  engineer: {
    id: string;
    name: string;
  };
  period: {
    start: string;
    end: string;
  };
  metrics: {
    pull_requests: number;
    reviews: number;
    issues: number;
    commits: number;
  };
  recent_events: Event[];
}

export interface OnboardingStatus {
  completed: boolean;
}

export interface ApiError {
  code: string;
  message: string;
  user_message: string;
  suggestions?: string[];
  docs_url?: string;
  context?: Record<string, unknown>;
}

export interface ExportSchedule {
  id: string;
  plugin_name: string;
  frequency: 'daily' | 'weekly' | 'monthly';
  day_of_week?: number;
  time_of_day: string;
  enabled: boolean;
  last_run?: string;
  next_run: string;
  created_at: string;
  updated_at: string;
}

export interface CreateScheduleRequest {
  plugin_name: string;
  frequency: 'daily' | 'weekly' | 'monthly';
  day_of_week?: number;
  time_of_day: string;
  enabled?: boolean;
}

export interface UpdateScheduleRequest {
  plugin_name?: string;
  frequency?: 'daily' | 'weekly' | 'monthly';
  day_of_week?: number;
  time_of_day?: string;
  enabled?: boolean;
}

export interface Role {
  id: string;
  name: string;
  target_score: number;
  expectations: Record<string, number>;
  created_at: string;
  updated_at: string;
}

export interface ScoringWeights {
  id: string;
  throughput_weight: number;
  quality_weight: number;
  speed_weight: number;
  collaboration_weight: number;
  impact_weight: number;
  updated_at: string;
}

export interface RawMetrics {
  throughput_prs_per_week: number;
  throughput_story_points: number;
  quality_bug_rate: number;
  quality_rework_rate: number;
  speed_cycle_time_days: number;
  speed_time_to_first_review: number;
  collaboration_reviews_given: number;
  collaboration_review_depth: number;
  impact_services_touched: number;
}

export interface PerformanceScore {
  id: string;
  engineer_id: string;
  week_start: string;
  total_score: number;
  throughput_score: number;
  quality_score: number;
  speed_score: number;
  collaboration_score: number;
  impact_score: number;
  raw_metrics: string;
  created_at: string;
}

export interface AttentionItem {
  engineer_id: string;
  engineer_name: string;
  issue: string;
  evidence: string[];
  suggested_one_on_one: string;
  severity: 'warning' | 'critical';
}

export interface AIInsight {
  observation: string;
  context: string;
  recommendation: string;
  evidence: string[];
}

export interface WeeklyBriefing {
  team_id: string;
  team_name: string;
  week_start: string;
  week_end: string;
  tldr: string;
  key_metrics: Array<{
    label: string;
    value: number;
    unit: string;
    change: number;
    change_direction: 'up' | 'down' | 'stable';
    positive_trend: boolean;
  }>;
  needs_attention: AttentionItem[];
  insights: AIInsight[];
  trending_up: string[];
  trending_down: string[];
  talking_points: string;
  generated_at: string;
  last_updated: string;
}

export interface Sprint {
  id: string;
  team_id: string;
  name: string;
  start_date: string;
  end_date: string;
  committed_points: number;
  completed_points: number;
  status: 'planning' | 'active' | 'completed';
  created_at: string;
  updated_at: string;
}

export interface VelocityTrend {
  team_id: string;
  average_points: number;
  std_dev: number;
  trend: 'increasing' | 'decreasing' | 'stable';
  confidence: number;
  recent_sprints: Array<{
    sprint_id: string;
    sprint_name: string;
    points: number;
    start_date: string;
    end_date: string;
  }>;
}

export interface SprintHealth {
  sprint_id: string;
  sprint_name: string;
  status: 'planning' | 'active' | 'completed';
  risk_level: 'on_track' | 'at_risk' | 'high_risk';
  committed_points: number;
  completed_points: number;
  remaining_points: number;
  completion_percentage: number;
  days_remaining: number;
  estimated_completion_points: number;
  velocity_comparison: {
    historical_avg: number;
    current_committed: number;
    difference: number;
    difference_percentage: number;
  };
  alerts: string[];
}

export interface TimelineEstimate {
  feature_name: string;
  estimated_points: number;
  team_id: string;
  team_name: string;
  average_velocity: number;
  best_case_sprints: number;
  likely_case_sprints: number;
  worst_case_sprints: number;
  best_case_weeks: number;
  likely_case_weeks: number;
  worst_case_weeks: number;
  confidence_level: string;
  assumptions: string[];
}

// Re-export all types from split files for backward compatibility
export type {
  Widget,
  VisualizationConfig,
  Dashboard,
  CreateDashboardRequest,
  UpdateDashboardRequest,
  MetricSnapshot,
} from './dashboard-types';

export type {
  MetricSummary,
  MetricAggregateRequest,
  MetricAggregateResponse,
  MetricTimeseriesRequest,
  MetricTimeseriesResponse,
  MetricCompareRequest,
  MetricCompareResponse,
  EngineersPerformanceRequest,
  EngineerPerformanceRow,
  EngineersPerformanceResponse,
  HealthStatusResponse,
  AlertsTimelineRequest,
  AlertMarker,
} from './metric-types';

export type {
  Team,
  TeamMembership,
  TeamPerformanceScore,
  TeamScorecard,
  TeamHierarchyNode,
  OrgScorecard,
} from './team-types';

export type {
  AlertRule,
  AlertInstance,
  AlertChannel,
  LiveMetrics,
  SprintBurndownData,
} from './alert-types';
