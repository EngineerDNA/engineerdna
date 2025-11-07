import { Component } from 'react';
import type { ReactNode } from 'react';

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
  showDetails: boolean;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null, showDetails: false };
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    this.setState({ errorInfo });
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null, errorInfo: null, showDetails: false });
  };

  handleReload = () => {
    window.location.reload();
  };

  toggleDetails = () => {
    this.setState((prevState) => ({ showDetails: !prevState.showDetails }));
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 p-4">
          <div className="max-w-2xl w-full bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg p-6">
            <div className="flex items-start">
              <div className="flex-shrink-0">
                <svg
                  className="h-8 w-8 text-red-600 dark:text-red-400"
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
              <div className="ml-4 flex-1">
                <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100 mb-2">
                  Something went wrong
                </h2>
                <p className="text-gray-700 dark:text-gray-300 mb-4">
                  {this.state.error?.message || 'An unexpected error occurred'}
                </p>

                <div className="bg-gray-50 dark:bg-gray-900 p-4 rounded-lg mb-4">
                  <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
                    The application encountered an error and needs to recover. You can try:
                  </p>
                  <ul className="text-sm text-gray-700 dark:text-gray-300 list-disc list-inside space-y-1">
                    <li>Reloading the page</li>
                    <li>Going back to the previous page</li>
                    <li>Checking your internet connection</li>
                  </ul>
                </div>

                {this.state.errorInfo && (
                  <div className="mb-4">
                    <button
                      onClick={this.toggleDetails}
                      className="text-sm font-medium text-blue-600 dark:text-blue-400 underline hover:no-underline focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800"
                    >
                      {this.state.showDetails ? 'Hide technical details' : 'Show technical details'}
                    </button>
                    {this.state.showDetails && (
                      <div className="mt-2">
                        <pre className="text-xs text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-900 p-3 rounded overflow-x-auto border border-gray-200 dark:border-gray-700">
                          {this.state.error?.stack}
                          {'\n\n'}
                          {this.state.errorInfo.componentStack}
                        </pre>
                      </div>
                    )}
                  </div>
                )}

                <div className="flex gap-3">
                  <button
                    onClick={this.handleReload}
                    className="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800"
                  >
                    Reload Page
                  </button>
                  <button
                    onClick={this.handleReset}
                    className="px-4 py-2 bg-gray-600 text-white text-sm font-medium rounded hover:bg-gray-700 dark:bg-gray-700 dark:hover:bg-gray-800 transition-colors focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800"
                  >
                    Try Again
                  </button>
                  <button
                    onClick={() => window.history.back()}
                    className="px-4 py-2 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 text-sm font-medium rounded border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800"
                  >
                    Go Back
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
