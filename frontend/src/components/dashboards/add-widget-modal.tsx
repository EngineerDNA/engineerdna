import { useState } from 'react';
import type { Widget } from '../../api/types';
import { WIDGET_DEFAULTS } from '../../constants';

interface AddWidgetModalProps {
  isOpen: boolean;
  onClose: () => void;
  onAdd: (widget: Omit<Widget, 'id'>) => void;
}

export function AddWidgetModal({ isOpen, onClose, onAdd }: AddWidgetModalProps) {
  const [widgetType, setWidgetType] = useState<Widget['type']>('number');
  const [title, setTitle] = useState('');
  const [metricType, setMetricType] = useState('prs_merged');
  const [entityType, setEntityType] = useState<'engineer' | 'team' | 'org'>('org');
  const [timeRange, setTimeRange] = useState('7d');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const newWidget: Omit<Widget, 'id'> = {
      type: widgetType,
      title: title || `New ${widgetType}`,
      data_source: `/metrics/${widgetType}`,
      query_params: {
        metric_type: metricType,
        entity_type: entityType,
        time_range: timeRange,
      },
      position: {
        x: 0,
        y: 0,
        w: WIDGET_DEFAULTS[widgetType]?.w || 6,
        h: WIDGET_DEFAULTS[widgetType]?.h || 3,
      },
    };

    onAdd(newWidget);
    onClose();
    setTitle('');
  };

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-[9999] flex items-center justify-center bg-black/50"
      onClick={(e) => {
        if (e.target === e.currentTarget) {
          onClose();
        }
      }}
    >
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full mx-4">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">Add Widget</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Widget Type
            </label>
            <select
              value={widgetType}
              onChange={(e) => setWidgetType(e.target.value as Widget['type'])}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500"
            >
              <option value="number">Number Card</option>
              <option value="timeseries">Timeseries Chart</option>
              <option value="bar">Bar Chart</option>
              <option value="table">Table</option>
              <option value="status">Status Indicator</option>
              <option value="feed">Activity Feed</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Title
            </label>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder={`New ${widgetType}`}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500"
            />
          </div>

          {widgetType !== 'feed' && widgetType !== 'status' && (
            <>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Metric Type
                </label>
                <select
                  value={metricType}
                  onChange={(e) => setMetricType(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500"
                >
                  <option value="prs_merged">PRs Merged</option>
                  <option value="cycle_time">Cycle Time</option>
                  <option value="throughput_score">Throughput Score</option>
                  <option value="quality_score">Quality Score</option>
                  <option value="total_score">Total Score</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Entity Type
                </label>
                <select
                  value={entityType}
                  onChange={(e) => setEntityType(e.target.value as 'engineer' | 'team' | 'org')}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500"
                >
                  <option value="org">Organization</option>
                  <option value="team">Team</option>
                  <option value="engineer">Engineer</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Time Range
                </label>
                <select
                  value={timeRange}
                  onChange={(e) => setTimeRange(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500"
                >
                  <option value="7d">Last 7 days</option>
                  <option value="30d">Last 30 days</option>
                  <option value="90d">Last 90 days</option>
                  <option value="1y">Last year</option>
                </select>
              </div>
            </>
          )}

          <div className="flex justify-end gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 text-white bg-blue-600 rounded-lg hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800"
            >
              Add Widget
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
