#!/bin/python
# Create a CloudEvent named network.notifi.lambda.job.start
# This event will be sent to the kaniko-jobs RabbitMQ queue
# and then triggered back to the knative-lambda-builder
import json
import uuid
import datetime
import requests
import os

# Create the CloudEvent for job start
def create_job_start_event(third_party_id, parser_id, priority=5):
    event = {
        # Required CloudEvent attributes
        "specversion": "1.0",
        "id": str(uuid.uuid4()),
        "source": f"network.notifi.kaniko-jobs",
        "subject": f"{parser_id}",
        "type": "network.notifi.lambda.job.start",
        "time": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        
        # Custom data payload for job start event
        "data": {
            "third_party_id": third_party_id,
            "parser_id": f"{parser_id}",
            "correlation_id": str(uuid.uuid4()),
            "job_name": f"kaniko-job-{third_party_id}-{parser_id}",
            "parameters": {
                "priority": priority,
                "build_type": "container",
                "runtime": "nodejs22",
                "source_url": f"https://github.com/notifi/parsers/{parser_id}",
                "build_timeout": 1800,
                "environment": {
                    "NODE_ENV": "production",
                    "BUILD_MODE": "optimized"
                }
            },
            "priority": priority
        },
        
        # Optional attributes
        "datacontenttype": "application/json"
    }
    
    return event

# Publish the CloudEvent to broker HTTP endpoint
def publish_to_broker(cloud_event, broker_url="http://localhost:8081", env=None):
    try:
        # Prepare headers for CloudEvent (CloudEvents HTTP protocol)
        headers = {
            "Content-Type": "application/cloudevents+json",
        }
        
        # Send the complete CloudEvent as the request body
        response = requests.post(
            broker_url,
            headers=headers,
            json=cloud_event,
            timeout=30
        )
        
        if response.status_code in [200, 202]:
            print(f"✅ Published Job Start CloudEvent: {cloud_event['id']} - HTTP {response.status_code}")
            print(f"   Event Type: {cloud_event['type']}")
            print(f"   Source: {cloud_event['source']}")
            print(f"   Third Party ID: {cloud_event['data']['third_party_id']}")
            print(f"   Parser ID: {cloud_event['data']['parser_id']}")
            print(f"   Priority: {cloud_event['data']['priority']}")
            return True
        else:
            print(f"❌ Failed to publish Job Start CloudEvent: {cloud_event['id']} - HTTP {response.status_code}")
            print(f"   Response: {response.text}")
            return False
            
    except Exception as e:
        print(f"❌ Failed to publish Job Start CloudEvent: {cloud_event['id']} - {e}")
        return False

if __name__ == "__main__":
    # Get environment from ENV variable, default to dev
    env = os.getenv("ENV", "dev")
    print(f"Using environment: {env}")
    
    # Configure broker URL based on environment
    broker_url = "http://0.0.0.0:8081"  # Use lambda broker for local too
    
    print(f"Broker URL: {broker_url}")
    print(f"Event Type: network.notifi.lambda.job.start")
    print(f"Queue: kaniko-jobs")
    print("=" * 60)
    
    # Create multiple job start events with different priorities for testing
    events_data = [
        {
            "third_party_id": "0307ea43639b4616b044d190310a26bd",
            "parser_id": "0197ad6c10b973b2b854a0e652155b7e",
            "priority": 1  # High priority
        },
        {
            "third_party_id": "0307ea43639b4616b044d190310a26bd",
            "parser_id": "c42d2e6ca3214f4b8d28a2cab47beecf",
            "priority": 5  # Medium priority
        },
        {
            "third_party_id": "0307ea43639b4616b044d190310a26bd",
            "parser_id": "e0a711bde5d748009a995432acbf590b",
            "priority": 10  # Low priority
        },
        {
            "third_party_id": "0307ea43639b4616b044d190310a26bd",
            "parser_id": "f1b822cef6e85911ab0aa543bd0c6a1c",
            "priority": 3  # High-medium priority
        },
    ]
    
    print(f"Creating {len(events_data)} job start events...")
    
    successful_events = 0
    failed_events = 0
    
    for i, event_data in enumerate(events_data, 1):
        print(f"\nCreating job start event {i}/{len(events_data)}...")
        event = create_job_start_event(
            event_data["third_party_id"], 
            event_data["parser_id"],
            event_data["priority"]
        )
        if publish_to_broker(event, broker_url=broker_url, env=env):
            successful_events += 1
        else:
            failed_events += 1
    
    print("\n" + "=" * 60)
    print(f"Job Start Events Summary:")
    print(f"✅ Successful: {successful_events}")
    print(f"❌ Failed: {failed_events}")
    print(f"📊 Total: {len(events_data)}")
    
    if successful_events > 0:
        print(f"\n🎯 Job start events sent to kaniko-jobs queue!")
        print(f"📋 These events will be processed by the knative-lambda-builder")
        print(f"⚡ The builder will create Kaniko jobs asynchronously")
        print(f"🔍 Check the builder logs to see job creation progress")
