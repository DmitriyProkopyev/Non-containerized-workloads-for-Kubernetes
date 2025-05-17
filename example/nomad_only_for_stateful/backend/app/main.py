from fastapi import FastAPI
from database import mongodb
from services import (
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
from monitoring import ReplicationMonitor

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

app = FastAPI(title="Distributed Database Deployment")

# Instrument FastAPI
FastAPIInstrumentor.instrument_app(app)

# Initialize replication monitor
replication_monitor = ReplicationMonitor()

@app.on_event("startup")
async def startup():
    await mongodb.connect()

@app.on_event("shutdown")
async def shutdown():
    await mongodb.close()

@app.post(
    "/create_document/",
    response_model=dict,
    summary="Create a new document",
    description="Creates a new document in the database with the provided data in json format.",
    response_description="Created document data with assigned id."
)
async def create_document(data: dict):
    return await create_doc(data)

@app.get(
    "/read_documents/",
    response_model=list[dict],
    summary="Get a list of documents",
    description="Retrieves multiple documents from the database. Accepts parameters int skip and limit.",
    response_description="List of documents."
)

async def read_documents(skip: int = 0, limit: int = 100):
    return await get_docs(skip=skip, limit=limit)

@app.get(
    "/read_document/{doc_id}",
    response_model=dict,
    summary="Get a document by ID",
    description="Retrieves a single document by its ID from the database.",
    response_description="Requested document data."
)
async def read_document(doc_id: str):
    return await get_doc(doc_id)

@app.put(
    "/update_document/{doc_id}",
    response_model=dict,
    summary="Update a document by ID",
    description="Updates an existing document with the provided data by its ID.",
    response_description="Updated document data."
)
async def update_document(doc_id: str, data: dict):
    return await update_doc(doc_id, data)

@app.get("/health")
async def check_health():
    return await health_check()
