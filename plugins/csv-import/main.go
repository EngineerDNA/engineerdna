package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/plugins/plugin-sdk"
	"github.com/google/uuid"
)

type ColumnMapping struct {
	Timestamp   interface{} `json:"timestamp"`   // Column index (int) or header name (string)
	Actor       interface{} `json:"actor"`       // Column index (int) or header name (string)
	Type        interface{} `json:"type"`        // Column index (int) or header name (string)
	Description interface{} `json:"description"` // Column index (int) or header name (string)
}

type CSVImportPlugin struct {
	filePath      string
	columnMapping *ColumnMapping
	dateFormats   []string
}

func (p *CSVImportPlugin) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		Name:        "csv-import",
		Version:     "0.1.0",
		Type:        "source",
		Description: "Import engineering metrics from CSV files",
		Author:      "EngineerDNA",
		ConfigFields: []sdk.ConfigField{
			{
				Name:        "file_path",
				Type:        "string",
				Required:    true,
				Description: "Path to CSV file",
				Secret:      false,
			},
			{
				Name:        "column_mapping",
				Type:        "string",
				Required:    false,
				Description: "JSON mapping of CSV columns to event fields. Example: {\"timestamp\":0,\"actor\":1,\"type\":2,\"description\":3} or use header names: {\"timestamp\":\"Date\",\"actor\":\"Person\",\"type\":\"Event Type\",\"description\":\"Details\"}. Defaults to columns 0,1,2,3.",
				Secret:      false,
			},
		},
		Anonymization: sdk.AnonymizationSpec{
			Required: false,
			Strategy: "sequential",
			Fields:   []string{"actor"},
		},
	}
}

func (p *CSVImportPlugin) Configure(config map[string]string) error {
	// Get and validate file path
	filePath, ok := config["file_path"]
	if !ok || filePath == "" {
		return fmt.Errorf("file_path is required")
	}

	// Path validation: prevent directory traversal and enforce working directory restriction
	// IMPORTANT: Reject absolute paths to prevent reading arbitrary system files
	// For V1 (localhost): Defense-in-depth to prevent accidental sensitive file access
	// For V2 (network): Critical security control to prevent remote file read attacks
	if filepath.IsAbs(filePath) {
		return fmt.Errorf("invalid file path: absolute paths not allowed (must be relative to current directory)")
	}

	// Check for directory traversal attempts BEFORE resolving
	if strings.Contains(filePath, "..") {
		return fmt.Errorf("invalid file path: directory traversal not allowed")
	}

	cleanPath := filepath.Clean(filePath)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Resolve symlinks to prevent symlink-based directory traversal
	resolvedPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If file doesn't exist yet, that's okay - we'll validate when it's accessed
		// But if it exists and we can't resolve it, that's an error
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to resolve symlinks: %w", err)
		}
		resolvedPath = absPath
	}

	// Ensure resolved path stays within working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	relPath, err := filepath.Rel(cwd, resolvedPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("invalid file path: must be within current directory")
	}

	p.filePath = resolvedPath

	// Parse column mapping if provided
	if mappingJSON, ok := config["column_mapping"]; ok && mappingJSON != "" {
		var mapping ColumnMapping
		if err := json.Unmarshal([]byte(mappingJSON), &mapping); err != nil {
			return fmt.Errorf("invalid column_mapping JSON: %w", err)
		}
		p.columnMapping = &mapping
	} else {
		// Default mapping: columns 0, 1, 2, 3
		p.columnMapping = &ColumnMapping{
			Timestamp:   0,
			Actor:       1,
			Type:        2,
			Description: 3,
		}
	}

	// Initialize supported date formats (try in order)
	p.dateFormats = []string{
		"2006-01-02",           // YYYY-MM-DD
		"01/02/2006",           // MM/DD/YYYY
		"2006-01-02 15:04:05",  // YYYY-MM-DD HH:MM:SS
		"2006-01-02T15:04:05Z", // ISO 8601
		"01-02-2006",           // MM-DD-YYYY
		"2/1/2006",             // M/D/YYYY
	}

	return nil
}

func (p *CSVImportPlugin) Health() sdk.HealthResult {
	if p.filePath == "" {
		return sdk.HealthResult{
			Healthy: false,
			Message: "Not configured",
		}
	}

	if _, err := os.Stat(p.filePath); os.IsNotExist(err) {
		return sdk.HealthResult{
			Healthy: false,
			Message: fmt.Sprintf("File not found: %s", p.filePath),
		}
	}

	return sdk.HealthResult{
		Healthy: true,
		Message: fmt.Sprintf("CSV file found: %s", p.filePath),
	}
}

