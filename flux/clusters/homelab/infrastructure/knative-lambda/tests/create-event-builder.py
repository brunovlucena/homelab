#!/bin/python
# Create a CloudEvent named network.notifi.lambda.build
# it shall have thirdPartyId and parserId as payload
import json
import uuid
import datetime
import requests
import os

# Create the CloudEvent
def create_cloud_event(third_party_id, parser_id):
    event = {
        # Required CloudEvent attributes
        "specversion": "1.0",
        "id": str(uuid.uuid4()),
        "source": f"network.notifi.{third_party_id}",
        "subject": f"{parser_id}",
        "type": "network.notifi.lambda.build.start",
        "time": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        
        # Custom data payload -- NOTE: We shouldn't need to use this, as the information is available
        # from source and subject, above
        "data": {
            "third_party_id": third_party_id,
            "parser_id": f"{parser_id}",
        },
        
        # Optional attributes
        "datacontenttype": "application/json"
    }
    
    return event

# Publish the CloudEvent to broker HTTP endpoint
def publish_to_broker(cloud_event, broker_url="http://localhost:8081", env=None):
    try:
        # Prepare headers for CloudEvent (CloudEvents HTTP protocol)
        # Do not specify ce-* headers, else CloudEvent libraries assume binary mode instead of structured
        headers = {
            "Content-Type": "application/cloudevents+json",
        }
        
        # Send the complete CloudEvent as the request body
        response = requests.post(
            broker_url,
            headers=headers,
            json=cloud_event,  # Removed ["data"] portion
            timeout=30  # Increase timeout for lambda cold start
        )
        
        if response.status_code in [200, 202]:
            print(f"✅ Published CloudEvent: {cloud_event['id']} - HTTP {response.status_code}")
            return True
        else:
            print(f"❌ Failed to publish CloudEvent: {cloud_event['id']} - HTTP {response.status_code}")
            return False
            
    except Exception as e:
        print(f"❌ Failed to publish CloudEvent: {cloud_event['id']} - {e}")
        return False

if __name__ == "__main__":
    # Get environment from ENV variable, default to dev
    env = os.getenv("ENV", "dev")
    print(f"Using environment: {env}")
    
    # Configure broker URL based on environment
    broker_url = "http://0.0.0.0:8081"  # Use lambda broker for local too
    
    print(f"Broker URL: {broker_url}")
    
    # Create multiple build events with unique IDs for testing
    events_data = [
        {
            "third_party_id": "0307ea43639b4616b044d190310a26bd",
            "parser_id": "0197ad6c10b973b2b854a0e652155b7e"
        },
        {
            "third_party_id": "0307ea43639b4616b044d190310a26bd",
            "parser_id": "c42d2e6ca3214f4b8d28a2cab47beecf"
        },
        {
            "third_party_id": "0307ea43639b4616b044d190310a26bd",
            "parser_id": "e0a711bde5d748009a995432acbf590b"
        },
    ]
    
    print(f"Creating {len(events_data)} build events...")
    
    successful_events = 0
    failed_events = 0
    
    for i, event_data in enumerate(events_data, 1):
        print(f"\nCreating event {i}/{len(events_data)}...")
        event = create_cloud_event(event_data["third_party_id"], event_data["parser_id"])
        if publish_to_broker(event, broker_url=broker_url, env=env):
            successful_events += 1
        else:
            failed_events += 1
    
    print(f"\n📊 RESULTS:")
    print(f"   ✅ Successful events: {successful_events}")
    print(f"   ❌ Failed events: {failed_events}")
    print(f"   📈 Total events: {successful_events + failed_events}")
    
    if successful_events > 0:
        print(f"\n✅ Successfully created and published {successful_events} build events!")
    else:
        print(f"\n❌ Failed to publish any build events!")
        print(f"\n💡 Troubleshooting:")
        print(f"   1. Make sure the broker is running: kubectl get pods -n knative-lambda-{env}")
        print(f"   2. Check if port forwarding is active: make pf-broker ENV={env}")
        print(f"   3. Verify the broker URL: {broker_url}")
        exit(1)