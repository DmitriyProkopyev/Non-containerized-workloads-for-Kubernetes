from pymongo import MongoClient
import time

client = MongoClient("mongodb://localhost:27017")
db = client["testdb"]
collection = db["sample"]

print("Inserting test data...")
result = collection.insert_one({"status": "post-failure-insert", "timestamp": time.time()})
print(f"Inserted ID: {result.inserted_id}")

print("Query result:")
for doc in collection.find({}):
    print(doc)

