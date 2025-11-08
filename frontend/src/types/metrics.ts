export interface MetricValue {
  id: string;
  metric_name: string;
  source: string;
  timestamp: string;
  granularity: 'hourly' | 'daily' | 'weekly' | 'monthly';
  value: number;
  unit?: string;
  dimensions?: Record<string, any>;
  created_at: string;
}

export interface EntityAttribute {
  id: string;
  entity_type: string;
  entity_id: string;
  attribute_name: string;
  value: string;
  value_type: 'string' | 'number' | 'boolean' | 'currency' | 'json';
  valid_from: string;
  valid_until?: string;
  source: string;
  created_at: string;
}

// Legacy interfaces for backward compatibility (v1.0)
// @deprecated Use MetricValue[] instead - will be removed in v2.0
export interface PerformanceScore {
  engineer_id: string;
  week_start: string;
  total_score?: number;
  throughput_score?: number;
  quality_score?: number;
  speed_score?: number;
  collaboration_score?: number;
  impact_score?: number;
}

export interface CostConfiguration {
  entity_type: string;
  entity_id: string;
  monthly_cost: number;
  currency: string;
  effective_from: string;
  effective_to?: string;
  notes?: string;
}

// API Response types (v1.3+)
export interface EngineerScoresResponse {
  engineer_id: string;
  start_date: string;
  end_date: string;
  metrics: MetricValue[];
}

// Legacy API Response types (v1.0)
// @deprecated Use EngineerScoresResponse instead - will be removed in v2.0
export interface PerformanceScoresResponse {
  engineer_id: string;
  start_date: string;
  end_date: string;
  scores: PerformanceScore[];
}

export interface EngineerCostResponse {
  engineer_id: string;
  monthly_cost: number;
  currency: string;
  message?: string;
}

export interface TeamCostResponse {
  team_id: string;
  monthly_cost: number;
  currency: string;
}
