export const DEFAULT_AUDIT_LOG_LIMIT = 50;
export const BYTES_PER_KB = 1024;
export const MILLISECONDS_PER_DAY = 24 * 60 * 60 * 1000;

// Query constants
export const QUERY_STALE_TIME_MS = 30000;
export const DEFAULT_TABLE_LIMIT = 10;

// Grid layout constants
export const GRID_BREAKPOINTS = { lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 } as const;
export const GRID_COLUMNS = { lg: 12, md: 10, sm: 6, xs: 4, xxs: 2 } as const;
export const GRID_ROW_HEIGHT = 100;

// Widget default dimensions
export const WIDGET_DEFAULTS = {
  number: { w: 3, h: 2 },
  timeseries: { w: 6, h: 3 },
  bar: { w: 6, h: 3 },
  table: { w: 12, h: 4 },
  status: { w: 6, h: 3 },
  feed: { w: 6, h: 3 },
  number_card: { w: 3, h: 2 },
  bar_chart: { w: 6, h: 3 },
  activity_feed: { w: 6, h: 4 },
  briefing_tldr: { w: 12, h: 3 },
  attention_items: { w: 6, h: 4 },
  talking_points: { w: 6, h: 4 },
  team_selector: { w: 4, h: 2 },
  team_member_cards: { w: 12, h: 4 },
  engineer_table: { w: 12, h: 5 },
  unresolved_identities: { w: 6, h: 4 },
  sprint_board: { w: 12, h: 6 },
  velocity_trend: { w: 8, h: 4 },
  timeline_estimator: { w: 8, h: 3 },
  plugin_list: { w: 6, h: 4 },
  event_stream: { w: 6, h: 5 },
  alert_list: { w: 6, h: 4 },
  goal_tracker: { w: 8, h: 5 },
  cost_breakdown: { w: 6, h: 4 },
  quick_actions: { w: 4, h: 3 },
} as const;
