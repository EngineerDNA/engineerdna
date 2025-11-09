import type {
  Event,
  Plugin,
  AnonymizationPolicy,
  AnonymizationMapping,
  Insight,
  AuditLogEntry,
  ThroughputData,
  TestResult,
  SystemInfo,
  Settings,
  Engineer,
  UnresolvedIdentity,
  MatchSuggestion,
  EngineerActivity,
  OnboardingStatus,
  ApiError as ApiErrorType,
  ExportSchedule,
  CreateScheduleRequest,
  UpdateScheduleRequest,
  Role,
  ScoringWeights,
  Team,
  TeamScorecard,
  TeamHierarchyNode,
  OrgScorecard,
  TeamPerformanceScore,
  WeeklyBriefing,
  Sprint,
  VelocityTrend,
  SprintHealth,
  TimelineEstimate,
  AlertRule,
  AlertInstance,
  AlertChannel,
  LiveMetrics,
  SprintBurndownData,
  Dashboard,
  DashboardTemplate,
} from './types';
import type { EngineerScoresResponse } from '../types/metrics';
import { DEFAULT_AUDIT_LOG_LIMIT } from '../constants';

const API_BASE = '/api';

export class ApiError extends Error implements ApiErrorType {
  code: string;
  user_message: string;
  suggestions?: string[];
  docs_url?: string;
  context?: Record<string, unknown>;

  constructor(error: ApiErrorType) {
    super(error.message);
    this.name = 'ApiError';
    this.code = error.code;
    this.user_message = error.user_message;
    this.suggestions = error.suggestions;
    this.docs_url = error.docs_url;
    this.context = error.context;
  }
}

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
      // Try JSON first (structured API errors)
      if (contentType?.includes('application/json')) {
        const errorData = await response.json();
        if (errorData.code && errorData.user_message) {
          throw new ApiError(errorData as ApiErrorType);
        }
      } else {
        // Plain text error message (http.Error responses)
        const textError = await response.text();
        if (textError && textError.trim()) {
          throw new Error(textError);
        }
      }
    } catch (parseError) {
      if (parseError instanceof ApiError || parseError instanceof Error) {
        throw parseError;
      }
      // If parsing fails, fallback to generic error
    }
    throw new Error(`API error: ${response.statusText}`);
  }

  return response.json();
}

