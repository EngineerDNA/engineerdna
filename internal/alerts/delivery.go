package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/db"
	"github.com/engineerdna/engineerdna/internal/models"
)

// Delivery handles sending alert notifications via configured channels
type Delivery struct {
	alertsStore *db.AlertsStore
}

// NewDelivery creates a new alert delivery service
func NewDelivery(alertsStore *db.AlertsStore) *Delivery {
	return &Delivery{
		alertsStore: alertsStore,
	}
}

// DeliverAlert sends an alert instance via all configured channels
func (d *Delivery) DeliverAlert(instance *models.AlertInstance) error {
	// Get channels for this alert's rule
	channels, err := d.alertsStore.GetAlertChannels(instance.RuleID)
	if err != nil {
		return fmt.Errorf("failed to get alert channels: %w", err)
	}

	if len(channels) == 0 {
		log.Printf("No delivery channels configured for alert rule %s", instance.RuleID)
		return nil
	}

	for _, channel := range channels {
		if !channel.Enabled {
			continue
		}

		// Check quiet hours
		if d.isQuietHours(channel) {
			d.recordDelivery(instance.ID, channel.ChannelType, "skipped_quiet_hours", nil)
			log.Printf("Skipped delivery to %s (quiet hours)", channel.ChannelType)
			continue
		}

		// Deliver via channel type
		var err error
		switch channel.ChannelType {
		case "slack":
			err = d.DeliverSlack(instance, channel)
		case "email":
			err = d.DeliverEmail(instance, channel)
		case "browser":
			err = d.DeliverBrowser(instance, channel)
		default:
			err = fmt.Errorf("unknown channel type: %s", channel.ChannelType)
		}

		// Record delivery attempt
		status := "sent"
		var errorMsg *string
		if err != nil {
			status = "failed"
			msg := err.Error()
			errorMsg = &msg
			log.Printf("Failed to deliver alert to %s: %v", channel.ChannelType, err)
		} else {
			log.Printf("Delivered alert to %s: %s", channel.ChannelType, instance.Title)
		}

		d.recordDelivery(instance.ID, channel.ChannelType, status, errorMsg)
	}

	return nil
}

// DeliverSlack sends alert to Slack webhook
func (d *Delivery) DeliverSlack(instance *models.AlertInstance, channel *models.AlertChannel) error {
	if channel.ChannelConfig == "" {
		return fmt.Errorf("channel config is empty")
	}

	var config models.AlertChannelConfig
	if err := json.Unmarshal([]byte(channel.ChannelConfig), &config); err != nil {
		return fmt.Errorf("failed to parse channel config: %w", err)
	}

	if config.WebhookURL == "" {
		return fmt.Errorf("webhook_url not configured")
	}

	// Build Slack message
	color := d.getSeverityColor(instance.Severity)

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"fallback": instance.Title,
				"color":    color,
				"title":    instance.Title,
				"text":     instance.Message,
				"fields": []map[string]interface{}{
					{
						"title": "Severity",
						"value": strings.ToUpper(instance.Severity),
						"short": true,
					},
					{
						"title": "Fired At",
						"value": instance.FiredAt.Format("2006-01-02 15:04 MST"),
						"short": true,
					},
				},
				"footer": "EngineerDNA Alert System",
				"ts":     instance.FiredAt.Unix(),
			},
		},
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	// Send webhook request
	resp, err := http.Post(config.WebhookURL, "application/json", bytes.NewReader(payloadJSON))
	if err != nil {
		return fmt.Errorf("failed to send Slack webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Slack webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeliverEmail sends alert via email
// Phase 2: Will integrate with SMTP or email destination plugin
func (d *Delivery) DeliverEmail(instance *models.AlertInstance, channel *models.AlertChannel) error {
	// Email delivery requires SMTP configuration or email destination plugin
	// For now, log that email would be sent and record as pending
	log.Printf("Email delivery pending (SMTP not configured): %s - %s", instance.Title, instance.Message)

	// Return error to indicate email delivery not available
	// This allows the system to gracefully handle missing email capability
	return fmt.Errorf("email delivery not configured - SMTP integration required (Phase 2)")
}

// DeliverBrowser stores alert for browser polling
func (d *Delivery) DeliverBrowser(instance *models.AlertInstance, channel *models.AlertChannel) error {
	// Browser delivery is just storing in database - already done when instance created
	// Frontend will poll /api/alerts/instances?status=active
	return nil
}

// isQuietHours checks if current time is within quiet hours
func (d *Delivery) isQuietHours(channel *models.AlertChannel) bool {
	if channel.QuietHoursStart == nil || channel.QuietHoursEnd == nil {
		return false
	}

	now := time.Now().UTC()
	currentTime := now.Format("15:04")

	start := *channel.QuietHoursStart
	end := *channel.QuietHoursEnd

	// Handle overnight quiet hours (e.g., 22:00 to 08:00)
	if start > end {
		return currentTime >= start || currentTime <= end
	}

	// Handle same-day quiet hours (e.g., 12:00 to 13:00)
	return currentTime >= start && currentTime <= end
}

// getSeverityColor returns Slack color for severity level
func (d *Delivery) getSeverityColor(severity string) string {
	switch severity {
	case "critical":
		return "danger"
	case "warning":
		return "warning"
	case "info":
		return "good"
	default:
		return "#808080"
	}
}

// recordDelivery records a delivery attempt in the database
func (d *Delivery) recordDelivery(instanceID, channelType, status string, errorMessage *string) {
	delivery := &models.AlertDelivery{
		InstanceID:   instanceID,
		ChannelType:  channelType,
		Status:       status,
		ErrorMessage: errorMessage,
	}

	if err := d.alertsStore.CreateAlertDelivery(delivery); err != nil {
		log.Printf("Failed to record delivery: %v", err)
	}
}

// DeliverPendingAlerts delivers all undelivered active alerts
func (d *Delivery) DeliverPendingAlerts() error {
	// Get active alerts that haven't been delivered yet
	instances, _, err := d.alertsStore.ListAlertInstances(map[string]string{
		"status": "active",
	}, 100, 0)
	if err != nil {
		return fmt.Errorf("failed to list alert instances: %w", err)
	}

	for _, instance := range instances {
		// Check if already delivered
		deliveries, err := d.alertsStore.GetAlertDeliveries(instance.ID)
		if err != nil {
			log.Printf("Error checking deliveries for alert %s: %v", instance.ID, err)
			continue
		}

		// If already delivered successfully, skip
		hasSuccessfulDelivery := false
		for _, delivery := range deliveries {
			if delivery.Status == "sent" {
				hasSuccessfulDelivery = true
				break
			}
		}

		if hasSuccessfulDelivery {
			continue
		}

		// Deliver the alert
		if err := d.DeliverAlert(instance); err != nil {
			log.Printf("Error delivering alert %s: %v", instance.ID, err)
		}
	}

	return nil
}
