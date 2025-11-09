import type { WidgetProps } from './widget-registry';

interface Goal {
  id: string;
  name: string;
  status: 'active' | 'completed' | 'at_risk' | 'off_track';
  progress: number;
  target: number;
  milestones: Array<{
    name: string;
    completed: boolean;
    target_value: number;
    current_value: number;
  }>;
}

export function GoalTrackerWidget({ config, onOpenModal }: WidgetProps) {
  const engineerId = config.engineer_id as string | undefined;
  const teamId = config.team_id as string | undefined;

  const mockGoals: Goal[] = [
    {
      id: '1',
      name: 'Improve Code Quality Score',
      status: 'active',
      progress: 75,
      target: 100,
      milestones: [
        { name: 'Reduce bug rate by 20%', completed: true, target_value: 20, current_value: 25 },
        {
          name: 'Increase test coverage to 80%',
          completed: false,
          target_value: 80,
          current_value: 65,
        },
        {
          name: 'Zero critical vulnerabilities',
          completed: true,
          target_value: 0,
          current_value: 0,
        },
      ],
    },
    {
      id: '2',
      name: 'Increase Team Velocity',
      status: 'at_risk',
      progress: 45,
      target: 100,
      milestones: [
        {
          name: 'Average 50 points per sprint',
          completed: false,
          target_value: 50,
          current_value: 38,
        },
        { name: 'Reduce cycle time by 30%', completed: false, target_value: 30, current_value: 15 },
      ],
    },
  ];

  const getStatusColor = (status: Goal['status']) => {
    switch (status) {
      case 'completed':
        return 'bg-green-100 dark:bg-green-900/20 text-green-800 dark:text-green-200';
      case 'active':
        return 'bg-blue-100 dark:bg-blue-900/20 text-blue-800 dark:text-blue-200';
      case 'at_risk':
        return 'bg-yellow-100 dark:bg-yellow-900/20 text-yellow-800 dark:text-yellow-200';
      case 'off_track':
        return 'bg-red-100 dark:bg-red-900/20 text-red-800 dark:text-red-200';
      default:
        return 'bg-gray-100 dark:bg-gray-900/20 text-gray-800 dark:text-gray-200';
    }
  };

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full flex flex-col">
      <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-4">Goals & OKRs</h3>

      <div className="flex-1 overflow-auto space-y-4">
        {mockGoals.length === 0 ? (
          <div className="flex items-center justify-center h-32 text-gray-500 dark:text-gray-400">
            No active goals
          </div>
        ) : (
          mockGoals.map((goal) => (
            <div
              key={goal.id}
              className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 cursor-pointer hover:shadow-md transition-shadow"
              onClick={() => onOpenModal?.('goal-detail', { goalId: goal.id })}
            >
              <div className="flex items-start justify-between mb-3">
                <h4 className="text-sm font-medium text-gray-900 dark:text-gray-100">
                  {goal.name}
                </h4>
                <span
                  className={`inline-flex px-2 py-1 rounded text-xs ${getStatusColor(goal.status)}`}
                >
                  {goal.status.replace('_', ' ')}
                </span>
              </div>

              <div className="mb-3">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs text-gray-600 dark:text-gray-400">Overall Progress</span>
                  <span className="text-xs font-medium text-gray-900 dark:text-gray-100">
                    {goal.progress}%
                  </span>
                </div>
                <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div
                    className={`h-2 rounded-full transition-all ${
                      goal.status === 'completed'
                        ? 'bg-green-600'
                        : goal.status === 'at_risk'
                          ? 'bg-yellow-600'
                          : goal.status === 'off_track'
                            ? 'bg-red-600'
                            : 'bg-blue-600'
                    }`}
                    style={{ width: `${Math.min(goal.progress, 100)}%` }}
                  />
                </div>
              </div>

              <div className="space-y-2">
                <p className="text-xs font-medium text-gray-700 dark:text-gray-300">Milestones:</p>
                {goal.milestones.map((milestone, idx) => (
                  <div key={idx} className="flex items-start gap-2">
                    <div className="flex-shrink-0 mt-0.5">
                      {milestone.completed ? (
                        <svg
                          className="w-4 h-4 text-green-600 dark:text-green-400"
                          fill="currentColor"
                          viewBox="0 0 20 20"
                        >
                          <path
                            fillRule="evenodd"
                            d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                            clipRule="evenodd"
                          />
                        </svg>
                      ) : (
                        <svg
                          className="w-4 h-4 text-gray-400 dark:text-gray-600"
                          fill="currentColor"
                          viewBox="0 0 20 20"
                        >
                          <path
                            fillRule="evenodd"
                            d="M10 18a8 8 0 100-16 8 8 0 000 16zm0-2a6 6 0 100-12 6 6 0 000 12z"
                            clipRule="evenodd"
                          />
                        </svg>
                      )}
                    </div>
                    <div className="flex-1">
                      <p
                        className={`text-xs ${milestone.completed ? 'text-gray-900 dark:text-gray-100 line-through' : 'text-gray-700 dark:text-gray-300'}`}
                      >
                        {milestone.name}
                      </p>
                      {!milestone.completed && (
                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                          {milestone.current_value} / {milestone.target_value}
                        </p>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))
        )}
      </div>

      {mockGoals.length > 0 && (
        <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700 text-center">
          <button
            onClick={() => onOpenModal?.('goals-list', { engineerId, teamId })}
            className="text-sm text-blue-600 dark:text-blue-400 hover:underline"
          >
            View All Goals
          </button>
        </div>
      )}
    </div>
  );
}
