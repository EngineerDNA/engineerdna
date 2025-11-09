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

// API Response types
export interface EngineerScoresResponse {
  engineer_id: string;
  start_date: string;
  end_date: string;
  metrics: MetricValue[];
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
