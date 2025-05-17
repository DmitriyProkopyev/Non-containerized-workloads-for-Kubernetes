from bson import json_util, ObjectId
from fastapi import HTTPException
import json

def to_json(data):
    if data is None:
        return {}
    return json.loads(json_util.dumps(data))

def validate_id(doc_id: str) -> ObjectId:
    if not ObjectId.is_valid(doc_id):
        raise HTTPException(400, "Invalid document ID format")
    return ObjectId(doc_id)
