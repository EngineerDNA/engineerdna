import { lazy, Suspense, useState } from 'react';
import { BrowserRouter, Routes, Route, Link, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query';
import { ErrorBoundary } from 'react-error-boundary';
import { useTheme } from './contexts/ThemeContext';
import { OnboardingModal } from './components/onboarding-modal';
import { api } from './api/client';

const DashboardPage = lazy(() =>
  import('./pages/DashboardPage').then((module) => ({ default: module.DashboardPage }))
);
const TeamPage = lazy(() =>
  import('./pages/TeamPage').then((module) => ({ default: module.TeamPage }))
);
const PluginsPage = lazy(() =>
  import('./pages/PluginsPage').then((module) => ({ default: module.PluginsPage }))
);
const EventsPage = lazy(() =>
  import('./pages/EventsPage').then((module) => ({ default: module.EventsPage }))
);
const SchedulesPage = lazy(() =>
  import('./pages/SchedulesPage').then((module) => ({ default: module.SchedulesPage }))
);
const SettingsPage = lazy(() =>
  import('./pages/SettingsPage').then((module) => ({ default: module.SettingsPage }))
);
const ScorecardPage = lazy(() =>
  import('./pages/scorecard-page').then((module) => ({ default: module.ScorecardPage }))
);
const TeamsPage = lazy(() =>
  import('./pages/teams-page').then((module) => ({ default: module.TeamsPage }))
);
const TeamDetailPage = lazy(() =>
  import('./pages/team-detail-page').then((module) => ({ default: module.TeamDetailPage }))
);
const BriefingPage = lazy(() =>
  import('./pages/briefing-page').then((module) => ({ default: module.BriefingPage }))
);
const PlanningPage = lazy(() =>
  import('./pages/planning-page').then((module) => ({ default: module.PlanningPage }))
);
const AlertsPage = lazy(() =>
  import('./pages/alerts-page').then((module) => ({ default: module.AlertsPage }))
);
const AlertRulesPage = lazy(() =>
  import('./pages/alert-rules-page').then((module) => ({ default: module.AlertRulesPage }))
);
const LiveDashboardPage = lazy(() =>
  import('./pages/live-dashboard-page').then((module) => ({ default: module.LiveDashboardPage }))
);
const DashboardsPage = lazy(() =>
  import('./pages/dashboards-page').then((module) => ({ default: module.DashboardsPage }))
);

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30000, // 30 seconds for events
      refetchOnWindowFocus: false, // Disable refetch on window focus for better UX
      retry: 1, // Retry failed queries once
    },
  },
});

function ErrorFallback({
  error,
  resetErrorBoundary,
}: {
  error: Error;
  resetErrorBoundary: () => void;
}) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100 dark:bg-gray-900">
      <div className="bg-white dark:bg-gray-800 p-8 rounded-lg shadow-lg max-w-md">
        <h2 className="text-2xl font-bold text-red-600 dark:text-red-400 mb-4">
          Something went wrong
        </h2>
        <pre className="text-sm text-gray-700 dark:text-gray-300 mb-4 whitespace-pre-wrap">
          {error.message}
        </pre>
        <button
          onClick={resetErrorBoundary}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800"
        >
          Retry
        </button>
      </div>
    </div>
  );
}

