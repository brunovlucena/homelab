#!/bin/python
# Create a CloudEvent named network.notifi.lambda.service.delete
# it shall have thirdPartyId and parserId as payload
import json
import uuid
import datetime
import requests
import os

# Create the CloudEvent for service deletion
def create_delete_cloud_event(third_party_id, parser_id, service_name=None, reason=None):
    event = {
        # Required CloudEvent attributes
        "specversion": "1.0",
        "id": str(uuid.uuid4()),
        "source": f"network.notifi.{third_party_id}",
        "subject": f"{parser_id}",
        "type": "network.notifi.lambda.service.delete",
        "time": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        
        # Custom data payload for service deletion
        "data": {
            "third_party_id": third_party_id,
            "parser_id": f"{parser_id}",
            "correlation_id": str(uuid.uuid4()),
        },
        
        # Optional attributes
        "datacontenttype": "application/json"
    }
    
    # Add optional fields if provided
    if service_name:
        event["data"]["service_name"] = service_name
    if reason:
        event["data"]["reason"] = reason
    
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
            timeout=30  # Increase timeout for lambda cold start
        )
        
        if response.status_code in [200, 202]:
            print(f"✅ Published Delete CloudEvent: {cloud_event['id']} - HTTP {response.status_code}")
            if response.text:
                print(f"   Response: {response.text}")
            return True
        else:
            print(f"❌ Failed to publish Delete CloudEvent: {cloud_event['id']} - HTTP {response.status_code}")
            if response.text:
                print(f"   Error Response: {response.text}")
            return False
            
    except Exception as e:
        print(f"❌ Failed to publish Delete CloudEvent: {cloud_event['id']} - {e}")
        return False

if __name__ == "__main__":
    # Get environment from ENV variable, default to dev
    env = os.getenv("ENV", "dev")
    print(f"Using environment: {env}")
    
    # Configure broker URL based on environment
    broker_url = "http://0.0.0.0:8081"  # Use lambda broker for local too
    
    print(f"Broker URL: {broker_url}")
    
    # Create delete events for existing services
    # Using the same third party ID and parser IDs from the running services
    delete_events_data = [
        {
            "third_party_id": "0307ea43639b4616b044d190310a26bd",
            "parser_id": "0197ad6c10b973b2b854a0e652155b7e",
            "service_name": None,  # Let it be auto-generated
            "reason": "Testing delete functionality"
        },
        {
            "third_party_id": "0307ea43639b4616b044d190310a26bd",
            "parser_id": "c42d2e6ca3214f4b8d28a2cab47beecf",
            "service_name": None,  # Let it be auto-generated
            "reason": "Testing delete functionality"
        },
        {
            "third_party_id": "0307ea43639b4616b044d190310a26bd",
            "parser_id": "e0a711bde5d748009a995432acbf590b",
            "service_name": None,  # Let it be auto-generated
            "reason": "Testing delete functionality"
        },
    ]
    
    print(f"Creating {len(delete_events_data)} delete events...")
    
    successful_events = 0
    failed_events = 0
    
    for i, event_data in enumerate(delete_events_data, 1):
        print(f"\nCreating delete event {i}/{len(delete_events_data)}...")
        print(f"   Third Party ID: {event_data['third_party_id']}")
        print(f"   Parser ID: {event_data['parser_id']}")
        print(f"   Service Name: {event_data['service_name'] or 'Auto-generated'}")
        print(f"   Reason: {event_data['reason']}")
        
        event = create_delete_cloud_event(
            event_data["third_party_id"], 
            event_data["parser_id"],
            event_data["service_name"],
            event_data["reason"]
        )
        
        if publish_to_broker(event, broker_url=broker_url, env=env):
            successful_events += 1
        else:
            failed_events += 1
    
    print(f"\n📊 RESULTS:")
    print(f"   ✅ Successful delete events: {successful_events}")
    print(f"   ❌ Failed delete events: {failed_events}")
    print(f"   📈 Total delete events: {successful_events + failed_events}")
    
    if successful_events > 0:
        print(f"\n✅ Successfully created and published {successful_events} delete events!")
        print(f"\n🔍 Next steps:")
        print(f"   1. Check if services were deleted: make check-services-dev")
        print(f"   2. Monitor logs: kubectl logs -f -n knative-lambda-{env} -l app=knative-lambda-builder")
        print(f"   3. Check for any remaining resources: kubectl get all -n knative-lambda-{env}")
    else:
        print(f"\n❌ Failed to publish any delete events!")
        print(f"\n💡 Troubleshooting:")
        print(f"   1. Make sure the broker is running: kubectl get pods -n knative-lambda-{env}")
        print(f"   2. Check if port forwarding is active: make pf-broker ENV={env}")
        print(f"   3. Verify the broker URL: {broker_url}")
        print(f"   4. Check if the service supports delete events")
        exit(1)
