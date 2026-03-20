package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SourceURL   string
	TableName   string
	KafkaBroker string
	KafkaTopic  string
	KafkaGroup  string
	PostgresURL string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		SourceURL:   os.Getenv("SOURCE_URL"),
		TableName:   os.Getenv("TABLE_NAME"),
		KafkaBroker: os.Getenv("KAFKA_BROKER"),
		KafkaTopic:  os.Getenv("KAFKA_TOPIC"),
		KafkaGroup:  os.Getenv("KAFKA_GROUP_ID"),
		PostgresURL: os.Getenv("POSTGRES_URL"),
	}

	if cfg.SourceURL == "" ||
		cfg.TableName == "" ||
		cfg.KafkaBroker == "" ||
		cfg.KafkaTopic == "" ||
		cfg.KafkaGroup == "" ||
		cfg.PostgresURL == "" {
		log.Fatal("missing required environment variables")
	}

	return cfg
}
