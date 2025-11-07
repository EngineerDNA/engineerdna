import { useState } from 'react';
import { ImportTeamModal } from '../ImportTeamModal';

interface TeamImportStepProps {
  onNext: () => void;
  onSkip: () => void;
  onBack: () => void;
}

export function TeamImportStep({ onNext, onSkip, onBack }: TeamImportStepProps) {
  const [showImportModal, setShowImportModal] = useState(false);

  const handleImportComplete = () => {
    setShowImportModal(false);
    onNext();
  };

  return (
    <>
      <div className="space-y-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">
            Import Your Team
          </h2>
          <p className="text-gray-600 dark:text-gray-400">
            Add your team members so EngineerDNA can match data to the right people
          </p>
        </div>

        <div className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <div className="space-y-4">
            <div>
              <h3 className="font-medium text-gray-900 dark:text-gray-100 mb-2">
                Why import team members?
              </h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                EngineerDNA uses team data to match identities across different tools (GitHub, Jira,
                Slack, etc.). This ensures accurate metrics and insights.
              </p>
            </div>

            <div>
              <h4 className="text-sm font-medium text-gray-900 dark:text-gray-100 mb-2">
                What you'll need:
              </h4>
              <ul className="text-sm text-gray-600 dark:text-gray-400 space-y-1 list-disc list-inside">
                <li>Team member names (required)</li>
                <li>Email addresses (recommended)</li>
                <li>GitHub usernames, Jira emails, or other identifiers (optional)</li>
              </ul>
            </div>
          </div>
        </div>

        <div className="flex gap-3">
          <button
            onClick={() => setShowImportModal(true)}
            className="flex-1 px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-900 font-medium"
          >
            Import from CSV
          </button>
          <button
            onClick={onNext}
            className="flex-1 px-6 py-3 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors font-medium"
          >
            Add Manually Later
          </button>
        </div>

        <div className="flex justify-between pt-4 border-t border-gray-200 dark:border-gray-700">
          <button
            onClick={onBack}
            className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 transition-colors"
          >
            Back
          </button>
          <button
            onClick={onSkip}
            className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 transition-colors"
          >
            Skip this step
          </button>
        </div>
      </div>

      {showImportModal && <ImportTeamModal onClose={handleImportComplete} />}
    </>
  );
}
