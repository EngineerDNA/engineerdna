import { Responsive, WidthProvider } from 'react-grid-layout';
import type { Layout } from 'react-grid-layout';
import { WidgetContainer } from './widget-container';
import { NumberCard } from '../widgets/number-card';
import { TimeseriesChart } from '../widgets/timeseries-chart';
import { BarChart } from '../widgets/bar-chart';
import { Table } from '../widgets/table';
import { StatusIndicator } from '../widgets/status-indicator';
import { ActivityFeed } from '../widgets/activity-feed';
import type { Widget, HealthStatusResponse } from '../../api/types';
import {
  GRID_BREAKPOINTS,
  GRID_COLUMNS,
  GRID_ROW_HEIGHT,
  DEFAULT_TABLE_LIMIT,
} from '../../constants';

const ResponsiveGridLayout = WidthProvider(Responsive);

interface DashboardGridProps {
  widgets: Widget[];
  widgetData: Map<string, unknown>;
  isEditing?: boolean;
  onLayoutChange?: (layout: Layout[]) => void;
  onWidgetEdit?: (widgetId: string) => void;
  onWidgetDelete?: (widgetId: string) => void;
}

export function DashboardGrid({
  widgets,
  widgetData,
  isEditing = false,
  onLayoutChange,
  onWidgetEdit,
  onWidgetDelete,
}: DashboardGridProps) {
  const layouts = {
    lg: widgets.map((w) => ({
      i: w.id,
      x: w.position.x,
      y: w.position.y,
      w: w.position.w,
      h: w.position.h,
    })),
  };

  const handleLayoutChange = (currentLayout: Layout[]) => {
    if (onLayoutChange && isEditing) {
      onLayoutChange(currentLayout);
    }
  };

  const renderWidget = (widget: Widget) => {
    const data = widgetData.get(widget.id);

    switch (widget.type) {
      case 'number': {
        const numberData = data as {
          value?: number;
          previous?: number;
          change_pct?: number;
          change?: number;
        };
        const changeDirection =
          numberData?.change === undefined || numberData?.change === 0
            ? 'stable'
            : numberData.change > 0
              ? 'up'
              : 'down';
        return (
          <NumberCard
            title={widget.title}
            value={numberData?.value ?? 0}
            previousValue={numberData?.previous}
            changePercentage={numberData?.change_pct}
            changeDirection={changeDirection}
            unit={(data as { unit?: string })?.unit}
            isLoading={!data}
          />
        );
      }
      case 'timeseries': {
        const timeseriesData = data as {
          data_points?: Array<{ period_start: string; period_end: string; value: number }>;
        };
        const transformedData =
          timeseriesData?.data_points?.map((point) => ({
            timestamp: point.period_start,
            value: point.value,
          })) ?? [];
        const dataRecord = data as Record<string, unknown>;
        return (
          <TimeseriesChart
            data={transformedData}
            unit={(data as { unit?: string })?.unit}
            alerts={
              Array.isArray(dataRecord?.alerts)
                ? (dataRecord.alerts as unknown as import('../../api/metric-types').AlertMarker[])
                : []
            }
            isLoading={!data}
          />
        );
      }
      case 'bar':
        return (
          <BarChart
            data={
              (data as { entities?: Array<{ entity_name: string; value: number }> })?.entities ?? []
            }
            unit={(data as { unit?: string })?.unit}
            isLoading={!data}
          />
        );
      case 'table': {
        const dataRecord = data as Record<string, unknown>;
        return (
          <Table
            data={
              Array.isArray(dataRecord?.engineers)
                ? (dataRecord.engineers as unknown as import('../../api/metric-types').EngineerPerformanceRow[])
                : []
            }
            isLoading={!data}
          />
        );
      }
      case 'status':
        return <StatusIndicator status={data as HealthStatusResponse} isLoading={!data} />;
      case 'feed': {
        const dataRecord = data as Record<string, unknown>;
        return (
          <ActivityFeed
            events={
              Array.isArray(dataRecord?.events)
                ? (dataRecord.events as unknown as import('../../api/types').Event[])
                : []
            }
            isLoading={!data}
            limit={
              widget.query_params?.limit ? parseInt(widget.query_params.limit) : DEFAULT_TABLE_LIMIT
            }
          />
        );
      }
      default:
        return <div className="p-4 text-gray-500">Unknown widget type: {widget.type}</div>;
    }
  };

  if (widgets.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 bg-white dark:bg-gray-800 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg">
        <div className="text-center">
          <svg
            className="mx-auto h-12 w-12 text-gray-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-gray-100">No widgets</h3>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Get started by adding a widget to your dashboard
          </p>
        </div>
      </div>
    );
  }

  return (
    <ResponsiveGridLayout
      className="layout"
      layouts={layouts}
      breakpoints={GRID_BREAKPOINTS}
      cols={GRID_COLUMNS}
      rowHeight={GRID_ROW_HEIGHT}
      isDraggable={isEditing}
      isResizable={isEditing}
      onLayoutChange={handleLayoutChange}
      draggableHandle=".widget-drag-handle"
    >
      {widgets.map((widget) => (
        <div key={widget.id}>
          <WidgetContainer
            title={widget.title}
            onEdit={isEditing ? () => onWidgetEdit?.(widget.id) : undefined}
            onDelete={isEditing ? () => onWidgetDelete?.(widget.id) : undefined}
          >
            {renderWidget(widget)}
          </WidgetContainer>
        </div>
      ))}
    </ResponsiveGridLayout>
  );
}
