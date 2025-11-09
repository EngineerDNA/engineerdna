import { useState } from 'react';
import { useRole, useUpdateRole } from '../../hooks/useSettings';
import type { UserRole } from '../../api/types';

const ROLE_OPTIONS: Array<{
  value: UserRole;
  name: string;
  description: string;
  dashboards: string[];
  features: string[];
}> = [
  {
    value: 'ic',
    name: 'Individual Contributor',
    description: 'Personal performance tracking and activity monitoring',
    dashboards: ['Personal Performance', 'My Activity'],
    features: ['Personal metrics', 'Activity tracking', 'Goal progress', 'Skills development'],
  },
  {
    value: 'manager',
    name: 'Team Lead / Engineering Manager',
    description: 'Team management, briefings, and sprint planning',
    dashboards: [
      'Command Center',
      'Weekly Briefing',
      'Team Performance',
      'Sprint Planning',
      'Team Management',
    ],
    features: [
      'Team briefings',
      'Sprint planning',
      'Team performance',
      'Identity resolution',
      'Manager notes',
    ],
  },
  {
    value: 'director',
    name: 'Director / VP of Engineering',
    description: 'Organization-wide metrics, cost analysis, and capacity planning',
    dashboards: [
      'Executive Overview',
      'Organization Scorecard',
      'Team Comparison',
      'Cost & ROI Analysis',
      'Capacity Planning',
    ],
    features: [
      'Org scorecard',
      'Multi-team comparison',
      'Cost & ROI tracking',
      'Capacity planning',
      'Forecasting',
    ],
  },
  {
    value: 'admin',
    name: 'Admin / Platform Owner',
    description: 'System administration and data management',
    dashboards: ['System Health', 'Data Management', 'Plus all other role dashboards'],
    features: [
      'System health',
      'Plugin management',
      'Data cleanup',
      'Advanced settings',
      'All role features',
    ],
  },
];

export function ProfileTab() {
  const { data: roleData, isLoading } = useRole();
  const updateRoleMutation = useUpdateRole();
  const [selectedRole, setSelectedRole] = useState<UserRole | null>(null);
  const [showComparison, setShowComparison] = useState(false);

  const currentRole = selectedRole || (roleData?.role as UserRole) || 'manager';
  const currentRoleInfo = ROLE_OPTIONS.find((r) => r.value === currentRole);

  const handleSave = () => {
    if (selectedRole && selectedRole !== roleData?.role) {
      updateRoleMutation.mutate(selectedRole);
    }
  };

  const hasChanges = selectedRole && selectedRole !== roleData?.role;

  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 animate-pulse">
        <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-32 mb-4"></div>
        <div className="space-y-3">
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Success Message */}
      {updateRoleMutation.isSuccess && (
        <div className="p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
          <p className="text-green-800 dark:text-green-200">
            Role updated successfully. New dashboards have been created for you.
          </p>
        </div>
      )}

      {/* Error Message */}
      {updateRoleMutation.error && (
        <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
          <p className="text-red-800 dark:text-red-200">
            Failed to update role: {(updateRoleMutation.error as Error).message}
          </p>
        </div>
      )}

      {/* Current Role */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Your Role</h2>

        <div className="space-y-4">
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">Current Role</p>
              <p className="text-lg font-medium text-gray-900 dark:text-gray-100">
                {currentRoleInfo?.name}
              </p>
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                {currentRoleInfo?.description}
              </p>
            </div>
          </div>

          <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
            <p className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Dashboards for this role:
            </p>
            <ul className="list-disc list-inside space-y-1">
              {currentRoleInfo?.dashboards.map((dashboard) => (
                <li key={dashboard} className="text-sm text-gray-600 dark:text-gray-400">
                  {dashboard}
                </li>
              ))}
            </ul>
          </div>

          <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
            <p className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Features:</p>
            <div className="flex flex-wrap gap-2">
              {currentRoleInfo?.features.map((feature) => (
                <span
                  key={feature}
                  className="px-3 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300 text-sm rounded-full"
                >
                  {feature}
                </span>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Role Selector */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Change Role</h2>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Select New Role
            </label>
            <select
              value={currentRole}
              onChange={(e) => setSelectedRole(e.target.value as UserRole)}
              className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              {ROLE_OPTIONS.map((role) => (
                <option key={role.value} value={role.value}>
                  {role.name}
                </option>
              ))}
            </select>
          </div>

          <button
            onClick={() => setShowComparison(true)}
            className="text-blue-600 dark:text-blue-400 hover:underline text-sm"
          >
            Compare All Roles
          </button>

          {hasChanges && (
            <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
              <p className="text-yellow-800 dark:text-yellow-200 text-sm">
                Changing your role will create {currentRoleInfo?.dashboards.length || 0} new
                dashboards for you. Your existing dashboards will remain unchanged.
              </p>
            </div>
          )}

          <div className="flex gap-2">
            <button
              onClick={handleSave}
              disabled={!hasChanges || updateRoleMutation.isPending}
              className="px-6 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {updateRoleMutation.isPending ? 'Saving...' : 'Save Changes'}
            </button>
            {hasChanges && (
              <button
                onClick={() => setSelectedRole(null)}
                className="px-6 py-2 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
              >
                Cancel
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Role Comparison Modal */}
      {showComparison && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                  Role Comparison
                </h2>
                <button
                  onClick={() => setShowComparison(false)}
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

              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                  <thead className="bg-gray-50 dark:bg-gray-900">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        Feature
                      </th>
                      {ROLE_OPTIONS.map((role) => (
                        <th
                          key={role.value}
                          className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
                        >
                          {role.value.toUpperCase()}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                    <tr>
                      <td className="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">
                        Dashboard Count
                      </td>
                      {ROLE_OPTIONS.map((role) => (
                        <td
                          key={role.value}
                          className="px-6 py-4 text-sm text-gray-600 dark:text-gray-400"
                        >
                          {role.dashboards.length}
                        </td>
                      ))}
                    </tr>
                    {Array.from(new Set(ROLE_OPTIONS.flatMap((r) => r.features))).map((feature) => (
                      <tr key={feature}>
                        <td className="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">
                          {feature}
                        </td>
                        {ROLE_OPTIONS.map((role) => (
                          <td key={role.value} className="px-6 py-4 text-sm text-center">
                            {role.features.includes(feature) || role.value === 'admin' ? (
                              <svg
                                className="w-5 h-5 text-green-500 dark:text-green-400 mx-auto"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                              >
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth={2}
                                  d="M5 13l4 4L19 7"
                                />
                              </svg>
                            ) : (
                              <span className="text-gray-300 dark:text-gray-600">-</span>
                            )}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
