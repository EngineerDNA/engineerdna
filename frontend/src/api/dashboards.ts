import type {
  Dashboard,
  CreateDashboardRequest,
  UpdateDashboardRequest,
  MetricAggregateRequest,
  MetricAggregateResponse,
  MetricTimeseriesRequest,
  MetricTimeseriesResponse,
  MetricCompareRequest,
  MetricCompareResponse,
  EngineersPerformanceRequest,
  EngineersPerformanceResponse,
  HealthStatusResponse,
  AlertsTimelineRequest,
  AlertMarker,
} from './types';
import type { MetricSnapshot } from './dashboard-types';

const API_BASE = '/api';

async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  });

  if (!response.ok) {
    const contentType = response.headers.get('content-type');

    try {
      if (contentType?.includes('application/json')) {
        const errorData = await response.json();
        if (errorData.code && errorData.user_message) {
          throw new Error(errorData.user_message);
        }
      } else {
        const textError = await response.text();
        if (textError && textError.trim()) {
          throw new Error(textError);
        }
      }
    } catch (parseError) {
      if (parseError instanceof Error) {
        throw parseError;
      }
    }
    throw new Error(`API error: ${response.statusText}`);
  }

  return response.json();
}

export const dashboardsApi = {
  // Dashboards (PDR-8)
  getDashboards: (params?: {
    persona?: string;
    templates?: boolean;
    limit?: number;
    offset?: number;
  }) => {
    if (!params) {
      return fetchAPI<{
        dashboards: Dashboard[];
        pagination: { total: number; limit: number; offset: number; hasMore: boolean };
      }>('/dashboards');
    }

    const queryParams = new URLSearchParams();
    if (params.persona) queryParams.set('persona', params.persona);
    if (params.templates !== undefined) queryParams.set('templates', String(params.templates));
    if (params.limit) queryParams.set('limit', String(params.limit));
    if (params.offset) queryParams.set('offset', String(params.offset));

    const query = queryParams.toString();
    return fetchAPI<{
      dashboards: Dashboard[];
      pagination: { total: number; limit: number; offset: number; hasMore: boolean };
    }>(`/dashboards${query ? `?${query}` : ''}`);
  },

  getDashboard: (id: string) => fetchAPI<Dashboard>(`/dashboards/${id}`),

  createDashboard: (data: CreateDashboardRequest) =>
    fetchAPI<Dashboard>('/dashboards', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  updateDashboard: (id: string, data: UpdateDashboardRequest) =>
    fetchAPI<Dashboard>(`/dashboards/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  deleteDashboard: (id: string) =>
    fetchAPI<{ message: string }>(`/dashboards/${id}`, {
      method: 'DELETE',
    }),

  cloneDashboard: (id: string, name?: string) =>
    fetchAPI<Dashboard>(`/dashboards/${id}/clone`, {
      method: 'POST',
      body: JSON.stringify({ name }),
    }),

  getMetricSnapshots: (params: {
    metric: string;
    entity_type: string;
    entity_id?: string;
    period_start: string;
    period_end: string;
    limit?: number;
    offset?: number;
  }) => {
    const query = new URLSearchParams();
    query.append('metric', params.metric);
    query.append('entity_type', params.entity_type);
    if (params.entity_id) query.append('entity_id', params.entity_id);
    query.append('start_date', params.period_start);
    query.append('end_date', params.period_end);
    if (params.limit) query.append('limit', params.limit.toString());
    if (params.offset) query.append('offset', params.offset.toString());

    return fetchAPI<{ snapshots: MetricSnapshot[]; count: number }>(`/metrics/snapshots?${query}`);
  },

  // Metrics
  getMetricAggregate: (params: MetricAggregateRequest) => {
    const queryParams = new URLSearchParams();
    queryParams.set('metric', params.metric_type);
    queryParams.set('entity_type', params.entity_type);
    if (params.entity_id) queryParams.set('entity_id', params.entity_id);
    queryParams.set('period', params.time_range);
    return fetchAPI<MetricAggregateResponse>(`/metrics/aggregate?${queryParams.toString()}`);
  },

  getMetricTimeseries: (params: MetricTimeseriesRequest) => {
    const queryParams = new URLSearchParams();
    queryParams.set('metric', params.metric_type);
    queryParams.set('entity_type', params.entity_type);
    if (params.entity_id) queryParams.set('entity_id', params.entity_id);
    queryParams.set('period', params.granularity || 'week');
    return fetchAPI<MetricTimeseriesResponse>(`/metrics/timeseries?${queryParams.toString()}`);
  },

  getMetricCompare: (params: MetricCompareRequest) => {
    const queryParams = new URLSearchParams();
    queryParams.set('metric', params.metric_type);
    queryParams.set('group_by', params.entity_type);
    queryParams.set('period', params.time_range);
    return fetchAPI<MetricCompareResponse>(`/metrics/compare?${queryParams.toString()}`);
  },

  getEngineersPerformance: (params: EngineersPerformanceRequest) => {
    const queryParams = new URLSearchParams();
    if (params.team_id) queryParams.set('team_id', params.team_id);
    queryParams.set('time_range', params.time_range);
    if (params.sort_by) queryParams.set('sort_by', params.sort_by);
    if (params.sort_order) queryParams.set('sort_order', params.sort_order);
    if (params.limit) queryParams.set('limit', String(params.limit));
    return fetchAPI<EngineersPerformanceResponse>(
      `/engineers/performance?${queryParams.toString()}`
    );
  },

  getHealthStatus: () => fetchAPI<HealthStatusResponse>('/metrics/health-status'),

  getAlertsTimeline: (params: AlertsTimelineRequest) => {
    const queryParams = new URLSearchParams();
    if (params.entity_type) queryParams.set('entity_type', params.entity_type);
    if (params.entity_id) queryParams.set('entity_id', params.entity_id);
    queryParams.set('start', params.start_date);
    queryParams.set('end', params.end_date);
    return fetchAPI<{ alerts: AlertMarker[] }>(`/alerts/timeline?${queryParams.toString()}`);
  },
};
