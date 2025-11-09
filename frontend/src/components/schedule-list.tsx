import { memo, useState } from 'react';
import type { ExportSchedule } from '../api/types';
import { useDeleteSchedule, useTriggerSchedule, useUpdateSchedule } from '../hooks/useSchedules';

interface ScheduleListProps {
  schedules: ExportSchedule[];
  onEdit: (schedule: ExportSchedule) => void;
}

export const ScheduleList = memo(function ScheduleList({ schedules, onEdit }: ScheduleListProps) {
  const deleteSchedule = useDeleteSchedule();
  const triggerSchedule = useTriggerSchedule();
  const updateSchedule = useUpdateSchedule();
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);

  const handleDelete = (id: string) => {
    if (deleteConfirm === id) {
      deleteSchedule.mutate(id);
      setDeleteConfirm(null);
    } else {
      setDeleteConfirm(id);
      setTimeout(() => setDeleteConfirm(null), 3000);
    }
  };

  const handleToggle = (schedule: ExportSchedule) => {
    updateSchedule.mutate({
      id: schedule.id,
      req: { enabled: !schedule.enabled },
    });
  };

  const handleRunNow = (id: string) => {
    triggerSchedule.mutate(id);
  };

  const formatNextRun = (nextRun: string): string => {
    const date = new Date(nextRun);
    const now = new Date();
    const diff = date.getTime() - now.getTime();
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const days = Math.floor(hours / 24);

    if (days > 1) {
      return `in ${days} days`;
    }
    if (hours > 0) {
      return `in ${hours} hours`;
    }
    if (diff > 0) {
      return `in ${Math.floor(diff / (1000 * 60))} minutes`;
    }
    return 'overdue';
  };

  const formatFrequency = (schedule: ExportSchedule): string => {
    if (schedule.frequency === 'weekly' && schedule.day_of_week !== undefined) {
      const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
      return `Weekly on ${days[schedule.day_of_week]}`;
    }
    if (!schedule.frequency) return 'Unknown';
    return schedule.frequency.charAt(0).toUpperCase() + schedule.frequency.slice(1);
  };

  if (schedules.length === 0) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-12 text-center">
        <div className="text-gray-500 dark:text-gray-400 mb-4">
          <svg className="mx-auto h-12 w-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
            />
          </svg>
        </div>
        <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
          No schedules configured
        </h3>
        <p className="text-gray-600 dark:text-gray-400">
          Create a schedule to automatically export data on a regular basis
        </p>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-900">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Plugin
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Frequency
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Time
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Next Run
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Status
              </th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            {schedules.map((schedule) => (
              <tr key={schedule.id} className={!schedule.enabled ? 'opacity-50' : ''}>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="text-sm font-medium text-gray-900 dark:text-gray-100">
                    {schedule.plugin_name}
                  </div>
                  {schedule.last_run && (
                    <div className="text-xs text-gray-500 dark:text-gray-400">
                      Last run: {new Date(schedule.last_run).toLocaleString()}
                    </div>
                  )}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
                  {formatFrequency(schedule)}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
                  {schedule.time_of_day}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
                  {formatNextRun(schedule.next_run)}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <button
                    onClick={() => handleToggle(schedule)}
                    className={`px-2 py-1 rounded text-xs font-medium transition-colors ${
                      schedule.enabled
                        ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200 hover:bg-green-200 dark:hover:bg-green-800'
                        : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
                    }`}
                  >
                    {schedule.enabled ? 'Enabled' : 'Disabled'}
                  </button>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <div className="flex items-center justify-end gap-2">
                    <button
                      onClick={() => handleRunNow(schedule.id)}
                      disabled={!schedule.enabled}
                      className="text-blue-600 dark:text-blue-400 hover:text-blue-900 dark:hover:text-blue-300 disabled:opacity-50 disabled:cursor-not-allowed"
                      title="Run now"
                    >
                      <svg
                        className="w-5 h-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                        />
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                        />
                      </svg>
                    </button>
                    <button
                      onClick={() => onEdit(schedule)}
                      className="text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-300"
                      title="Edit"
                    >
                      <svg
                        className="w-5 h-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                        />
                      </svg>
                    </button>
                    <button
                      onClick={() => handleDelete(schedule.id)}
                      className={`${
                        deleteConfirm === schedule.id
                          ? 'text-red-600 dark:text-red-400'
                          : 'text-gray-600 dark:text-gray-400 hover:text-red-600 dark:hover:text-red-400'
                      }`}
                      title={deleteConfirm === schedule.id ? 'Click again to confirm' : 'Delete'}
                    >
                      <svg
                        className="w-5 h-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                        />
                      </svg>
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
});