function Navigation() {
  const { theme, toggleTheme } = useTheme();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const navLinks = [
    { to: '/', label: 'Dashboard' },
    { to: '/live', label: 'Live' },
    { to: '/dashboards', label: 'Dashboards' },
    { to: '/briefing', label: 'Briefing' },
    { to: '/alerts', label: 'Alerts' },
    { to: '/team', label: 'Team' },
    { to: '/teams', label: 'Teams' },
    { to: '/scorecard', label: 'Scorecard' },
    { to: '/planning', label: 'Planning' },
    { to: '/plugins', label: 'Plugins' },
    { to: '/events', label: 'Events' },
    { to: '/schedules', label: 'Schedules' },
    { to: '/settings', label: 'Settings' },
  ];

  return (
    <nav className="bg-white dark:bg-gray-800 shadow">
      <div className="max-w-7xl mx-auto px-4 py-4">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">EngineerDNA</h1>

          {/* Desktop Navigation */}
          <div className="hidden lg:flex items-center gap-4">
            {navLinks.map((link) => (
              <Link
                key={link.to}
                to={link.to}
                className="text-gray-700 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400 transition-colors whitespace-nowrap"
              >
                {link.label}
              </Link>
            ))}
            <button
              onClick={toggleTheme}
              className="p-2 rounded-lg bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors flex-shrink-0"
              aria-label="Toggle theme"
            >
              {theme === 'light' ? (
                <svg
                  className="w-5 h-5 text-gray-700 dark:text-gray-300"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
                  />
                </svg>
              ) : (
                <svg
                  className="w-5 h-5 text-gray-700 dark:text-gray-300"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
                  />
                </svg>
              )}
            </button>
          </div>

          {/* Mobile Menu Button */}
          <div className="lg:hidden flex items-center gap-2">
            <button
              onClick={toggleTheme}
              className="p-2 rounded-lg bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
              aria-label="Toggle theme"
            >
              {theme === 'light' ? (
                <svg
                  className="w-5 h-5 text-gray-700 dark:text-gray-300"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
                  />
                </svg>
              ) : (
                <svg
                  className="w-5 h-5 text-gray-700 dark:text-gray-300"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
                  />
                </svg>
              )}
            </button>
            <button
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              className="p-2 rounded-lg bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
              aria-label="Toggle menu"
            >
              {mobileMenuOpen ? (
                <svg
                  className="w-6 h-6 text-gray-700 dark:text-gray-300"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              ) : (
                <svg
                  className="w-6 h-6 text-gray-700 dark:text-gray-300"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 6h16M4 12h16M4 18h16"
                  />
                </svg>
              )}
            </button>
          </div>
        </div>

        {/* Mobile Navigation */}
        {mobileMenuOpen && (
          <div className="lg:hidden mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
            <div className="flex flex-col gap-2">
              {navLinks.map((link) => (
                <Link
                  key={link.to}
                  to={link.to}
                  onClick={() => setMobileMenuOpen(false)}
                  className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
                >
                  {link.label}
                </Link>
              ))}
            </div>
          </div>
        )}
      </div>
    </nav>
  );
}

function AppContent() {
  const [showOnboarding, setShowOnboarding] = useState(false);

  const { data: onboardingStatus, isLoading } = useQuery({
    queryKey: ['onboarding-status'],
    queryFn: api.getOnboardingStatus,
  });

  // Show onboarding modal if not completed
  const shouldShowOnboarding =
    !isLoading && onboardingStatus && !onboardingStatus.completed && !showOnboarding;

  return (
    <>
      <div className="min-h-screen bg-gray-100 dark:bg-gray-900 transition-colors">
        <Navigation />

        <main className="max-w-7xl mx-auto px-4 py-8">
          <Suspense
            fallback={
              <div className="flex items-center justify-center h-screen">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
              </div>
            }
          >
            <Routes>
              <Route path="/" element={<DashboardPage />} />
              <Route path="/live" element={<LiveDashboardPage />} />
              <Route path="/dashboards" element={<DashboardsPage />} />
              <Route path="/briefing" element={<BriefingPage />} />
              <Route path="/alerts" element={<AlertsPage />} />
              <Route path="/alerts/rules" element={<AlertRulesPage />} />
              <Route path="/alert-rules" element={<Navigate to="/alerts/rules" replace />} />
              <Route path="/team" element={<TeamPage />} />
              <Route path="/teams" element={<TeamsPage />} />
              <Route path="/teams/:id" element={<TeamDetailPage />} />
              <Route path="/scorecard" element={<ScorecardPage />} />
              <Route path="/planning" element={<PlanningPage />} />
              <Route path="/plugins" element={<PluginsPage />} />
              <Route path="/events" element={<EventsPage />} />
              <Route path="/schedules" element={<SchedulesPage />} />
              <Route path="/settings" element={<SettingsPage />} />
            </Routes>
          </Suspense>
        </main>
      </div>

      {shouldShowOnboarding && <OnboardingModal onComplete={() => setShowOnboarding(true)} />}
    </>
  );
}

export default function App() {
  return (
    <ErrorBoundary FallbackComponent={ErrorFallback}>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <AppContent />
        </BrowserRouter>
      </QueryClientProvider>
    </ErrorBoundary>
  );
}
