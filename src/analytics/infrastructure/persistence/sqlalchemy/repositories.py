"""SQLAlchemy repository implementations."""
from __future__ import annotations

from datetime import datetime, timezone
from typing import Callable, Dict, Iterable, Sequence

from sqlalchemy import Select, func, select
from sqlalchemy.ext.asyncio import AsyncSession

from analytics.core.domain.model.analytics_event import AnalyticsEvent
from analytics.core.domain.repositories.analytics_event_repository import (
    AnalyticsEventRepository,
)
from analytics.infrastructure.persistence.sqlalchemy.models import AnalyticsEventRecord
from analytics.infrastructure.persistence.sqlalchemy.session import SessionFactory


class SQLAlchemyAnalyticsEventRepository(AnalyticsEventRepository):
    """Concrete repository for analytics events using SQLAlchemy."""

    def __init__(
        self,
        session_factory: Callable[[], AsyncSession] = SessionFactory,  # type: ignore[misc]
    ) -> None:
        self._session_factory = session_factory

    async def save(self, event: AnalyticsEvent) -> None:
        async with self._session_factory() as session:  # type: ignore[misc]
            await self._save_one(session, event)
            await session.commit()

    async def bulk_save(self, events: Sequence[AnalyticsEvent]) -> None:
        if not events:
            return
        async with self._session_factory() as session:  # type: ignore[misc]
            for event in events:
                await self._save_one(session, event)
            await session.commit()

    async def get_recent_events(self, *, limit: int = 100) -> Iterable[AnalyticsEvent]:
        stmt: Select[tuple[AnalyticsEventRecord]] = (
            select(AnalyticsEventRecord)
            .order_by(AnalyticsEventRecord.occurred_at.desc())
            .limit(limit)
        )
        async with self._session_factory() as session:  # type: ignore[misc]
            results = await session.execute(stmt)
            rows = results.scalars().all()

        return [self._to_domain(row) for row in rows]

    async def count_by_type(self, *, start: datetime, end: datetime) -> Dict[str, int]:
        stmt = (
            select(AnalyticsEventRecord.event_type, func.count())
            .where(AnalyticsEventRecord.occurred_at.between(start, end))
            .group_by(AnalyticsEventRecord.event_type)
        )
        async with self._session_factory() as session:  # type: ignore[misc]
            results = await session.execute(stmt)
        return {event_type: count for event_type, count in results.all()}

    async def time_series_count(
        self,
        *,
        start: datetime,
        end: datetime,
        interval_minutes: int = 60,
    ) -> Sequence[tuple[datetime, int]]:
        bucket_size = max(interval_minutes, 1)
        seconds_per_bucket = bucket_size * 60
        epoch_value = func.extract("epoch", AnalyticsEventRecord.occurred_at)
        bucket_expr = func.to_timestamp(
            func.floor(epoch_value / seconds_per_bucket) * seconds_per_bucket
        )

        stmt = (
            select(bucket_expr.label("bucket"), func.count())
            .where(AnalyticsEventRecord.occurred_at.between(start, end))
            .group_by("bucket")
            .order_by("bucket")
        )
        async with self._session_factory() as session:  # type: ignore[misc]
            results = await session.execute(stmt)
            rows = results.all()
        return [
            (
                bucket.replace(tzinfo=timezone.utc) if bucket.tzinfo is None else bucket,
                count,
            )
            for bucket, count in rows
        ]

    async def _save_one(self, session: AsyncSession, event: AnalyticsEvent) -> None:
        record = AnalyticsEventRecord(
            id=event.id,
            event_type=event.event_type,
            source=event.source,
            occurred_at=event.occurred_at,
            ingested_at=event.ingested_at,
            tenant_id=event.tenant_id,
            payload=event.payload,
        )
        session.add(record)

    @staticmethod
    def _to_domain(record: AnalyticsEventRecord) -> AnalyticsEvent:
        return AnalyticsEvent(
            id=record.id,
            event_type=record.event_type,
            source=record.source,
            occurred_at=record.occurred_at,
            payload=record.payload,
            tenant_id=record.tenant_id,
            ingested_at=record.ingested_at,
        )
