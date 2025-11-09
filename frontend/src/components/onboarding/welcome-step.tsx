interface WelcomeStepProps {
  onNext: () => void;
  onSkip: () => void;
}

export function WelcomeStep({ onNext, onSkip }: WelcomeStepProps) {
  return (
    <div className="space-y-6">
      <div className="text-center">
        <h2 className="text-3xl font-bold text-gray-900 dark:text-gray-100 mb-2">
          Welcome to EngineerDNA
        </h2>
        <p className="text-lg text-gray-600 dark:text-gray-400">
          AI Chief of Staff for engineering leaders
        </p>
      </div>

      <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
          What makes EngineerDNA different:
        </h3>
        <div className="space-y-3">
          <div className="flex items-start">
            <svg
              className="w-5 h-5 text-blue-600 dark:text-blue-400 mr-3 mt-0.5 flex-shrink-0"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
              />
            </svg>
            <div>
              <span className="font-medium text-gray-900 dark:text-gray-100">
                Privacy-first: Runs locally on your machine
              </span>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                All data stays on your computer. No cloud, no external servers.
              </p>
            </div>
          </div>
          <div className="flex items-start">
            <svg
              className="w-5 h-5 text-blue-600 dark:text-blue-400 mr-3 mt-0.5 flex-shrink-0"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"
              />
            </svg>
            <div>
              <span className="font-medium text-gray-900 dark:text-gray-100">
                Single binary for macOS, Linux, Windows
              </span>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                One cross-platform executable. No dependencies, no complex setup.
              </p>
            </div>
          </div>
          <div className="flex items-start">
            <svg
              className="w-5 h-5 text-blue-600 dark:text-blue-400 mr-3 mt-0.5 flex-shrink-0"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M11 4a2 2 0 114 0v1a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-1a2 2 0 100 4h1a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-1a2 2 0 10-4 0v1a1 1 0 01-1 1H7a1 1 0 01-1-1v-3a1 1 0 00-1-1H4a2 2 0 110-4h1a1 1 0 001-1V7a1 1 0 011-1h3a1 1 0 001-1V4z"
              />
            </svg>
            <div>
              <span className="font-medium text-gray-900 dark:text-gray-100">
                Extensible plugin architecture
              </span>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Connect to GitHub, Jira, Slack, and custom tools via plugins.
              </p>
            </div>
          </div>
          <div className="flex items-start">
            <svg
              className="w-5 h-5 text-blue-600 dark:text-blue-400 mr-3 mt-0.5 flex-shrink-0"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
              />
            </svg>
            <div>
              <span className="font-medium text-gray-900 dark:text-gray-100">
                100% open source, free forever
              </span>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Apache 2.0 licensed. No subscriptions, no usage limits.
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
        <h4 className="font-medium text-gray-900 dark:text-gray-100 mb-2">What you'll do next:</h4>
        <ol className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
          <li>1. Choose your role (IC, Manager, Director, or Admin)</li>
          <li>2. Optionally configure a data source</li>
          <li>3. Run your first sync and create dashboards</li>
        </ol>
      </div>

      <div className="flex justify-between pt-4">
        <button
          onClick={onSkip}
          className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 transition-colors"
        >
          Skip setup
        </button>
        <button
          onClick={onNext}
          className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-900"
        >
          Get Started
        </button>
      </div>
    </div>
  );
}
