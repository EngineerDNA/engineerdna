import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';
import { MetricsSummary } from '../components/metrics-summary';
import { AttentionCard } from '../components/attention-card';
import { InsightCard } from '../components/insight-card';

function formatDate(date: Date): string {
  return date.toISOString().split('T')[0];
}

function getCurrentWeekStart(): Date {
  const now = new Date();
  const day = now.getDay();
  const diff = now.getDate() - day + (day === 0 ? -6 : 1);
  const monday = new Date(now.setDate(diff));
  monday.setHours(0, 0, 0, 0);
  return monday;
}

function formatWeekRange(weekStart: string): string {
  const start = new Date(weekStart);
  const end = new Date(start);
  end.setDate(end.getDate() + 6);

  const options: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric', year: 'numeric' };
  return `${start.toLocaleDateString('en-US', options)} - ${end.toLocaleDateString('en-US', options)}`;
}

export function BriefingPage() {
  const [selectedTeam, setSelectedTeam] = useState<string>('');
  const [weekStart, setWeekStart] = useState<Date>(getCurrentWeekStart());
  const [isGenerating, setIsGenerating] = useState(false);

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
    data: briefing,
    isLoading: briefingLoading,
    error: briefingError,
    refetch: refetchBriefing,
  } = useQuery({
    queryKey: ['briefing', selectedTeam, formatDate(weekStart)],
    queryFn: () => api.getWeeklyBriefing(selectedTeam, formatDate(weekStart)),
    enabled: !!selectedTeam,
    staleTime: 300000,
    retry: false,
  });

  const handlePreviousWeek = () => {
    const prev = new Date(weekStart);
    prev.setDate(prev.getDate() - 7);
    setWeekStart(prev);
  };

  const handleNextWeek = () => {
    const next = new Date(weekStart);
    next.setDate(next.getDate() + 7);
    setWeekStart(next);
  };

  const handleRefresh = async () => {
    if (!selectedTeam) return;
    setIsGenerating(true);
    try {
      await api.generateWeeklyBriefing(selectedTeam, formatDate(weekStart));
      await refetchBriefing();
    } catch (error) {
      console.error('Failed to generate briefing:', error);
    } finally {
      setIsGenerating(false);
    }
  };

  const handleCopyTalkingPoints = () => {
    if (briefing?.talking_points) {
      navigator.clipboard.writeText(briefing.talking_points);
    }
  };

  if (teamsLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
      </div>
    );
  }

  if (teams.length === 0) {
    return (
      <div className="text-center py-12">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">No Teams</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-6">
          Create a team to start using weekly briefings.
        </p>
        <a
          href="/teams"
          className="inline-block px-4 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded hover:bg-blue-700 dark:hover:bg-blue-800 transition-colors"
        >
          Go to Teams
        </a>
      </div>
    );
  }

  const isLoading = briefingLoading || isGenerating;
  const hasError = briefingError && !isLoading;
  const noBriefing = !briefing && !isLoading && !hasError;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Weekly Briefing</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            AI-powered insights for engineering managers
          </p>
        </div>

        <button
          onClick={handleRefresh}
          disabled={isLoading || !selectedTeam}
          className="px-4 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded hover:bg-blue-700 dark:hover:bg-blue-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {isGenerating ? 'Generating...' : 'Refresh'}
        </button>
      </div>

      <div className="flex items-center gap-4">
        {teams.length > 1 && (
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Team
            </label>
            <select
              value={selectedTeam}
              onChange={(e) => setSelectedTeam(e.target.value)}
              className="px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400"
            >
              {teams.map((team) => (
                <option key={team.id} value={team.id}>
                  {team.name}
                </option>
              ))}
            </select>
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Week
          </label>
          <div className="flex items-center gap-2">
            <button
              onClick={handlePreviousWeek}
              className="p-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
              aria-label="Previous week"
            >
              <svg
                className="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M15 19l-7-7 7-7"
                />
              </svg>
            </button>
            <div className="px-4 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded min-w-[280px] text-center">
              {formatWeekRange(formatDate(weekStart))}
            </div>
            <button
              onClick={handleNextWeek}
              disabled={weekStart >= getCurrentWeekStart()}
              className="p-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              aria-label="Next week"
            >
              <svg
                className="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 5l7 7-7 7"
                />
              </svg>
            </button>
          </div>
        </div>
      </div>

      {isLoading && (
        <div className="flex justify-center items-center h-64">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400 mx-auto mb-4" />
            <p className="text-gray-600 dark:text-gray-400">
              {isGenerating ? 'Generating briefing...' : 'Loading briefing...'}
            </p>
          </div>
        </div>
      )}

      {hasError && (
        <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded">
          <p className="text-red-800 dark:text-red-200">
            Failed to load briefing. {(briefingError as Error).message}
          </p>
        </div>
      )}

      {noBriefing && (
        <div className="text-center py-12 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
          <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100 mb-4">
            No Briefing Available
          </h2>
          <p className="text-gray-600 dark:text-gray-400 mb-6">
            Generate a briefing for this week to get started.
          </p>
          <button
            onClick={handleRefresh}
            className="px-4 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded hover:bg-blue-700 dark:hover:bg-blue-800 transition-colors"
          >
            Generate Briefing
          </button>
        </div>
      )}

      {briefing && !isLoading && (
        <>
          <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-6">
            <div className="flex items-center gap-2 mb-3">
              <svg
                className="w-6 h-6 text-blue-600 dark:text-blue-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">TL;DR</h2>
            </div>
            <p className="text-gray-700 dark:text-gray-300">{briefing.tldr}</p>
            <div className="mt-3 text-sm text-gray-500 dark:text-gray-400">
              Last updated: {new Date(briefing.last_updated).toLocaleString()}
            </div>
          </div>

          <MetricsSummary metrics={briefing.key_metrics} />

          {briefing.needs_attention.length > 0 && (
            <section>
              <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">
                Needs Your Attention ({briefing.needs_attention.length})
              </h2>
              <div className="space-y-4">
                {briefing.needs_attention.map((item, idx) => (
                  <AttentionCard key={idx} item={item} />
                ))}
              </div>
            </section>
          )}

          {briefing.insights.length > 0 && (
            <section>
              <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">
                AI Insights ({briefing.insights.length})
              </h2>
              <div className="space-y-4">
                {briefing.insights.map((insight, idx) => (
                  <InsightCard key={idx} insight={insight} />
                ))}
              </div>
            </section>
          )}

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {briefing.trending_up.length > 0 && (
              <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
                <div className="flex items-center gap-2 mb-4">
                  <svg
                    className="w-5 h-5 text-green-600 dark:text-green-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    aria-hidden="true"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
                    />
                  </svg>
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                    Trending Up
                  </h3>
                </div>
                <ul className="space-y-2">
                  {briefing.trending_up.map((item, idx) => (
                    <li key={idx} className="text-sm text-gray-700 dark:text-gray-300">
                      <span className="text-green-600 dark:text-green-400">•</span> {item}
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {briefing.trending_down.length > 0 && (
              <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
                <div className="flex items-center gap-2 mb-4">
                  <svg
                    className="w-5 h-5 text-red-600 dark:text-red-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    aria-hidden="true"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M13 17h8m0 0V9m0 8l-8-8-4 4-6-6"
                    />
                  </svg>
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                    Trending Down
                  </h3>
                </div>
                <ul className="space-y-2">
                  {briefing.trending_down.map((item, idx) => (
                    <li key={idx} className="text-sm text-gray-700 dark:text-gray-300">
                      <span className="text-red-600 dark:text-red-400">•</span> {item}
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>

          {briefing.talking_points && (
            <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">
                  Talking Points for Standup
                </h2>
                <button
                  onClick={handleCopyTalkingPoints}
                  className="px-3 py-1.5 text-sm bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
                >
                  Copy
                </button>
              </div>
              <pre className="text-sm text-gray-700 dark:text-gray-300 whitespace-pre-wrap">
                {briefing.talking_points}
              </pre>
            </div>
          )}
        </>
      )}
    </div>
  );
}
