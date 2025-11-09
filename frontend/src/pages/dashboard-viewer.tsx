import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import type { Layout } from 'react-grid-layout';
import { dashboardsApi } from '../api/dashboards';
import { useDashboardMutations } from '../hooks/useDashboardMutations';
import { useWidgetData } from '../hooks/useWidgetData';
import { DashboardGrid } from '../components/dashboards/dashboard-grid';
import { DashboardToolbar } from '../components/dashboards/dashboard-toolbar';
import { AddWidgetModal } from '../components/dashboards/add-widget-modal';
import { EditDashboardModal } from '../components/dashboards/edit-dashboard-modal';
import {
  TeamDetailModal,
  EngineerDetailDrawer,
  SprintDetailModal,
  AlertDetailModal,
  EventDetailDrawer,
  GoalDetailModal,
} from '../components/modals';
import type { Widget } from '../api/types';

interface ModalState {
  team: { isOpen: boolean; teamId: string | null };
  engineer: { isOpen: boolean; engineerId: string | null };
  sprint: { isOpen: boolean; sprintId: string | null };
  alert: { isOpen: boolean; alertId: string | null };
  event: { isOpen: boolean; eventId: string | null };
  goal: { isOpen: boolean; goalId: string | null };
}

export function DashboardViewer() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [isEditing, setIsEditing] = useState(false);
  const [showAddWidget, setShowAddWidget] = useState(false);
  const [showEditDashboard, setShowEditDashboard] = useState(false);
  const [modals, setModals] = useState<ModalState>({
    team: { isOpen: false, teamId: null },
    engineer: { isOpen: false, engineerId: null },
    sprint: { isOpen: false, sprintId: null },
    alert: { isOpen: false, alertId: null },
    event: { isOpen: false, eventId: null },
    goal: { isOpen: false, goalId: null },
  });

  const {
    data: dashboard,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['dashboard', id],
    queryFn: () => dashboardsApi.getDashboard(id!),
    enabled: !!id,
  });

  const { data: dashboardsData } = useQuery({
    queryKey: ['dashboards'],
    queryFn: () => dashboardsApi.getDashboards(),
  });

  const { updateDashboard, deleteDashboard, cloneDashboard } = useDashboardMutations();

  const dashboards = dashboardsData?.dashboards || [];

  // Fetch data for all widgets on current dashboard
  const widgetDataQueries = useWidgetData(dashboard);

  // Helper to serialize dashboard layout for backend
  const serializeDashboard = (widgets: Widget[]) => {
    const layout = JSON.stringify({ widgets });
    return { layout };
  };

  const handleLayoutChange = (layout: Layout[]) => {
    if (!dashboard) return;

    const updatedWidgets = dashboard.widgets.map((widget: Widget) => {
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
      id: dashboard.id,
      data: {
        layout: serializeDashboard(updatedWidgets).layout,
        name: dashboard.name,
      },
    });
  };

  const handleAddWidget = (newWidget: Omit<Widget, 'id'>) => {
    if (!dashboard) return;

    const widget: Widget = {
      ...newWidget,
      id: crypto.randomUUID(),
    };

    const updatedWidgets = [...dashboard.widgets, widget];
    updateDashboard.mutate({
      id: dashboard.id,
      data: {
        layout: serializeDashboard(updatedWidgets).layout,
        name: dashboard.name,
      },
    });
  };

  const handleWidgetDelete = (widgetId: string) => {
    if (!dashboard) return;

    const updatedWidgets = dashboard.widgets.filter((w: Widget) => w.id !== widgetId);
    updateDashboard.mutate({
      id: dashboard.id,
      data: {
        layout: serializeDashboard(updatedWidgets).layout,
        name: dashboard.name,
      },
    });
  };

  const handleEditDashboard = (data: { name: string; description?: string }) => {
    if (!dashboard) return;
    updateDashboard.mutate({
      id: dashboard.id,
      data: {
        name: data.name,
        description: data.description,
        layout: dashboard.layout,
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
          navigate(`/dashboards/${newDashboard.id}`);
        },
      }
    );
  };

  const handleDeleteDashboard = (dashboardId: string) => {
    deleteDashboard.mutate(dashboardId, {
      onSuccess: () => {
        navigate('/dashboards');
      },
    });
  };

  const handleDashboardChange = (newDashboardId: string | null) => {
    if (newDashboardId) {
      navigate(`/dashboards/${newDashboardId}`);
    }
  };

  const handleOpenModal = (modalType: string, data: any) => {
    switch (modalType) {
      case 'team':
      case 'team-detail':
        setModals((prev) => ({
          ...prev,
          team: { isOpen: true, teamId: data.teamId || data.team_id },
        }));
        break;
      case 'engineer':
      case 'engineer-detail':
        setModals((prev) => ({
          ...prev,
          engineer: { isOpen: true, engineerId: data.engineerId || data.engineer_id },
        }));
        break;
      case 'sprint':
      case 'sprint-detail':
        setModals((prev) => ({
          ...prev,
          sprint: { isOpen: true, sprintId: data.sprintId || data.sprint_id },
        }));
        break;
      case 'alert':
      case 'alert-detail':
        setModals((prev) => ({
          ...prev,
          alert: { isOpen: true, alertId: data.alertId || data.alert_id },
        }));
        break;
      case 'event':
      case 'event-detail':
        setModals((prev) => ({
          ...prev,
          event: { isOpen: true, eventId: data.eventId || data.event_id },
        }));
        break;
      case 'goal':
      case 'goal-detail':
        setModals((prev) => ({
          ...prev,
          goal: { isOpen: true, goalId: data.goalId || data.goal_id },
        }));
        break;
      default:
        console.warn(`Unknown modal type: ${modalType}`);
    }
  };

  const handleCloseModal = (modalType: keyof ModalState) => {
    setModals((prev) => ({
      ...prev,
      [modalType]: {
        isOpen: false,
        [`${modalType}Id`]:
          prev[modalType][`${modalType}Id` as keyof (typeof prev)[typeof modalType]],
      },
    }));
  };

  const handleUpdateWidget = (widgetId: string, updates: Record<string, any>) => {
    if (!dashboard) return;

    const updatedWidgets = dashboard.widgets.map((w: Widget) =>
      w.id === widgetId ? { ...w, config: { ...w.config, ...updates } } : w
    );

    updateDashboard.mutate({
      id: dashboard.id,
      data: {
        layout: serializeDashboard(updatedWidgets).layout,
        name: dashboard.name,
      },
    });
  };

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
        <h2 className="text-xl font-bold text-red-900 dark:text-red-100 mb-2">Error</h2>
        <p className="text-red-700 dark:text-red-300">
          {error instanceof Error ? error.message : 'Dashboard not found'}
        </p>
        <button
          onClick={() => navigate('/dashboards')}
          className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800"
        >
          Back to Dashboards
        </button>
      </div>
    );
  }

  if (!dashboard) {
    return (
      <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
        <h2 className="text-xl font-bold text-yellow-900 dark:text-yellow-100 mb-2">
          Dashboard not found
        </h2>
        <button
          onClick={() => navigate('/dashboards')}
          className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800"
        >
          Back to Dashboards
        </button>
      </div>
    );
  }

  return (
    <div>
      <DashboardToolbar
        dashboards={dashboards}
        currentDashboard={dashboard}
        currentDashboardId={dashboard.id}
        isEditing={isEditing}
        onDashboardChange={handleDashboardChange}
        onToggleEdit={() => setIsEditing(!isEditing)}
        onShowAddWidget={() => setShowAddWidget(true)}
        onShowEditDashboard={() => setShowEditDashboard(true)}
        onShowCreateDashboard={() => navigate('/dashboards')}
        onCloneDashboard={handleCloneDashboard}
        onDeleteDashboard={handleDeleteDashboard}
      />

      <DashboardGrid
        widgets={dashboard.widgets}
        widgetData={widgetDataQueries.data || new Map()}
        isEditing={isEditing}
        onLayoutChange={handleLayoutChange}
        onWidgetDelete={handleWidgetDelete}
        onOpenModal={handleOpenModal}
        onUpdateWidget={handleUpdateWidget}
      />

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
          dashboard ? { name: dashboard.name, description: dashboard.description } : undefined
        }
      />

      {modals.team.isOpen && modals.team.teamId && (
        <TeamDetailModal
          team_id={modals.team.teamId}
          isOpen={modals.team.isOpen}
          onClose={() => handleCloseModal('team')}
        />
      )}

      {modals.engineer.isOpen && modals.engineer.engineerId && (
        <EngineerDetailDrawer
          engineer_id={modals.engineer.engineerId}
          isOpen={modals.engineer.isOpen}
          onClose={() => handleCloseModal('engineer')}
        />
      )}

      {modals.sprint.isOpen && modals.sprint.sprintId && (
        <SprintDetailModal
          sprint_id={modals.sprint.sprintId}
          isOpen={modals.sprint.isOpen}
          onClose={() => handleCloseModal('sprint')}
        />
      )}

      {modals.alert.isOpen && modals.alert.alertId && (
        <AlertDetailModal
          alert_id={modals.alert.alertId}
          isOpen={modals.alert.isOpen}
          onClose={() => handleCloseModal('alert')}
        />
      )}

      {modals.event.isOpen && modals.event.eventId && (
        <EventDetailDrawer
          event_id={modals.event.eventId}
          isOpen={modals.event.isOpen}
          onClose={() => handleCloseModal('event')}
        />
      )}

      {modals.goal.isOpen && modals.goal.goalId && (
        <GoalDetailModal
          goal_id={modals.goal.goalId}
          isOpen={modals.goal.isOpen}
          onClose={() => handleCloseModal('goal')}
        />
      )}
    </div>
  );
}
