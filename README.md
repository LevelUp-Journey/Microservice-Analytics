# Analytics Microservice

FastAPI-based analytics bounded context designed around domain-driven design (DDD) and CQRS principles. The service ingests domain events through Apache Kafka, persists them in PostgreSQL, and exposes read-oriented projections via HTTP APIs. Eureka integration enables service discovery within JVM-centric service meshes.

## Architecture Highlights
- **Bounded context:** `analytics` aggregate manages immutable `AnalyticsEvent` entities.
- **CQRS:** write operations flow through command handlers, while read models are produced by query handlers and a dedicated aggregation service.
- **Messaging integration:** an `aiokafka` consumer deserializes events and dispatches them to the command handler, ensuring resilience with retry-friendly loops.
- **Event normalization:** upstream payloads (IAM, Guides, Execution, Community, Challenges) are normalized into canonical analytics events before persistence.
- **Persistence:** SQLAlchemy Async ORM targets PostgreSQL using JSONB payload storage for schemaless event bodies.
- **Service discovery:** optional registration against Spring Cloud Eureka via `py-eureka-client`.
- **API ergonomics:** versioned FastAPI routes (`/v1/analytics/...`) plus Scalar documentation rendering.

## Project Layout

```
src/
  analytics/
	 application/        # Command & query handlers (CQRS)
	 config.py           # Centralized environment-backed settings
	 core/               # Domain model, repositories, domain services
	 infrastructure/     # Kafka, SQLAlchemy, Eureka adapters
	 interfaces/         # HTTP APIs and lifecycle orchestration
main.py                 # FastAPI application factory & router wiring
.env.example            # Environment variable template
```

## Prerequisites
- Python 3.11+
- PostgreSQL 14+ (or managed instance)
- Kafka 3.x cluster
- Optional: Eureka server (e.g., Spring Cloud Netflix Eureka)

## Local Setup
1. **Create environment & install dependencies**
	```bash
	python -m venv .venv
	source .venv/bin/activate
	pip install -U pip
	pip install -e .
	```
2. **Copy environment template**
	```bash
	cp .env.example .env
	```
	Adjust PostgreSQL, Kafka, and Eureka values as needed.
3. **Apply database schema**
	```sql
	CREATE TABLE IF NOT EXISTS analytics_events (
		 id UUID PRIMARY KEY,
		 event_type VARCHAR(120) NOT NULL,
		 source VARCHAR(120) NOT NULL,
		 occurred_at TIMESTAMPTZ NOT NULL,
		 ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		 tenant_id VARCHAR(64),
		 payload JSONB NOT NULL
	);
	```
	(Alembic migrations can be layered on later.)
4. **Run the application**
	```bash
	uvicorn main:app --reload
	```

## Running With Docker
> TODO: add Dockerfile/compose definition if containerization is required.

## Observability & Health
- `GET /health` – lightweight readiness probe.
- `GET /docs` – Scalar-generated interactive API explorer.

## Testing the API
```bash
curl -X POST http://localhost:8000/v1/analytics/events \
	  -H 'Content-Type: application/json' \
	  -d '{
			  "eventType": "challenge.completed",
			  "source": "challenge-service",
			  "payload": {"challengeId": "123", "studentId": "42"}
			}'

curl "http://localhost:8000/v1/analytics/metrics/summary?windowMinutes=60"

curl http://localhost:8000/v1/analytics/events/recent
```

## Next Steps
- Add Alembic migrations for schema evolution.
- Extend query-side projections (e.g., leaderboard endpoints).
- Harden Kafka consumer with dead-letter strategy and structured logging.
- Instrument with OpenTelemetry for distributed tracing.
