// Dashboard System Types

export interface Widget {
  id: string;
  type:
    | 'number'
    | 'timeseries'
    | 'bar'
    | 'table'
    | 'status'
    | 'feed'
    | 'number_card'
    | 'bar_chart'
    | 'activity_feed'
    | 'briefing_tldr'
    | 'attention_items'
    | 'talking_points'
    | 'team_selector'
    | 'team_member_cards'
    | 'engineer_table'
    | 'unresolved_identities'
    | 'sprint_board'
    | 'velocity_trend'
    | 'timeline_estimator'
    | 'plugin_list'
    | 'event_stream'
    | 'alert_list'
    | 'goal_tracker'
    | 'cost_breakdown'
    | 'quick_actions';
  title: string;
  data_source: string;
  query_params?: Record<string, string>;
  config?: Record<string, unknown>;
  visualization_config?: VisualizationConfig;
  position: {
    x: number;
    y: number;
    w: number;
    h: number;
  };
}

export interface VisualizationConfig {
  primary_color?: string;
  secondary_color?: string;
  show_legend?: boolean;
  show_grid?: boolean;
  y_axis_label?: string;
  x_axis_label?: string;
  time_format?: string;
  number_format?: string;
  columns?: Array<{ key: string; label: string; sortable?: boolean; format?: string }>;
  threshold?: {
    good_min: number;
    warning_min: number;
    critical_max: number;
  };
  additional?: Record<string, unknown>;
}

export interface Dashboard {
  id: string;
  name: string;
  description?: string;
  persona?: string;
  is_template: boolean;
  is_system: boolean;
  widgets: Widget[];
  layout: string;
  filters?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateDashboardRequest {
  name: string;
  description?: string;
  is_template?: boolean;
  layout: string; // JSON string of { widgets: Widget[] }
}

export interface UpdateDashboardRequest {
  name?: string;
  description?: string;
  layout?: string; // JSON string of { widgets: Widget[] }
}

export interface MetricSnapshot {
  metric_name: string;
  entity_type: string;
  entity_id?: string;
  entity_name?: string;
  period_start: string;
  period_end: string;
  value: number;
  metadata?: string;
  computed_at: string;
}
