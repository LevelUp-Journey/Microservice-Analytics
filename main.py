"""FastAPI application entry-point for the analytics microservice."""
from __future__ import annotations

from fastapi import FastAPI, Response
from fastapi.responses import RedirectResponse
from scalar_fastapi import get_scalar_api_reference

from analytics.config import get_settings
from analytics.interfaces.api.v1.routers import router as analytics_router
from analytics.interfaces.lifecycle.startup import register_lifecycle


def create_app() -> FastAPI:
    """Application factory that wires dependencies and routers."""

    settings = get_settings()

    app = FastAPI(
        title="Analytics Microservice",
        version="0.1.0",
        docs_url=None,
        redoc_url=None,
        openapi_url=f"/{settings.api_version}/openapi.json",
    )

    app.include_router(analytics_router, prefix=f"/{settings.api_version}")

    register_lifecycle(app, settings)

    @app.get("/", include_in_schema=False)
    async def root() -> RedirectResponse:
        return RedirectResponse(url="/docs", status_code=307)

    @app.get("/docs", include_in_schema=False)
    async def scalar_html() -> Response:
        return get_scalar_api_reference(
            openapi_url=app.openapi_url,
            title=app.title,
        )

    @app.get("/health", include_in_schema=False)
    async def health() -> dict[str, str]:
        return {"status": "UP"}

    return app


app = create_app()


if __name__ == "__main__":
    import uvicorn

    settings = get_settings()
    uvicorn.run(
        "main:app",
        host=settings.service_host,
        port=settings.service_port,
        reload=False,
    )
