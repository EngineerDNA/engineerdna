import { useState } from 'react';
import { RoleComparisonTable } from './role-comparison-table';
import type { UserRole } from '../../api/types';

interface RoleSelectionStepProps {
  selectedRole: UserRole | null;
  onRoleSelect: (role: UserRole) => void;
  onNext: () => void;
  onBack: () => void;
}

interface RoleCardInfo {
  id: UserRole;
  name: string;
  description: string;
  dashboards: string[];
  features: string[];
  bestFor: string;
}

const roles: RoleCardInfo[] = [
  {
    id: 'ic',
    name: 'Individual Contributor',
    description: 'Focus: Personal performance, goals, activity',
    dashboards: ['Personal Performance', 'My Activity'],
    features: ['Goal tracking', 'Skill development tracking'],
    bestFor: 'Software engineers, designers, individual contributors tracking personal work',
  },
  {
    id: 'manager',
    name: 'Team Lead / Engineering Manager',
    description: 'Focus: Team health, 1-on-1 prep, sprint work',
    dashboards: [
      'Command Center',
      'Weekly Briefing',
      'Team Performance',
      'Sprint Planning',
      'Team Management',
    ],
    features: ['Team briefing', 'Alert management', 'Sprint tracking'],
    bestFor: 'Team leads, engineering managers managing 3-12 people',
  },
  {
    id: 'director',
    name: 'Director / VP of Engineering',
    description: 'Focus: Org health, team comparison, strategy',
    dashboards: [
      'Executive Overview',
      'Organization Scorecard',
      'Team Comparison',
      'Cost & ROI Analysis',
      'Capacity Planning',
    ],
    features: ['Org metrics', 'Cost analysis', 'Forecasting'],
    bestFor: 'Directors, VPs, CTOs managing multiple teams',
  },
  {
    id: 'admin',
    name: 'Admin / Platform Owner',
    description: 'Focus: System setup, data management',
    dashboards: ['All dashboards above', 'System Health', 'Data Management'],
    features: ['Full system access', 'Advanced configuration'],
    bestFor: 'Platform engineers, DevOps, system administrators',
  },
];

export function RoleSelectionStep({
  selectedRole,
  onRoleSelect,
  onNext,
  onBack,
}: RoleSelectionStepProps) {
  const [showComparison, setShowComparison] = useState(false);

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">
          Choose Your Role
        </h2>
        <p className="text-gray-600 dark:text-gray-400">
          Select the role that best describes your position. We'll create the right dashboards for
          you.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {roles.map((role) => (
          <button
            key={role.id}
            onClick={() => onRoleSelect(role.id)}
            className={`text-left p-5 rounded-lg border-2 transition-all ${
              selectedRole === role.id
                ? 'border-blue-600 dark:border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                : 'border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 hover:border-blue-400 dark:hover:border-blue-600'
            }`}
          >
            <div className="flex items-start justify-between mb-3">
              <div className="flex-1">
                <h3 className="font-semibold text-gray-900 dark:text-gray-100 mb-1">{role.name}</h3>
                <p className="text-sm text-gray-600 dark:text-gray-400">{role.description}</p>
              </div>
              {selectedRole === role.id && (
                <div className="flex-shrink-0 ml-3">
                  <svg
                    className="w-6 h-6 text-blue-600 dark:text-blue-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                </div>
              )}
            </div>

            <div className="space-y-3">
              <div>
                <h4 className="text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">
                  You'll get:
                </h4>
                <ul className="space-y-1">
                  {role.dashboards.map((dashboard, idx) => (
                    <li
                      key={idx}
                      className="text-xs text-gray-600 dark:text-gray-400 flex items-start"
                    >
                      <span className="mr-2">•</span>
                      <span>{dashboard}</span>
                    </li>
                  ))}
                </ul>
              </div>

              <div>
                <h4 className="text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Best for:
                </h4>
                <p className="text-xs text-gray-600 dark:text-gray-400">{role.bestFor}</p>
              </div>
            </div>
          </button>
        ))}
      </div>

      <div className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
        <button
          onClick={() => setShowComparison(!showComparison)}
          className="flex items-center justify-between w-full text-left"
        >
          <h4 className="text-sm font-medium text-gray-900 dark:text-gray-100">
            Compare Roles & Features
          </h4>
          <svg
            className={`w-5 h-5 text-gray-500 dark:text-gray-400 transition-transform ${
              showComparison ? 'rotate-180' : ''
            }`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>
        {showComparison && (
          <div className="mt-4">
            <RoleComparisonTable />
            <p className="mt-3 text-xs text-gray-600 dark:text-gray-400">
              Don't worry - you can change your role anytime in Settings.
            </p>
          </div>
        )}
      </div>

      <div className="flex justify-between pt-4 border-t border-gray-200 dark:border-gray-700">
        <button
          onClick={onBack}
          className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 transition-colors"
        >
          Back
        </button>
        <button
          onClick={onNext}
          disabled={!selectedRole}
          className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-900 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Continue
        </button>
      </div>
    </div>
  );
}
