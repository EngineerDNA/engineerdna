import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';
import { ScoreCard } from '../components/score-card';
import type { PerformanceScore, Role } from '../api/types';

export function ScorecardPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<
    'all' | 'warning' | 'below' | 'on-track' | 'exceeding'
  >('all');
  const [roleFilter, setRoleFilter] = useState('all');
  const {
    data: engineersData,
    isLoading: engineersLoading,
    error: engineersError,
  } = useQuery({
    queryKey: ['engineers'],
    queryFn: api.getEngineers,
    staleTime: 30000,
  });

  const { data: rolesData, isLoading: rolesLoading } = useQuery({
    queryKey: ['roles'],
    queryFn: api.getRoles,
    staleTime: 60000,
  });

  const engineers = engineersData?.engineers || [];
  const roles = rolesData?.roles || [];

  const engineerScoresQueries = useQuery({
    queryKey: ['all-performance-scores', engineers.map((e) => e.id)],
    queryFn: async () => {
      const scorePromises = engineers.map(async (engineer) => {
        try {
          const result = await api.getPerformanceScores(engineer.id);
          return { engineerId: engineer.id, scores: result.scores };
        } catch {
          return { engineerId: engineer.id, scores: [] };
        }
      });
      return Promise.all(scorePromises);
    },
    enabled: engineers.length > 0,
    staleTime: 30000,
  });

  const engineerScoresMap = useMemo(() => {
    if (!engineerScoresQueries.data) return new Map<string, PerformanceScore | null>();
    const map = new Map<string, PerformanceScore | null>();
    engineerScoresQueries.data.forEach(({ engineerId, scores }) => {
      map.set(engineerId, scores && scores.length > 0 ? scores[0] : null);
    });
    return map;
  }, [engineerScoresQueries.data]);

  const rolesMap = useMemo(() => {
    const map = new Map<string, Role>();
    roles.forEach((role) => {
      map.set(role.id, role);
    });
    return map;
  }, [roles]);

  const sortedEngineers = useMemo(() => {
    return [...engineers].sort((a, b) => {
      const scoreA = engineerScoresMap.get(a.id);
      const scoreB = engineerScoresMap.get(b.id);
      const totalA = scoreA?.total_score || 0;
      const totalB = scoreB?.total_score || 0;
      return totalB - totalA;
    });
  }, [engineers, engineerScoresMap]);

  const getScoreStatus = (
    score: number | null | undefined
  ): 'warning' | 'below' | 'on-track' | 'exceeding' => {
    if (!score) return 'warning';
    if (score < 80) return 'warning';
    if (score < 90) return 'below';
    if (score <= 110) return 'on-track';
    return 'exceeding';
  };

  const filteredEngineers = useMemo(() => {
    let result = [...sortedEngineers];

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      result = result.filter(
        (engineer) =>
          engineer.name.toLowerCase().includes(query) ||
          engineer.email?.toLowerCase().includes(query)
      );
    }

    if (statusFilter !== 'all') {
      result = result.filter((engineer) => {
        const score = engineerScoresMap.get(engineer.id);
        return getScoreStatus(score?.total_score) === statusFilter;
      });
    }

    if (roleFilter !== 'all') {
      result = result.filter((engineer) => engineer.role_id === roleFilter);
    }

    return result;
  }, [sortedEngineers, searchQuery, statusFilter, roleFilter, engineerScoresMap]);

  if (engineersLoading || rolesLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
      </div>
    );
  }

  if (engineersError) {
    return (
      <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded">
        <p className="text-red-800 dark:text-red-200">
          Failed to load scorecard data: {(engineersError as Error).message}
        </p>
      </div>
    );
  }

  if (engineers.length === 0) {
    return (
      <div className="text-center py-12">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">
          No Team Members
        </h2>
        <p className="text-gray-600 dark:text-gray-400 mb-6">
          Add team members to see their performance scores.
        </p>
        <a
          href="/team"
          className="inline-block px-4 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded hover:bg-blue-700 dark:hover:bg-blue-800 transition-colors"
        >
          Go to Team Page
        </a>
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Team Scorecard</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Target baseline: 100 (scores calculated weekly)
          </p>
        </div>
      </div>

      <div className="mb-6 space-y-4">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1">
            <input
              type="text"
              placeholder="Search by name or email..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent"
            />
          </div>
          <div className="flex gap-4">
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as typeof statusFilter)}
              className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent"
            >
              <option value="all">All Statuses</option>
              <option value="warning">Warning (&lt; 80)</option>
              <option value="below">Below Target (80-89)</option>
              <option value="on-track">On Track (90-110)</option>
              <option value="exceeding">Exceeding (&gt; 110)</option>
            </select>
            <select
              value={roleFilter}
              onChange={(e) => setRoleFilter(e.target.value)}
              className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent"
            >
              <option value="all">All Roles</option>
              {roles.map((role) => (
                <option key={role.id} value={role.id}>
                  {role.name}
                </option>
              ))}
            </select>
          </div>
        </div>
        {(searchQuery || statusFilter !== 'all' || roleFilter !== 'all') && (
          <div className="text-sm text-gray-600 dark:text-gray-400">
            Showing {filteredEngineers.length} of {sortedEngineers.length} engineers
          </div>
        )}
      </div>

      {engineerScoresQueries.isLoading ? (
        <div className="flex justify-center items-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
        </div>
      ) : filteredEngineers.length === 0 ? (
        <div className="text-center py-12 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
          <p className="text-gray-600 dark:text-gray-400">No engineers match your filters</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-6">
          {filteredEngineers.map((engineer) => {
            const score = engineerScoresMap.get(engineer.id) || null;
            const role = engineer.role_id ? rolesMap.get(engineer.role_id) || null : null;
            return <ScoreCard key={engineer.id} engineer={engineer} score={score} role={role} />;
          })}
        </div>
      )}
    </div>
  );
}
