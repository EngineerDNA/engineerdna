import { useState } from 'react';
import { usePrimaryDashboard, useUpdatePrimaryDashboard } from '../../hooks/useSettings';
import { useDashboards } from '../../hooks/useDashboards';
import {
  useDashboardTemplates,
  useCreateDashboardFromTemplate,
} from '../../hooks/useDashboardTemplates';
import type { DashboardTemplate } from '../../api/types';

export function DashboardsTab() {
  const { data: primaryData, isLoading: primaryLoading } = usePrimaryDashboard();
  const { data: dashboardsData, isLoading: dashboardsLoading } = useDashboards();
  const { data: templatesData, isLoading: templatesLoading } = useDashboardTemplates();
  const updatePrimaryMutation = useUpdatePrimaryDashboard();
  const createFromTemplateMutation = useCreateDashboardFromTemplate();

  const [selectedPrimary, setSelectedPrimary] = useState<string | null>(null);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [showTemplateModal, setShowTemplateModal] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<DashboardTemplate | null>(null);
  const [newDashboardName, setNewDashboardName] = useState('');

  const dashboards = dashboardsData?.dashboards || [];
  const templates = templatesData?.templates || [];
  const primaryDashboardId = selectedPrimary || primaryData?.primary_dashboard_id || null;

  const categories = Array.from(new Set(templates.map((t) => t.category).filter(Boolean)));

  const filteredTemplates =
    selectedCategory === 'all'
      ? templates
      : templates.filter((t) => t.category === selectedCategory);

  const handleSetPrimary = () => {
    if (selectedPrimary && selectedPrimary !== primaryData?.primary_dashboard_id) {
      updatePrimaryMutation.mutate(selectedPrimary);
    }
  };

  const handleCreateFromTemplate = () => {
    if (selectedTemplate) {
      createFromTemplateMutation.mutate(
        { templateId: selectedTemplate.id, name: newDashboardName || undefined },
        {
          onSuccess: () => {
            setShowTemplateModal(false);
            setSelectedTemplate(null);
            setNewDashboardName('');
          },
        }
      );
    }
  };

  const hasChanges = selectedPrimary && selectedPrimary !== primaryData?.primary_dashboard_id;

  if (primaryLoading || dashboardsLoading || templatesLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 animate-pulse">
        <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-32 mb-4"></div>
        <div className="space-y-3">
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Success Message */}
      {updatePrimaryMutation.isSuccess && (
        <div className="p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
          <p className="text-green-800 dark:text-green-200">
            Primary dashboard updated successfully.
          </p>
        </div>
      )}

      {/* Primary Dashboard Selector */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
          Primary Dashboard
        </h2>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          Your primary dashboard loads automatically when you visit the Dashboards page.
        </p>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Select Primary Dashboard
            </label>
            <select
              value={primaryDashboardId || ''}
              onChange={(e) => setSelectedPrimary(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="">No primary dashboard</option>
              {dashboards.map((dashboard) => (
                <option key={dashboard.id} value={dashboard.id}>
                  {dashboard.name}
                </option>
              ))}
            </select>
          </div>

          {hasChanges && (
            <div className="flex gap-2">
              <button
                onClick={handleSetPrimary}
                disabled={updatePrimaryMutation.isPending}
                className="px-6 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {updatePrimaryMutation.isPending ? 'Saving...' : 'Save Changes'}
              </button>
              <button
                onClick={() => setSelectedPrimary(null)}
                className="px-6 py-2 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
              >
                Cancel
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Favorite Dashboards */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
          Your Dashboards
        </h2>

        {dashboards.length === 0 ? (
          <div className="text-center py-8 text-gray-500 dark:text-gray-400">
            <p>No dashboards yet. Create one from a template below.</p>
          </div>
        ) : (
          <div className="space-y-2">
            {dashboards.map((dashboard) => (
              <div
                key={dashboard.id}
                className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-900 rounded-lg"
              >
                <div>
                  <p className="font-medium text-gray-900 dark:text-gray-100">{dashboard.name}</p>
                  {dashboard.description && (
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      {dashboard.description}
                    </p>
                  )}
                </div>
                {dashboard.id === primaryDashboardId && (
                  <span className="px-3 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300 text-sm rounded-full">
                    Primary
                  </span>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Template Browser */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
            Dashboard Templates
          </h2>
          <select
            value={selectedCategory}
            onChange={(e) => setSelectedCategory(e.target.value)}
            className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="all">All Categories</option>
            {categories.map((category) => (
              <option key={category} value={category}>
                {category}
              </option>
            ))}
          </select>
        </div>

        {filteredTemplates.length === 0 ? (
          <div className="text-center py-8 text-gray-500 dark:text-gray-400">
            <p>No templates available.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {filteredTemplates.map((template) => (
              <div
                key={template.id}
                className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:border-blue-500 dark:hover:border-blue-500 transition-colors"
              >
                <div className="flex items-start justify-between mb-2">
                  <h3 className="font-medium text-gray-900 dark:text-gray-100">{template.name}</h3>
                  {template.is_system && (
                    <span className="px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 text-xs rounded">
                      System
                    </span>
                  )}
                </div>
                {template.description && (
                  <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
                    {template.description}
                  </p>
                )}
                <div className="flex items-center justify-between">
                  {template.role && (
                    <span className="px-2 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300 text-xs rounded">
                      {template.role.toUpperCase()}
                    </span>
                  )}
                  <button
                    onClick={() => {
                      setSelectedTemplate(template);
                      setNewDashboardName(template.name);
                      setShowTemplateModal(true);
                    }}
                    className="px-3 py-1 bg-blue-600 dark:bg-blue-700 text-white text-sm rounded hover:bg-blue-700 dark:hover:bg-blue-800 transition-colors"
                  >
                    Create Dashboard
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Create Dashboard Modal */}
      {showTemplateModal && selectedTemplate && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg max-w-md w-full p-6">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">
              Create Dashboard from Template
            </h2>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Template
                </label>
                <p className="text-gray-900 dark:text-gray-100">{selectedTemplate.name}</p>
                {selectedTemplate.description && (
                  <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                    {selectedTemplate.description}
                  </p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Dashboard Name
                </label>
                <input
                  type="text"
                  value={newDashboardName}
                  onChange={(e) => setNewDashboardName(e.target.value)}
                  placeholder="My Dashboard"
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400"
                />
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  Leave blank to use template name
                </p>
              </div>

              <div className="flex justify-end gap-2 pt-4">
                <button
                  onClick={() => {
                    setShowTemplateModal(false);
                    setSelectedTemplate(null);
                    setNewDashboardName('');
                  }}
                  className="px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={handleCreateFromTemplate}
                  disabled={createFromTemplateMutation.isPending}
                  className="px-4 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded hover:bg-blue-700 dark:hover:bg-blue-800 disabled:opacity-50 transition-colors"
                >
                  {createFromTemplateMutation.isPending ? 'Creating...' : 'Create Dashboard'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
