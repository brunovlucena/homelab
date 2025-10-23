# LocalStack Quick Start Guide

## 🚀 Deploy LocalStack

LocalStack is configured to deploy automatically via Flux GitOps in phase5-mocks.

### Manual Deployment (if needed)

```bash
# From the homelab directory
kubectl apply -k flux/clusters/homelab/infrastructure/localstack/
```

## ✅ Verify Deployment

```bash
# Check if LocalStack pod is running
kubectl get pods -n localstack

# Expected output:
# NAME                          READY   STATUS    RESTARTS   AGE
# localstack-xxxxxxxx-xxxxx     1/1     Running   0          2m

# Check logs
kubectl logs -n localstack -l app.kubernetes.io/name=localstack -f

# Check service
kubectl get svc -n localstack
```

## 🔌 Access LocalStack

### Option 1: Port Forward (Recommended for testing)

```bash
# Forward LocalStack port to localhost
kubectl port-forward -n localstack svc/localstack 4566:4566
```

Keep this terminal open and use `http://localhost:4566` in another terminal.

### Option 2: From within Kubernetes

Use the service DNS name from any pod in the cluster:
```
http://localstack.localstack.svc.cluster.local:4566
```

## 🧪 Test DynamoDB

### Using AWS CLI

```bash
# Set environment variables
export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1

# Create a table
aws dynamodb create-table \
  --table-name test-table \
  --attribute-definitions \
    AttributeName=id,AttributeType=S \
  --key-schema \
    AttributeName=id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --endpoint-url $AWS_ENDPOINT_URL

# Put an item
aws dynamodb put-item \
  --table-name test-table \
  --item '{
    "id": {"S": "user-001"},
    "name": {"S": "Bruno"},
    "email": {"S": "bruno@example.com"}
  }' \
  --endpoint-url $AWS_ENDPOINT_URL

# Get the item
aws dynamodb get-item \
  --table-name test-table \
  --key '{"id": {"S": "user-001"}}' \
  --endpoint-url $AWS_ENDPOINT_URL

# Scan table
aws dynamodb scan \
  --table-name test-table \
  --endpoint-url $AWS_ENDPOINT_URL

# List tables
aws dynamodb list-tables --endpoint-url $AWS_ENDPOINT_URL
```

### Using Python (boto3)

```python
import boto3

# Create client
dynamodb = boto3.client(
    'dynamodb',
    endpoint_url='http://localhost:4566',
    aws_access_key_id='test',
    aws_secret_access_key='test',
    region_name='us-east-1'
)

# Create table
response = dynamodb.create_table(
    TableName='users',
    KeySchema=[
        {'AttributeName': 'userId', 'KeyType': 'HASH'},
    ],
    AttributeDefinitions=[
        {'AttributeName': 'userId', 'AttributeType': 'S'},
    ],
    BillingMode='PAY_PER_REQUEST'
)

print(f"Table created: {response['TableDescription']['TableArn']}")

# Put item
dynamodb.put_item(
    TableName='users',
    Item={
        'userId': {'S': '123'},
        'name': {'S': 'Bruno Lucena'},
        'role': {'S': 'DevOps Engineer'}
    }
)

# Get item
response = dynamodb.get_item(
    TableName='users',
    Key={'userId': {'S': '123'}}
)

print(f"Item: {response['Item']}")
```

### Using Node.js (AWS SDK v3)

```javascript
const { DynamoDBClient, CreateTableCommand, PutItemCommand, GetItemCommand } = require("@aws-sdk/client-dynamodb");

const client = new DynamoDBClient({
  region: "us-east-1",
  endpoint: "http://localhost:4566",
  credentials: {
    accessKeyId: "test",
    secretAccessKey: "test"
  }
});

// Create table
async function createTable() {
  const command = new CreateTableCommand({
    TableName: "products",
    KeySchema: [
      { AttributeName: "productId", KeyType: "HASH" }
    ],
    AttributeDefinitions: [
      { AttributeName: "productId", AttributeType: "S" }
    ],
    BillingMode: "PAY_PER_REQUEST"
  });
  
  const response = await client.send(command);
  console.log("Table created:", response.TableDescription.TableArn);
}

// Put item
async function putItem() {
  const command = new PutItemCommand({
    TableName: "products",
    Item: {
      productId: { S: "prod-001" },
      name: { S: "Laptop" },
      price: { N: "999.99" }
    }
  });
  
  await client.send(command);
  console.log("Item added");
}

createTable().then(() => putItem());
```

## 🧪 Test S3

