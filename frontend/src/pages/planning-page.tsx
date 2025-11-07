import { useState, useEffect } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { api } from '../api/client';
import type { Sprint, Team } from '../api/types';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';

export function PlanningPage() {
  const [selectedTeam, setSelectedTeam] = useState<string>('');
  const [showEstimator, setShowEstimator] = useState(false);
  const [estimatorForm, setEstimatorForm] = useState({
    feature_name: '',
    estimated_points: '',
  });

  const { data: teamsData, isLoading: teamsLoading } = useQuery({
    queryKey: ['teams'],
    queryFn: api.getTeams,
    staleTime: 60000,
  });

  const teams = teamsData?.teams || [];

  useEffect(() => {
    if (teams.length > 0 && !selectedTeam) {
      setSelectedTeam(teams[0].id);
    }
  }, [teams, selectedTeam]);

  const {
    data: sprintsData,
    isLoading: sprintsLoading,
    error: sprintsError,
  } = useQuery({
    queryKey: ['sprints', selectedTeam],
    queryFn: () => api.getSprints(selectedTeam, 'active'),
    enabled: !!selectedTeam,
  });

  const { data: velocity, isLoading: velocityLoading } = useQuery({
    queryKey: ['velocity', selectedTeam],
    queryFn: () => api.getVelocity(selectedTeam, 12),
    enabled: !!selectedTeam,
  });

  const estimateMutation = useMutation({
    mutationFn: (data: { feature_name: string; estimated_points: number; team_id: string }) =>
      api.estimateTimeline(data),
  });

  const handleEstimate = () => {
    if (!estimatorForm.feature_name || !estimatorForm.estimated_points || !selectedTeam) {
      return;
    }
    estimateMutation.mutate({
      feature_name: estimatorForm.feature_name,
      estimated_points: parseInt(estimatorForm.estimated_points, 10),
      team_id: selectedTeam,
    });
  };

  if (teamsLoading) {
    return (
      <div className="flex justify-center items-center h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
      </div>
    );
  }

  const sprints = sprintsData?.sprints || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Planning</h1>
        <div className="flex items-center gap-4">
          <select
            value={selectedTeam}
            onChange={(e) => setSelectedTeam(e.target.value)}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400"
          >
            {teams.map((team: Team) => (
              <option key={team.id} value={team.id}>
                {team.name}
              </option>
            ))}
          </select>
          <button
            onClick={() => setShowEstimator(!showEstimator)}
            className="px-4 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-800 transition-colors"
          >
            {showEstimator ? 'Hide' : 'Timeline Estimator'}
          </button>
        </div>
      </div>

      {showEstimator && (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow border border-gray-200 dark:border-gray-700">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4">
            Timeline Estimator
          </h2>
          <div className="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Feature Name
              </label>
              <input
                type="text"
                value={estimatorForm.feature_name}
                onChange={(e) =>
                  setEstimatorForm({ ...estimatorForm, feature_name: e.target.value })
                }
                className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                placeholder="e.g., User Dashboard Redesign"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Estimated Points
              </label>
              <input
                type="number"
                value={estimatorForm.estimated_points}
                onChange={(e) =>
                  setEstimatorForm({ ...estimatorForm, estimated_points: e.target.value })
                }
                className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                placeholder="e.g., 84"
              />
            </div>
          </div>
          <button
            onClick={handleEstimate}
            disabled={estimateMutation.isPending}
            className="px-6 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-800 disabled:bg-gray-400 dark:disabled:bg-gray-600 transition-colors"
          >
            {estimateMutation.isPending ? 'Estimating...' : 'Estimate Timeline'}
          </button>

          {estimateMutation.data && (
            <div className="mt-6 p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg">
              <h3 className="font-semibold text-blue-900 dark:text-blue-100 mb-3">
                Estimated Timeline for {estimateMutation.data.feature_name}
              </h3>
              <div className="grid grid-cols-3 gap-4 mb-4">
                <div className="text-center">
                  <div className="text-sm text-blue-700 dark:text-blue-300 mb-1">Best Case</div>
                  <div className="text-2xl font-bold text-blue-900 dark:text-blue-100">
                    {estimateMutation.data.best_case_weeks}w
                  </div>
                  <div className="text-xs text-blue-600 dark:text-blue-400">
                    ({estimateMutation.data.best_case_sprints} sprints)
                  </div>
                </div>
                <div className="text-center">
                  <div className="text-sm text-blue-700 dark:text-blue-300 mb-1">Likely Case</div>
                  <div className="text-2xl font-bold text-blue-900 dark:text-blue-100">
                    {estimateMutation.data.likely_case_weeks}w
                  </div>
                  <div className="text-xs text-blue-600 dark:text-blue-400">
                    ({estimateMutation.data.likely_case_sprints} sprints)
                  </div>
                </div>
                <div className="text-center">
                  <div className="text-sm text-blue-700 dark:text-blue-300 mb-1">Worst Case</div>
                  <div className="text-2xl font-bold text-blue-900 dark:text-blue-100">
                    {estimateMutation.data.worst_case_weeks}w
                  </div>
                  <div className="text-xs text-blue-600 dark:text-blue-400">
                    ({estimateMutation.data.worst_case_sprints} sprints)
                  </div>
                </div>
              </div>
              <div className="text-sm text-blue-800 dark:text-blue-200 mb-2">
                Based on {estimateMutation.data.average_velocity} points/sprint average
              </div>
              {estimateMutation.data.assumptions &&
                estimateMutation.data.assumptions.length > 0 && (
                  <div className="mt-3">
                    <div className="text-sm font-medium text-blue-900 dark:text-blue-100 mb-1">
                      Assumptions:
                    </div>
                    <ul className="text-sm text-blue-800 dark:text-blue-200 list-disc list-inside">
                      {estimateMutation.data.assumptions.map((assumption, i) => (
                        <li key={i}>{assumption}</li>
                      ))}
                    </ul>
                  </div>
                )}
            </div>
          )}
        </div>
      )}

      {velocity && !velocityLoading && (
        <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow border border-gray-200 dark:border-gray-700">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4">
            Team Velocity
          </h2>
          <div className="grid grid-cols-3 gap-4 mb-4">
            <div>
              <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Average Velocity</div>
              <div className="text-3xl font-bold text-gray-900 dark:text-gray-100">
                {velocity.average_points}
              </div>
              <div className="text-xs text-gray-500 dark:text-gray-500">points/sprint</div>
            </div>
            <div>
              <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Trend</div>
              <div className="text-3xl font-bold capitalize text-gray-900 dark:text-gray-100">
                {velocity.trend}
              </div>
              <div className="text-xs text-gray-500 dark:text-gray-500">
                {typeof velocity.confidence === 'number'
                  ? `${Math.round(velocity.confidence * 100)}% confidence`
                  : `${velocity.confidence} confidence`}
              </div>
            </div>
            <div>
              <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Std. Deviation</div>
              <div className="text-3xl font-bold text-gray-900 dark:text-gray-100">
                {(velocity.std_dev ?? 0).toFixed(1)}
              </div>
              <div className="text-xs text-gray-500 dark:text-gray-500">points</div>
            </div>
          </div>
          <div className="mt-4">
            <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Recent Sprints
            </h3>
            {velocity.recent_sprints && velocity.recent_sprints.length > 0 ? (
              <ResponsiveContainer width="100%" height={200}>
                <BarChart data={velocity.recent_sprints}>
                  <XAxis
                    dataKey="sprint_name"
                    tick={{ fontSize: 12 }}
                    angle={-45}
                    textAnchor="end"
                    height={80}
                  />
                  <YAxis tick={{ fontSize: 12 }} />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: 'var(--tooltip-bg, #fff)',
                      border: '1px solid var(--tooltip-border, #ccc)',
                    }}
                  />
                  <Bar dataKey="points" fill="#3b82f6" />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="text-sm text-gray-500 dark:text-gray-400">
                No recent sprint data available
              </div>
            )}
          </div>
        </div>
      )}

      <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow border border-gray-200 dark:border-gray-700">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4">
          Active Sprints
        </h2>

        {sprintsLoading && (
          <div className="flex justify-center items-center h-32">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 dark:border-blue-400" />
          </div>
        )}

        {sprintsError && (
          <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
            <p className="text-red-800 dark:text-red-200">
              Failed to load sprints: {(sprintsError as Error).message}
            </p>
          </div>
        )}

        {!sprintsLoading && !sprintsError && sprints.length === 0 && (
          <div className="text-center py-8 text-gray-500 dark:text-gray-400">
            No active sprints found. Sprints will appear here once created via API.
          </div>
        )}

        {!sprintsLoading && !sprintsError && sprints.length > 0 && (
          <div className="space-y-4">
            {sprints.map((sprint: Sprint) => (
              <SprintCard key={sprint.id} sprint={sprint} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function SprintCard({ sprint }: { sprint: Sprint }) {
  const { data: health, isLoading } = useQuery({
    queryKey: ['sprint-health', sprint.id],
    queryFn: () => api.getSprintHealth(sprint.id),
    staleTime: 30000,
  });

  // Calculate completion percentage from sprint data if health data is missing or 0
  const completionPercentage =
    health?.completion_percentage ||
    (sprint.committed_points > 0 ? (sprint.completed_points / sprint.committed_points) * 100 : 0);

  const getRiskColor = (riskLevel: string) => {
    switch (riskLevel) {
      case 'on_track':
        return 'bg-green-100 dark:bg-green-900/20 border-green-300 dark:border-green-800 text-green-800 dark:text-green-200';
      case 'at_risk':
        return 'bg-yellow-100 dark:bg-yellow-900/20 border-yellow-300 dark:border-yellow-800 text-yellow-800 dark:text-yellow-200';
      case 'high_risk':
        return 'bg-red-100 dark:bg-red-900/20 border-red-300 dark:border-red-800 text-red-800 dark:text-red-200';
      default:
        return 'bg-gray-100 dark:bg-gray-700 border-gray-300 dark:border-gray-600 text-gray-800 dark:text-gray-200';
    }
  };

  return (
    <div className="border border-gray-300 dark:border-gray-600 rounded-lg p-4 bg-gray-50 dark:bg-gray-700">
      <div className="flex justify-between items-start mb-4">
        <div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">{sprint.name}</h3>
          <div className="text-sm text-gray-600 dark:text-gray-400">
            {new Date(sprint.start_date).toLocaleDateString()} -{' '}
            {new Date(sprint.end_date).toLocaleDateString()}
          </div>
        </div>
        {health && (
          <span
            className={`px-3 py-1 rounded-full text-sm font-medium border ${getRiskColor(health.risk_level)}`}
          >
            {health.risk_level?.replace('_', ' ').toUpperCase() || 'UNKNOWN'}
          </span>
        )}
      </div>

      {isLoading && (
        <div className="flex justify-center items-center h-20">
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600 dark:border-blue-400" />
        </div>
      )}

      {health && !isLoading && (
        <div className="space-y-3">
          <div className="grid grid-cols-3 gap-4">
            <div>
              <div className="text-xs text-gray-600 dark:text-gray-400 mb-1">Committed</div>
              <div className="text-xl font-bold text-gray-900 dark:text-gray-100">
                {health.committed_points}
              </div>
            </div>
            <div>
              <div className="text-xs text-gray-600 dark:text-gray-400 mb-1">Completed</div>
              <div className="text-xl font-bold text-green-600 dark:text-green-400">
                {health.completed_points}
              </div>
            </div>
            <div>
              <div className="text-xs text-gray-600 dark:text-gray-400 mb-1">Remaining</div>
              <div className="text-xl font-bold text-blue-600 dark:text-blue-400">
                {health.remaining_points}
              </div>
            </div>
          </div>

          <div className="w-full bg-gray-200 dark:bg-gray-600 rounded-full h-2">
            <div
              className="bg-blue-600 dark:bg-blue-500 h-2 rounded-full transition-all"
              style={{ width: `${completionPercentage}%` }}
            />
          </div>

          <div className="text-sm text-gray-700 dark:text-gray-300">
            {completionPercentage.toFixed(0)}% complete · {health.days_remaining} days remaining
          </div>

          {health.velocity_comparison && (
            <div className="text-sm text-gray-600 dark:text-gray-400">
              Committed {health.velocity_comparison.difference > 0 ? 'above' : 'below'} historical
              avg by{' '}
              <span className="font-semibold">
                {(Math.abs(health.velocity_comparison.difference_percentage ?? 0) ?? 0).toFixed(0)}%
              </span>
            </div>
          )}

          {health.alerts && health.alerts.length > 0 && (
            <div className="mt-3 p-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
              <div className="text-sm font-medium text-yellow-900 dark:text-yellow-100 mb-1">
                Alerts:
              </div>
              <ul className="text-sm text-yellow-800 dark:text-yellow-200 list-disc list-inside">
                {health.alerts.map((alert, i) => (
                  <li key={i}>{alert}</li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
