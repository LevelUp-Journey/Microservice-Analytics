# Analytics Event Ingestion Playbook

This document outlines the end-to-end flow for ingesting analytics events via Kafka, persisting them in PostgreSQL, and exposing query-friendly read models through the HTTP API.

## 1. Kafka Subscription

1. **Topic**: `ANALYTICS_KAFKA_TOPIC` (default `analytics.events`).
2. **Bootstrap servers**: configured via `ANALYTICS_KAFKA_BOOTSTRAP_SERVERS`.
3. **Consumer group**: `ANALYTICS_KAFKA_CONSUMER_GROUP` (default `analytics-service`).
4. **Configuration**: `auto_offset_reset=earliest` ensures fresh environments replay historical events. Optional SASL/SSL knobs exist for secure clusters.
5. **Runtime flow**:
   - During FastAPI startup the `KafkaAnalyticsConsumer` is instantiated.
   - The consumer subscribes to the configured topic and begins streaming records.
   - Each record value must be a UTF-8 JSON document matching the command contract below.

## 2. Event Payload Schema

Kafka record values may be submitted either in the canonical command format or in one
of the recognized upstream shapes described below. The consumer normalizes supported
payloads into the internal `RecordAnalyticsEventCommand` contract:

```json
{
  "eventType": "challenge.completed",
  "source": "challenge-service",
  "occurredAt": "2025-11-16T18:14:00Z",
  "tenantId": "academy-123",
  "payload": {
    "challengeId": "c12f4a13-6a8f-48ba-bb7c-6ed15b98c1b9",
    "studentId": "s-1029",
    "score": 97,
    "submittedAt": "2025-11-16T18:13:45Z"
  }
}
```

- `eventType` – domain semantic (bounded context + action), used for aggregation.
- `source` – emitting microservice or bounded context.
- `occurredAt` – ISO 8601 timestamp (UTC). Defaults to ingestion time if omitted.
- `tenantId` – optional multi-tenant discriminator.
- `payload` – schemaless JSON blob stored in PostgreSQL for later enrichment.

## 3. Persistence Flow

1. Kafka consumer deserializes the JSON payload.
2. The payload is validated against the command schema.
3. The `RecordAnalyticsEventHandler` creates an immutable `AnalyticsEvent` aggregate.
4. `SQLAlchemyAnalyticsEventRepository` persists the event:
   - Table: `analytics_events`.
   - Columns: `id`, `event_type`, `source`, `occurred_at`, `ingested_at`, `tenant_id`, `payload`.
   - `payload` leverages PostgreSQL `JSONB` for efficient querying.
5. Transactions commit per event batch (single event or bulk) to ensure idempotence.

### Supported external events

| Domain emission                            | Canonical event type                | Source             | Notes |
|-------------------------------------------|-------------------------------------|--------------------|-------|
| `iam registered`                          | `iam.user-registered`               | `iam-service`      | `registeredAt` arrays are converted to ISO timestamps and stored as `registeredAtIso` |
| `guides.challenges`                       | `guides.challenge-linked`           | `guides-service`   | Epoch floats translated to UTC `occurredAt` + `occurredAtIso` |
| `excecution analytics`                    | `execution.analytics`               | `execution-service`| Nanosecond precision ISO strings truncated safely to microseconds |
| `community registration`                  | `community.user-registered`         | `community-service`| `profileUrl` is null on registrations and used to disambiguate from updates |
| `community profile updated`               | `community.profile-updated`         | `community-service`| Updates require non-null `profileUrl` |
| `challenges completion analytics`         | `challenges.solution-completed`     | `challenges-service`| Experience/scoring metrics stored as-is with ISO-normalized `occurredOnIso` |

If a payload does not match any known signature it is skipped with a structured log so
unknown events never poison the consumer group.

## 4. Query/Reporting Patterns

- **Recent events**: `GET /v1/analytics/events/recent?limit=50` – returns chronological snapshots for debugging, dashboards, or audit trails.
- **Aggregated metrics**: `GET /v1/analytics/metrics/summary?windowMinutes=60`
  - Response contains counts per event type and time-bucketed histograms derived from `EventAggregationService`.
  - Window defaults to the `ANALYTICS_METRICS_WINDOW_MINUTES` setting.

## 5. Extending the Flow

1. **Additional projections**
   - Create query handlers under `src/analytics/application/queries/`.
   - Extend `EventAggregationService` or compose new domain services.
   - Expose via versioned routers in `interfaces/api/v1`.
2. **Enriched persistence**
   - Add calculated columns or materialized views in PostgreSQL for common aggregations.
   - Introduce Alembic migrations to automate schema evolution.
3. **Data products**
   - Export read models to downstream consumers (e.g., Kafka projections, REST, or gRPC) using outbox patterns.

## 6. Operational Considerations

- **Idempotency**: If upstream producers can redeliver messages, consider adding a natural key within `payload` (e.g., `eventId`) and enforcing uniqueness constraints or deduplication logic in the repository.
- **Schema evolution**: Version `eventType` names or embed `payload.version` to coordinate changes across producers and consumers.
- **Dead-letter queues**: Extend the consumer to route invalid events to a companion topic for later analysis rather than dropping them.
- **Monitoring**: Wire metrics and traces (e.g., OpenTelemetry) around consumer lag, handler throughput, and database latency.
