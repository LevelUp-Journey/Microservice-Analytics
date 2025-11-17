"""Domain service that orchestrates analytics aggregations."""
from __future__ import annotations

from datetime import datetime, timedelta, timezone
from typing import Dict, Sequence

from analytics.core.domain.repositories.analytics_event_repository import (
    AnalyticsEventRepository,
)


class EventAggregationService:
    """Provides aggregated insights over captured analytics events."""

    def __init__(self, repository: AnalyticsEventRepository) -> None:
        self._repository = repository

    async def summary(
        self,
        *,
        window_minutes: int,
        reference: datetime | None = None,
    ) -> Dict[str, object]:
        """Return summary statistics for the provided time window."""

        now = reference or datetime.now(tz=timezone.utc)
        window_start = now - timedelta(minutes=window_minutes)

        per_type = await self._repository.count_by_type(start=window_start, end=now)
        time_series = await self._repository.time_series_count(
            start=window_start, end=now, interval_minutes=max(window_minutes // 6, 5)
        )

        return {
            "windowStart": window_start.isoformat(),
            "windowEnd": now.isoformat(),
            "eventsPerType": per_type,
            "eventsPerInterval": [
                {"bucketStart": bucket.isoformat(), "count": count}
                for bucket, count in time_series
            ],
        }

    async def recent_events(self, *, limit: int = 50) -> Sequence[dict[str, object]]:
        """Return recent events in a shape tailored for the read model."""

        events = await self._repository.get_recent_events(limit=limit)
        return [
            {
                "id": str(event.id),
                "eventType": event.event_type,
                "source": event.source,
                "occurredAt": event.occurred_at.isoformat(),
                "tenantId": event.tenant_id,
                "payload": event.payload,
            }
            for event in events
        ]