export const api = {
  // Health & System
  health: () => fetchAPI<{ status: string }>('/health'),
  version: () => fetchAPI<{ version: string }>('/version'),
  getSystemInfo: () => fetchAPI<SystemInfo>('/system/info'),

  // Settings
  getSettings: () => fetchAPI<Settings>('/settings'),

  // Events
  getEvents: (params?: {
    since?: string;
    type?: string;
    source?: string;
    limit?: number;
    offset?: number;
  }) => {
    const queryParams: Record<string, string> = {};
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          queryParams[key] = String(value);
        }
      });
    }
    const query = new URLSearchParams(queryParams).toString();
    return fetchAPI<{ events: Event[]; total: number; limit: number; offset: number }>(
      `/events?${query}`
    );
  },

  // Plugins
  getPlugins: () => fetchAPI<{ plugins: Plugin[] }>('/plugins'),
  getPlugin: (name: string) => fetchAPI<Plugin>(`/plugins/${name}`),
  configurePlugin: (name: string, config: Record<string, string>) =>
    fetchAPI(`/plugins/${name}/configure`, {
      method: 'POST',
      body: JSON.stringify(config),
    }),
  syncPlugin: (name: string) => fetchAPI(`/plugins/${name}/sync`, { method: 'POST' }),
  testPlugin: (name: string) => fetchAPI<TestResult>(`/plugins/${name}/test`, { method: 'POST' }),

  // Anonymization
  getAnonymizationPolicies: () =>
    fetchAPI<{ policies: AnonymizationPolicy[] }>('/anonymization/policies'),
  updateAnonymizationPolicy: (pluginName: string, policy: AnonymizationPolicy) =>
    fetchAPI(`/anonymization/policies/${pluginName}`, {
      method: 'POST',
      body: JSON.stringify(policy),
    }),
  getAnonymizationMappings: () =>
    fetchAPI<{ mappings: AnonymizationMapping[] }>('/anonymization/mappings'),
  getAuditLog: (limit?: number) =>
    fetchAPI<{ entries: AuditLogEntry[] }>(
      `/anonymization/audit?limit=${limit || DEFAULT_AUDIT_LOG_LIMIT}`
    ),

  // Insights
  getInsights: () => fetchAPI<{ insights: Insight[] }>('/insights'),
  generateInsights: (pluginName: string, period: { start: string; end: string }) =>
    fetchAPI('/insights/generate', {
      method: 'POST',
      body: JSON.stringify({ plugin: pluginName, period }),
    }),
  dismissInsight: (id: string) => fetchAPI(`/insights/${id}/dismiss`, { method: 'PUT' }),

  // Metrics
  getThroughput: (days: number) => fetchAPI<ThroughputData>(`/metrics/throughput?days=${days}`),

  // Engineers
  getEngineers: () => fetchAPI<{ engineers: Engineer[] }>('/engineers'),
  getEngineer: (id: string) => fetchAPI<{ engineer: Engineer }>(`/engineers/${id}`),
  createEngineer: (data: Partial<Engineer>) =>
    fetchAPI('/engineers', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  updateEngineer: (id: string, data: Partial<Engineer>) =>
    fetchAPI(`/engineers/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  deleteEngineer: (id: string) =>
    fetchAPI(`/engineers/${id}`, {
      method: 'DELETE',
    }),
  getEngineerActivity: (id: string, days: number) =>
    fetchAPI<EngineerActivity>(`/engineers/${id}/activity?days=${days}`),
  importEngineers: async (file: File, columnMapping: Record<string, string>) => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('column_mapping', JSON.stringify(columnMapping));

    const response = await fetch(`${API_BASE}/engineers/import`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`API error: ${response.statusText}`);
    }

    return response.json();
  },

  // Identity Resolution
  getUnresolved: () => fetchAPI<{ unresolved: UnresolvedIdentity[] }>('/identity/unresolved'),
  getIgnoredIdentities: () => fetchAPI<{ ignored: UnresolvedIdentity[] }>('/identity/ignored'),
  resolveIdentity: (unresolvedId: string, engineerId: string | null, name?: string) =>
    fetchAPI('/identity/resolve', {
      method: 'POST',
      body: JSON.stringify({ unresolved_id: unresolvedId, engineer_id: engineerId, name }),
    }),
  getSuggestions: () => fetchAPI<{ suggestions: MatchSuggestion[] }>('/identity/suggestions'),
  mergeEngineers: (keepId: string, mergeId: string) =>
    fetchAPI('/identity/merge', {
      method: 'POST',
      body: JSON.stringify({ keep_id: keepId, merge_id: mergeId }),
    }),
  ignoreIdentity: (unresolvedId: string, ignored: boolean) =>
    fetchAPI('/identity/ignore', {
      method: 'POST',
      body: JSON.stringify({ unresolved_id: unresolvedId, ignored }),
    }),

  // Onboarding
  getOnboardingStatus: () => fetchAPI<OnboardingStatus>('/onboarding/status'),
  completeOnboarding: () =>
    fetchAPI<{ status: string }>('/onboarding/complete', { method: 'POST' }),

  // Export Schedules
  getSchedules: () => fetchAPI<{ schedules: ExportSchedule[] }>('/exports/schedules'),
  getSchedule: (id: string) => fetchAPI<ExportSchedule>(`/exports/schedules/${id}`),
  createSchedule: (req: CreateScheduleRequest) =>
    fetchAPI<ExportSchedule>('/exports/schedules', {
      method: 'POST',
      body: JSON.stringify(req),
    }),
  updateSchedule: (id: string, req: UpdateScheduleRequest) =>
    fetchAPI<ExportSchedule>(`/exports/schedules/${id}`, {
      method: 'PUT',
      body: JSON.stringify(req),
    }),
  deleteSchedule: (id: string) =>
    fetchAPI<{ message: string }>(`/exports/schedules/${id}`, {
      method: 'DELETE',
    }),
  triggerSchedule: (id: string) =>
    fetchAPI<{ message: string }>(`/exports/schedules/${id}/run`, {
      method: 'POST',
    }),

  // Roles
  getRoles: () => fetchAPI<{ roles: Role[] }>('/roles'),
  getRole: (id: string) => fetchAPI<Role>(`/roles/${id}`),
  createRole: (data: {
    name: string;
    target_score: number;
    expectations: Record<string, number>;
  }) =>
    fetchAPI<Role>('/roles', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  updateRole: (
    id: string,
    data: { name: string; target_score: number; expectations: Record<string, number> }
  ) =>
    fetchAPI<Role>(`/roles/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  deleteRole: (id: string) =>
    fetchAPI<{ status: string }>(`/roles/${id}`, {
      method: 'DELETE',
    }),

  // Scoring Weights
  getScoringWeights: () => fetchAPI<ScoringWeights>('/scoring/weights'),
  updateScoringWeights: (weights: {
    throughput_weight: number;
    quality_weight: number;
    speed_weight: number;
    collaboration_weight: number;
    impact_weight: number;
  }) =>
    fetchAPI<ScoringWeights>('/scoring/weights', {
      method: 'PUT',
      body: JSON.stringify(weights),
    }),

  // Performance Scores
  getPerformanceScores: (engineerId: string, startDate?: string, endDate?: string) => {
    const params = new URLSearchParams();
    if (startDate) params.set('start_date', startDate);
    if (endDate) params.set('end_date', endDate);
    const query = params.toString();
    return fetchAPI<EngineerScoresResponse>(
      `/performance/individual/${engineerId}${query ? `?${query}` : ''}`
    );
  },

  // Teams
  getTeams: () => fetchAPI<{ teams: Team[] }>('/teams'),
  getTeam: (id: string) => fetchAPI<Team>(`/teams/${id}`),
  getTeamHierarchy: () => fetchAPI<{ hierarchy: TeamHierarchyNode[] }>('/teams/hierarchy'),
  getOrgScorecard: () => fetchAPI<OrgScorecard>('/teams/org/scorecard'),
  getTeamMembers: (id: string) => fetchAPI<{ members: Engineer[] }>(`/teams/${id}/members`),
  getTeamPerformance: (id: string) => fetchAPI<TeamPerformanceScore>(`/teams/${id}/performance`),
  getTeamScorecard: (id: string) => fetchAPI<TeamScorecard>(`/teams/${id}/scorecard`),

  // Briefing
  getWeeklyBriefing: (teamId: string, weekStart?: string) => {
    const params = weekStart ? `?week_start=${weekStart}` : '';
    return fetchAPI<WeeklyBriefing>(`/briefing/weekly/${teamId}${params}`);
  },
  generateWeeklyBriefing: (teamId: string, weekStart?: string) =>
    fetchAPI<WeeklyBriefing>(`/briefing/weekly/${teamId}/generate`, {
      method: 'POST',
      body: JSON.stringify({ week_start: weekStart }),
    }),

  // Planning
  getSprints: (teamId?: string, status?: string) => {
    const params = new URLSearchParams();
    if (teamId) params.set('team_id', teamId);
    if (status) params.set('status', status);
    const query = params.toString();
    return fetchAPI<{ sprints: Sprint[] }>(`/planning/sprints${query ? `?${query}` : ''}`);
  },
  getSprint: (sprintId: string) => fetchAPI<Sprint>(`/planning/sprints/${sprintId}`),
  createSprint: (data: {
    team_id: string;
    name: string;
    start_date: string;
    end_date: string;
    committed_points: number;
  }) =>
    fetchAPI<Sprint>('/planning/sprints', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  updateSprint: (
    sprintId: string,
    data: {
      name?: string;
      committed_points?: number;
      completed_points?: number;
      status?: string;
    }
  ) =>
    fetchAPI<Sprint>(`/planning/sprints/${sprintId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  deleteSprint: (sprintId: string) =>
    fetchAPI<{ message: string }>(`/planning/sprints/${sprintId}`, {
      method: 'DELETE',
    }),
  getSprintHealth: (sprintId: string) =>
    fetchAPI<SprintHealth>(`/planning/sprints/${sprintId}/health`),
  getVelocity: (teamId: string, numSprints?: number) => {
    const params = numSprints ? `?num_sprints=${numSprints}` : '';
    return fetchAPI<VelocityTrend>(`/planning/velocity/${teamId}${params}`);
  },
  estimateTimeline: (data: { feature_name: string; estimated_points: number; team_id: string }) =>
    fetchAPI<TimelineEstimate>('/planning/estimate', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  // Alert Instances
  getAlertInstances: (filters?: {
    rule_id?: string;
    severity?: string;
    status?: string;
    entity_type?: string;
    entity_id?: string;
  }) => {
    const params = new URLSearchParams();
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined) {
          params.set(key, String(value));
        }
      });
    }
    const query = params.toString();
    return fetchAPI<{ alerts: AlertInstance[] }>(`/alerts/instances${query ? `?${query}` : ''}`);
  },
  getAlertInstance: (id: string) => fetchAPI<AlertInstance>(`/alerts/instances/${id}`),
  acknowledgeAlert: (id: string) =>
    fetchAPI<{ message: string }>(`/alerts/instances/${id}/acknowledge`, {
      method: 'POST',
    }),
  snoozeAlert: (id: string, durationMinutes: number) =>
    fetchAPI<{ message: string }>(`/alerts/instances/${id}/snooze`, {
      method: 'POST',
      body: JSON.stringify({ duration_minutes: durationMinutes }),
    }),
  dismissAlert: (id: string) =>
    fetchAPI<{ message: string }>(`/alerts/instances/${id}/dismiss`, {
      method: 'POST',
    }),
  resolveAlert: (id: string) =>
    fetchAPI<{ message: string }>(`/alerts/instances/${id}/resolve`, {
      method: 'POST',
    }),

  // Alert Rules
  getAlertRules: () => fetchAPI<{ rules: AlertRule[] }>('/alerts/rules'),
  getAlertRule: (id: string) => fetchAPI<AlertRule>(`/alerts/rules/${id}`),
  createAlertRule: (data: Partial<AlertRule>) =>
    fetchAPI<AlertRule>('/alerts/rules', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  updateAlertRule: (id: string, data: Partial<AlertRule>) =>
    fetchAPI<AlertRule>(`/alerts/rules/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  deleteAlertRule: (id: string) =>
    fetchAPI<{ message: string }>(`/alerts/rules/${id}`, {
      method: 'DELETE',
    }),

  // Alert Channels
  getAlertChannels: (ruleId?: string) => {
    const params = ruleId ? `?rule_id=${ruleId}` : '';
    return fetchAPI<{ channels: AlertChannel[] }>(`/alerts/channels${params}`);
  },
  createAlertChannel: (data: Partial<AlertChannel>) =>
    fetchAPI<AlertChannel>('/alerts/channels', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  updateAlertChannel: (id: string, data: Partial<AlertChannel>) =>
    fetchAPI<AlertChannel>(`/alerts/channels/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  deleteAlertChannel: (id: string) =>
    fetchAPI<{ message: string }>(`/alerts/channels/${id}`, {
      method: 'DELETE',
    }),

  // Live Metrics
  getTodayMetrics: () => fetchAPI<LiveMetrics>('/metrics/today'),
  getSprintBurndown: (sprintId?: string) => {
    const params = sprintId ? `?sprint_id=${sprintId}` : '';
    return fetchAPI<SprintBurndownData>(`/planning/sprint-burndown${params}`);
  },
  getTodayActivity: () => {
    return fetchAPI<{ events: Event[] }>(`/activity/today`);
  },

  // Dashboard Templates
  getDashboardTemplates: (role?: string) => {
    const params = role ? `?role=${role}` : '';
    return fetchAPI<{ templates: DashboardTemplate[] }>(`/dashboard-templates${params}`);
  },
  getDashboardTemplate: (id: string) => fetchAPI<DashboardTemplate>(`/dashboard-templates/${id}`),
  createDashboardFromTemplate: (templateId: string, name?: string) =>
    fetchAPI<Dashboard>(`/dashboards/from-template/${templateId}`, {
      method: 'POST',
      body: JSON.stringify({ name }),
    }),

  // Settings - Role
  getUserRole: () => fetchAPI<{ role: string }>('/settings/role'),
  setUserRole: (role: string) =>
    fetchAPI<{ message: string }>('/settings/role', {
      method: 'PUT',
      body: JSON.stringify({ role }),
    }),

  // Settings - Primary Dashboard
  getPrimaryDashboard: () =>
    fetchAPI<{ primary_dashboard_id: string | null }>('/settings/primary-dashboard'),
  setPrimaryDashboard: (dashboardId: string) =>
    fetchAPI<{ message: string }>('/settings/primary-dashboard', {
      method: 'PUT',
      body: JSON.stringify({ dashboard_id: dashboardId }),
    }),
};
