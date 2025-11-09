// Metric API Response Types

export interface MetricSummary {
  label: string;
  value: number;
  unit: string;
  change: number;
  change_direction: 'up' | 'down' | 'stable';
  positive_trend: boolean;
}

export interface MetricAggregateRequest {
  metric_type: string;
  entity_type: 'engineer' | 'team' | 'org';
  entity_id?: string;
  time_range: string;
  comparison_period?: string;
}

export interface MetricAggregateResponse {
  metric: string;
  entity_type: string;
  entity_id?: string;
  period: string;
  period_start: string;
  period_end: string;
  value: number;
  previous: number;
  change: number;
  change_pct: number;
}

export interface MetricTimeseriesRequest {
  metric_type: string;
  entity_type: 'engineer' | 'team' | 'org';
  entity_id?: string;
  start_date: string;
  end_date: string;
  granularity?: 'day' | 'week' | 'month';
}

export interface MetricTimeseriesResponse {
  data_points: Array<{
    period_start: string;
    period_end: string;
    value: number;
  }>;
  metric: string;
  entity_type: string;
  entity_id?: string;
  period: string;
  count: number;
}

export interface MetricCompareRequest {
  metric_type: string;
  entity_type: 'engineer' | 'team';
  entity_ids: string[];
  time_range: string;
}

export interface MetricCompareResponse {
  entities: Array<{
    entity_id: string;
    entity_name: string;
    value: number;
  }>;
  metric: string;
  group_by: string;
  period: string;
}

export interface EngineersPerformanceRequest {
  team_id?: string;
  time_range: string;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
  limit?: number;
}

export interface EngineerPerformanceRow {
  engineer_id: string;
  engineer_name: string;
  total_score: number;
  throughput_score: number;
  quality_score: number;
  speed_score: number;
  collaboration_score: number;
  impact_score: number;
  prs_merged: number;
  cycle_time_days: number;
  reviews_given: number;
}

export interface EngineersPerformanceResponse {
  engineers: EngineerPerformanceRow[];
  total: number;
}

export interface HealthStatusResponse {
  status: 'ok' | 'warning' | 'critical';
  checks: Array<{
    name: string;
    status: 'ok' | 'warning' | 'critical';
    message?: string;
    value?: number;
    threshold?: number;
  }>;
}

export interface AlertsTimelineRequest {
  entity_type?: 'engineer' | 'team' | 'org';
  entity_id?: string;
  start_date: string;
  end_date: string;
}

export interface AlertMarker {
  timestamp: string;
  alert_id: string;
  title: string;
  severity: 'info' | 'warning' | 'critical';
}
