# Centralized Logging System - Golang Microservices (Assignment)

This repository contains a minimal but complete implementation for the "Centralized Logging System"
assignment. It includes three microservices:

- `client` - simulates log-generating clients that send JSON logs to the log collector over TCP.
- `log-collector` - accepts logs over TCP, parses and enriches them, forwards to log-server.
- `log-server` - central logging service that stores logs (file-based) and provides querying APIs.

## How to run (using Docker Compose)

Make sure you have Docker and docker-compose installed.

Build and start all services:
```bash
docker-compose up --build
```

Services:
- Log server API: http://localhost:8081
  - POST /ingest  -> accept JSON log entries
  - GET /logs -> query logs, supports `service`, `level`, `username`, `is.blacklisted`, `limit`, `sort=timestamp`
  - GET /metrics -> metrics: total_logs, grouped counts

Clients will continuously generate logs and send them to the collector.

## Files
- `log-server/` - central logging microservice
- `log-collector/` - receiver and forwarder
- `client/` - log generator
- `docker-compose.yml` - to launch all services
- `README.md` - this help

## Notes & Tests
- Storage is file-based (`/data/logs.jsonl`) and abstracted via `Storage` interface.
- Basic unit tests are included for parsing/storage functions in production code as examples.

## Example curl
Ingest a synthetic log directly:
```bash
curl -X POST -H "Content-Type: application/json" -d '{"timestamp":"2025-07-29T12:35:24Z","event.category":"login.audit","username":"root","hostname":"aiops9242","severity":"INFO","raw.message":"<86> aiops9242 sudo: session opened for user root"}' http://localhost:8081/ingest
```

Query logs:
```bash
curl 'http://localhost:8081/logs?username=root&limit=10'
```
