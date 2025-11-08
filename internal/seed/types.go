package seed

import (
	"time"

	"github.com/engineerdna/engineerdna/internal/config"
	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/identity"
	"github.com/engineerdna/engineerdna/internal/models"
)

// SeedData is the data structure passed to all seed helper functions
type SeedData struct {
	Database         *config.Database
	EventStore       *db.EventStore
	IdentityService  *identity.Service
	TeamStore        *db.TeamStore
	ScoringStore     *db.ScoringStore
	BriefingStore    *db.BriefingsStore
	PlanningStore    *db.PlanningStore
	AlertsStore      *db.AlertsStore
	GoalsStore       *db.GoalsStore
	SkillsStore      *db.SkillsStore
	CostROIStore     *db.CostROIStore
	ContextStore     *db.ContextStore
	ForecastingStore *db.ForecastingStore
	ActionsStore     *db.ActionsStore
	Now              time.Time
	EngineeringTeam  *models.Team
	BackendTeam      *models.Team
	FrontendTeam     *models.Team
	PlatformTeam     *models.Team
	EngineerIDs      []string
	EngineerMap      map[string]EngineerDef
	RoleMap          map[string]string
	SprintIDs        []string
}

// EngineerDef defines an engineer for seeding
type EngineerDef struct {
	Name        string
	Email       string
	Manager     string
	Role        string
	TeamID      string
	Identifiers map[string]string
}

// GetStringValue returns the string value or empty string if nil
func GetStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
