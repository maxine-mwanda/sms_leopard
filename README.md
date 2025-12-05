# SMSLeopard - RabbitMQ + MySQL (Full)

This repository contains a small Go backend implementing campaign management, using:
- MySQL for persistent storage
- RabbitMQ for queueing outbound message work
- Project layout inspired by maxine-mwanda/lipad

Quickstart (for local dev using real MySQL & RabbitMQ)
1. Start MySQL and RabbitMQ (docker-compose or system services).
2. Set env:
   - DSN (e.g. user:pass@tcp(localhost:3306)/smsleopard?parseTime=true)
   - AMQP_URL (e.g. amqp://guest:guest@localhost:5672/)
3. `go run ./cmd/server`
4. Use the HTTP endpoints:
   - POST /campaigns { name, template }
   - POST /campaigns/send { campaign_id }
   - GET /campaigns?limit=&offset=
   - POST /preview { template, customer }

Tests
- Unit tests live in `tests/` and use sqlmock so they run without real MySQL or RabbitMQ.
- Run: `go test ./...`

