from fastapi import HTTPException
from app.database import mongodb
from app.config import settings
from app.utils import to_json, validate_id
import uuid


async def create_doc(data: dict) -> dict:
    col = mongodb.get_collection()
    result = await col.insert_one(data)
    return await get_doc(str(result.inserted_id))


async def get_doc(doc_id: str) -> dict:
    col = mongodb.get_collection()
    doc = await col.find_one({"_id": validate_id(doc_id)})
    if not doc:
        raise HTTPException(404, "Document not found")
    return to_json(doc)

async def get_docs(skip: int = 0, limit: int = 100) -> list[dict]:
    col = mongodb.get_collection()
    cursor = col.find().skip(skip).limit(limit)
    docs = await cursor.to_list(length=limit)
    return to_json(docs)


async def update_doc(doc_id: str, data: dict) -> dict:
    col = mongodb.get_collection()
    await col.replace_one({"_id": validate_id(doc_id)}, data, upsert=True)
    return await get_doc(doc_id)


async def health_check() -> dict:
    try:
        await mongodb.db.command('ping')
        return {
            "status": "healthy",
            "mongo": "ok"
        }
    except Exception as e:
        return {
            "status": "unhealthy",
            "error": str(e)
        }
