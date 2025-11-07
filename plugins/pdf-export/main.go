package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/engineerdna/engineerdna/plugins/plugin-sdk"
	"github.com/jung-kurt/gofpdf"
)

type PDFExportPlugin struct {
	outputDirectory string
	includeCharts   bool
	pageSize        string
}

func (p *PDFExportPlugin) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		Name:        "pdf-export",
		Version:     "1.0.0",
		Type:        "destination",
		Description: "Export metrics to PDF reports",
		Author:      "EngineerDNA",
		ConfigFields: []sdk.ConfigField{
			{
				Name:        "output_directory",
				Type:        "string",
				Required:    false,
				Description: "Output directory for PDF files (defaults to ~/.engineerdna/exports/)",
				Secret:      false,
				Default:     "",
			},
			{
				Name:        "include_charts",
				Type:        "boolean",
				Required:    false,
				Description: "Include charts in PDF report",
				Secret:      false,
				Default:     "true",
			},
			{
				Name:        "page_size",
				Type:        "select",
				Required:    false,
				Description: "PDF page size",
				Secret:      false,
				Default:     "A4",
				Options:     []string{"A4", "Letter"},
			},
		},
		Anonymization: sdk.AnonymizationSpec{
			Required: false,
			Strategy: "sequential",
			Fields:   []string{"actor", "author"},
		},
	}
}

func (p *PDFExportPlugin) Configure(config map[string]string) error {
	// Output directory
	outputDir := config["output_directory"]
	if outputDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		outputDir = filepath.Join(home, ".engineerdna", "exports")
	}
	p.outputDirectory = outputDir

	// Include charts
	includeCharts := config["include_charts"]
	if includeCharts == "" {
		includeCharts = "true"
	}
	p.includeCharts = includeCharts == "true"

	// Page size
	pageSize := config["page_size"]
	if pageSize == "" {
		pageSize = "A4"
	}
	if pageSize != "A4" && pageSize != "Letter" {
		return fmt.Errorf("invalid page_size: %s (must be A4 or Letter)", pageSize)
	}
	p.pageSize = pageSize

	return nil
}

func (p *PDFExportPlugin) Health() sdk.HealthResult {
	// Check if output directory exists and is writable
	if err := os.MkdirAll(p.outputDirectory, 0755); err != nil {
		return sdk.HealthResult{
			Healthy: false,
			Message: fmt.Sprintf("Cannot create output directory: %v", err),
		}
	}

	// Try to create a test file
	testFile := filepath.Join(p.outputDirectory, ".health-check-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return sdk.HealthResult{
			Healthy: false,
			Message: fmt.Sprintf("Cannot write to output directory: %v", err),
		}
	}
	os.Remove(testFile)

	return sdk.HealthResult{
		Healthy: true,
		Message: fmt.Sprintf("Output directory: %s", p.outputDirectory),
	}
}

func (p *PDFExportPlugin) Test() sdk.TestResult {
	// Check if output directory is writable
	if err := os.MkdirAll(p.outputDirectory, 0755); err != nil {
		return sdk.TestResult{
			Connected:   false,
			Destination: fmt.Sprintf("Error creating directory: %v", err),
		}
	}

	testFile := filepath.Join(p.outputDirectory, ".test-write")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return sdk.TestResult{
			Connected:   false,
			Destination: fmt.Sprintf("Cannot write to directory: %v", err),
		}
	}
	os.Remove(testFile)

	return sdk.TestResult{
		Connected:   true,
		Destination: fmt.Sprintf("Directory: %s", p.outputDirectory),
	}
}

