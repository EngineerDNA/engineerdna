package plugin

import (
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/models"
)

// SyncMetricPlugin executes a metric source plugin sync operation
// This is a stub implementation for PDR-9 Phase 1.
// Full implementation will be added in Phase 2 when metric_source plugins are available.
func (e *Executor) SyncMetricPlugin(pluginName string, since time.Time) ([]*models.MetricValue, error) {
	return nil, fmt.Errorf("metric source plugins not yet implemented (PDR-9 Phase 2)")
}
