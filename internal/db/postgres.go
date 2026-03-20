package db

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"table-pipeline/internal/model"
	"table-pipeline/internal/schema"

	"github.com/jackc/pgx/v5"
)

func Connect(ctx context.Context, dbURL string) (*pgx.Conn, error) {
	return pgx.Connect(ctx, dbURL)
}

func EnsureTable(ctx context.Context, conn *pgx.Conn, tableName string, cols []schema.ColumnSchema) error {
	createParts := []string{
		"row_hash TEXT PRIMARY KEY",
	}

	for _, col := range cols {
		createParts = append(createParts, fmt.Sprintf("%s %s", col.Name, mapDBType(col.Type)))
	}

	query := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (%s)",
		tableName,
		strings.Join(createParts, ", "),
	)

	_, err := conn.Exec(ctx, query)
	return err
}

func EnsureColumns(ctx context.Context, conn *pgx.Conn, tableName string, cols []schema.ColumnSchema) error {
	existing, err := getExistingColumns(ctx, conn, tableName)
	if err != nil {
		return err
	}

	for _, col := range cols {
		if existing[col.Name] {
			continue
		}

		query := fmt.Sprintf(
			"ALTER TABLE %s ADD COLUMN %s %s",
			tableName,
			col.Name,
			mapDBType(col.Type),
		)

		log.Printf("adding missing column %s to %s", col.Name, tableName)

		if _, err := conn.Exec(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

func InsertRow(ctx context.Context, conn *pgx.Conn, msg model.RowMessage) error {
	columnNames := []string{"row_hash"}
	placeholders := []string{"$1"}
	values := []interface{}{msg.RowHash}

	index := 2
	for _, col := range msg.Schema {
		columnNames = append(columnNames, col.Name)
		values = append(values, getValueForSanitizedColumn(msg.Data, col.Name))
		placeholders = append(placeholders, fmt.Sprintf("$%d", index))
		index++
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (row_hash) DO NOTHING",
		msg.TableName,
		strings.Join(columnNames, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := conn.Exec(ctx, query, values...)
	return err
}

func InsertRowsBatch(ctx context.Context, conn *pgx.Conn, batch []model.RowMessage) error {
	if len(batch) == 0 {
		return nil
	}

	first := batch[0]

	columnNames := []string{"row_hash"}
	for _, col := range first.Schema {
		columnNames = append(columnNames, col.Name)
	}

	var (
		allValues     []interface{}
		valueGroups   []string
		placeholderIx = 1
	)

	for _, msg := range batch {
		rowPlaceholders := []string{}

		allValues = append(allValues, msg.RowHash)
		rowPlaceholders = append(rowPlaceholders, fmt.Sprintf("$%d", placeholderIx))
		placeholderIx++

		for _, col := range msg.Schema {
			raw := getValueForSanitizedColumn(msg.Data, col.Name)
			normalized := normalizeValue(raw, col.Type)

			allValues = append(allValues, normalized)
			rowPlaceholders = append(rowPlaceholders, fmt.Sprintf("$%d", placeholderIx))
			placeholderIx++
		}

		valueGroups = append(valueGroups, fmt.Sprintf("(%s)", strings.Join(rowPlaceholders, ", ")))
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s ON CONFLICT (row_hash) DO NOTHING",
		first.TableName,
		strings.Join(columnNames, ", "),
		strings.Join(valueGroups, ", "),
	)

	_, err := conn.Exec(ctx, query, allValues...)
	return err
}

func getExistingColumns(ctx context.Context, conn *pgx.Conn, tableName string) (map[string]bool, error) {
	query := `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_name = $1
	`

	rows, err := conn.Query(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		result[name] = true
	}

	return result, nil
}

func getValueForSanitizedColumn(data map[string]string, sanitized string) string {
	for original, value := range data {
		if schema.SanitizeColumnName(original) == sanitized {
			return value
		}
	}
	return ""
}

func normalizeValue(raw string, dataType string) interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	switch dataType {
	case "INT":
		clean := cleanNumeric(raw)
		if clean == "" {
			return nil
		}
		v, err := strconv.ParseInt(clean, 10, 64)
		if err != nil {
			return nil
		}
		return v

	case "FLOAT":
		clean := cleanNumeric(raw)
		if clean == "" {
			return nil
		}
		v, err := strconv.ParseFloat(clean, 64)
		if err != nil {
			return nil
		}
		return v

	default:
		return raw
	}
}

func cleanNumeric(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, "%", "")
	s = strings.ReplaceAll(s, "−", "-")
	s = strings.ReplaceAll(s, "—", "")
	s = strings.ReplaceAll(s, "–", "")
	return strings.TrimSpace(s)
}


func mapDBType(t string) string {
	switch t {
	case "INT":
		return "BIGINT"
	case "FLOAT":
		return "DOUBLE PRECISION"
	case "TIMESTAMP":
		return "TIMESTAMP"
	default:
		return "TEXT"
	}
}