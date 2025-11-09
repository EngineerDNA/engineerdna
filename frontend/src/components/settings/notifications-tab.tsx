export function NotificationsTab() {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
          Notification Preferences
        </h2>
        <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
          Configure how and when you receive notifications
        </p>
      </div>

      {/* Email Notifications */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h3 className="text-base font-semibold text-gray-900 dark:text-gray-100 mb-4">
          Email Notifications
        </h3>
        <div className="space-y-4">
          <label className="flex items-center">
            <input
              type="checkbox"
              className="w-4 h-4 text-blue-600 border-gray-300 dark:border-gray-600 rounded focus:ring-blue-500 dark:focus:ring-blue-400"
            />
            <span className="ml-2 text-sm text-gray-900 dark:text-gray-100">
              Alert notifications
            </span>
          </label>
          <label className="flex items-center">
            <input
              type="checkbox"
              className="w-4 h-4 text-blue-600 border-gray-300 dark:border-gray-600 rounded focus:ring-blue-500 dark:focus:ring-blue-400"
            />
            <span className="ml-2 text-sm text-gray-900 dark:text-gray-100">
              Weekly briefing summaries
            </span>
          </label>
          <label className="flex items-center">
            <input
              type="checkbox"
              className="w-4 h-4 text-blue-600 border-gray-300 dark:border-gray-600 rounded focus:ring-blue-500 dark:focus:ring-blue-400"
            />
            <span className="ml-2 text-sm text-gray-900 dark:text-gray-100">
              Export completion notifications
            </span>
          </label>
        </div>
      </div>

      {/* Browser Notifications */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h3 className="text-base font-semibold text-gray-900 dark:text-gray-100 mb-4">
          Browser Notifications
        </h3>
        <div className="space-y-4">
          <label className="flex items-center">
            <input
              type="checkbox"
              className="w-4 h-4 text-blue-600 border-gray-300 dark:border-gray-600 rounded focus:ring-blue-500 dark:focus:ring-blue-400"
            />
            <span className="ml-2 text-sm text-gray-900 dark:text-gray-100">
              Enable browser notifications
            </span>
          </label>
          <label className="flex items-center">
            <input
              type="checkbox"
              className="w-4 h-4 text-blue-600 border-gray-300 dark:border-gray-600 rounded focus:ring-blue-500 dark:focus:ring-blue-400"
            />
            <span className="ml-2 text-sm text-gray-900 dark:text-gray-100">
              Critical alerts only
            </span>
          </label>
        </div>
      </div>

      {/* Quiet Hours */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h3 className="text-base font-semibold text-gray-900 dark:text-gray-100 mb-4">
          Quiet Hours
        </h3>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          Suppress non-critical notifications during specified hours
        </p>
        <div className="space-y-4">
          <label className="flex items-center">
            <input
              type="checkbox"
              className="w-4 h-4 text-blue-600 border-gray-300 dark:border-gray-600 rounded focus:ring-blue-500 dark:focus:ring-blue-400"
            />
            <span className="ml-2 text-sm text-gray-900 dark:text-gray-100">
              Enable quiet hours
            </span>
          </label>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Start Time
              </label>
              <input
                type="time"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400"
                defaultValue="22:00"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                End Time
              </label>
              <input
                type="time"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400"
                defaultValue="08:00"
              />
            </div>
          </div>
        </div>
      </div>

      {/* Channel Preferences */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h3 className="text-base font-semibold text-gray-900 dark:text-gray-100 mb-4">
          Channel Preferences
        </h3>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          Choose which channels to use for different types of notifications
        </p>
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-900 dark:text-gray-100">Critical Alerts</span>
            <div className="flex gap-2">
              <span className="px-2 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300 text-xs rounded">
                Email
              </span>
              <span className="px-2 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300 text-xs rounded">
                Browser
              </span>
            </div>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-900 dark:text-gray-100">Warning Alerts</span>
            <div className="flex gap-2">
              <span className="px-2 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300 text-xs rounded">
                Browser
              </span>
            </div>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-900 dark:text-gray-100">Info Alerts</span>
            <div className="flex gap-2">
              <span className="px-2 py-1 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 text-xs rounded">
                None
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Save Button */}
      <div className="flex justify-end">
        <button className="px-6 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-800 transition-colors">
          Save Preferences
        </button>
      </div>
    </div>
  );
}
