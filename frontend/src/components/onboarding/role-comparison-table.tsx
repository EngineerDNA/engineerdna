interface FeatureRow {
  feature: string;
  ic: boolean;
  manager: boolean;
  director: boolean;
  admin: boolean;
}

const features: FeatureRow[] = [
  {
    feature: 'Personal Performance Dashboard',
    ic: true,
    manager: true,
    director: true,
    admin: true,
  },
  {
    feature: 'My Activity Dashboard',
    ic: true,
    manager: true,
    director: true,
    admin: true,
  },
  {
    feature: 'Goal Tracking',
    ic: true,
    manager: true,
    director: true,
    admin: true,
  },
  {
    feature: 'Command Center',
    ic: false,
    manager: true,
    director: true,
    admin: true,
  },
  {
    feature: 'Weekly Team Briefing',
    ic: false,
    manager: true,
    director: true,
    admin: true,
  },
  {
    feature: 'Team Performance Dashboard',
    ic: false,
    manager: true,
    director: true,
    admin: true,
  },
  {
    feature: 'Sprint Planning',
    ic: false,
    manager: true,
    director: true,
    admin: true,
  },
  {
    feature: 'Team Management',
    ic: false,
    manager: true,
    director: true,
    admin: true,
  },
  {
    feature: 'Executive Overview',
    ic: false,
    manager: false,
    director: true,
    admin: true,
  },
  {
    feature: 'Organization Scorecard',
    ic: false,
    manager: false,
    director: true,
    admin: true,
  },
  {
    feature: 'Team Comparison',
    ic: false,
    manager: false,
    director: true,
    admin: true,
  },
  {
    feature: 'Cost & ROI Analysis',
    ic: false,
    manager: false,
    director: true,
    admin: true,
  },
  {
    feature: 'Capacity Planning',
    ic: false,
    manager: false,
    director: true,
    admin: true,
  },
  {
    feature: 'System Health Dashboard',
    ic: false,
    manager: false,
    director: false,
    admin: true,
  },
  {
    feature: 'Data Management Dashboard',
    ic: false,
    manager: false,
    director: false,
    admin: true,
  },
  {
    feature: 'Plugin Configuration',
    ic: false,
    manager: false,
    director: false,
    admin: true,
  },
];

function CheckIcon() {
  return (
    <svg
      className="w-4 h-4 text-green-600 dark:text-green-400"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
    </svg>
  );
}

function MinusIcon() {
  return (
    <svg
      className="w-4 h-4 text-gray-300 dark:text-gray-600"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 12H4" />
    </svg>
  );
}

export function RoleComparisonTable() {
  return (
    <div className="overflow-x-auto">
      <table className="min-w-full text-xs">
        <thead>
          <tr className="border-b border-gray-200 dark:border-gray-700">
            <th className="text-left py-2 pr-4 font-medium text-gray-900 dark:text-gray-100">
              Feature
            </th>
            <th className="text-center px-2 py-2 font-medium text-gray-900 dark:text-gray-100">
              IC
            </th>
            <th className="text-center px-2 py-2 font-medium text-gray-900 dark:text-gray-100">
              Manager
            </th>
            <th className="text-center px-2 py-2 font-medium text-gray-900 dark:text-gray-100">
              Director
            </th>
            <th className="text-center px-2 py-2 font-medium text-gray-900 dark:text-gray-100">
              Admin
            </th>
          </tr>
        </thead>
        <tbody>
          {features.map((row, idx) => (
            <tr key={idx} className="border-b border-gray-100 dark:border-gray-800 last:border-0">
              <td className="py-2 pr-4 text-gray-700 dark:text-gray-300">{row.feature}</td>
              <td className="text-center px-2 py-2">{row.ic ? <CheckIcon /> : <MinusIcon />}</td>
              <td className="text-center px-2 py-2">
                {row.manager ? <CheckIcon /> : <MinusIcon />}
              </td>
              <td className="text-center px-2 py-2">
                {row.director ? <CheckIcon /> : <MinusIcon />}
              </td>
              <td className="text-center px-2 py-2">{row.admin ? <CheckIcon /> : <MinusIcon />}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
