import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDashboards } from '../hooks/useDashboards';
import { usePrimaryDashboard } from '../hooks/useSettings';
import { useDashboardMutations } from '../hooks/useDashboardMutations';
import { EditDashboardModal } from '../components/dashboards/edit-dashboard-modal';

export function DashboardsPage() {
  const navigate = useNavigate();
  const { data: primaryData } = usePrimaryDashboard();
  const { data: dashboardsData, isLoading: dashboardsLoading } = useDashboards();
  const { createDashboard } = useDashboardMutations();
  const [showCreateDashboard, setShowCreateDashboard] = useState(false);

  const dashboards = dashboardsData?.dashboards || [];

  // Auto-redirect to primary dashboard if set, or first dashboard if available
  useEffect(() => {
    if (primaryData?.primary_dashboard_id) {
      navigate(`/dashboards/${primaryData.primary_dashboard_id}`, { replace: true });
    } else if (dashboards.length > 0) {
      // If no primary, redirect to first dashboard
      navigate(`/dashboards/${dashboards[0].id}`, { replace: true });
    }
  }, [primaryData, dashboards, navigate]);

  const handleCreateDashboard = (data: { name: string; description?: string }) => {
    createDashboard.mutate(
      {
        name: data.name,
        description: data.description,
        layout: JSON.stringify({ widgets: [] }),
        is_template: false,
      },
      {
        onSuccess: (newDashboard) => {
          navigate(`/dashboards/${newDashboard.id}`);
        },
      }
    );
  };

  if (dashboardsLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
      </div>
    );
  }

  // If no dashboards exist, show the "create dashboard" screen
  if (dashboards.length === 0) {
    return (
      <div className="container mx-auto py-8">
        <div className="text-center py-12 bg-white dark:bg-gray-800 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">
            No Dashboards
          </h2>
          <p className="text-gray-600 dark:text-gray-400 mb-6">
            Create your first dashboard to get started
          </p>
          <button
            onClick={() => setShowCreateDashboard(true)}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800"
          >
            Create Dashboard
          </button>
        </div>

        <EditDashboardModal
          isOpen={showCreateDashboard}
          onClose={() => setShowCreateDashboard(false)}
          onSave={handleCreateDashboard}
        />
      </div>
    );
  }

  // While redirecting to dashboard, show loading
  return (
    <div className="flex justify-center items-center h-64">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
    </div>
  );
}