```bash
# Create bucket
aws s3 mb s3://my-test-bucket --endpoint-url $AWS_ENDPOINT_URL

# Upload file
echo "Hello LocalStack" > test.txt
aws s3 cp test.txt s3://my-test-bucket/ --endpoint-url $AWS_ENDPOINT_URL

# List buckets
aws s3 ls --endpoint-url $AWS_ENDPOINT_URL

# List objects in bucket
aws s3 ls s3://my-test-bucket/ --endpoint-url $AWS_ENDPOINT_URL

# Download file
aws s3 cp s3://my-test-bucket/test.txt downloaded.txt --endpoint-url $AWS_ENDPOINT_URL
```

## 🧪 Test SQS

```bash
# Create queue
aws sqs create-queue \
  --queue-name my-queue \
  --endpoint-url $AWS_ENDPOINT_URL

# Send message
aws sqs send-message \
  --queue-url http://localhost:4566/000000000000/my-queue \
  --message-body "Hello from LocalStack" \
  --endpoint-url $AWS_ENDPOINT_URL

# Receive message
aws sqs receive-message \
  --queue-url http://localhost:4566/000000000000/my-queue \
  --endpoint-url $AWS_ENDPOINT_URL

# List queues
aws sqs list-queues --endpoint-url $AWS_ENDPOINT_URL
```

## 🔍 Debug and Troubleshooting

### Check LocalStack Health

```bash
# Health check endpoint
curl http://localhost:4566/_localstack/health | jq

# Should return status of all services
```

### View Logs

```bash
# Follow logs
kubectl logs -n localstack -l app.kubernetes.io/name=localstack -f

# Last 100 lines
kubectl logs -n localstack -l app.kubernetes.io/name=localstack --tail=100
```

### Check Persistence

```bash
# Check PVC
kubectl get pvc -n localstack

# Check if data directory is mounted
kubectl exec -n localstack -it $(kubectl get pod -n localstack -l app.kubernetes.io/name=localstack -o jsonpath='{.items[0].metadata.name}') -- ls -la /tmp/localstack
```

### Common Issues

**Pod CrashLoopBackOff:**
```bash
kubectl describe pod -n localstack -l app.kubernetes.io/name=localstack
```

**Service not responding:**
- Check if service is enabled in SERVICES env var
- Verify port-forward is active
- Check firewall/network policies

**Data not persisting:**
- Verify PVC is created: `kubectl get pvc -n localstack`
- Check PERSISTENCE=1 environment variable
- Verify mount path is correct

## 🔄 Using LocalStack in Your Apps

### Kubernetes Deployment Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    spec:
      containers:
      - name: app
        image: my-app:latest
        env:
        - name: AWS_ENDPOINT_URL
          value: "http://localstack.localstack.svc.cluster.local:4566"
        - name: AWS_ACCESS_KEY_ID
          value: "test"
        - name: AWS_SECRET_ACCESS_KEY
          value: "test"
        - name: AWS_DEFAULT_REGION
          value: "us-east-1"
        # DynamoDB specific
        - name: DYNAMODB_ENDPOINT
          value: "http://localstack.localstack.svc.cluster.local:4566"
        # S3 specific
        - name: S3_ENDPOINT
          value: "http://localstack.localstack.svc.cluster.local:4566"
```

### Docker Compose Example (for local development)

```yaml
version: '3.8'
services:
  app:
    image: my-app:latest
    environment:
      - AWS_ENDPOINT_URL=http://localstack:4566
      - AWS_ACCESS_KEY_ID=test
      - AWS_SECRET_ACCESS_KEY=test
      - AWS_DEFAULT_REGION=us-east-1
    depends_on:
      - localstack
  
  localstack:
    image: localstack/localstack:latest
    ports:
      - "4566:4566"
    environment:
      - SERVICES=dynamodb,s3,sqs,sns
      - DEBUG=1
    volumes:
      - "./localstack-data:/tmp/localstack"
```

## 📚 Resources

- [LocalStack Documentation](https://docs.localstack.cloud/)
- [AWS CLI Reference](https://docs.aws.amazon.com/cli/latest/)
- [DynamoDB Examples](https://docs.localstack.cloud/user-guide/aws/dynamodb/)
- [Supported Services](https://docs.localstack.cloud/user-guide/aws/feature-coverage/)

## 🚫 Cleanup

```bash
# Delete LocalStack (will preserve PVC)
kubectl delete -k flux/clusters/homelab/infrastructure/localstack/

# Delete PVC (will delete all data)
kubectl delete pvc -n localstack --all

# Delete namespace
kubectl delete namespace localstack
```

