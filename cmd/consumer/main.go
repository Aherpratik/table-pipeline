package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"table-pipeline/internal/config"
	"table-pipeline/internal/db"
	"table-pipeline/internal/kafka"
	"table-pipeline/internal/model"
)

const (
	batchSize     = 50
	flushInterval = 30 * time.Second
)

func main() {
	cfg := config.Load()
	log.Println("starting consumer")

	ctx := context.Background()

	conn, err := db.Connect(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("db connection failed: %v", err)
	}
	defer conn.Close(ctx)

	reader := kafka.NewReader(cfg.KafkaBroker, cfg.KafkaTopic, cfg.KafkaGroup)
	defer reader.Close()

	var batch []model.RowMessage
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}

		first := batch[0]

		if err := db.EnsureTable(ctx, conn, first.TableName, first.Schema); err != nil {
			log.Printf("ensure table failed: %v", err)
			return
		}

		if err := db.EnsureColumns(ctx, conn, first.TableName, first.Schema); err != nil {
			log.Printf("ensure columns failed: %v", err)
			return
		}

		if err := db.InsertRowsBatch(ctx, conn, batch); err != nil {
			log.Printf("batch insert failed: %v", err)
			return
		}

		log.Printf("inserted batch of %d rows", len(batch))
		batch = nil
	}

	for {
		select {
		case <-ticker.C:
			flush()

		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("kafka read error: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			var rowMsg model.RowMessage
			if err := json.Unmarshal(msg.Value, &rowMsg); err != nil {
				log.Printf("json unmarshal failed: %v", err)
				continue
			}

			batch = append(batch, rowMsg)

			if len(batch) >= batchSize {
				flush()
			}
		}
	}
}