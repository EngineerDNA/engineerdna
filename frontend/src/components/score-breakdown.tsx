interface ScoreBreakdownProps {
  score: {
    total_score?: number;
    throughput_score?: number;
    quality_score?: number;
    speed_score?: number;
    collaboration_score?: number;
    impact_score?: number;
  };
}

export function ScoreBreakdown({ score }: ScoreBreakdownProps) {
  // Score components - raw_metrics removed in PDR-9 schema migration
  // Description now shows just the score value
  const components = [
    {
      name: 'Throughput',
      score: score.throughput_score ?? 0,
      description: `Score: ${(score.throughput_score ?? 0).toFixed(0)}`,
      color: 'blue',
    },
    {
      name: 'Quality',
      score: score.quality_score ?? 0,
      description: `Score: ${(score.quality_score ?? 0).toFixed(0)}`,
      color: 'green',
    },
    {
      name: 'Speed',
      score: score.speed_score ?? 0,
      description: `Score: ${(score.speed_score ?? 0).toFixed(0)}`,
      color: 'purple',
    },
    {
      name: 'Collaboration',
      score: score.collaboration_score ?? 0,
      description: `Score: ${(score.collaboration_score ?? 0).toFixed(0)}`,
      color: 'orange',
    },
    {
      name: 'Impact',
      score: score.impact_score ?? 0,
      description: `Score: ${(score.impact_score ?? 0).toFixed(0)}`,
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
              style={{ width: `${Math.min(component.score ?? 0, 100)}%` }}
            />
          </div>
        </div>
      ))}
    </div>
  );
}
