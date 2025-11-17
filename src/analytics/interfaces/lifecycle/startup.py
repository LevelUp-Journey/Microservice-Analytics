"""Application lifecycle wiring."""
from __future__ import annotations

import logging

from fastapi import FastAPI

from analytics.application.commands.record_analytics_event import (
    RecordAnalyticsEventHandler,
)
from analytics.config import Settings
from analytics.infrastructure.messaging.kafka.consumer import KafkaAnalyticsConsumer
from analytics.infrastructure.persistence.sqlalchemy.session import (
    dispose_engine,
    init_db,
)
from analytics.infrastructure.service_discovery.eureka import EurekaServiceRegistry
from analytics.interfaces.api.dependencies import get_record_event_handler

_logger = logging.getLogger(__name__)


class ApplicationLifecycle:
    """Encapsulates startup and shutdown orchestration."""

    def __init__(self, settings: Settings) -> None:
        self._settings = settings
        self._record_event_handler: RecordAnalyticsEventHandler = get_record_event_handler()
        self._kafka_consumer = KafkaAnalyticsConsumer(settings, self._record_event_handler)
        self._eureka_registry = EurekaServiceRegistry(settings)

    async def on_startup(self) -> None:
        _logger.info("Starting up analytics service")
        await init_db()
        await self._kafka_consumer.start()
        await self._eureka_registry.register()

    async def on_shutdown(self) -> None:
        _logger.info("Shutting down analytics service")
        await self._kafka_consumer.stop()
        await self._eureka_registry.deregister()
        await dispose_engine()


def register_lifecycle(app: FastAPI, settings: Settings) -> ApplicationLifecycle:
    """Attach startup and shutdown hooks to the FastAPI application."""

    lifecycle = ApplicationLifecycle(settings)

    @app.on_event("startup")
    async def _startup() -> None:  # pragma: no cover - framework integration
        await lifecycle.on_startup()

    @app.on_event("shutdown")
    async def _shutdown() -> None:  # pragma: no cover - framework integration
        await lifecycle.on_shutdown()

    return lifecycle
