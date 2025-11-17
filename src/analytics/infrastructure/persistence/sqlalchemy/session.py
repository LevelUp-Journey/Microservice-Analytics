"""SQLAlchemy session and engine management."""
from __future__ import annotations

from contextlib import asynccontextmanager
from typing import AsyncIterator

from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from analytics.config import get_settings

_settings = get_settings()

_engine = create_async_engine(
    _settings.database_dsn,
    echo=False,
    pool_pre_ping=True,
    pool_size=5,
    max_overflow=10,
)

SessionFactory = async_sessionmaker(
    _engine,
    expire_on_commit=False,
    autoflush=False,
)


@asynccontextmanager
async def session_scope() -> AsyncIterator[AsyncSession]:
    """Provide a transactional scope for async operations."""

    session: AsyncSession = SessionFactory()
    try:
        yield session
        await session.commit()
    except Exception:
        await session.rollback()
        raise
    finally:
        await session.close()


async def init_db() -> None:
    """Ensure database connectivity during startup."""

    async with _engine.begin() as conn:  # pragma: no cover - startup sanity check
        await conn.run_sync(lambda _: None)


async def dispose_engine() -> None:
    """Release database resources during shutdown."""

    await _engine.dispose()
