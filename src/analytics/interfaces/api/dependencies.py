"""Dependency wiring for the API layer."""
from __future__ import annotations

from functools import lru_cache

from analytics.application.commands.record_analytics_event import (
    RecordAnalyticsEventHandler,
)
from analytics.application.queries.fetch_event_metrics import (
    FetchEventMetricsHandler,
    FetchRecentEventsHandler,
)
from analytics.config import Settings, get_settings
from analytics.core.services.aggregation_service import EventAggregationService
from analytics.infrastructure.persistence.sqlalchemy.repositories import (
    SQLAlchemyAnalyticsEventRepository,
)


@lru_cache(maxsize=1)
def get_repository() -> SQLAlchemyAnalyticsEventRepository:
    return SQLAlchemyAnalyticsEventRepository()


@lru_cache(maxsize=1)
def get_aggregation_service() -> EventAggregationService:
    return EventAggregationService(repository=get_repository())


@lru_cache(maxsize=1)
def get_record_event_handler() -> RecordAnalyticsEventHandler:
    return RecordAnalyticsEventHandler(repository=get_repository())


@lru_cache(maxsize=1)
def get_fetch_metrics_handler() -> FetchEventMetricsHandler:
    return FetchEventMetricsHandler(service=get_aggregation_service())


@lru_cache(maxsize=1)
def get_fetch_recent_events_handler() -> FetchRecentEventsHandler:
    return FetchRecentEventsHandler(service=get_aggregation_service())


def get_app_settings() -> Settings:
    return get_settings()
