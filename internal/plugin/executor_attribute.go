package plugin

import (
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
)

// SyncAttributePlugin executes an attribute source plugin sync operation
// This is a stub implementation for PDR-9 Phase 1.
// Full implementation will be added in Phase 2 when attribute_source plugins are available.
func (e *Executor) SyncAttributePlugin(pluginName string, since time.Time) ([]*models.EntityAttribute, error) {
	return nil, fmt.Errorf("attribute source plugins not yet implemented (PDR-9 Phase 2)")
}
