# LocalStack

LocalStack provides a fully functional local AWS cloud stack. This deployment includes emulation for:

- **DynamoDB** - NoSQL database
- **S3** - Object storage
- **SNS/SQS** - Messaging services
- **Lambda** - Serverless functions
- **Kinesis** - Data streaming
- **CloudWatch** - Monitoring and logs
- **EventBridge** - Event bus

## Configuration

### Accessing LocalStack

LocalStack is exposed as a ClusterIP service within the cluster:

```bash
# From within the cluster
http://localstack.localstack.svc.cluster.local:4566
```

### Port Forwarding for Local Access

```bash
kubectl port-forward -n localstack svc/localstack 4566:4566
```

Then access at: `http://localhost:4566`

### Using AWS CLI with LocalStack

```bash
# Set the endpoint URL
export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1

# Example: Create DynamoDB table
aws dynamodb create-table \
  --table-name users \
  --attribute-definitions AttributeName=id,AttributeType=S \
  --key-schema AttributeName=id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --endpoint-url $AWS_ENDPOINT_URL

# Example: Put item
aws dynamodb put-item \
  --table-name users \
  --item '{"id": {"S": "1"}, "name": {"S": "Bruno"}}' \
  --endpoint-url $AWS_ENDPOINT_URL

# Example: Scan table
aws dynamodb scan \
  --table-name users \
  --endpoint-url $AWS_ENDPOINT_URL
```

### Using AWS SDK (Python boto3)

```python
import boto3

# Create DynamoDB client
dynamodb = boto3.client(
    'dynamodb',
    endpoint_url='http://localhost:4566',
    aws_access_key_id='test',
    aws_secret_access_key='test',
    region_name='us-east-1'
)

# Create table
dynamodb.create_table(
    TableName='users',
    KeySchema=[
        {'AttributeName': 'id', 'KeyType': 'HASH'}
    ],
    AttributeDefinitions=[
        {'AttributeName': 'id', 'AttributeType': 'S'}
    ],
    BillingMode='PAY_PER_REQUEST'
)
```

### Using from within Kubernetes

Configure your application pods with these environment variables:

```yaml
env:
  - name: AWS_ENDPOINT_URL
    value: "http://localstack.localstack.svc.cluster.local:4566"
  - name: AWS_ACCESS_KEY_ID
    value: "test"
  - name: AWS_SECRET_ACCESS_KEY
    value: "test"
  - name: AWS_DEFAULT_REGION
    value: "us-east-1"
```

## Features

### Persistence

Data is persisted to a PersistentVolumeClaim (10Gi), so tables and data survive pod restarts.

### Available Services

The deployment enables these AWS services:
- dynamodb
- s3
- sns
- sqs
- lambda
- kinesis
- cloudwatch
- logs
- events

### DynamoDB Specific Settings

- `DYNAMODB_SHARE_DB=1` - Share DynamoDB database across all services
- `DYNAMODB_IN_MEMORY=0` - Persist data to disk (not in-memory)

## Troubleshooting

### Check LocalStack status

```bash
# Get pods
kubectl get pods -n localstack

# Check logs
kubectl logs -n localstack -l app.kubernetes.io/name=localstack -f

# Check service endpoints
kubectl get svc -n localstack

# Describe the pod
kubectl describe pod -n localstack -l app.kubernetes.io/name=localstack
```

### Test connectivity

```bash
# From within the cluster
kubectl run -it --rm debug --image=amazon/aws-cli --restart=Never -- \
  dynamodb list-tables \
  --endpoint-url http://localstack.localstack.svc.cluster.local:4566 \
  --region us-east-1
```

### Common Issues

1. **Pod won't start**: Check resource limits and PVC availability
2. **Services not responding**: Verify service is listed in SERVICES env var
3. **Data not persisting**: Check PVC mount and PERSISTENCE=1

## Resources

- [LocalStack Docs](https://docs.localstack.cloud/)
- [AWS CLI with LocalStack](https://docs.localstack.cloud/user-guide/integrations/aws-cli/)
- [Supported Services](https://docs.localstack.cloud/user-guide/aws/feature-coverage/)

