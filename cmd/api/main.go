package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"table-pipeline/internal/config"

	"github.com/jackc/pgx/v5"
)

type DatasetResponse struct {
	TableName      string                   `json:"table_name"`
	Columns        []ColumnInfo             `json:"columns"`
	Rows           []map[string]interface{} `json:"rows"`
	RowCount       int                      `json:"row_count"`
	NumericColumns []string                 `json:"numeric_columns"`
	TextColumns    []string                 `json:"text_columns"`
}

type ColumnInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func main() {
	cfg := config.Load()
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("db connection failed: %v", err)
	}
	defer conn.Close(ctx)

	http.HandleFunc("/api/dataset", func(w http.ResponseWriter, r *http.Request) {
		tableName := r.URL.Query().Get("table")
		if tableName == "" {
			tableName = getenvOrDefault("ACTIVE_TABLE_NAME", cfg.TableName)
		}
		if tableName == "" {
			tableName = "railway_stations"
		}

		columns, err := getColumns(ctx, conn, tableName)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to load columns: %v", err), http.StatusInternalServerError)
			return
		}

		rows, err := getRows(ctx, conn, tableName, columns)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to load rows: %v", err), http.StatusInternalServerError)
			return
		}

		rowCount, err := getRowCount(ctx, conn, tableName)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to load row count: %v", err), http.StatusInternalServerError)
			return
		}

		numericCols, textCols := splitColumns(columns, rows)

		resp := DatasetResponse{
			TableName:      tableName,
			Columns:        columns,
			Rows:           rows,
			RowCount:       rowCount,
			NumericColumns: numericCols,
			TextColumns:    textCols,
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(resp)
	})

	log.Println("Dynamic API server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getColumns(ctx context.Context, conn *pgx.Conn, tableName string) ([]ColumnInfo, error) {
	query := `
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = $1
		ORDER BY ordinal_position
	`

	rows, err := conn.Query(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []ColumnInfo
	for rows.Next() {
		var c ColumnInfo
		if err := rows.Scan(&c.Name, &c.Type); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, nil
}

func getRows(ctx context.Context, conn *pgx.Conn, tableName string, columns []ColumnInfo) ([]map[string]interface{}, error) {
	colNames := make([]string, 0, len(columns))
	for _, c := range columns {
		colNames = append(colNames, c.Name)
	}

	query := fmt.Sprintf("SELECT %s FROM %s LIMIT 500", strings.Join(colNames, ", "), tableName)

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	fieldDescriptions := rows.FieldDescriptions()

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}

		record := make(map[string]interface{})
		for i, fd := range fieldDescriptions {
			record[string(fd.Name)] = normalizeValue(values[i])
		}
		result = append(result, record)
	}

	return result, nil
}

func getRowCount(ctx context.Context, conn *pgx.Conn, tableName string) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	var count int
	err := conn.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func splitColumns(columns []ColumnInfo, rows []map[string]interface{}) ([]string, []string) {
	var numericCols []string
	var textCols []string

	for _, col := range columns {
		if col.Name == "row_hash" {
			continue
		}

		isNumeric := false
		for _, row := range rows {
			if v, ok := row[col.Name]; ok && v != nil {
				switch v.(type) {
				case int, int32, int64, float32, float64:
					isNumeric = true
				}
				if _, ok := tryParseNumber(fmt.Sprintf("%v", v)); ok {
					isNumeric = true
				}
				break
			}
		}

		if isNumeric {
			numericCols = append(numericCols, col.Name)
		} else {
			textCols = append(textCols, col.Name)
		}
	}

	return numericCols, textCols
}

func normalizeValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case string:
		clean := regexp.MustCompile(`\\[\\d+\\]`).ReplaceAllString(val, "")
		clean = strings.TrimSpace(clean)
		if num, ok := tryParseNumber(clean); ok {
			return num
		}
		return clean
	default:
		return v
	}
}

func tryParseNumber(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "%", "")
	s = strings.ReplaceAll(s, "$", "")
	s = regexp.MustCompile(`\\[\\d+\\]`).ReplaceAllString(s, "")
	if s == "" {
		return 0, false
	}

	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func getenvOrDefault(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}