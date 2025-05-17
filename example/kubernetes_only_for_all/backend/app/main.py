from fastapi import FastAPI, APIRouter
from app.database import mongodb
from app.services import (
    create_doc,
    get_doc,
    get_docs,
    update_doc,
    health_check
)
import asyncio
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from app.monitoring import ReplicationMonitor

# Configure OpenTelemetry
resource = Resource(attributes={
    "service.name": "backend",
    "service.version": "1.0.0"
})
trace.set_tracer_provider(TracerProvider(resource=resource))
jaeger_exporter = JaegerExporter(
    collector_endpoint="http://jaeger:14268/api/traces",
)
span_processor = BatchSpanProcessor(jaeger_exporter)
trace.get_tracer_provider().add_span_processor(span_processor)

router = APIRouter(prefix="/api")

@router.post("/create_document/")
async def create_document(data: dict):
    return await create_doc(data)

@router.get("/read_documents/")
async def read_documents(skip: int = 0, limit: int = 100):
    return await get_docs(skip=skip, limit=limit)

@router.get("/read_document/{doc_id}")
async def read_document(doc_id: str):
    return await get_doc(doc_id)

@router.put("/update_document/{doc_id}")
async def update_document(doc_id: str, data: dict):
    return await update_doc(doc_id, data)

@router.get("/health")
async def check_health():
    return await health_check()

app = FastAPI(title="Distributed Database Deployment")
FastAPIInstrumentor.instrument_app(app)
replication_monitor = ReplicationMonitor()

@app.on_event("startup")
async def startup():
    await mongodb.connect()

@app.on_event("shutdown")
async def shutdown():
    await mongodb.close()

# Include router with prefix /api
app.include_router(router)
