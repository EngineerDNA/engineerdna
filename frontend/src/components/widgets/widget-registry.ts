import { BriefingTldrWidget } from './briefing-tldr-widget';
import { AttentionItemsWidget } from './attention-items-widget';
import { TalkingPointsWidget } from './talking-points-widget';
import { TeamSelectorWidget } from './team-selector-widget';
import { TeamMemberCardsWidget } from './team-member-cards-widget';
import { EngineerTableWidget } from './engineer-table-widget';
import { UnresolvedIdentitiesWidget } from './unresolved-identities-widget';
import { SprintBoardWidget } from './sprint-board-widget';
import { VelocityTrendWidget } from './velocity-trend-widget';
import { TimelineEstimatorWidget } from './timeline-estimator-widget';
import { PluginListWidget } from './plugin-list-widget';
import { EventStreamWidget } from './event-stream-widget';
import { AlertListWidget } from './alert-list-widget';
import { GoalTrackerWidget } from './goal-tracker-widget';
import { CostBreakdownWidget } from './cost-breakdown-widget';
import { QuickActionsWidget } from './quick-actions-widget';

export interface WidgetProps {
  widgetId: string;
  config: Record<string, any>;
  onOpenModal?: (modalType: string, data: any) => void;
  onUpdateWidget?: (widgetId: string, updates: Record<string, any>) => void;
}

const widgetRegistry = {
  briefing_tldr: BriefingTldrWidget,
  attention_items: AttentionItemsWidget,
  talking_points: TalkingPointsWidget,
  team_selector: TeamSelectorWidget,
  team_member_cards: TeamMemberCardsWidget,
  engineer_table: EngineerTableWidget,
  unresolved_identities: UnresolvedIdentitiesWidget,
  sprint_board: SprintBoardWidget,
  velocity_trend: VelocityTrendWidget,
  timeline_estimator: TimelineEstimatorWidget,
  plugin_list: PluginListWidget,
  event_stream: EventStreamWidget,
  alert_list: AlertListWidget,
  goal_tracker: GoalTrackerWidget,
  cost_breakdown: CostBreakdownWidget,
  quick_actions: QuickActionsWidget,
} as const;

type WidgetType = keyof typeof widgetRegistry;

export function getWidget(type: WidgetType) {
  return widgetRegistry[type];
}
