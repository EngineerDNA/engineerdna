export const THROUGHPUT_DAYS = 30;
export const CHART_WIDTH = 500;
export const CHART_HEIGHT = 300;
export const DASHBOARD_EVENT_LIMIT = 50;
export const RECENT_EVENTS_COUNT = 5;
export const DEFAULT_AUDIT_LOG_LIMIT = 50;
export const EVENTS_PER_PAGE = 20;
export const BYTES_PER_KB = 1024;
export const HIGH_CONFIDENCE_THRESHOLD = 0.8;

// Time constants
export const HOURS_PER_DAY = 24;
export const MINUTES_PER_HOUR = 60;
export const SECONDS_PER_MINUTE = 60;
export const MILLISECONDS_PER_SECOND = 1000;
export const MILLISECONDS_PER_DAY =
  HOURS_PER_DAY * MINUTES_PER_HOUR * SECONDS_PER_MINUTE * MILLISECONDS_PER_SECOND;

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
} as const;
