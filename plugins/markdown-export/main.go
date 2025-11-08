package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	sdk "github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

type MarkdownExportPlugin struct {
	outputFile  string
	includeTOC  bool
	formatStyle string
}

func (p *MarkdownExportPlugin) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		Name:        "markdown-export",
		Version:     "0.1.0",
		Type:        "destination",
		Description: "Export metrics to Markdown format",
		Author:      "EngineerDNA",
		ConfigFields: []sdk.ConfigField{
			{
				Name:        "output_file",
				Type:        "string",
				Required:    false,
				Description: "Output file path (defaults to report-YYYY-MM-DD.md in ~/.engineerdna/exports/)",
				Secret:      false,
				Default:     "",
			},
			{
				Name:        "include_toc",
				Type:        "boolean",
				Required:    false,
				Description: "Include table of contents",
				Secret:      false,
				Default:     "true",
			},
			{
				Name:        "format_style",
				Type:        "select",
				Required:    false,
				Description: "Markdown formatting style",
				Secret:      false,
				Default:     "github",
				Options:     []string{"github", "gitlab", "plain"},
			},
		},
		Anonymization: sdk.AnonymizationSpec{
			Required: false,
			Strategy: "sequential",
			Fields:   []string{"actor", "author"},
		},
	}
}

func (p *MarkdownExportPlugin) Configure(config map[string]string) error {
	// Output file
	outputFile := config["output_file"]
	if outputFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		outputDir := filepath.Join(home, ".engineerdna", "exports")
		outputFile = filepath.Join(outputDir, fmt.Sprintf("report-%s.md", time.Now().UTC().Format("2006-01-02")))
	}
	p.outputFile = outputFile

	// Include TOC
	includeTOC := config["include_toc"]
	if includeTOC == "" {
		includeTOC = "true"
	}
	p.includeTOC = includeTOC == "true"

	// Format style
	formatStyle := config["format_style"]
	if formatStyle == "" {
		formatStyle = "github"
	}
	if formatStyle != "github" && formatStyle != "gitlab" && formatStyle != "plain" {
		return fmt.Errorf("invalid format_style: %s (must be github, gitlab, or plain)", formatStyle)
	}
	p.formatStyle = formatStyle

	return nil
}

func (p *MarkdownExportPlugin) Health() sdk.HealthResult {
	// Check if output directory exists and is writable
	outputDir := filepath.Dir(p.outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return sdk.HealthResult{
			Healthy: false,
			Message: fmt.Sprintf("Cannot create output directory: %v", err),
		}
	}

	// Try to create a test file
	testFile := filepath.Join(outputDir, ".health-check-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return sdk.HealthResult{
			Healthy: false,
			Message: fmt.Sprintf("Cannot write to output directory: %v", err),
		}
	}
	os.Remove(testFile)

	return sdk.HealthResult{
		Healthy: true,
		Message: fmt.Sprintf("Output file: %s", p.outputFile),
	}
}

func (p *MarkdownExportPlugin) Test() sdk.TestResult {
	// Check if output directory is writable
	outputDir := filepath.Dir(p.outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return sdk.TestResult{
			Connected:   false,
			Destination: fmt.Sprintf("Error creating directory: %v", err),
		}
	}

	testFile := filepath.Join(outputDir, ".test-write")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return sdk.TestResult{
			Connected:   false,
			Destination: fmt.Sprintf("Cannot write to directory: %v", err),
		}
	}
	os.Remove(testFile)

	return sdk.TestResult{
		Connected:   true,
		Destination: fmt.Sprintf("File: %s", p.outputFile),
	}
}

