# Create a CloudEvent named network.notifi.lambda.build
# it shall have thirdPartyId and parserId as payload
import json
import uuid
import datetime
import requests
import threading
import time

# Create the CloudEvent
def create_cloud_event(third_party_id, parser_id, ctx, blkId):
    event = {
        # Required CloudEvent attributes
        "specversion": "1.0",
        "id": str(uuid.uuid4()),
        "source": f"network.notifi.{third_party_id}",
        "subject": f"{parser_id}",
        "type": "network.notifi.lambda.parser.start",
        "time": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        
        # Custom data payload
        "data": {
            "contextId": ctx,
            "parameters": {"blockId": blkId,"blockchainType":52,"urlForBlob":"redis://ephemeralblock/Botanix/659780","logIndices":[0]},
        },
        
        # Optional attributes
        "datacontenttype": "application/json"
    }
    
    return event

# Publish the CloudEvent to broker HTTP endpoint
def publish_to_broker(cloud_event, broker_url="http://localhost:8081", env=None, third_party_id=None, parser_id=None):
    try:
        # Prepare headers for CloudEvent in structured mode
        headers = {
            "Content-Type": "application/cloudevents+json"
        }
        
        # Send the complete CloudEvent as JSON (structured mode)
        response = requests.post(
            broker_url,
            headers=headers,
            json=cloud_event,  # Send the complete CloudEvent
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

# Usage example
if __name__ == "__main__":
    import os
    
    # Get environment from ENV variable, default to dev
    env = os.getenv("ENV", "dev")
    print(f"Using environment: {env}")
    
    broker_url = "http://127.0.0.1:8081"  # Port-forwarded service broker endpoint
    
    print(f"Broker URL: {broker_url}")
    
    third_party_id = "0307ea43639b4616b044d190310a26bd"
    
    # Define all parser IDs
    parser_ids = [
        "0197ad6c10b973b2b854a0e652155b7e",
        "c42d2e6ca3214f4b8d28a2cab47beecf",
        "e0a711bde5d748009a995432acbf590b"
    ]

    ctx1 = "1FV29-cJHSIOOuk-gJPFHrnAhhEBFA36Rs90PlyvrlcEvke0ZHvoDP90c117FEgYPH1699JMgv2soU0Vz49qH-ZK_EFvwMD_wo_u0KQo5PYNV1F9UG1Wjb0nynbAyS1bIH-w7b--TGYLYZHo1rLAsl4EeRAl2OZYRWDhwJy7"
    ctx2 = "fLk7FvaIsE-bhBvyf7M3u2eOC-tCiUonUwTlpjK6MtGfQm4chBNTdY7dS6z3hVr8C0NBX5bCsPxgosJUKJUnaEtI_fFh1tDxTnUMJNkBwda_yqqgV0fh6XzAeq5FvZoP7ThPGRZhpKVeF0rotd6wfkhx-ZBl7IM_ZgQeY5wE"
    ctx3 = "j1LbXVXBHnvT574YvpHcB6pJfQWJAr0YBdU7NVhFgDLOVyKC3LfvYDPwTJJRFb9VpddeF4p9HbYIrKJ995YQaUkr23ymIpVRsRw1yVjidPYbgRrQvP9o7a8iScI5SxJd0Ey7vZRtbhI1Ep9soCpIqzwWiKJFqiOYY5tSBVvP"
    
    successful_events = 0
    failed_events = 0
    
    # Function to send events for a specific parser
    def send_events_for_parser(parser_id, start_batch, num_batches, results):
        local_success = 0
        local_failed = 0
        
        for i in range(start_batch, start_batch + num_batches):
            base_block_id = 659780 + (i * 18)
            
            # Send 6 events per batch (2 per context)
            contexts = [ctx1, ctx2, ctx3]
            for ctx_idx, ctx in enumerate(contexts):
                for ctx_alt in [ctx1, ctx2]:  # Send 2 events per context
                    event = create_cloud_event(third_party_id, parser_id, ctx_alt, str(base_block_id + ctx_idx))
                    if publish_to_broker(event, broker_url=broker_url, env=env, third_party_id=third_party_id, parser_id=parser_id):
                        local_success += 1
                    else:
                        local_failed += 1
                    
                    # Small delay to prevent overwhelming
                    time.sleep(0.1)
        
        # Store results
        results[parser_id] = {'success': local_success, 'failed': local_failed}
    
    # Send high load to ALL services to trigger scaling
    print("🔥 Starting high load test with concurrent processing...")
    
    # Create threads for each parser to send events concurrently
    threads = []
    batches_per_parser = 20  # Each parser gets 20 batches
    results = {}
    
    for i, parser_id in enumerate(parser_ids):
        start_batch = i * batches_per_parser
        thread = threading.Thread(
            target=send_events_for_parser, 
            args=(parser_id, start_batch, batches_per_parser, results)
        )
        threads.append(thread)
        thread.start()
    
    # Wait for all threads to complete
    for thread in threads:
        thread.join()
    
    # Sum up results
    for parser_id, result in results.items():
        successful_events += result['success']
        failed_events += result['failed']

    print(f"📊 RESULTS:")
    print(f"   ✅ Successful events: {successful_events}")
    print(f"   ❌ Failed events: {failed_events}")
    print(f"   📈 Total events: {successful_events + failed_events} (3 parsers × 6 contexts)")