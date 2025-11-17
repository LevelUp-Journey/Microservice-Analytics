"""Version 1 API routes for analytics."""
from __future__ import annotations

from fastapi import APIRouter, Depends, status

from analytics.application.commands.record_analytics_event import (
    RecordAnalyticsEventCommand,
    RecordAnalyticsEventHandler,
)
from analytics.application.queries.fetch_event_metrics import (
    FetchEventMetricsHandler,
    FetchEventMetricsQuery,
    FetchRecentEventsHandler,
    FetchRecentEventsQuery,
)
from analytics.interfaces.api.dependencies import (
    get_fetch_metrics_handler,
    get_fetch_recent_events_handler,
    get_record_event_handler,
)

router = APIRouter(prefix="/analytics", tags=["analytics"])


@router.post("/events", status_code=status.HTTP_201_CREATED)
async def ingest_event(
    command: RecordAnalyticsEventCommand,
    handler: RecordAnalyticsEventHandler = Depends(get_record_event_handler),
) -> dict[str, object]:
    """Persist an analytics event via the command handler."""

    event = await handler(command)
    return {
        "id": str(event.id),
        "eventType": event.event_type,
        "source": event.source,
        "occurredAt": event.occurred_at.isoformat(),
        "ingestedAt": event.ingested_at.isoformat(),
        "tenantId": event.tenant_id,
    }


@router.get("/metrics/summary")
async def summarize_metrics(
    query: FetchEventMetricsQuery = Depends(),
    handler: FetchEventMetricsHandler = Depends(get_fetch_metrics_handler),
) -> dict[str, object]:
    """Expose aggregated analytics metrics over a configurable window."""

    return await handler(query)


@router.get("/events/recent")
async def recent_events(
    query: FetchRecentEventsQuery = Depends(),
    handler: FetchRecentEventsHandler = Depends(get_fetch_recent_events_handler),
) -> dict[str, object]:
    """List the most recent analytics events for exploratory analysis."""

    events = await handler(query)
    return {"items": events, "count": len(events)}
