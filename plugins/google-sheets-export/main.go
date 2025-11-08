package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/engineerdna/engineerdna/internal/config"
	"github.com/engineerdna/engineerdna/plugins/plugin-sdk"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsPlugin struct {
	credentialsJSON string
	spreadsheetID   string
	sheetName       string
	exportMode      string
	service         *sheets.Service
}

func (p *GoogleSheetsPlugin) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		Name:        "google-sheets-export",
		Version:     "1.0.0",
		Type:        "destination",
		Description: "Export metrics to Google Sheets",
		Author:      "EngineerDNA",
		OAuth: &sdk.OAuthSpec{
			Provider: "google",
			Scopes:   []string{"https://www.googleapis.com/auth/spreadsheets"},
		},
		Anonymization: sdk.AnonymizationSpec{
			Required: false,
			Reason:   "Data exported to external Google services - recommend manual review before export",
			Strategy: "sequential",
			Fields:   []string{"actor", "author", "assignee"},
		},
		ConfigFields: []sdk.ConfigField{
			{
				Name:        "credentials_json",
				Type:        "password",
				Required:    true,
				Description: "OAuth client credentials JSON",
				Secret:      true,
			},
			{
				Name:        "spreadsheet_id",
				Type:        "string",
				Required:    true,
				Description: "Google Sheets document ID",
			},
			{
				Name:        "sheet_name",
				Type:        "string",
				Required:    false,
				Default:     "EngineerDNA",
				Description: "Sheet tab name",
			},
			{
				Name:        "export_mode",
				Type:        "select",
				Options:     []string{"raw", "aggregated"},
				Required:    true,
				Description: "Export format: raw events or aggregated metrics",
			},
		},
		Schedule: &sdk.ScheduleSpec{
			Supported: true,
			Default:   "weekly",
		},
	}
}

func (p *GoogleSheetsPlugin) Configure(config map[string]string) error {
	credentialsJSON, ok := config["credentials_json"]
	if !ok || credentialsJSON == "" {
		return fmt.Errorf("credentials_json is required")
	}
	p.credentialsJSON = credentialsJSON

	spreadsheetID, ok := config["spreadsheet_id"]
	if !ok || spreadsheetID == "" {
		return fmt.Errorf("spreadsheet_id is required")
	}
	p.spreadsheetID = spreadsheetID

	p.sheetName = config["sheet_name"]
	if p.sheetName == "" {
		p.sheetName = "EngineerDNA"
	}

	exportMode, ok := config["export_mode"]
	if !ok || exportMode == "" {
		return fmt.Errorf("export_mode is required")
	}
	if exportMode != "raw" && exportMode != "aggregated" {
		return fmt.Errorf("export_mode must be 'raw' or 'aggregated', got: %s", exportMode)
	}
	p.exportMode = exportMode

	// Initialize Google Sheets client
	ctx := context.Background()
	creds, err := google.CredentialsFromJSON(ctx, []byte(credentialsJSON), sheets.SpreadsheetsScope)
	if err != nil {
		return fmt.Errorf("failed to parse credentials: %w", err)
	}

	service, err := sheets.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return fmt.Errorf("failed to create sheets service: %w", err)
	}
	p.service = service

	return nil
}

func (p *GoogleSheetsPlugin) Health() sdk.HealthResult {
	if p.service == nil {
		return sdk.HealthResult{
			Healthy: false,
			Message: "Not configured",
		}
	}

	// Try to read the spreadsheet to verify access (with timeout)
	ctx, cancel := context.WithTimeout(context.Background(), config.PluginTimeout)
	defer cancel()
	spreadsheet, err := p.service.Spreadsheets.Get(p.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return sdk.HealthResult{
			Healthy: false,
			Message: fmt.Sprintf("Cannot access spreadsheet: %v", err),
		}
	}

	return sdk.HealthResult{
		Healthy: true,
		Message: fmt.Sprintf("Connected to: %s", spreadsheet.Properties.Title),
	}
}

func (p *GoogleSheetsPlugin) Test() sdk.TestResult {
	if p.service == nil {
		return sdk.TestResult{
			Connected:   false,
			Destination: "Not configured",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.PluginTimeout)
	defer cancel()
	spreadsheet, err := p.service.Spreadsheets.Get(p.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return sdk.TestResult{
			Connected:   false,
			Destination: fmt.Sprintf("Error: %v", err),
		}
	}

	return sdk.TestResult{
		Connected:   true,
		Destination: fmt.Sprintf("Spreadsheet: %s", spreadsheet.Properties.Title),
	}
}

func (p *GoogleSheetsPlugin) Export(params sdk.ExportParams) (sdk.ExportResult, error) {
	if p.service == nil {
		return sdk.ExportResult{}, fmt.Errorf("plugin not configured")
	}

	// Use timeout for entire export operation (5 minutes)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Ensure sheet exists
	if err := p.ensureSheetExists(ctx); err != nil {
		return sdk.ExportResult{}, fmt.Errorf("failed to ensure sheet exists: %w", err)
	}

	var rowsWritten int
	var err error

	switch p.exportMode {
	case "raw":
		rowsWritten, err = p.exportRaw(ctx, params.Data)
	case "aggregated":
		rowsWritten, err = p.exportAggregated(ctx, params.Data)
	default:
		return sdk.ExportResult{}, fmt.Errorf("unsupported export mode: %s", p.exportMode)
	}

	if err != nil {
		return sdk.ExportResult{}, err
	}

	url := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s", p.spreadsheetID)
	return sdk.ExportResult{
		Status:      "success",
		URL:         url,
		RowsWritten: rowsWritten,
	}, nil
}

func (p *GoogleSheetsPlugin) ensureSheetExists(ctx context.Context) error {
	spreadsheet, err := p.service.Spreadsheets.Get(p.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return err
	}

	// Check if sheet exists
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == p.sheetName {
			return nil // Already exists
		}
	}

	// Create new sheet
	req := &sheets.Request{
		AddSheet: &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{
				Title: p.sheetName,
			},
		},
	}

	batchUpdate := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{req},
	}

	_, err = p.service.Spreadsheets.BatchUpdate(p.spreadsheetID, batchUpdate).Context(ctx).Do()
	return err
}

