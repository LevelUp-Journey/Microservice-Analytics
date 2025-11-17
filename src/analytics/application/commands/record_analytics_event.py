"""Command handler for recording analytics events."""
from __future__ import annotations

from datetime import datetime, timezone
from typing import Any, Dict, Optional

from pydantic import BaseModel, Field, field_validator

from analytics.core.domain.model.analytics_event import AnalyticsEvent
from analytics.core.domain.repositories.analytics_event_repository import (
    AnalyticsEventRepository,
)


class RecordAnalyticsEventCommand(BaseModel):
    """Command definition for persisting a single analytics event."""

    event_type: str = Field(..., min_length=1, max_length=120, alias="eventType")
    source: str = Field(..., min_length=1, max_length=120)
    occurred_at: Optional[datetime] = Field(default=None, alias="occurredAt")
    payload: Dict[str, Any]
    tenant_id: Optional[str] = Field(default=None, alias="tenantId", max_length=64)

    @field_validator("occurred_at")
    @classmethod
    def _ensure_timezone(cls, value: Optional[datetime]) -> Optional[datetime]:
        if value is None:
            return value
        return value if value.tzinfo else value.replace(tzinfo=timezone.utc)

    model_config = {
        "populate_by_name": True,
        "json_schema_extra": {
            "example": {
                "eventType": "challenge.completed",
                "source": "challenge-service",
                "occurredAt": "2025-11-16T18:14:00Z",
                "tenantId": "academy-123",
                "payload": {
                    "challengeId": "c12f4a13-6a8f-48ba-bb7c-6ed15b98c1b9",
                    "studentId": "s-1029",
                    "score": 97,
                    "submittedAt": "2025-11-16T18:13:45Z",
                },
            }
        },
    }


class RecordAnalyticsEventHandler:
    """Application service responsible for handling the record event command."""

    def __init__(self, repository: AnalyticsEventRepository) -> None:
        self._repository = repository

    async def __call__(self, command: RecordAnalyticsEventCommand) -> AnalyticsEvent:
        event = AnalyticsEvent.new(
            event_type=command.event_type,
            source=command.source,
            payload=command.payload,
            occurred_at=command.occurred_at,
            tenant_id=command.tenant_id,
        )
        await self._repository.save(event)
        return event
