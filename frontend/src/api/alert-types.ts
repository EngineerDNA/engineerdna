// Alert and Monitoring Types (PDR-7)

export interface AlertRule {
  id: string;
  name: string;
  description?: string;
  alert_type: string;
  enabled: boolean;
  threshold_value?: number;
  threshold_operator?: string;
  severity: 'info' | 'warning' | 'critical';
  target_entity?: string;
  target_id?: string;
  created_at: string;
  updated_at: string;
}

export interface AlertInstance {
  id: string;
  rule_id: string;
  title: string;
  message: string;
  severity: 'info' | 'warning' | 'critical';
  entity_type?: string;
  entity_id?: string;
  context?: Record<string, unknown>;
  fired_at: string;
  acknowledged_at?: string;
  acknowledged_by?: string;
  snoozed_until?: string;
  dismissed_at?: string;
  dismissed_by?: string;
  resolved_at?: string;
}

export interface AlertChannel {
  id: string;
  rule_id: string;
  channel_type: 'slack' | 'email' | 'browser';
  channel_config?: Record<string, unknown>;
  enabled: boolean;
  quiet_hours_start?: string;
  quiet_hours_end?: string;
}

export interface LiveMetrics {
  prs_merged_today: number;
  prs_opened_today: number;
  prs_in_review: number;
  avg_review_time_today: number;
  active_alerts: number;
  sprint_progress: number;
}

export interface SprintBurndownData {
  sprint_id: string;
  sprint_name: string;
  start_date: string;
  end_date: string;
  total_points: number;
  completed_points: number;
  risk_level: 'on_track' | 'at_risk' | 'high_risk';
  burndown_points: Array<{
    date: string;
    ideal: number;
    actual: number;
  }>;
}
