package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Bulk Import Handler

func (s *Server) handleEngineersImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form with size limit
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse form: %v", err), nil)
		return
	}

	// Get file
	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("No file uploaded: %v", err), nil)
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > MaxUploadSize {
		respondError(w, http.StatusBadRequest, "File too large (max 10MB)", nil)
		return
	}

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		respondError(w, http.StatusBadRequest, "Only CSV files allowed", nil)
		return
	}

	// Parse column mapping
	mappingJSON := r.FormValue("column_mapping")
	var columnMapping map[string]string
	if err := json.Unmarshal([]byte(mappingJSON), &columnMapping); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid column mapping: %v", err), nil)
		return
	}

	// Read CSV
	csvReader := csv.NewReader(file)
	csvReader.TrimLeadingSpace = true

	headers, err := csvReader.Read()
	if err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Failed to read CSV headers: %v", err), nil)
		return
	}

	// Build index map for column mapping
	headerIndex := make(map[string]int)
	for i, h := range headers {
		headerIndex[h] = i
	}

	imported := 0
	skipped := 0
	var errors []map[string]interface{}
	rowNum := 1 // Start from 1 (header is row 0)

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		rowNum++

		// Enforce maximum row limit to prevent resource exhaustion
		if rowNum > MaxImportRows {
			errors = append(errors, map[string]interface{}{
				"row":   rowNum,
				"error": fmt.Sprintf("Import limit exceeded: maximum %d rows allowed", MaxImportRows),
			})
			break
		}

		if err != nil {
			errors = append(errors, map[string]interface{}{
				"row":   rowNum,
				"error": fmt.Sprintf("CSV read error: %v", err),
			})
			skipped++
			continue
		}

		// Extract engineer data using column mapping
		name := getCSVValue(row, headerIndex, columnMapping["name"])
		email := getCSVValue(row, headerIndex, columnMapping["email"])
		manager := getCSVValue(row, headerIndex, columnMapping["manager"])

		if name == "" {
			errors = append(errors, map[string]interface{}{
				"row":   rowNum,
				"error": "Missing required field: name",
			})
			skipped++
			continue
		}

		// Build identifiers map from other mappings (with source validation)
		identifiers := make(map[string]string)
		for key, colName := range columnMapping {
			if key != "name" && key != "email" && key != "manager" && colName != "" {
				// Validate source name against whitelist
				if !ValidIdentifierSources[key] {
					errors = append(errors, map[string]interface{}{
						"row":   rowNum,
						"error": fmt.Sprintf("Invalid identifier source: %s (allowed: github, jira, gitlab, slack, zoom, csv)", key),
					})
					skipped++
					continue
				}

				value := getCSVValue(row, headerIndex, colName)
				if value != "" {
					identifiers[key] = value
				}
			}
		}

		// Create engineer
		_, err = s.identityService.CreateEngineer(name, email, manager, identifiers)
		if err != nil {
			errors = append(errors, map[string]interface{}{
				"row":   rowNum,
				"error": err.Error(),
			})
			skipped++
		} else {
			imported++
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"imported": imported,
		"skipped":  skipped,
		"errors":   errors,
	})
}

// Helper function to get CSV value by column name
func getCSVValue(row []string, headerIndex map[string]int, columnName string) string {
	if columnName == "" {
		return ""
	}
	idx, ok := headerIndex[columnName]
	if !ok || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}
