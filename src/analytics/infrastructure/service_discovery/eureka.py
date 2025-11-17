"""Integration with Spring Cloud Eureka for service discovery."""
from __future__ import annotations

import asyncio
import logging
from typing import Optional

from py_eureka_client import eureka_client

from analytics.config import Settings

_logger = logging.getLogger(__name__)


class EurekaServiceRegistry:
    """Registers the analytics service instance with Eureka."""

    def __init__(self, settings: Settings) -> None:
        self._settings = settings
        self._client: Optional[eureka_client.EurekaClient] = None

    async def register(self) -> None:
        if not self._settings.eureka_enabled:
            _logger.info("Eureka registration disabled via configuration")
            return
        if not self._settings.eureka_service_url:
            _logger.warning("Eureka registration skipped: service URL not configured")
            return

        instance_ip = str(
            self._settings.eureka_instance_ip or self._settings.service_host
        )
        instance_host = self._settings.eureka_instance_host or instance_ip

        self._client = eureka_client.EurekaClient(
            eureka_server=self._settings.eureka_service_url,
            app_name=self._settings.eureka_app_name,
            instance_port=self._settings.service_port,
            instance_host=instance_host,
            instance_ip=instance_ip,
            status_page_url=f"http://{instance_host}:{self._settings.service_port}/health",
            heartbeat_interval=self._settings.eureka_heartbeat_interval,
            registry_fetch_interval=self._settings.eureka_registry_fetch_interval,
        )

        _logger.info("Registering service instance with Eureka")
        await asyncio.to_thread(self._client.start)

    async def deregister(self) -> None:
        if self._client is None:
            return
        _logger.info("Deregistering service instance from Eureka")
        await asyncio.to_thread(self._client.stop)
        self._client = None
