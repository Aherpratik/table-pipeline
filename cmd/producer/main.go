package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"table-pipeline/internal/config"
	"table-pipeline/internal/fetcher"
	"table-pipeline/internal/kafka"
	"table-pipeline/internal/model"
	"table-pipeline/internal/parser"
	"table-pipeline/internal/schema"
	"table-pipeline/internal/util"
)

func main() {
	cfg := config.Load()
	log.Println("starting producer")

	html, err := fetcher.FetchHTML(cfg.SourceURL)
	if err != nil {
		log.Fatalf("fetch failed: %v", err)
	}

	parsed, err := parser.ParseBestTable(html)
	if err != nil {
		log.Fatalf("parse failed: %v", err)
	}

	log.Printf("parsed %d rows and %d headers", len(parsed.Rows), len(parsed.Headers))

	inferredSchema := schema.InferSchema(parsed.Rows, parsed.Headers)
	log.Printf("inferred schema for %d columns", len(inferredSchema))

	writer := kafka.NewWriter(cfg.KafkaBroker, cfg.KafkaTopic)
	defer writer.Close()

	for i, row := range parsed.Rows {
		rowHash, err := util.HashRow(row)
		if err != nil {
			log.Fatalf("hash generation failed: %v", err)
		}

		msg := model.RowMessage{
			TableName: cfg.TableName,
			Schema:    inferredSchema,
			Data:      row,
			RowHash:   rowHash,
		}

		payload, err := json.Marshal(msg)
		if err != nil {
			log.Fatalf("json marshal failed: %v", err)
		}

		writeCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		err = kafka.WriteMessage(writeCtx, writer, []byte(rowHash), payload)
		cancel()

		if err != nil {
			log.Fatalf("kafka write failed at row %d: %v", i, err)
		}
	}

	log.Printf("published %d rows to kafka topic %s", len(parsed.Rows), cfg.KafkaTopic)
}