func (p *PDFExportPlugin) Export(params sdk.ExportParams) (sdk.ExportResult, error) {
	// Create PDF
	orientation := "P" // Portrait
	unit := "mm"
	size := p.pageSize
	pdf := gofpdf.New(orientation, unit, size, "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Engineering Metrics Report")
	pdf.Ln(12)

	// Date range
	pdf.SetFont("Arial", "", 12)
	periodStart := ""
	periodEnd := ""
	if ps, ok := params.Data["period_start"].(string); ok {
		periodStart = ps
	}
	if pe, ok := params.Data["period_end"].(string); ok {
		periodEnd = pe
	}

	if periodStart != "" && periodEnd != "" {
		pdf.Cell(0, 6, fmt.Sprintf("Period: %s to %s", formatDate(periodStart), formatDate(periodEnd)))
		pdf.Ln(8)
	}

	// Summary metrics
	if metricsData, ok := params.Data["metrics"]; ok {
		if metrics, ok := metricsData.(map[string]interface{}); ok && len(metrics) > 0 {
			pdf.SetFont("Arial", "B", 14)
			pdf.Cell(0, 10, "Summary")
			pdf.Ln(8)

			pdf.SetFont("Arial", "", 12)

			// Sort metrics by key for consistent output
			keys := make([]string, 0, len(metrics))
			for k := range metrics {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, key := range keys {
				value := metrics[key]
				pdf.Cell(80, 6, fmt.Sprintf("%s:", formatMetricName(key)))
				pdf.Cell(40, 6, fmt.Sprintf("%v", value))
				pdf.Ln(6)
			}
			pdf.Ln(4)
		}
	}

	// Events table (if provided)
	if eventsData, ok := params.Data["events"]; ok {
		if eventsList, ok := eventsData.([]interface{}); ok && len(eventsList) > 0 {
			pdf.SetFont("Arial", "B", 14)
			pdf.Cell(0, 10, "Recent Events")
			pdf.Ln(8)

			// Table headers
			pdf.SetFont("Arial", "B", 10)
			pdf.Cell(30, 6, "Type")
			pdf.Cell(50, 6, "Actor")
			pdf.Cell(50, 6, "Timestamp")
			pdf.Ln(6)

			// Table rows (limit to 20 to fit on page)
			pdf.SetFont("Arial", "", 10)
			rowCount := 0
			maxRows := 20

			for _, eventData := range eventsList {
				if rowCount >= maxRows {
					break
				}

				event, ok := eventData.(map[string]interface{})
				if !ok {
					continue
				}

				eventType := ""
				if t, ok := event["type"].(string); ok {
					eventType = t
				}

				actor := ""
				if a, ok := event["actor"].(string); ok {
					actor = a
				}

				timestamp := ""
				if ts, ok := event["timestamp"].(string); ok {
					timestamp = formatTimestamp(ts)
				}

				pdf.Cell(30, 6, truncate(eventType, 15))
				pdf.Cell(50, 6, truncate(actor, 25))
				pdf.Cell(50, 6, timestamp)
				pdf.Ln(6)
				rowCount++
			}

			if len(eventsList) > maxRows {
				pdf.Ln(4)
				pdf.SetFont("Arial", "I", 9)
				pdf.Cell(0, 6, fmt.Sprintf("Showing %d of %d events", maxRows, len(eventsList)))
				pdf.Ln(6)
			}
		}
	}

	// Engineer breakdown table
	if byEngineerData, ok := params.Data["by_engineer"]; ok {
		if byEngineer, ok := byEngineerData.([]interface{}); ok && len(byEngineer) > 0 {
			pdf.AddPage()
			pdf.SetFont("Arial", "B", 14)
			pdf.Cell(0, 10, "Engineer Activity")
			pdf.Ln(8)

			// Table headers
			pdf.SetFont("Arial", "B", 10)
			pdf.Cell(60, 6, "Engineer")
			pdf.Cell(30, 6, "PRs")
			pdf.Cell(30, 6, "Reviews")
			pdf.Cell(30, 6, "Commits")
			pdf.Ln(6)

			// Table rows
			pdf.SetFont("Arial", "", 10)
			for _, engData := range byEngineer {
				eng, ok := engData.(map[string]interface{})
				if !ok {
					continue
				}

				name := ""
				if n, ok := eng["name"].(string); ok {
					name = n
				}

				prs := formatValue(eng["prs"])
				reviews := formatValue(eng["reviews"])
				commits := formatValue(eng["commits"])

				pdf.Cell(60, 6, truncate(name, 30))
				pdf.Cell(30, 6, prs)
				pdf.Cell(30, 6, reviews)
				pdf.Cell(30, 6, commits)
				pdf.Ln(6)
			}
		}
	}

	// Output to file
	if err := os.MkdirAll(p.outputDirectory, 0755); err != nil {
		return sdk.ExportResult{}, fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := filepath.Join(p.outputDirectory, fmt.Sprintf("report-%s.pdf", time.Now().UTC().Format("2006-01-02-150405")))

	if err := pdf.OutputFileAndClose(filename); err != nil {
		return sdk.ExportResult{}, fmt.Errorf("failed to save PDF: %w", err)
	}

	// Count rows written (events + engineers)
	rowsWritten := 0
	if eventsData, ok := params.Data["events"]; ok {
		if eventsList, ok := eventsData.([]interface{}); ok {
			rowsWritten += len(eventsList)
		}
	}
	if byEngineerData, ok := params.Data["by_engineer"]; ok {
		if byEngineer, ok := byEngineerData.([]interface{}); ok {
			rowsWritten += len(byEngineer)
		}
	}

	return sdk.ExportResult{
		Status:      "success",
		URL:         filename,
		RowsWritten: rowsWritten,
	}, nil
}

// Helper functions

func formatDate(dateStr string) string {
	// Try parsing RFC3339 format
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("2006-01-02")
}

func formatTimestamp(tsStr string) string {
	// Try parsing RFC3339 format
	t, err := time.Parse(time.RFC3339, tsStr)
	if err != nil {
		return tsStr
	}
	return t.Format("2006-01-02 15:04")
}

func formatMetricName(name string) string {
	// Convert snake_case to Title Case
	// e.g., "total_prs" -> "Total PRs"
	result := ""
	capitalize := true
	for _, char := range name {
		if char == '_' {
			result += " "
			capitalize = true
		} else {
			if capitalize {
				result += string(char - 32) // Convert to uppercase
				capitalize = false
			} else {
				result += string(char)
			}
		}
	}
	return result
}

func formatValue(value interface{}) string {
	if value == nil {
		return "0"
	}

	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', 0, 64)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func main() {
	sdk.Serve(&PDFExportPlugin{})
}
