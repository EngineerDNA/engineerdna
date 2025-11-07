import { useState } from 'react';
import { ApiError } from '../api/client';

interface ErrorAlertProps {
  error: Error | ApiError | null;
  onRetry?: () => void;
  onDismiss?: () => void;
  className?: string;
}

export function ErrorAlert({ error, onRetry, onDismiss, className = '' }: ErrorAlertProps) {
  const [showDetails, setShowDetails] = useState(false);

  if (!error) return null;

  const isApiError = error instanceof ApiError;
  const errorCode = isApiError ? error.code : 'unknown';
  const userMessage = isApiError ? error.user_message : error.message;
  const suggestions = isApiError ? error.suggestions : undefined;
  const docsUrl = isApiError ? error.docs_url : undefined;
  const context = isApiError ? error.context : undefined;

  const getSeverityStyles = () => {
    switch (errorCode) {
      case 'auth_failed':
      case 'permission_denied':
        return {
          bg: 'bg-red-50 dark:bg-red-900/20',
          border: 'border-red-200 dark:border-red-800',
          text: 'text-red-800 dark:text-red-200',
          icon: 'text-red-600 dark:text-red-400',
        };
      case 'rate_limit':
      case 'timeout':
        return {
          bg: 'bg-yellow-50 dark:bg-yellow-900/20',
          border: 'border-yellow-200 dark:border-yellow-800',
          text: 'text-yellow-800 dark:text-yellow-200',
          icon: 'text-yellow-600 dark:text-yellow-400',
        };
      case 'network_error':
      case 'config_invalid':
        return {
          bg: 'bg-orange-50 dark:bg-orange-900/20',
          border: 'border-orange-200 dark:border-orange-800',
          text: 'text-orange-800 dark:text-orange-200',
          icon: 'text-orange-600 dark:text-orange-400',
        };
      default:
        return {
          bg: 'bg-red-50 dark:bg-red-900/20',
          border: 'border-red-200 dark:border-red-800',
          text: 'text-red-800 dark:text-red-200',
          icon: 'text-red-600 dark:text-red-400',
        };
    }
  };

  const styles = getSeverityStyles();

  return (
    <div className={`${styles.bg} border ${styles.border} rounded-lg p-4 ${className}`}>
      <div className="flex items-start">
        <div className="flex-shrink-0">
          <svg
            className={`h-6 w-6 ${styles.icon}`}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
        </div>
        <div className="ml-3 flex-1">
          <h3 className={`text-sm font-medium ${styles.text}`}>Error</h3>
          <div className={`mt-2 text-sm ${styles.text}`}>
            <p>{userMessage}</p>
          </div>

          {suggestions && suggestions.length > 0 && (
            <div className="mt-3">
              <p className={`text-sm font-medium ${styles.text}`}>Suggestions:</p>
              <ul className={`mt-2 text-sm ${styles.text} list-disc list-inside space-y-1`}>
                {suggestions.map((suggestion, index) => (
                  <li key={index}>{suggestion}</li>
                ))}
              </ul>
            </div>
          )}

          {context && Object.keys(context).length > 0 && (
            <div className="mt-3">
              <button
                onClick={() => setShowDetails(!showDetails)}
                className={`text-sm font-medium ${styles.text} underline hover:no-underline focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 dark:focus:ring-offset-gray-800`}
              >
                {showDetails ? 'Hide details' : 'Show details'}
              </button>
              {showDetails && (
                <pre
                  className={`mt-2 text-xs ${styles.text} bg-black/5 dark:bg-black/30 p-2 rounded overflow-x-auto`}
                >
                  {JSON.stringify(context, null, 2)}
                </pre>
              )}
            </div>
          )}

          <div className="mt-4 flex gap-2">
            {onRetry && (
              <button
                onClick={onRetry}
                className="px-3 py-2 bg-blue-600 text-white text-sm font-medium rounded hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800"
              >
                Retry
              </button>
            )}
            {docsUrl && (
              <a
                href={docsUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="px-3 py-2 bg-gray-600 text-white text-sm font-medium rounded hover:bg-gray-700 dark:bg-gray-700 dark:hover:bg-gray-800 transition-colors focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800"
              >
                View Documentation
              </a>
            )}
            {onDismiss && (
              <button
                onClick={onDismiss}
                className={`px-3 py-2 ${styles.bg} ${styles.text} text-sm font-medium rounded border ${styles.border} hover:bg-opacity-75 dark:hover:bg-opacity-75 transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 dark:focus:ring-offset-gray-800`}
              >
                Dismiss
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