func (p *GoogleSheetsPlugin) exportRaw(ctx context.Context, data map[string]interface{}) (int, error) {
	// Build rows from raw events
	var rows [][]interface{}

	// Header row
	rows = append(rows, []interface{}{"Timestamp", "Type", "Actor", "Source", "Data"})

	// Extract events array
	eventsData, ok := data["events"]
	if !ok {
		return 0, fmt.Errorf("no events field in data")
	}

	events, ok := eventsData.([]interface{})
	if !ok {
		return 0, fmt.Errorf("events field is not an array")
	}

	// Enforce max row limit to prevent memory/API issues
	const maxRows = 10000
	if len(events) > maxRows {
		return 0, fmt.Errorf("too many events to export: %d (max %d rows)", len(events), maxRows)
	}

	// Process each event
	for _, eventData := range events {
		event, ok := eventData.(map[string]interface{})
		if !ok {
			continue // Skip invalid events
		}

		// Safely extract fields with defaults
		timestamp := ""
		if ts, ok := event["timestamp"].(string); ok {
			timestamp = ts
		}

		eventType := ""
		if t, ok := event["type"].(string); ok {
			eventType = t
		}

		actor := ""
		if a, ok := event["actor"].(string); ok {
			actor = a
		}

		source := ""
		if s, ok := event["source"].(string); ok {
			source = s
		}

		// Stringify the data field as JSON instead of Go format
		dataStr := ""
		if d, ok := event["data"]; ok {
			if jsonBytes, err := json.Marshal(d); err == nil {
				dataStr = string(jsonBytes)
			} else {
				dataStr = fmt.Sprintf("%v", d) // Fallback to Go format if JSON fails
			}
		}

		rows = append(rows, []interface{}{timestamp, eventType, actor, source, dataStr})
	}

	// Write to sheet
	valueRange := &sheets.ValueRange{
		Values: rows,
	}

	_, err := p.service.Spreadsheets.Values.Append(
		p.spreadsheetID,
		p.sheetName,
		valueRange,
	).ValueInputOption("RAW").Context(ctx).Do()

	if err != nil {
		return 0, fmt.Errorf("failed to append data: %w", err)
	}

	return len(rows) - 1, nil // Subtract header row
}

func (p *GoogleSheetsPlugin) exportAggregated(ctx context.Context, data map[string]interface{}) (int, error) {
	// Build rows from aggregated data
	var rows [][]interface{}

	// Header row
	rows = append(rows, []interface{}{"Metric", "Value", "Period"})

	// Extract period (with safe type check and default)
	period := ""
	if p, ok := data["period"].(string); ok {
		period = p
	}

	// Extract metrics (with safe type check)
	if metricsData, ok := data["metrics"]; ok {
		if metrics, ok := metricsData.(map[string]interface{}); ok {
			for key, value := range metrics {
				rows = append(rows, []interface{}{key, value, period})
			}
		}
	}

	// Add engineer breakdown if present (with safe type checks)
	if byEngineerData, ok := data["by_engineer"]; ok {
		if byEngineer, ok := byEngineerData.([]interface{}); ok {
			rows = append(rows, []interface{}{""}) // Empty row
			rows = append(rows, []interface{}{"Engineer", "PRs", "Reviews"})

			for _, eng := range byEngineer {
				if engMap, ok := eng.(map[string]interface{}); ok {
					// Safely extract fields with defaults
					name := ""
					if n, ok := engMap["name"]; ok {
						name = fmt.Sprintf("%v", n)
					}

					prs := ""
					if p, ok := engMap["prs"]; ok {
						prs = fmt.Sprintf("%v", p)
					}

					reviews := ""
					if r, ok := engMap["reviews"]; ok {
						reviews = fmt.Sprintf("%v", r)
					}

					rows = append(rows, []interface{}{name, prs, reviews})
				}
			}
		}
	}

	// If no rows were added beyond the header, return error
	if len(rows) <= 1 {
		return 0, fmt.Errorf("no valid data to export")
	}

	// Enforce max row limit
	const maxRows = 10000
	if len(rows) > maxRows {
		return 0, fmt.Errorf("too many rows to export: %d (max %d rows)", len(rows), maxRows)
	}

	// Write to sheet
	valueRange := &sheets.ValueRange{
		Values: rows,
	}

	_, err := p.service.Spreadsheets.Values.Append(
		p.spreadsheetID,
		p.sheetName,
		valueRange,
	).ValueInputOption("RAW").Context(ctx).Do()

	if err != nil {
		return 0, fmt.Errorf("failed to append data: %w", err)
	}

	return len(rows) - 1, nil // Subtract header row
}

func main() {
	sdk.Serve(&GoogleSheetsPlugin{})
}
