import { useQuery } from '@tanstack/react-query';
import { dashboardsApi } from '../api/dashboards';
import { api } from '../api/client';
import type { Widget } from '../api/types';
import { MILLISECONDS_PER_DAY, DEFAULT_TABLE_LIMIT, QUERY_STALE_TIME_MS } from '../constants';

interface Dashboard {
  widgets: Widget[];
}

/**
 * Fetches data for all widgets on a dashboard
 * Returns a Map of widgetId -> data for efficient lookup
 */
export function useWidgetData(dashboard: Dashboard | undefined) {
  return useQuery({
    queryKey: ['widget-data', dashboard?.widgets],
    queryFn: async () => {
      if (!dashboard) return new Map();

      const dataMap = new Map<string, unknown>();

      await Promise.all(
        dashboard.widgets.map(async (widget: Widget) => {
          try {
            let data;
            const params = widget.query_params || {};

            switch (widget.type) {
              case 'number':
                if (params.metric) {
                  data = await dashboardsApi.getMetricAggregate({
                    metric_type: params.metric,
                    entity_type: (params.entity_type as 'engineer' | 'team' | 'org') || 'org',
                    entity_id: params.entity_id,
                    time_range: params.time_range || '30d',
                    comparison_period: params.comparison_period,
                  });
                }
                break;

              case 'timeseries':
                if (params.metric) {
                  const now = new Date();
                  const daysAgo = parseInt(params.time_range?.replace('d', '') || '30');
                  const startDate = new Date(now.getTime() - daysAgo * MILLISECONDS_PER_DAY);

                  data = await dashboardsApi.getMetricTimeseries({
                    metric_type: params.metric,
                    entity_type: (params.entity_type as 'engineer' | 'team' | 'org') || 'org',
                    entity_id: params.entity_id,
                    start_date: startDate.toISOString(),
                    end_date: now.toISOString(),
                    granularity: 'day',
                  });

                  if (params.show_alerts === 'true') {
                    const alerts = await dashboardsApi.getAlertsTimeline({
                      entity_type: params.entity_type as 'engineer' | 'team' | 'org' | undefined,
                      entity_id: params.entity_id,
                      start_date: startDate.toISOString(),
                      end_date: now.toISOString(),
                    });
                    data = { ...data, alerts: alerts.alerts };
                  }
                }
                break;

              case 'bar':
                if (params.metric && params.entity_type !== 'org') {
                  data = await dashboardsApi.getMetricCompare({
                    metric_type: params.metric,
                    entity_type: params.entity_type as 'engineer' | 'team',
                    entity_ids: [],
                    time_range: params.time_range || '30d',
                  });
                }
                break;

              case 'table':
                data = await dashboardsApi.getEngineersPerformance({
                  team_id: params.entity_id,
                  time_range: params.time_range || '30d',
                  sort_by: params.sort_by,
                  sort_order: params.sort_order as 'asc' | 'desc' | undefined,
                  limit: params.limit ? parseInt(params.limit) : DEFAULT_TABLE_LIMIT,
                });
                break;

              case 'status':
                data = await dashboardsApi.getHealthStatus();
                break;

              case 'feed':
                const activityData = await api.getTodayActivity();
                data = { events: activityData.events };
                break;
            }

            if (data) {
              dataMap.set(widget.id, data);
            }
          } catch (error) {
            // Error is caught, data remains empty for widget
          }
        })
      );

      return dataMap;
    },
    enabled: !!dashboard,
    staleTime: QUERY_STALE_TIME_MS,
  });
}
