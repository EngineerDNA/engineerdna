import { Link, Outlet, useLocation, Navigate } from 'react-router-dom';

export function SettingsPage() {
  const location = useLocation();
  const currentTab = location.pathname.split('/')[2] || 'profile';

  const tabs = [
    { id: 'profile', label: 'Profile', path: '/settings/profile' },
    { id: 'dashboards', label: 'Dashboards', path: '/settings/dashboards' },
    { id: 'plugins', label: 'Plugins', path: '/settings/plugins' },
    { id: 'alerts', label: 'Alerts', path: '/settings/alerts' },
    { id: 'schedules', label: 'Schedules', path: '/settings/schedules' },
    { id: 'notifications', label: 'Notifications', path: '/settings/notifications' },
    { id: 'system', label: 'System', path: '/settings/system' },
  ];

  // Redirect /settings to /settings/profile
  if (location.pathname === '/settings') {
    return <Navigate to="/settings/profile" replace />;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Settings</h1>
        <p className="mt-2 text-gray-600 dark:text-gray-400">
          Configure your preferences and system settings
        </p>
      </div>

      {/* Tab Navigation - Desktop */}
      <div className="hidden sm:block border-b border-gray-200 dark:border-gray-700">
        <nav className="flex space-x-8" aria-label="Settings tabs">
          {tabs.map((tab) => {
            const isActive = currentTab === tab.id;
            return (
              <Link
                key={tab.id}
                to={tab.path}
                className={`
                  py-4 px-1 border-b-2 font-medium text-sm whitespace-nowrap
                  ${
                    isActive
                      ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300 dark:hover:border-gray-600'
                  }
                `}
              >
                {tab.label}
              </Link>
            );
          })}
        </nav>
      </div>

      {/* Tab Navigation - Mobile */}
      <div className="sm:hidden">
        <select
          value={currentTab}
          onChange={(e) => {
            const tab = tabs.find((t) => t.id === e.target.value);
            if (tab) {
              window.location.href = tab.path;
            }
          }}
          className="block w-full rounded-md border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        >
          {tabs.map((tab) => (
            <option key={tab.id} value={tab.id}>
              {tab.label}
            </option>
          ))}
        </select>
      </div>

      {/* Tab Content */}
      <div className="mt-6">
        <Outlet />
      </div>
    </div>
  );
}
