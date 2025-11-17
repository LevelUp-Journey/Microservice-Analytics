"""Kafka consumer responsible for ingesting analytics events."""
from __future__ import annotations

import asyncio
import contextlib
import json
import logging
from typing import Any, Optional

from aiokafka import AIOKafkaConsumer
from aiokafka.errors import ConsumerStoppedError

from analytics.application.commands.record_analytics_event import (
    RecordAnalyticsEventCommand,
    RecordAnalyticsEventHandler,
)
from analytics.config import Settings
from analytics.interfaces.messaging.transformers import (
    UnknownEventError,
    map_external_event,
)

_logger = logging.getLogger(__name__)


class KafkaAnalyticsConsumer:
    """Coordinates consumption of analytics events from Kafka."""

    def __init__(
        self,
        settings: Settings,
        command_handler: RecordAnalyticsEventHandler,
    ) -> None:
        self._settings = settings
        self._command_handler = command_handler
        self._consumer: Optional[AIOKafkaConsumer] = None
        self._task: Optional[asyncio.Task[None]] = None

    async def start(self) -> None:
        if self._consumer is not None:
            return

        common_config = self._settings.kafka_common_config
        self._consumer = AIOKafkaConsumer(
            self._settings.kafka_topic,
            group_id=self._settings.kafka_consumer_group,
            enable_auto_commit=True,
            auto_offset_reset=self._settings.kafka_auto_offset_reset,
            value_deserializer=lambda message: message.decode("utf-8"),
            **common_config,
        )

        await self._consumer.start()
        _logger.info(
            "Kafka consumer started", extra={"topic": self._settings.kafka_topic}
        )
        self._task = asyncio.create_task(self._consume_loop())

    async def stop(self) -> None:
        if self._consumer is None:
            return

        if self._task:
            self._task.cancel()
            with contextlib.suppress(asyncio.CancelledError):
                await self._task

        try:
            await self._consumer.stop()
        finally:
            self._consumer = None
            self._task = None
            _logger.info("Kafka consumer stopped")

    async def _consume_loop(self) -> None:
        assert self._consumer is not None

        while True:
            try:
                async for message in self._consumer:
                    await self._process_message(message.value)
            except asyncio.CancelledError:  # pragma: no cover - cooperative shutdown
                raise
            except ConsumerStoppedError:  # pragma: no cover - occurs on manual stops
                _logger.debug("Kafka consumer stopped")
                break
            except Exception:  # pragma: no cover - log and retry
                _logger.exception("Kafka consumption loop crashed")
                await asyncio.sleep(2)
            else:
                break

    async def _process_message(self, raw_value: str) -> None:
        try:
            raw_payload: Any = json.loads(raw_value)
            try:
                command_payload = map_external_event(raw_payload)
            except UnknownEventError:
                _logger.warning(
                    "Unsupported analytics payload skipped", extra={"payload": raw_payload}
                )
                return

            command = RecordAnalyticsEventCommand.model_validate(command_payload)
        except Exception:  # pragma: no cover - defensive logging
            _logger.exception("Failed to decode analytics event", extra={"raw": raw_value})
            return

        try:
            await self._command_handler(command)
        except Exception:  # pragma: no cover - avoid poisoning the consumer group
            _logger.exception(
                "Failed to persist analytics event",
                extra={"eventType": command.event_type, "source": command.source},
            )