func (p *MarkdownExportPlugin) Export(params sdk.ExportParams) (sdk.ExportResult, error) {
	var md strings.Builder

	// Title
	md.WriteString("# Engineering Metrics Report\n\n")

	// Date range
	periodStart := ""
	periodEnd := ""
	if ps, ok := params.Data["period_start"].(string); ok {
		periodStart = ps
	}
	if pe, ok := params.Data["period_end"].(string); ok {
		periodEnd = pe
	}

	if periodStart != "" && periodEnd != "" {
		md.WriteString(fmt.Sprintf("**Period:** %s to %s\n\n", formatDate(periodStart), formatDate(periodEnd)))
	}

	// Table of Contents (optional)
	if p.includeTOC {
		md.WriteString("## Table of Contents\n\n")
		md.WriteString("- [Summary](#summary)\n")

		// Check if we have engineer data
		if byEngineerData, ok := params.Data["by_engineer"]; ok {
			if byEngineer, ok := byEngineerData.([]interface{}); ok && len(byEngineer) > 0 {
				md.WriteString("- [By Engineer](#by-engineer)\n")
			}
		}

		// Check if we have events
		if eventsData, ok := params.Data["events"]; ok {
			if eventsList, ok := eventsData.([]interface{}); ok && len(eventsList) > 0 {
				md.WriteString("- [Recent Events](#recent-events)\n")
			}
		}
		md.WriteString("\n")
	}

	// Summary
	if metricsData, ok := params.Data["metrics"]; ok {
		if metrics, ok := metricsData.(map[string]interface{}); ok && len(metrics) > 0 {
			md.WriteString("## Summary\n\n")
			md.WriteString("| Metric | Value |\n")
			md.WriteString("|--------|-------|\n")

			// Sort metrics by key for consistent output
			keys := make([]string, 0, len(metrics))
			for k := range metrics {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, key := range keys {
				value := metrics[key]
				md.WriteString(fmt.Sprintf("| %s | %v |\n", formatMetricName(key), value))
			}
			md.WriteString("\n")
		}
	}

	// By Engineer (if provided)
	if byEngineerData, ok := params.Data["by_engineer"]; ok {
		if byEngineer, ok := byEngineerData.([]interface{}); ok && len(byEngineer) > 0 {
			md.WriteString("## By Engineer\n\n")
			md.WriteString("| Engineer | PRs | Reviews | Commits |\n")
			md.WriteString("|----------|-----|---------|----------|\n")

			for _, engData := range byEngineer {
				eng, ok := engData.(map[string]interface{})
				if !ok {
					continue
				}

				name := formatValue(eng["name"])
				prs := formatValue(eng["prs"])
				reviews := formatValue(eng["reviews"])
				commits := formatValue(eng["commits"])

				md.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", name, prs, reviews, commits))
			}
			md.WriteString("\n")
		}
	}

	// Recent Events (limit to 20)
	if eventsData, ok := params.Data["events"]; ok {
		if eventsList, ok := eventsData.([]interface{}); ok && len(eventsList) > 0 {
			md.WriteString("## Recent Events\n\n")
			md.WriteString("| Type | Actor | Timestamp |\n")
			md.WriteString("|------|-------|------------|\n")

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

				md.WriteString(fmt.Sprintf("| %s | %s | %s |\n", eventType, actor, timestamp))
				rowCount++
			}

			if len(eventsList) > maxRows {
				md.WriteString("\n")
				md.WriteString(fmt.Sprintf("*Showing %d of %d events*\n", maxRows, len(eventsList)))
			}
			md.WriteString("\n")
		}
	}

	// Save to file
	outputDir := filepath.Dir(p.outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return sdk.ExportResult{}, fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(p.outputFile, []byte(md.String()), 0644); err != nil {
		return sdk.ExportResult{}, fmt.Errorf("failed to write Markdown file: %w", err)
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
		URL:         p.outputFile,
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
	// e.g., "total_prs" -> "Total Prs"
	parts := strings.Split(name, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

func formatValue(value interface{}) string {
	if value == nil {
		return "0"
	}
	return fmt.Sprintf("%v", value)
}

func main() {
	sdk.Serve(&MarkdownExportPlugin{})
}
