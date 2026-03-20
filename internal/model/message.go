package model

import "table-pipeline/internal/schema"

type RowMessage struct {
	TableName string                `json:"table_name"`
	Schema    []schema.ColumnSchema `json:"schema"`
	Data      map[string]string     `json:"data"`
	RowHash   string                `json:"row_hash"`
}
