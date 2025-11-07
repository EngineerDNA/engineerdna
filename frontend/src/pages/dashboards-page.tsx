import { useState, useEffect } from 'react';
import type { Layout } from 'react-grid-layout';
import { useDashboards } from '../hooks/useDashboards';
import { useDashboardMutations } from '../hooks/useDashboardMutations';
import { useWidgetData } from '../hooks/useWidgetData';
import { DashboardGrid } from '../components/dashboards/dashboard-grid';
import { DashboardToolbar } from '../components/dashboards/dashboard-toolbar';
import { AddWidgetModal } from '../components/dashboards/add-widget-modal';
import { EditDashboardModal } from '../components/dashboards/edit-dashboard-modal';
import type { Widget } from '../api/types';

export function DashboardsPage() {
  const [currentDashboardId, setCurrentDashboardId] = useState<string | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [showAddWidget, setShowAddWidget] = useState(false);
  const [showEditDashboard, setShowEditDashboard] = useState(false);
  const [showCreateDashboard, setShowCreateDashboard] = useState(false);

  const { data: dashboardsData, isLoading: dashboardsLoading } = useDashboards();
  const { createDashboard, updateDashboard, deleteDashboard, cloneDashboard } =
    useDashboardMutations();

  const dashboards = dashboardsData?.dashboards || [];
  const currentDashboard = dashboards.find((d: { id: string }) => d.id === currentDashboardId);

  // Set first dashboard as current if none selected
  useEffect(() => {
    if (!currentDashboardId && dashboards.length > 0) {
      setCurrentDashboardId(dashboards[0].id);
    }
  }, [dashboards, currentDashboardId]);

  // Fetch data for all widgets on current dashboard
  const widgetDataQueries = useWidgetData(currentDashboard);

  // Helper to serialize dashboard layout for backend
  const serializeDashboard = (widgets: Widget[]) => {
    const layout = JSON.stringify({ widgets });
    return { layout };
  };

  const handleLayoutChange = (layout: Layout[]) => {
    if (!currentDashboard) return;

    const updatedWidgets = currentDashboard.widgets.map((widget: Widget) => {
      const layoutItem = layout.find((l) => l.i === widget.id);
      if (layoutItem) {
        return {
          ...widget,
          position: {
            x: layoutItem.x,
            y: layoutItem.y,
            w: layoutItem.w,
            h: layoutItem.h,
          },
        };
      }
      return widget;
    });

    updateDashboard.mutate({
      id: currentDashboard.id,
      data: {
        layout: serializeDashboard(updatedWidgets).layout,
        name: currentDashboard.name,
      },
    });
  };

  const handleAddWidget = (newWidget: Omit<Widget, 'id'>) => {
    if (!currentDashboard) return;

    const widget: Widget = {
      ...newWidget,
      id: crypto.randomUUID(),
    };

    const updatedWidgets = [...currentDashboard.widgets, widget];
    updateDashboard.mutate({
      id: currentDashboard.id,
      data: {
        layout: serializeDashboard(updatedWidgets).layout,
        name: currentDashboard.name,
      },
    });
  };

  const handleWidgetDelete = (widgetId: string) => {
    if (!currentDashboard) return;

    const updatedWidgets = currentDashboard.widgets.filter((w: Widget) => w.id !== widgetId);
    updateDashboard.mutate({
      id: currentDashboard.id,
      data: {
        layout: serializeDashboard(updatedWidgets).layout,
        name: currentDashboard.name,
      },
    });
  };

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
          setCurrentDashboardId(newDashboard.id);
        },
      }
    );
  };

  const handleEditDashboard = (data: { name: string; description?: string }) => {
    if (!currentDashboard) return;
    updateDashboard.mutate({
      id: currentDashboard.id,
      data: {
        name: data.name,
        description: data.description,
        layout: currentDashboard.layout, // Keep existing layout
      },
    });
  };

  const handleCloneDashboard = (dashboardId: string) => {
    const sourceDashboard = dashboards.find(
      (d: { id: string; name: string }) => d.id === dashboardId
    );
    const newName = sourceDashboard ? `${sourceDashboard.name} (Copy)` : 'Dashboard Copy';

    cloneDashboard.mutate(
      { id: dashboardId, name: newName },
      {
        onSuccess: (newDashboard) => {
          setCurrentDashboardId(newDashboard.id);
        },
      }
    );
  };

  const handleDeleteDashboard = (dashboardId: string) => {
    deleteDashboard.mutate(dashboardId, {
      onSuccess: () => {
        setCurrentDashboardId(null);
      },
    });
  };

  if (dashboardsLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
      </div>
    );
  }

  return (
    <div>
      <DashboardToolbar
        dashboards={dashboards}
        currentDashboard={currentDashboard}
        currentDashboardId={currentDashboardId}
        isEditing={isEditing}
        onDashboardChange={setCurrentDashboardId}
        onToggleEdit={() => setIsEditing(!isEditing)}
        onShowAddWidget={() => setShowAddWidget(true)}
        onShowEditDashboard={() => setShowEditDashboard(true)}
        onShowCreateDashboard={() => setShowCreateDashboard(true)}
        onCloneDashboard={handleCloneDashboard}
        onDeleteDashboard={handleDeleteDashboard}
      />

      {currentDashboard ? (
        <DashboardGrid
          widgets={currentDashboard.widgets}
          widgetData={widgetDataQueries.data || new Map()}
          isEditing={isEditing}
          onLayoutChange={handleLayoutChange}
          onWidgetDelete={handleWidgetDelete}
        />
      ) : (
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
      )}

      <AddWidgetModal
        isOpen={showAddWidget}
        onClose={() => setShowAddWidget(false)}
        onAdd={handleAddWidget}
      />

      <EditDashboardModal
        isOpen={showEditDashboard}
        onClose={() => setShowEditDashboard(false)}
        onSave={handleEditDashboard}
        initialData={
          currentDashboard
            ? { name: currentDashboard.name, description: currentDashboard.description }
            : undefined
        }
      />

      <EditDashboardModal
        isOpen={showCreateDashboard}
        onClose={() => setShowCreateDashboard(false)}
        onSave={handleCreateDashboard}
      />
    </div>
  );
}
