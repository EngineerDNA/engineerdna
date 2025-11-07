import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';
import { OrgScorecardComponent } from '../components/org-scorecard';
import { TeamList } from '../components/team-list';
import type { TeamPerformanceScore } from '../api/types';
import { useMemo, useState } from 'react';

export function TeamsPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<
    'all' | 'unconfigured' | 'warning' | 'below' | 'on-track' | 'exceeding'
  >('all');
  const {
    data: orgScorecard,
    isLoading: orgLoading,
    error: orgError,
  } = useQuery({
    queryKey: ['org-scorecard'],
    queryFn: api.getOrgScorecard,
    staleTime: 30000,
  });

  const {
    data: teamsData,
    isLoading: teamsLoading,
    error: teamsError,
  } = useQuery({
    queryKey: ['teams'],
    queryFn: api.getTeams,
    staleTime: 30000,
  });

  const teams = teamsData?.teams || [];

  const teamPerformanceQueries = useQuery({
    queryKey: ['team-performances', teams.map((t) => t.id)],
    queryFn: async () => {
      const performancePromises = teams.map(async (team) => {
        try {
          const performance = await api.getTeamPerformance(team.id);
          return { teamId: team.id, performance };
        } catch {
          return { teamId: team.id, performance: null };
        }
      });
      return Promise.all(performancePromises);
    },
    enabled: teams.length > 0,
    staleTime: 30000,
  });

  const performancesMap = useMemo(() => {
    if (!teamPerformanceQueries.data) return new Map<string, TeamPerformanceScore>();
    const map = new Map<string, TeamPerformanceScore>();
    teamPerformanceQueries.data.forEach(({ teamId, performance }) => {
      if (performance) {
        map.set(teamId, performance);
      }
    });
    return map;
  }, [teamPerformanceQueries.data]);

  const getTeamStatus = (
    performance: TeamPerformanceScore | undefined
  ): 'unconfigured' | 'warning' | 'below' | 'on-track' | 'exceeding' => {
    if (!performance || performance.member_count === 0) return 'unconfigured';
    const score = performance.total_score || 0;
    if (score < 80) return 'warning';
    if (score < 90) return 'below';
    if (score <= 110) return 'on-track';
    return 'exceeding';
  };

  const filteredTeams = useMemo(() => {
    let result = [...teams];

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      result = result.filter(
        (team) =>
          team.name.toLowerCase().includes(query) ||
          team.manager?.name.toLowerCase().includes(query)
      );
    }

    if (statusFilter !== 'all') {
      result = result.filter((team) => {
        const performance = performancesMap.get(team.id);
        return getTeamStatus(performance) === statusFilter;
      });
    }

    return result;
  }, [teams, searchQuery, statusFilter, performancesMap]);

  const isLoading = orgLoading || teamsLoading || teamPerformanceQueries.isLoading;
  const error = orgError || teamsError;

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded">
        <p className="text-red-800 dark:text-red-200">
          Failed to load teams data: {(error as Error).message}
        </p>
      </div>
    );
  }

  if (!orgScorecard) {
    return (
      <div className="text-center py-12">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">
          No Organization Data
        </h2>
        <p className="text-gray-600 dark:text-gray-400 mb-6">
          Organization scorecard data is not available yet.
        </p>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">
          Team Performance Overview
        </h1>
        <p className="text-gray-600 dark:text-gray-400 mt-1">
          Organization-wide team metrics and performance
        </p>
      </div>

      <OrgScorecardComponent scorecard={orgScorecard} />

      <div className="mt-8">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Teams</h2>

        <div className="mb-6 space-y-4">
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="flex-1">
              <input
                type="text"
                placeholder="Search by team name or manager..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent"
              />
            </div>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as typeof statusFilter)}
              className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent"
            >
              <option value="all">All Statuses</option>
              <option value="unconfigured">Unconfigured</option>
              <option value="warning">Warning (&lt; 80)</option>
              <option value="below">Below Target (80-89)</option>
              <option value="on-track">On Track (90-110)</option>
              <option value="exceeding">Exceeding (&gt; 110)</option>
            </select>
          </div>
          {(searchQuery || statusFilter !== 'all') && (
            <div className="text-sm text-gray-600 dark:text-gray-400">
              Showing {filteredTeams.length} of {teams.length} teams
            </div>
          )}
        </div>

        {filteredTeams.length === 0 ? (
          <div className="text-center py-12 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
            <p className="text-gray-600 dark:text-gray-400">No teams match your filters</p>
          </div>
        ) : (
          <TeamList teams={filteredTeams} performances={performancesMap} />
        )}
      </div>
    </div>
  );
}
