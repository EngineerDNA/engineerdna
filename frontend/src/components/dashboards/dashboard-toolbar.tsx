import { DashboardSelector } from './dashboard-selector';

interface Dashboard {
  id: string;
  name: string;
  description?: string;
  is_template: boolean;
}

interface DashboardToolbarProps {
  dashboards: Dashboard[];
  currentDashboard: Dashboard | undefined;
  currentDashboardId: string | null;
  isEditing: boolean;
  onDashboardChange: (dashboardId: string) => void;
  onToggleEdit: () => void;
  onShowAddWidget: () => void;
  onShowEditDashboard: () => void;
  onShowCreateDashboard: () => void;
  onCloneDashboard: (dashboardId: string) => void;
  onDeleteDashboard: (dashboardId: string) => void;
}

/**
 * Dashboard toolbar component providing dashboard management controls.
 * Includes dashboard selector, edit mode toggle, add widget button, and actions (clone, delete).
 */
export function DashboardToolbar({
  dashboards,
  currentDashboard,
  currentDashboardId,
  isEditing,
  onDashboardChange,
  onToggleEdit,
  onShowAddWidget,
  onShowEditDashboard,
  onShowCreateDashboard,
  onCloneDashboard,
  onDeleteDashboard,
}: DashboardToolbarProps) {
  return (
    <div className="mb-6">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Dashboards</h1>
        <div className="flex gap-3">
          <button
            onClick={onShowCreateDashboard}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 flex items-center gap-2"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 4v16m8-8H4"
              />
            </svg>
            Create Dashboard
          </button>
        </div>
      </div>

      <div className="flex items-center gap-4 mb-4">
        <DashboardSelector
          dashboards={dashboards}
          currentDashboardId={currentDashboardId}
          onDashboardChange={onDashboardChange}
        />

        {currentDashboard && (
          <>
            <button
              onClick={onToggleEdit}
              className={`px-4 py-2 rounded-lg flex items-center gap-2 ${
                isEditing
                  ? 'bg-green-600 text-white hover:bg-green-700 dark:bg-green-700 dark:hover:bg-green-800'
                  : 'bg-gray-200 text-gray-700 hover:bg-gray-300 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'
              }`}
            >
              {isEditing ? 'Done Editing' : 'Edit Layout'}
            </button>

            {isEditing && (
              <button
                onClick={onShowAddWidget}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 flex items-center gap-2"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 4v16m8-8H4"
                  />
                </svg>
                Add Widget
              </button>
            )}

            <button
              onClick={onShowEditDashboard}
              className="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
            >
              Edit Info
            </button>

            {currentDashboard.is_template ? (
              <button
                onClick={() => onCloneDashboard(currentDashboard.id)}
                className="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
              >
                Clone Template
              </button>
            ) : (
              <button
                onClick={() => {
                  if (confirm('Are you sure you want to delete this dashboard?')) {
                    onDeleteDashboard(currentDashboard.id);
                  }
                }}
                className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 dark:bg-red-700 dark:hover:bg-red-800"
              >
                Delete
              </button>
            )}
          </>
        )}
      </div>

      {currentDashboard?.description && (
        <p className="text-gray-600 dark:text-gray-400">{currentDashboard.description}</p>
      )}
    </div>
  );
}
