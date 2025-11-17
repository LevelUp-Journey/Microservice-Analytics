"""Application configuration leveraging environment variables."""
from __future__ import annotations

from functools import lru_cache
from typing import Any, Dict, Optional

from pydantic import IPvAnyAddress, computed_field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Centralized application configuration."""

    app_name: str = "analytics-service"
    environment: str = "local"
    api_version: str = "v1"

    service_host: str = "0.0.0.0"
    service_port: int = 8000
    public_base_url: Optional[str] = None

    postgres_user: str
    postgres_password: str
    postgres_host: str = "localhost"
    postgres_port: int = 5432
    postgres_db: str

    kafka_bootstrap_servers: str = "localhost:9092"
    kafka_topic: str = "analytics.events"
    kafka_consumer_group: str = "analytics-service"
    kafka_auto_offset_reset: str = "earliest"
    kafka_security_protocol: Optional[str] = None
    kafka_sasl_mechanism: Optional[str] = None
    kafka_sasl_username: Optional[str] = None
    kafka_sasl_password: Optional[str] = None

    eureka_enabled: bool = False
    eureka_service_url: Optional[str] = None
    eureka_app_name: str = "analytics-service"
    eureka_instance_ip: Optional[IPvAnyAddress] = None
    eureka_instance_host: Optional[str] = None
    eureka_heartbeat_interval: int = 30
    eureka_registry_fetch_interval: int = 30

    metrics_window_minutes: int = 60

    model_config = SettingsConfigDict(
        env_file=(".env", ".env.local"),
        env_prefix="ANALYTICS_",
        case_sensitive=False,
        extra="ignore",
    )

    @computed_field  # type: ignore[misc]
    @property
    def database_dsn(self) -> str:
        """Construct the async database connection string."""

        return (
            f"postgresql+asyncpg://{self.postgres_user}:{self.postgres_password}"
            f"@{self.postgres_host}:{self.postgres_port}/{self.postgres_db}"
        )

    @property
    def kafka_common_config(self) -> Dict[str, Any]:
        """Base keyword arguments shared by Kafka clients."""

        config: Dict[str, Any] = {
            "bootstrap_servers": self.kafka_bootstrap_servers.split(","),
            "security_protocol": self.kafka_security_protocol,
            "sasl_mechanism": self.kafka_sasl_mechanism,
            "sasl_plain_username": self.kafka_sasl_username,
            "sasl_plain_password": self.kafka_sasl_password,
        }
        return {k: v for k, v in config.items() if v is not None}


@lru_cache(maxsize=1)
def get_settings() -> Settings:
    """Return a cached settings instance to avoid repeated disk I/O."""

    return Settings()
