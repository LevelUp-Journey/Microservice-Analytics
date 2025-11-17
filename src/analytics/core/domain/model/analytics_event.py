"""Domain model representing an analytics event within the system."""
from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Any, Dict, Optional
from uuid import UUID, uuid4


@dataclass(slots=True)
class AnalyticsEvent:
    """Immutable representation of a captured analytics event."""

    id: UUID
    event_type: str
    source: str
    occurred_at: datetime
    payload: Dict[str, Any]
    tenant_id: Optional[str] = None
    ingested_at: datetime = field(default_factory=lambda: datetime.now(tz=timezone.utc))

    @classmethod
    def new(
        cls,
        event_type: str,
        source: str,
        payload: Dict[str, Any],
        occurred_at: Optional[datetime] = None,
        tenant_id: Optional[str] = None,
    ) -> "AnalyticsEvent":
        """Factory that enforces invariants when creating a domain entity."""

        event_time = occurred_at or datetime.now(tz=timezone.utc)
        if event_time.tzinfo is None:
            event_time = event_time.replace(tzinfo=timezone.utc)

        return cls(
            id=uuid4(),
            event_type=event_type,
            source=source,
            occurred_at=event_time,
            payload=payload,
            tenant_id=tenant_id,
        )
