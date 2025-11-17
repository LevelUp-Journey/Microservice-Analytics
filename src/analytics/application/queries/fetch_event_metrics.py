"""Query handlers for analytics read models."""
from __future__ import annotations

from datetime import datetime, timezone
from typing import Dict, Sequence

from pydantic import BaseModel, Field, field_validator

from analytics.core.services.aggregation_service import EventAggregationService


class FetchEventMetricsQuery(BaseModel):
    """Query definition for aggregated analytics metrics."""

    window_minutes: int = Field(default=60, alias="windowMinutes", ge=5, le=7 * 24 * 60)
    reference: datetime | None = Field(default=None)

    @field_validator("reference")
    @classmethod
    def _ensure_timezone(cls, value: datetime | None) -> datetime | None:
        if value is None:
            return value
        return value if value.tzinfo else value.replace(tzinfo=timezone.utc)

    model_config = {"populate_by_name": True}


class FetchEventMetricsHandler:
    """Handles metrics aggregation queries."""

    def __init__(self, service: EventAggregationService) -> None:
        self._service = service

    async def __call__(self, query: FetchEventMetricsQuery) -> Dict[str, object]:
        return await self._service.summary(
            window_minutes=query.window_minutes,
            reference=query.reference,
        )


class FetchRecentEventsQuery(BaseModel):
    """Query for the latest ingested analytics events."""

    limit: int = Field(default=50, ge=1, le=500)


class FetchRecentEventsHandler:
    """Returns a list of recent analytics events."""

    def __init__(self, service: EventAggregationService) -> None:
        self._service = service

    async def __call__(self, query: FetchRecentEventsQuery) -> Sequence[dict[str, object]]:
        return await self._service.recent_events(limit=query.limit)
