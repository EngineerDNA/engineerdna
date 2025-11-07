import { useEffect, useRef } from 'react';
import type { ExportSchedule, Plugin } from '../api/types';
import { ScheduleForm } from './schedule-form';
import { useCreateSchedule, useUpdateSchedule } from '../hooks/useSchedules';

interface ScheduleModalProps {
  isOpen: boolean;
  onClose: () => void;
  plugins: Plugin[];
  schedule?: ExportSchedule;
  preSelectedPlugin?: string;
}

export function ScheduleModal({
  isOpen,
  onClose,
  plugins,
  schedule,
  preSelectedPlugin,
}: ScheduleModalProps) {
  const modalRef = useRef<HTMLDivElement>(null);
  const createSchedule = useCreateSchedule();
  const updateSchedule = useUpdateSchedule();

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = 'unset';
    };
  }, [isOpen, onClose]);

  useEffect(() => {
    if (createSchedule.isSuccess || updateSchedule.isSuccess) {
      onClose();
      createSchedule.reset();
      updateSchedule.reset();
    }
  }, [createSchedule.isSuccess, updateSchedule.isSuccess, onClose, createSchedule, updateSchedule]);

  const handleSubmit = (data: {
    plugin_name: string;
    frequency: 'daily' | 'weekly' | 'monthly';
    day_of_week?: number;
    time_of_day: string;
    enabled: boolean;
  }) => {
    if (schedule) {
      updateSchedule.mutate({ id: schedule.id, req: data });
    } else {
      createSchedule.mutate(data);
    }
  };

  if (!isOpen) {
    return null;
  }

  const initialScheduleWithPlugin = schedule
    ? schedule
    : preSelectedPlugin
      ? ({
          plugin_name: preSelectedPlugin,
          frequency: 'daily',
          time_of_day: '09:00',
          enabled: true,
        } as ExportSchedule)
      : undefined;

  return (
    <div
      className="fixed inset-0 z-50 overflow-y-auto"
      aria-labelledby="modal-title"
      role="dialog"
      aria-modal="true"
    >
      <div className="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <div
          className="fixed inset-0 bg-gray-500 dark:bg-gray-900 bg-opacity-75 dark:bg-opacity-85 transition-opacity"
          aria-hidden="true"
          onClick={onClose}
        />

        <span className="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">
          &#8203;
        </span>

        <div
          ref={modalRef}
          className="inline-block align-bottom bg-white dark:bg-gray-800 rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full"
        >
          <div className="bg-white dark:bg-gray-800 px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
            <div className="sm:flex sm:items-start">
              <div className="mt-3 text-center sm:mt-0 sm:text-left w-full">
                <h3
                  className="text-lg leading-6 font-medium text-gray-900 dark:text-gray-100 mb-4"
                  id="modal-title"
                >
                  {schedule ? 'Edit Schedule' : 'Create Schedule'}
                </h3>
                <ScheduleForm
                  plugins={plugins}
                  initialSchedule={initialScheduleWithPlugin}
                  onSubmit={handleSubmit}
                  onCancel={onClose}
                  isSubmitting={createSchedule.isPending || updateSchedule.isPending}
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
