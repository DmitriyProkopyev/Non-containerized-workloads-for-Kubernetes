from motor.motor_asyncio import AsyncIOMotorClient
from app.config import settings

class MongoDB:
    def __init__(self):
        self.client = None
        self.db = None
        
    async def connect(self):
        self.client = AsyncIOMotorClient(settings.MONGO_URI)
        self.db = self.client[settings.MONGO_DB_NAME]
        await self.db.command("ping")
        await self.db["processed_requests"].create_index(
            "timestamp", 
            expireAfterSeconds=24*3600
        )
        print("Connected to MongoDB")

    async def close(self):
        if self.client:
            self.client.close()
            print("MongoDB connection closed")
    
    def get_processed_requests_collection(self):
        return self.db["processed_requests"]

    def get_collection(self, name="sample"):
        return self.db[name]

mongodb = MongoDB()
