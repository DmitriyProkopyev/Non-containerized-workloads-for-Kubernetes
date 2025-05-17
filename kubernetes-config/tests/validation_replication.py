from pymongo import MongoClient, ReadPreference
import time

def connect_to_member(host, port=27017, prefer_secondary=False):
    if prefer_secondary:
        client = MongoClient(
            host,
            port,
            read_preference=ReadPreference.SECONDARY_PREFERRED,
            directConnection=True,
            serverSelectionTimeoutMS=3000
        )
    else:
        client = MongoClient(
            host,
            port,
            directConnection=True,
            serverSelectionTimeoutMS=3000
        )
    return client

def fetch_all_docs(client):
    db = client["testdb"]
    return list(db["sample"].find({}, {"_id": 0}))  # Ignore _id for comparison

primary_host = "mongodb-0.mongodb-headless"
secondary_host = "mongodb-1.mongodb-headless"  # можно протестировать и mongodb-2

print("Connecting to primary...")
primary = connect_to_member(primary_host)
primary_docs = fetch_all_docs(primary)
print(f"Primary has {len(primary_docs)} documents.")

print("Connecting to secondary...")
secondary = connect_to_member(secondary_host, prefer_secondary=True)
secondary_docs = fetch_all_docs(secondary)
print(f"Secondary has {len(secondary_docs)} documents.")

print("\nComparing documents...")
if primary_docs == secondary_docs:
    print("✅ Secondary is in sync with primary!")
else:
    print("❌ Replication inconsistency detected.")
    print("Primary docs:")
    for doc in primary_docs:
        print(doc)
    print("Secondary docs:")
    for doc in secondary_docs:
        print(doc)
