#!/bin/bash

set -e

echo "Inserting initial test data..."
python3 tests/insert_and_query.py

echo "Simulating replica failure (mongodb-1)..."
kubectl delete pod mongodb-1

echo "Verifying writes still work with one replica down..."
python3 tests/insert_and_query.py

echo "Verifying replication status across remaining nodes..."
python3 tests/validation_replication.py

echo "Waiting for MongoDB secondary to be rescheduled and Ready..."
kubectl wait --for=condition=Ready pod/mongodb-1 --timeout=120s

echo "Waiting for replication (10s)..."
sleep 10

echo "Verifying replication after pod recovery..."
python3 tests/validation_replication.py

echo "Checking logs from reanimated secondary (mongodb-1)..."
kubectl logs mongodb-1