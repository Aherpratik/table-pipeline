# Dynamic HTML Table Ingestion Pipeline

## Overview

This project implements an end-to-end data ingestion pipeline in Go that extracts structured data from HTML tables, processes it, and stores it in a relational database using a streaming architecture.

The system is designed to work with any publicly available HTML table by dynamically inferring its schema and adapting storage and visualization accordingly.

In addition to the ingestion pipeline, the project also includes a lightweight API and frontend dashboard to explore and visualize the ingested data.

---

## Architecture

The system follows a streaming pipeline design:

Source URL
→ HTML Fetcher
→ Table Parser
→ Schema Inference
→ Kafka Producer
→ Kafka Topic
→ Kafka Consumer
→ PostgreSQL
→ API Layer
→ Frontend Dashboard

---

## Features

### 1. Dynamic HTML Table Extraction

* Fetches data from any configurable URL
* Parses HTML tables using GoQuery
* No hardcoded schema or structure

---

### 2. Schema Inference

Automatically detects column types:

* INT
* FLOAT
* TEXT
* TIMESTAMP (basic support)

Handles real-world messy values such as:

* commas (1,000,000)
* symbols ($, %)
* citation artifacts ([1], [2])

---

### 3. Kafka-Based Streaming

* Producer publishes each row as a JSON message
* Consumer reads from Kafka and processes data asynchronously
* Decouples ingestion from persistence

---

### 4. Dynamic Database Handling

* Automatically creates tables if they do not exist
* Adds missing columns when schema evolves
* Supports flexible datasets without migrations

---

### 5. Idempotent Writes

* Each row is assigned a unique `row_hash`
* Duplicate rows are ignored using:

  ```
  ON CONFLICT (row_hash) DO NOTHING
  ```

---

### 6. Batch Inserts

* Consumer buffers messages before writing to DB
* Reduces database round trips
* Improves throughput and efficiency

---

### 7. Configuration via Environment Variables

All configuration is externalized:

* Source URL
* Kafka settings
* Database connection
* Active dataset

---

### 8. API Layer

* Exposes `/api/dataset`
* Returns:

  * table name
  * schema
  * rows
  * detected numeric/text columns

---

### 9. Dynamic Frontend Dashboard

* Automatically adapts to any dataset
* No hardcoded columns
* Features:

  * search across all fields
  * dynamic charts (based on inferred schema)
  * table view
  * schema visualization

---

## How to Run

### 1. Start Kafka & Postgres

```bash
docker compose up -d
```

### 2. Setup environment

```bash
cp .env.example .env
```

Update values if needed.

---

### 3. Run Producer

```bash
go run ./cmd/producer
```

---

### 4. Run Consumer

```bash
go run ./cmd/consumer
```

---

### 5. Run API

```bash
go run ./cmd/api
```

---

### 6. Run Frontend

```bash
cd frontend
npm install
npm run dev
```

Open:

```
http://localhost:5173
```

---

## Example Dataset

Tested with:

* Wikipedia railway stations dataset
* Population dataset

The system adapts automatically to different schemas.

---

## Design Decisions

* **Kafka** used for decoupling ingestion and persistence
* **Schema inference before DB write** avoids rigid schemas
* **Dynamic table creation** enables flexibility
* **Batch inserts** improve performance
* **Idempotency via hashing** ensures data consistency
* **Frontend driven by schema** allows zero hardcoding

---

## Tradeoffs

* Schema inference is heuristic-based (not perfect for all edge cases)
* Producer currently sends messages sequentially (can be optimized)
* UI is optimized for clarity over responsiveness

---

## Future Improvements

* Stronger type inference (dates, currencies, units)
* Producer-side batching
* Pagination and filtering in frontend
* Support multiple datasets simultaneously
* Deploy as microservices

---

## Summary

This project demonstrates a complete data pipeline from unstructured web data to structured storage and visualization. It emphasizes flexibility, scalability, and clean system design while handling real-world data inconsistencies.