// Helper: parse timestamp with multiple format support
func (p *CSVImportPlugin) parseTimestamp(value string) (time.Time, error) {
	var lastErr error
	for _, format := range p.dateFormats {
		t, err := time.Parse(format, value)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, fmt.Errorf("failed to parse timestamp '%s' with any known format: %w", value, lastErr)
}

// Helper: get column value by index or header name
func (p *CSVImportPlugin) getColumnValue(record []string, headers []string, mapping interface{}) (string, error) {
	switch v := mapping.(type) {
	case float64: // JSON numbers are float64
		// Check for integer overflow/underflow before conversion
		if v < 0 || v > float64(len(record)-1) || v != float64(int(v)) {
			return "", fmt.Errorf("invalid column index: %v (must be non-negative integer)", v)
		}
		idx := int(v)
		if idx >= len(record) {
			return "", fmt.Errorf("column index %d out of range", idx)
		}
		return record[idx], nil
	case string: // Header name
		for i, header := range headers {
			if header == v {
				if i >= len(record) {
					return "", fmt.Errorf("column index %d out of range for header '%s'", i, v)
				}
				return record[i], nil
			}
		}
		return "", fmt.Errorf("header '%s' not found", v)
	default:
		return "", fmt.Errorf("invalid mapping type: %T", v)
	}
}

func (p *CSVImportPlugin) Sync(params sdk.SyncParams) (sdk.SyncResult, error) {
	if p.filePath == "" {
		return sdk.SyncResult{}, fmt.Errorf("plugin not configured")
	}

	file, err := os.Open(p.filePath)
	if err != nil {
		return sdk.SyncResult{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Check file size immediately after opening to minimize race window
	const maxFileSize = 100 * 1024 * 1024 // 100 MB limit
	fileInfo, err := file.Stat()
	if err != nil {
		return sdk.SyncResult{}, fmt.Errorf("failed to stat file: %w", err)
	}
	if fileInfo.Size() > maxFileSize {
		return sdk.SyncResult{}, fmt.Errorf("file too large: %d bytes (max %d bytes / 100MB)", fileInfo.Size(), maxFileSize)
	}

	// Use LimitReader as failsafe to prevent memory exhaustion even if file grows during read
	limitedReader := io.LimitReader(file, maxFileSize+1)

	reader := csv.NewReader(limitedReader)
	records, err := reader.ReadAll()
	if err != nil {
		return sdk.SyncResult{}, fmt.Errorf("failed to read CSV: %w", err)
	}

	// Limit number of rows to prevent memory issues
	const maxRows = 100000
	if len(records) > maxRows {
		return sdk.SyncResult{}, fmt.Errorf("too many rows: %d (max %d rows)", len(records), maxRows)
	}

	if len(records) < 2 {
		return sdk.SyncResult{Events: []sdk.Event{}}, nil
	}

	// First row is headers
	headers := records[0]
	var events []sdk.Event
	var warnings []string

	for i, record := range records[1:] {
		rowNum := i + 2 // Account for 0-index and header row

		// Validate cell sizes to prevent memory exhaustion
		const maxCellSize = 1024 * 1024 // 1MB per cell
		for cellIdx, cell := range record {
			if len(cell) > maxCellSize {
				warnings = append(warnings, fmt.Sprintf("Row %d, column %d: cell too large (%d bytes, max 1MB), truncating", rowNum, cellIdx, len(cell)))
				record[cellIdx] = cell[:maxCellSize]
			}
		}

		// Get column values using mapping
		timestampStr, err := p.getColumnValue(record, headers, p.columnMapping.Timestamp)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Row %d: failed to get timestamp: %v", rowNum, err))
			continue
		}

		actorStr, err := p.getColumnValue(record, headers, p.columnMapping.Actor)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Row %d: failed to get actor: %v", rowNum, err))
			continue
		}

		typeStr, err := p.getColumnValue(record, headers, p.columnMapping.Type)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Row %d: failed to get type: %v", rowNum, err))
			continue
		}

		descriptionStr, err := p.getColumnValue(record, headers, p.columnMapping.Description)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Row %d: failed to get description: %v", rowNum, err))
			continue
		}

		// Parse timestamp with multiple format support
		timestamp, err := p.parseTimestamp(timestampStr)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Row %d: invalid timestamp '%s': %v", rowNum, timestampStr, err))
			continue
		}

		// Only include events after the 'since' parameter
		if timestamp.Before(params.Since) {
			continue
		}

		event := sdk.Event{
			ID:        uuid.New().String(),
			Type:      typeStr,
			Source:    "csv-import",
			SourceID:  fmt.Sprintf("csv-%d", i),
			Timestamp: timestamp,
			Actor:     actorStr,
			Data: map[string]interface{}{
				"description": descriptionStr,
			},
		}

		events = append(events, event)
	}

	// Log summary to stderr for debugging
	if len(warnings) > 0 {
		fmt.Fprintf(os.Stderr, "CSV import completed with %d warnings (skipped %d rows)\n", len(warnings), len(warnings))
	}

	return sdk.SyncResult{
		Events:   events,
		Warnings: warnings,
	}, nil
}

func main() {
	sdk.Serve(&CSVImportPlugin{})
}
