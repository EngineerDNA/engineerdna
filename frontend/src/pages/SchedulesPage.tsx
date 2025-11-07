import { useState, useMemo } from 'react';
import { useSchedules } from '../hooks/useSchedules';
import { usePlugins } from '../hooks/usePlugins';
import { ScheduleList } from '../components/schedule-list';
import { ScheduleModal } from '../components/schedule-modal';
import type { ExportSchedule } from '../api/types';

export function SchedulesPage() {
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
      <div className="flex items-center justify-center h-96">
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
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Export Schedules</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">
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
            Markdown, or PDF export). Please configure and enable a destination plugin to get
            started.
          </p>
          <a
            href="/plugins"
            className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:focus:ring-offset-gray-900 transition-colors"
          >
            <svg className="mr-2 h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
              />
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
              />
            </svg>
            Go to Plugins
          </a>
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
