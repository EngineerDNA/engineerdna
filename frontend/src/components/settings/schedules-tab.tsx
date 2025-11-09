import { useState, useMemo } from 'react';
import { useSchedules } from '../../hooks/useSchedules';
import { usePlugins } from '../../hooks/usePlugins';
import { ScheduleList } from '../schedule-list';
import { ScheduleModal } from '../schedule-modal';
import type { ExportSchedule } from '../../api/types';

export function SchedulesTab() {
  const {
    data: schedulesData,
    isLoading: schedulesLoading,
    error: schedulesError,
  } = useSchedules();
  const { data: pluginsData, isLoading: pluginsLoading } = usePlugins();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingSchedule, setEditingSchedule] = useState<ExportSchedule | undefined>(undefined);

  const plugins = pluginsData?.plugins || [];
  const hasEnabledDestinationPlugins = useMemo(
    () => plugins.some((p) => p.type === 'destination' && p.enabled),
    [plugins]
  );

  const handleCreateNew = () => {
    setEditingSchedule(undefined);
    setIsModalOpen(true);
  };

  const handleEdit = (schedule: ExportSchedule) => {
    setEditingSchedule(schedule);
    setIsModalOpen(true);
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
    setEditingSchedule(undefined);
  };

  if (schedulesLoading || pluginsLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
      </div>
    );
  }

  if (schedulesError) {
    return (
      <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
        <div className="flex">
          <div className="flex-shrink-0">
            <svg
              className="h-5 w-5 text-red-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </div>
          <div className="ml-3">
            <h3 className="text-sm font-medium text-red-800 dark:text-red-300">
              Error loading schedules
            </h3>
            <div className="mt-2 text-sm text-red-700 dark:text-red-400">
              {schedulesError instanceof Error ? schedulesError.message : 'Unknown error'}
            </div>
          </div>
        </div>
      </div>
    );
  }

  const schedules = schedulesData?.schedules || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
            Export Schedules
          </h2>
          <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
            Automatically export data to external systems on a schedule
          </p>
        </div>
        <button
          onClick={handleCreateNew}
          disabled={!hasEnabledDestinationPlugins}
          title={
            hasEnabledDestinationPlugins
              ? 'Create a new export schedule'
              : 'Please enable a destination plugin first'
          }
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-900 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-blue-600 dark:disabled:hover:bg-blue-700"
        >
          Create Schedule
        </button>
      </div>

      {!hasEnabledDestinationPlugins && schedules.length === 0 && (
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-8 text-center">
          <div className="text-blue-500 dark:text-blue-400 mb-4">
            <svg
              className="mx-auto h-16 w-16"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
          </div>
          <h3 className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-2">
            No Destination Plugins Enabled
          </h3>
          <p className="text-gray-600 dark:text-gray-400 mb-6 max-w-md mx-auto">
            Export schedules require at least one enabled destination plugin (like Google Sheets,
            Markdown, or PDF export). Please configure and enable a destination plugin in the
            Plugins tab to get started.
          </p>
        </div>
      )}

      <ScheduleList schedules={schedules} onEdit={handleEdit} />

      <ScheduleModal
        isOpen={isModalOpen}
        onClose={handleCloseModal}
        plugins={plugins}
        schedule={editingSchedule}
      />
    </div>
  );
}
