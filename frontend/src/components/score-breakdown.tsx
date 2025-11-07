import type { PerformanceScore, RawMetrics } from '../api/types';

interface ScoreBreakdownProps {
  score: PerformanceScore;
}

export function ScoreBreakdown({ score }: ScoreBreakdownProps) {
  const rawMetrics: RawMetrics = score.raw_metrics
    ? JSON.parse(score.raw_metrics)
    : {
        throughput_prs_per_week: 0,
        throughput_story_points: 0,
        quality_bug_rate: 0,
        quality_rework_rate: 0,
        speed_cycle_time_days: 0,
        speed_time_to_first_review: 0,
        collaboration_reviews_given: 0,
        collaboration_review_depth: 0,
        impact_services_touched: 0,
      };

  const components = [
    {
      name: 'Throughput',
      score: score.throughput_score,
      description: `${(rawMetrics.throughput_prs_per_week ?? 0).toFixed(1)} PRs/week`,
      color: 'blue',
    },
    {
      name: 'Quality',
      score: score.quality_score,
      description: `${(rawMetrics.quality_bug_rate ?? 0).toFixed(1)}% bug rate`,
      color: 'green',
    },
    {
      name: 'Speed',
      score: score.speed_score,
      description: `${(rawMetrics.speed_cycle_time_days ?? 0).toFixed(1)} day cycle time`,
      color: 'purple',
    },
    {
      name: 'Collaboration',
      score: score.collaboration_score,
      description: `${(rawMetrics.collaboration_reviews_given ?? 0).toFixed(0)} reviews given`,
      color: 'orange',
    },
    {
      name: 'Impact',
      score: score.impact_score,
      description: `${(rawMetrics.impact_services_touched ?? 0).toFixed(0)} services touched`,
      color: 'pink',
    },
  ];

  return (
    <div className="mt-4 space-y-3">
      <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">Score Breakdown</h4>
      {components.map((component) => (
        <div key={component.name} className="space-y-1">
          <div className="flex items-center justify-between text-sm">
            <span className="text-gray-700 dark:text-gray-300">{component.name}</span>
            <div className="flex items-center gap-2">
              <span className="text-gray-500 dark:text-gray-400 text-xs">
                {component.description}
              </span>
              <span className="font-semibold text-gray-900 dark:text-gray-100">
                {(component.score ?? 0).toFixed(0)}/100
              </span>
            </div>
          </div>
          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
            <div
              className={`bg-${component.color}-500 dark:bg-${component.color}-400 h-2 rounded-full transition-all`}
              style={{ width: `${Math.min(component.score, 100)}%` }}
            />
          </div>
        </div>
      ))}
    </div>
  );
}
