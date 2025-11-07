interface Dashboard {
  id: string;
  name: string;
  is_template: boolean;
}

interface DashboardSelectorProps {
  dashboards: Dashboard[];
  currentDashboardId: string | null;
  onDashboardChange: (dashboardId: string) => void;
}

/**
 * Dashboard selector dropdown component.
 * Groups dashboards into "My Dashboards" and "Templates" optgroups.
 */
export function DashboardSelector({
  dashboards,
  currentDashboardId,
  onDashboardChange,
}: DashboardSelectorProps) {
  const templates = dashboards.filter((d) => d.is_template);
  const userDashboards = dashboards.filter((d) => !d.is_template);

  return (
    <select
      value={currentDashboardId || ''}
      onChange={(e) => onDashboardChange(e.target.value)}
      className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500"
    >
      {userDashboards.length > 0 && (
        <optgroup label="My Dashboards">
          {userDashboards.map((d) => (
            <option key={d.id} value={d.id}>
              {d.name}
            </option>
          ))}
        </optgroup>
      )}
      {templates.length > 0 && (
        <optgroup label="Templates">
          {templates.map((d) => (
            <option key={d.id} value={d.id}>
              {d.name}
            </option>
          ))}
        </optgroup>
      )}
    </select>
  );
}
