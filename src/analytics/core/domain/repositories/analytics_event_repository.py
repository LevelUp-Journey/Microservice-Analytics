"""Repository contract for analytics events."""
from __future__ import annotations

from abc import abstractmethod
from datetime import datetime
from typing import Dict, Iterable, Protocol, Sequence

from analytics.core.domain.model.analytics_event import AnalyticsEvent


class AnalyticsEventRepository(Protocol):
    """Interface describing persistence operations for analytics events."""

    @abstractmethod
    async def save(self, event: AnalyticsEvent) -> None:
        """Persist a new analytics event."""

    @abstractmethod
    async def bulk_save(self, events: Sequence[AnalyticsEvent]) -> None:
        """Persist multiple analytics events efficiently."""

    @abstractmethod
    async def get_recent_events(self, *, limit: int = 100) -> Iterable[AnalyticsEvent]:
        """Return the most recent events for exploratory queries."""

    @abstractmethod
    async def count_by_type(self, *, start: datetime, end: datetime) -> Dict[str, int]:
        """Aggregate events by type within the provided interval."""

    @abstractmethod
    async def time_series_count(
        self,
        *,
        start: datetime,
        end: datetime,
        interval_minutes: int = 60,
    ) -> Sequence[tuple[datetime, int]]:
        """Return aggregated counts grouped by a time window sized in minutes."""
