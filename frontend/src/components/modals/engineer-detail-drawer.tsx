import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/client';
import type { Event as EventType } from '../../api/types';

interface EngineerDetailDrawerProps {
  engineer_id: string;
  isOpen: boolean;
  onClose: () => void;
}

type TabType = 'performance' | 'activity' | 'goals' | 'skills';

interface Goal {
  id: string;
  name: string;
  status: 'active' | 'completed' | 'at_risk' | 'off_track';
  progress: number;
  target: number;
}

interface Skill {
  name: string;
  level: number;
  category: string;
}

export function EngineerDetailDrawer({ engineer_id, isOpen, onClose }: EngineerDetailDrawerProps) {
  const [activeTab, setActiveTab] = useState<TabType>('performance');

  const { data: engineerData, isLoading: engineerLoading } = useQuery({
    queryKey: ['engineer', engineer_id],
    queryFn: () => api.getEngineer(engineer_id),
    enabled: isOpen,
  });

  const { data: scoresData, isLoading: scoresLoading } = useQuery({
    queryKey: ['engineer-scores', engineer_id],
    queryFn: () => api.getPerformanceScores(engineer_id),
    enabled: isOpen && activeTab === 'performance',
  });

  const { data: activityData, isLoading: activityLoading } = useQuery({
    queryKey: ['engineer-activity', engineer_id],
    queryFn: () => api.getEngineerActivity(engineer_id, 30),
    enabled: isOpen && activeTab === 'activity',
  });

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      return () => document.removeEventListener('keydown', handleEscape);
    }
  }, [isOpen, onClose]);

  const engineer = engineerData?.engineer;

  const tabs: { id: TabType; label: string }[] = [
    { id: 'performance', label: 'Performance' },
    { id: 'activity', label: 'Activity' },
    { id: 'goals', label: 'Goals' },
    { id: 'skills', label: 'Skills' },
  ];

  const mockGoals: Goal[] = [
    { id: '1', name: 'Improve Code Quality', status: 'active', progress: 75, target: 100 },
    { id: '2', name: 'Complete Certification', status: 'at_risk', progress: 30, target: 100 },
  ];

  const mockSkills: Skill[] = [
    { name: 'TypeScript', level: 85, category: 'Technical' },
    { name: 'React', level: 90, category: 'Technical' },
    { name: 'Leadership', level: 70, category: 'Soft Skills' },
  ];

  return (
    <>
      {isOpen && <div className="fixed inset-0 bg-black bg-opacity-50 z-40" onClick={onClose} />}

      <div
        className={`fixed right-0 top-0 h-full w-full sm:w-[32rem] bg-white dark:bg-gray-800 shadow-xl transform transition-transform duration-300 z-50 ${
          isOpen ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        <div className="flex flex-col h-full">
          <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
            <div className="flex-1 min-w-0">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 truncate">
                {engineerLoading ? 'Loading...' : engineer?.name || 'Engineer Details'}
              </h2>
              {engineer?.email && (
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1 truncate">
                  {engineer.email}
                </p>
              )}
            </div>
            <button
              onClick={onClose}
              className="ml-4 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors flex-shrink-0"
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
            <nav className="flex space-x-4 px-6 overflow-x-auto" aria-label="Tabs">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`py-4 px-1 border-b-2 font-medium text-sm whitespace-nowrap transition-colors ${
                    activeTab === tab.id
                      ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </nav>
          </div>

          <div className="flex-1 overflow-y-auto p-6">
            {activeTab === 'performance' && (
              <PerformanceTab scores={scoresData} isLoading={scoresLoading} />
            )}
            {activeTab === 'activity' && (
              <ActivityTab activity={activityData} isLoading={activityLoading} />
            )}
            {activeTab === 'goals' && <GoalsTab goals={mockGoals} />}
            {activeTab === 'skills' && <SkillsTab skills={mockSkills} />}
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
    </>
  );
}

function PerformanceTab({ scores, isLoading }: { scores?: any; isLoading: boolean }) {
  if (isLoading) {
    return <LoadingSkeleton />;
  }

  if (!scores) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-gray-400">
        No performance data available
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <ScoreCard label="Total Score" value={scores.total_score?.toFixed(1) || 'N/A'} />
        <ScoreCard label="Throughput" value={scores.throughput_score?.toFixed(1) || 'N/A'} />
        <ScoreCard label="Quality" value={scores.quality_score?.toFixed(1) || 'N/A'} />
        <ScoreCard label="Speed" value={scores.speed_score?.toFixed(1) || 'N/A'} />
      </div>
    </div>
  );
}

