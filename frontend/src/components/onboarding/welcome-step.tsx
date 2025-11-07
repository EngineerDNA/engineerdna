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
          Let's get you set up in just a few steps
        </p>
      </div>

      <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
          What you'll do:
        </h3>
        <ol className="space-y-3">
          <li className="flex items-start">
            <span className="flex-shrink-0 w-6 h-6 bg-blue-600 dark:bg-blue-500 text-white rounded-full flex items-center justify-center text-sm font-medium mr-3">
              1
            </span>
            <div>
              <span className="font-medium text-gray-900 dark:text-gray-100">Import your team</span>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Add team members from a CSV file or manually
              </p>
            </div>
          </li>
          <li className="flex items-start">
            <span className="flex-shrink-0 w-6 h-6 bg-blue-600 dark:bg-blue-500 text-white rounded-full flex items-center justify-center text-sm font-medium mr-3">
              2
            </span>
            <div>
              <span className="font-medium text-gray-900 dark:text-gray-100">
                Configure data sources
              </span>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Connect GitHub, Jira, or other tools to sync data
              </p>
            </div>
          </li>
          <li className="flex items-start">
            <span className="flex-shrink-0 w-6 h-6 bg-blue-600 dark:bg-blue-500 text-white rounded-full flex items-center justify-center text-sm font-medium mr-3">
              3
            </span>
            <div>
              <span className="font-medium text-gray-900 dark:text-gray-100">
                Run your first sync
              </span>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Pull in your engineering data and start analyzing
              </p>
            </div>
          </li>
        </ol>
      </div>

      <div className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
        <h4 className="font-medium text-gray-900 dark:text-gray-100 mb-2">Privacy First</h4>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          All your data stays on your machine. EngineerDNA runs locally and never sends data to
          external servers unless you explicitly configure export destinations.
        </p>
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
