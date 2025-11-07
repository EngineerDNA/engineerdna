import { useState, useEffect, useMemo } from 'react';
import { Link } from 'react-router';
import type { ExportSchedule, Plugin } from '../api/types';

interface ScheduleFormProps {
  plugins: Plugin[];
  initialSchedule?: ExportSchedule;
  onSubmit: (data: {
    plugin_name: string;
    frequency: 'daily' | 'weekly' | 'monthly';
    day_of_week?: number;
    time_of_day: string;
    enabled: boolean;
  }) => void;
  onCancel: () => void;
  isSubmitting: boolean;
}

export function ScheduleForm({
  plugins,
  initialSchedule,
  onSubmit,
  onCancel,
  isSubmitting,
}: ScheduleFormProps) {
  const [pluginName, setPluginName] = useState(initialSchedule?.plugin_name || '');
  const [frequency, setFrequency] = useState<'daily' | 'weekly' | 'monthly'>(
    initialSchedule?.frequency || 'daily'
  );
  const [dayOfWeek, setDayOfWeek] = useState<number>(initialSchedule?.day_of_week ?? 1);
  const [timeOfDay, setTimeOfDay] = useState(initialSchedule?.time_of_day || '09:00');
  const [enabled, setEnabled] = useState(initialSchedule?.enabled ?? true);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const destinationPlugins = useMemo(
    () => plugins.filter((p) => p.type === 'destination' && p.enabled),
    [plugins]
  );

  useEffect(() => {
    if (!pluginName && destinationPlugins.length > 0) {
      setPluginName(destinationPlugins[0].name);
    }
  }, [pluginName, destinationPlugins]);

  const calculateNextRun = (): string | null => {
    if (!timeOfDay || !/^\d{2}:\d{2}$/.test(timeOfDay)) {
      return null;
    }

    const now = new Date();
    const [hours, minutes] = timeOfDay.split(':').map(Number);
    const next = new Date(now);
    next.setHours(hours, minutes, 0, 0);

    if (frequency === 'daily') {
      if (next <= now) {
        next.setDate(next.getDate() + 1);
      }
    } else if (frequency === 'weekly') {
      const currentDay = now.getDay();
      const daysUntilTarget = (dayOfWeek - currentDay + 7) % 7;
      next.setDate(next.getDate() + daysUntilTarget);
      if (next <= now) {
        next.setDate(next.getDate() + 7);
      }
    } else if (frequency === 'monthly') {
      next.setMonth(next.getMonth() + 1, 1);
      if (next <= now) {
        next.setMonth(next.getMonth() + 1, 1);
      }
    }

    return next.toLocaleString();
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!pluginName) {
      newErrors.plugin_name = 'Plugin is required';
    }

    if (!timeOfDay) {
      newErrors.time_of_day = 'Time is required';
    } else if (!/^\d{2}:\d{2}$/.test(timeOfDay)) {
      newErrors.time_of_day = 'Time must be in HH:MM format';
    }

    if (frequency === 'weekly' && (dayOfWeek < 0 || dayOfWeek > 6)) {
      newErrors.day_of_week = 'Invalid day of week';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) {
      return;
    }

    onSubmit({
      plugin_name: pluginName,
      frequency,
      day_of_week: frequency === 'weekly' ? dayOfWeek : undefined,
      time_of_day: timeOfDay,
      enabled,
    });
  };

  const nextRun = calculateNextRun();

  if (destinationPlugins.length === 0) {
    return (
      <div className="p-6 text-center">
        <div className="text-gray-500 dark:text-gray-400 mb-4">
          <svg className="mx-auto h-12 w-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
        </div>
        <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
          No destination plugins available
        </h3>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Please configure and enable a destination plugin before creating a schedule
        </p>
        <Link
          to="/plugins"
          className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:focus:ring-offset-gray-800"
        >
          <svg className="mr-2 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
        </Link>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div>
        <label
          htmlFor="plugin"
          className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
        >
          Destination Plugin
        </label>
        <select
          id="plugin"
          value={pluginName}
          onChange={(e) => setPluginName(e.target.value)}
          className="block w-full rounded-md border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm focus:border-blue-500 focus:ring-blue-500"
          disabled={isSubmitting}
        >
          {destinationPlugins.map((plugin) => (
            <option key={plugin.name} value={plugin.name}>
              {plugin.name}
            </option>
          ))}
        </select>
        {errors.plugin_name && (
          <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.plugin_name}</p>
        )}
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
          Frequency
        </label>
        <div className="space-y-2">
          {(['daily', 'weekly', 'monthly'] as const).map((freq) => (
            <label key={freq} className="flex items-center cursor-pointer">
              <input
                type="radio"
                value={freq}
                checked={frequency === freq}
                onChange={(e) => setFrequency(e.target.value as typeof frequency)}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 dark:border-gray-600"
                disabled={isSubmitting}
              />
              <span className="ml-2 text-sm text-gray-900 dark:text-gray-100 capitalize">
                {freq}
              </span>
            </label>
          ))}
        </div>
      </div>

      {frequency === 'weekly' && (
        <div>
          <label
            htmlFor="dayOfWeek"
            className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
          >
            Day of Week
          </label>
          <select
            id="dayOfWeek"
            value={dayOfWeek}
            onChange={(e) => setDayOfWeek(Number(e.target.value))}
            className="block w-full rounded-md border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm focus:border-blue-500 focus:ring-blue-500"
            disabled={isSubmitting}
          >
            <option value={0}>Sunday</option>
            <option value={1}>Monday</option>
            <option value={2}>Tuesday</option>
            <option value={3}>Wednesday</option>
            <option value={4}>Thursday</option>
            <option value={5}>Friday</option>
            <option value={6}>Saturday</option>
          </select>
          {errors.day_of_week && (
            <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.day_of_week}</p>
          )}
        </div>
      )}

      <div>
        <label
          htmlFor="timeOfDay"
          className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
        >
          Time of Day
        </label>
        <input
          id="timeOfDay"
          type="time"
          value={timeOfDay}
          onChange={(e) => setTimeOfDay(e.target.value)}
          className="block w-full rounded-md border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm focus:border-blue-500 focus:ring-blue-500"
          disabled={isSubmitting}
        />
        {errors.time_of_day && (
          <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.time_of_day}</p>
        )}
      </div>

      <div className="flex items-center">
        <input
          id="enabled"
          type="checkbox"
          checked={enabled}
          onChange={(e) => setEnabled(e.target.checked)}
          className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 dark:border-gray-600 rounded"
          disabled={isSubmitting}
        />
        <label htmlFor="enabled" className="ml-2 text-sm text-gray-900 dark:text-gray-100">
          Enable schedule
        </label>
      </div>

      {nextRun && (
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg
                className="h-5 w-5 text-blue-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-blue-800 dark:text-blue-300">
                Next run scheduled
              </h3>
              <div className="mt-1 text-sm text-blue-700 dark:text-blue-400">{nextRun}</div>
            </div>
          </div>
        </div>
      )}

      <div className="flex justify-end gap-3">
        <button
          type="button"
          onClick={onCancel}
          disabled={isSubmitting}
          className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isSubmitting}
          className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isSubmitting ? 'Saving...' : initialSchedule ? 'Update Schedule' : 'Create Schedule'}
        </button>
      </div>
    </form>
  );
}
