package schema

import (
	"regexp"
	"strconv"
	"strings"
)

type ColumnSchema struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

var datePattern = regexp.MustCompile(`^\d{4}([-/]\d{1,2}([-/]\d{1,2})?)?$`)

func InferSchema(rows []map[string]string, headers []string) []ColumnSchema {
	columns := make([]ColumnSchema, 0, len(headers))

	for _, header := range headers {
		columns = append(columns, ColumnSchema{
			Name: SanitizeColumnName(header),
			Type: inferColumnType(rows, header),
		})
	}

	return columns
}

func inferColumnType(rows []map[string]string, column string) string {
	hasValue := false
	allInt := true
	allFloat := true
	allDate := true

	for _, row := range rows {
		raw := strings.TrimSpace(row[column])
		if raw == "" {
			continue
		}

		hasValue = true
		cleaned := cleanNumeric(raw)

		if _, err := strconv.Atoi(cleaned); err != nil {
			allInt = false
		}

		if _, err := strconv.ParseFloat(cleaned, 64); err != nil {
			allFloat = false
		}

		if !datePattern.MatchString(raw) {
			allDate = false
		}
	}

	if !hasValue {
		return "TEXT"
	}

	switch {
	case allInt:
		return "INT"
	case allFloat:
		return "FLOAT"
	case allDate:
		return "TIMESTAMP"
	default:
		return "TEXT"
	}
}

func cleanNumeric(s string) string {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, "%", "")
	s = strings.ReplaceAll(s, "−", "-")
	s = strings.TrimSpace(s)
	return s
}

func SanitizeColumnName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`_+`).ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")

	if s == "" {
		return "column"
	}
	return s
}
