from opentelemetry import trace
from opentelemetry.trace import SpanKind
from motor.motor_asyncio import AsyncIOMotorClient
import asyncio
import time
import os

tracer = trace.get_tracer(__name__)

class ReplicationMonitor:
    def __init__(self):
        # Get MongoDB addresses from environment variables or use defaults
        self.mongo_addresses = [
            os.getenv("MONGO_0_ADDRESS", "mongo-0.mongo-headless:27017"),
            os.getenv("MONGO_1_ADDRESS", "mongo-1.mongo-headless:27017"),
            os.getenv("MONGO_2_ADDRESS", "mongo-2.mongo-headless:27017")
        ]
        self.clients = {
            addr: AsyncIOMotorClient(f"mongodb://{addr}")
            for addr in self.mongo_addresses
        }
        self.running = False

    async def _get_primary_and_secondaries(self):
        """Determine which node is primary and which are secondaries"""
        primary = None
        secondaries = []
        
        for addr, client in self.clients.items():
            try:
                status = await client.admin.command("replSetGetStatus")
                for member in status['members']:
                    if member['name'] == addr:
                        if member['stateStr'] == 'PRIMARY':
                            primary = addr
                        elif member['stateStr'] == 'SECONDARY':
                            secondaries.append(addr)
            except Exception as e:
                print(f"Error getting status from {addr}: {e}")
        
        return primary, secondaries

    async def start_monitoring(self):
        self.running = True
        while self.running:
            with tracer.start_as_current_span("replication_lag_check", kind=SpanKind.INTERNAL) as span:
                try:
                    # Get current primary and secondaries
                    primary_addr, secondary_addrs = await self._get_primary_and_secondaries()
                    
                    if not primary_addr or not secondary_addrs:
                        span.set_attribute("error", "Could not determine primary or secondary nodes")
                        print("Could not determine primary or secondary nodes")
                        await asyncio.sleep(5)
                        continue

                    # Get primary's oplog timestamp
                    primary_status = await self.clients[primary_addr].admin.command("replSetGetStatus")
                    primary_optime = primary_status['members'][0]['optime']['ts'].time

                    # Track lag for each secondary
                    for secondary_addr in secondary_addrs:
                        secondary_status = await self.clients[secondary_addr].admin.command("replSetGetStatus")
                        secondary_optime = secondary_status['members'][0]['optime']['ts'].time
                        
                        # Calculate lag in seconds
                        lag_seconds = primary_optime - secondary_optime

                        # Add lag information to the span
                        span.set_attribute(f"replication.lag.{secondary_addr}.seconds", lag_seconds)
                        span.set_attribute(f"primary.optime", primary_optime)
                        span.set_attribute(f"secondary.{secondary_addr}.optime", secondary_optime)

                        # Log the lag
                        print(f"Replication lag for {secondary_addr}: {lag_seconds} seconds")

                except Exception as e:
                    span.record_exception(e)
                    print(f"Error monitoring replication: {e}")

                await asyncio.sleep(5)  # Check every 5 seconds

    def stop_monitoring(self):
        self.running = False
        # Close all MongoDB connections
        for client in self.clients.values():
            client.close() 