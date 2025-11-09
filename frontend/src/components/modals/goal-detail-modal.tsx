import { useState, useEffect } from 'react';

interface GoalDetailModalProps {
  goal_id: string;
  isOpen: boolean;
  onClose: () => void;
}

interface Goal {
  id: string;
  name: string;
  description?: string;
  owner_type: 'individual' | 'team' | 'organization';
  owner_id: string;
  owner_name: string;
  status: 'active' | 'completed' | 'at_risk' | 'off_track' | 'archived';
  progress: number;
  target: number;
  start_date: string;
  target_date: string;
  milestones: Milestone[];
  dependencies: Dependency[];
  blockers: Blocker[];
  created_at: string;
  updated_at: string;
}

interface Milestone {
  id: string;
  name: string;
  description?: string;
  target_value: number;
  current_value: number;
  completed: boolean;
  target_date?: string;
  completed_at?: string;
}

interface Dependency {
  id: string;
  goal_id: string;
  goal_name: string;
  status: string;
}

interface Blocker {
  id: string;
  description: string;
  severity: 'low' | 'medium' | 'high';
  created_at: string;
  resolved_at?: string;
}

export function GoalDetailModal({ goal_id, isOpen, onClose }: GoalDetailModalProps) {
  const [activeTab, setActiveTab] = useState<'overview' | 'milestones' | 'dependencies'>(
    'overview'
  );

  const mockGoal: Goal = {
    id: goal_id,
    name: 'Improve Code Quality Score',
    description: 'Increase overall code quality metrics across the team to meet industry standards',
    owner_type: 'team',
    owner_id: 'team-1',
    owner_name: 'Engineering Team',
    status: 'active',
    progress: 75,
    target: 100,
    start_date: '2025-01-01T00:00:00Z',
    target_date: '2025-03-31T23:59:59Z',
    milestones: [
      {
        id: 'm1',
        name: 'Reduce bug rate by 20%',
        description: 'Lower production bugs from 5/week to 4/week',
        target_value: 20,
        current_value: 25,
        completed: true,
        target_date: '2025-01-31T23:59:59Z',
        completed_at: '2025-01-28T10:30:00Z',
      },
      {
        id: 'm2',
        name: 'Increase test coverage to 80%',
        description: 'Bring unit test coverage from 65% to 80%',
        target_value: 80,
        current_value: 72,
        completed: false,
        target_date: '2025-02-28T23:59:59Z',
      },
      {
        id: 'm3',
        name: 'Zero critical vulnerabilities',
        description: 'Address all critical security vulnerabilities',
        target_value: 0,
        current_value: 0,
        completed: true,
        target_date: '2025-01-15T23:59:59Z',
        completed_at: '2025-01-12T14:20:00Z',
      },
    ],
    dependencies: [
      {
        id: 'd1',
        goal_id: 'goal-2',
        goal_name: 'Complete CI/CD Pipeline',
        status: 'completed',
      },
    ],
    blockers: [
      {
        id: 'b1',
        description: 'Waiting for security team approval on new testing framework',
        severity: 'medium',
        created_at: '2025-01-20T09:00:00Z',
      },
    ],
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-25T15:30:00Z',
  };

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      return () => document.removeEventListener('keydown', handleEscape);
    }
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const goal = mockGoal;
  const completedMilestones = goal.milestones.filter((m) => m.completed).length;
  const totalMilestones = goal.milestones.length;
  const daysRemaining = Math.ceil(
    (new Date(goal.target_date).getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24)
  );

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="fixed inset-0 bg-black bg-opacity-50" onClick={onClose} />

      <div className="relative min-h-screen flex items-center justify-center p-4">
        <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] flex flex-col">
          <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
            <div className="flex-1">
              <div className="flex items-center gap-3">
                <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{goal.name}</h2>
                <span
                  className={`px-3 py-1 rounded-full text-sm font-medium ${
                    goal.status === 'active'
                      ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400'
                      : goal.status === 'completed'
                        ? 'bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400'
                        : goal.status === 'at_risk'
                          ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400'
                          : goal.status === 'off_track'
                            ? 'bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400'
                            : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-400'
                  }`}
                >
                  {goal.status.replace('_', ' ').toUpperCase()}
                </span>
              </div>
              {goal.description && (
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-2">{goal.description}</p>
              )}
            </div>
            <button
              onClick={onClose}
              className="ml-4 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
              aria-label="Close"
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

          <div className="border-b border-gray-200 dark:border-gray-700">
            <nav className="flex space-x-8 px-6" aria-label="Tabs">
              {['overview', 'milestones', 'dependencies'].map((tab) => (
                <button
                  key={tab}
                  onClick={() => setActiveTab(tab as typeof activeTab)}
                  className={`py-4 px-1 border-b-2 font-medium text-sm capitalize transition-colors ${
                    activeTab === tab
                      ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
                  }`}
                >
                  {tab}
                </button>
              ))}
            </nav>
          </div>

          <div className="flex-1 overflow-y-auto p-6">
            {activeTab === 'overview' && (
              <div className="space-y-6">
                <div>
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
                    Progress
                  </h3>
                  <div className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                    <div className="flex items-center justify-between mb-3">
                      <span className="text-sm text-gray-600 dark:text-gray-400">
                        Overall Progress
                      </span>
                      <span className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                        {goal.progress}%
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
                      <div
                        className="bg-blue-600 dark:bg-blue-500 h-3 rounded-full transition-all"
                        style={{ width: `${goal.progress}%` }}
                      />
                    </div>
                    <div className="grid grid-cols-3 gap-4 mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
                      <div>
                        <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                          Milestones
                        </dt>
                        <dd className="mt-1 text-lg font-semibold text-gray-900 dark:text-gray-100">
                          {completedMilestones}/{totalMilestones}
                        </dd>
                      </div>
                      <div>
                        <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                          Days Remaining
                        </dt>
                        <dd className="mt-1 text-lg font-semibold text-gray-900 dark:text-gray-100">
                          {daysRemaining}
                        </dd>
                      </div>
                      <div>
                        <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                          Owner
                        </dt>
                        <dd className="mt-1 text-lg font-semibold text-gray-900 dark:text-gray-100 capitalize">
                          {goal.owner_name}
                        </dd>
                      </div>
                    </div>
                  </div>
                </div>

                {goal.blockers.filter((b) => !b.resolved_at).length > 0 && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
                      Active Blockers
                    </h3>
                    <div className="space-y-3">
                      {goal.blockers
                        .filter((b) => !b.resolved_at)
                        .map((blocker) => (
                          <div
                            key={blocker.id}
                            className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
                          >
                            <div className="flex items-start justify-between">
                              <p className="text-sm text-gray-900 dark:text-gray-100 flex-1">
                                {blocker.description}
                              </p>
                              <span
                                className={`ml-3 px-2 py-1 text-xs rounded ${
                                  blocker.severity === 'high'
                                    ? 'bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400'
                                    : blocker.severity === 'medium'
                                      ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400'
                                      : 'bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400'
                                }`}
                              >
                                {blocker.severity}
                              </span>
                            </div>
                            <p className="text-xs text-gray-500 dark:text-gray-500 mt-2">
                              Created {new Date(blocker.created_at).toLocaleDateString()}
                            </p>
                          </div>
                        ))}
                    </div>
                  </div>
                )}
              </div>
            )}

            {activeTab === 'milestones' && (
              <div className="space-y-4">
                {goal.milestones.map((milestone) => (
                  <div
                    key={milestone.id}
                    className={`p-4 border rounded-lg ${
                      milestone.completed
                        ? 'border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-900/10'
                        : 'border-gray-200 dark:border-gray-700'
                    }`}
                  >
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex-1">
                        <h4 className="font-medium text-gray-900 dark:text-gray-100">
                          {milestone.name}
                        </h4>
                        {milestone.description && (
                          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                            {milestone.description}
                          </p>
                        )}
                      </div>
                      {milestone.completed ? (
                        <span className="ml-3 px-2 py-1 text-xs bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400 rounded">
                          Completed
                        </span>
                      ) : (
                        <span className="ml-3 px-2 py-1 text-xs bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400 rounded">
                          In Progress
                        </span>
                      )}
                    </div>

                    <div className="mb-3">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm text-gray-600 dark:text-gray-400">Progress</span>
                        <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
                          {milestone.current_value} / {milestone.target_value}
                        </span>
                      </div>
                      <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                        <div
                          className={`h-2 rounded-full ${
                            milestone.completed
                              ? 'bg-green-600 dark:bg-green-500'
                              : 'bg-blue-600 dark:bg-blue-500'
                          }`}
                          style={{
                            width: `${Math.min((milestone.current_value / milestone.target_value) * 100, 100)}%`,
                          }}
                        />
                      </div>
                    </div>

                    <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-500">
                      {milestone.target_date && (
                        <span>Target: {new Date(milestone.target_date).toLocaleDateString()}</span>
                      )}
                      {milestone.completed_at && (
                        <span>
                          Completed: {new Date(milestone.completed_at).toLocaleDateString()}
                        </span>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}

            {activeTab === 'dependencies' && (
              <div className="space-y-6">
                {goal.dependencies.length > 0 ? (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
                      Depends On
                    </h3>
                    <div className="space-y-3">
                      {goal.dependencies.map((dep) => (
                        <div
                          key={dep.id}
                          className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
                        >
                          <div className="flex items-center justify-between">
                            <span className="text-gray-900 dark:text-gray-100">
                              {dep.goal_name}
                            </span>
                            <span
                              className={`px-2 py-1 text-xs rounded capitalize ${
                                dep.status === 'completed'
                                  ? 'bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400'
                                  : dep.status === 'active'
                                    ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400'
                                    : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-400'
                              }`}
                            >
                              {dep.status}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                ) : (
                  <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                    No dependencies
                  </div>
                )}
              </div>
            )}
          </div>

          <div className="flex gap-3 justify-end p-6 border-t border-gray-200 dark:border-gray-700">
            <button
              onClick={onClose}
              className="px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
