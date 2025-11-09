// Team and Organization Types

import type { Engineer } from './types';

export interface Team {
  id: string;
  name: string;
  description?: string;
  manager_id?: string;
  manager?: Engineer;
  parent_team_id?: string;
  created_at: string;
  updated_at: string;
}

export interface TeamMembership {
  team_id: string;
  engineer_id: string;
  joined_at: string;
}

export interface TeamPerformanceScore {
  team_id: string;
  week_start: string;
  total_score: number;
  member_count: number;
  score_change: number;
  velocity_prs_per_week: number;
  velocity_change: number;
  cycle_time_days: number;
  cycle_time_change: number;
}

export interface TeamScorecard {
  team: Team;
  performance: TeamPerformanceScore;
  members: Engineer[];
}

export interface TeamHierarchyNode {
  team: Team;
  performance: TeamPerformanceScore;
  children: TeamHierarchyNode[];
}

export interface OrgScorecard {
  total_engineers: number;
  total_score: number;
  score_change: number;
  velocity_prs_per_week: number;
  velocity_change: number;
  cycle_time_days: number;
  cycle_time_change: number;
  teams_needing_attention: number;
}
