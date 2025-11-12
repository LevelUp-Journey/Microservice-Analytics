from fastapi import FastAPI
from fastapi.responses import RedirectResponse
from scalar_fastapi import get_scalar_api_reference

app = FastAPI(
    title="Microservice Analytics",
    version="0.1.0",
    docs_url=None,  # Deshabilitar Swagger UI
    redoc_url=None  # Deshabilitar ReDoc
)

@app.get("/hello")
async def hello_world():
    """
    Hello World endpoint

    Returns a simple greeting message.
    """
    return {"message": "Hello World"}

@app.get("/", include_in_schema=False)
async def root():
    return RedirectResponse(url="/docs")

@app.get("/docs", include_in_schema=False)
async def scalar_html():
    return get_scalar_api_reference(
        openapi_url=app.openapi_url,
        title=app.title,
    )

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