function ActivityTab({ activity, isLoading }: { activity?: any; isLoading: boolean }) {
  if (isLoading) {
    return <LoadingSkeleton />;
  }

  if (!activity) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-gray-400">
        No activity data available
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Metrics</h3>
        <div className="grid grid-cols-2 gap-4">
          <MetricCard label="Pull Requests" value={activity.metrics?.pull_requests || 0} />
          <MetricCard label="Reviews" value={activity.metrics?.reviews || 0} />
          <MetricCard label="Issues" value={activity.metrics?.issues || 0} />
          <MetricCard label="Commits" value={activity.metrics?.commits || 0} />
        </div>
      </div>

      {activity.recent_events && activity.recent_events.length > 0 && (
        <div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
            Recent Events
          </h3>
          <div className="space-y-3">
            {activity.recent_events.slice(0, 5).map((event: EventType) => (
              <div
                key={event.id}
                className="p-3 border border-gray-200 dark:border-gray-700 rounded-lg"
              >
                <p className="text-sm font-medium text-gray-900 dark:text-gray-100">{event.type}</p>
                <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
                  {new Date(event.timestamp).toLocaleString()}
                </p>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function GoalsTab({ goals }: { goals: Goal[] }) {
  if (goals.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-gray-400">No goals assigned</div>
    );
  }

  return (
    <div className="space-y-4">
      {goals.map((goal) => (
        <div key={goal.id} className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
          <div className="flex items-start justify-between mb-3">
            <h4 className="font-medium text-gray-900 dark:text-gray-100">{goal.name}</h4>
            <span
              className={`px-2 py-1 text-xs rounded ${
                goal.status === 'active'
                  ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400'
                  : goal.status === 'completed'
                    ? 'bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400'
                    : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400'
              }`}
            >
              {goal.status.replace('_', ' ')}
            </span>
          </div>
          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
            <div
              className="bg-blue-600 dark:bg-blue-500 h-2 rounded-full"
              style={{ width: `${goal.progress}%` }}
            />
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-2">
            {goal.progress}% of {goal.target}
          </p>
        </div>
      ))}
    </div>
  );
}

function SkillsTab({ skills }: { skills: Skill[] }) {
  if (skills.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-gray-400">No skills recorded</div>
    );
  }

  const categories = [...new Set(skills.map((s) => s.category))];

  return (
    <div className="space-y-6">
      {categories.map((category) => (
        <div key={category}>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
            {category}
          </h3>
          <div className="space-y-4">
            {skills
              .filter((s) => s.category === category)
              .map((skill) => (
                <div key={skill.name}>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
                      {skill.name}
                    </span>
                    <span className="text-sm text-gray-600 dark:text-gray-400">
                      {skill.level}/100
                    </span>
                  </div>
                  <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                    <div
                      className="bg-blue-600 dark:bg-blue-500 h-2 rounded-full"
                      style={{ width: `${skill.level}%` }}
                    />
                  </div>
                </div>
              ))}
          </div>
        </div>
      ))}
    </div>
  );
}

function ScoreCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
      <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">{label}</dt>
      <dd className="mt-2 text-2xl font-semibold text-gray-900 dark:text-gray-100">{value}</dd>
    </div>
  );
}

function MetricCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
      <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">{label}</dt>
      <dd className="mt-2 text-2xl font-semibold text-gray-900 dark:text-gray-100">{value}</dd>
    </div>
  );
}

function LoadingSkeleton() {
  return (
    <div className="space-y-4">
      {[1, 2, 3].map((i) => (
        <div key={i} className="animate-pulse">
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/4 mb-2" />
          <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded w-full" />
        </div>
      ))}
    </div>
  );
}
