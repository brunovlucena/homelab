"""
🚀 Serverless LanceDB API with MinIO Storage

A serverless REST API for vector search and storage using LanceDB and MinIO.
Designed to run on Knative with auto-scaling capabilities.
"""

import os
import logging
import time
from typing import List, Dict, Any, Optional
from datetime import datetime

from flask import Flask, request, jsonify
import lancedb
from prometheus_client import Counter, Histogram, Gauge, generate_latest, CONTENT_TYPE_LATEST

# ============================================================================
# 🔧 Configuration
# ============================================================================

LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")
AWS_ENDPOINT = os.getenv("AWS_ENDPOINT", "http://minio-service.minio.svc.cluster.local:9000")
AWS_REGION = os.getenv("AWS_DEFAULT_REGION", "us-east-1")
LANCEDB_BUCKET = os.getenv("LANCEDB_BUCKET", "lancedb")

# Configure logging with timestamps and levels
logging.basicConfig(
    level=getattr(logging, LOG_LEVEL),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# ============================================================================
# 📊 Prometheus Metrics
# ============================================================================

REQUEST_COUNT = Counter(
    'lancedb_requests_total',
    'Total number of requests',
    ['method', 'endpoint', 'status']
)

REQUEST_DURATION = Histogram(
    'lancedb_request_duration_seconds',
    'Request duration in seconds',
    ['method', 'endpoint']
)

ERROR_COUNT = Counter(
    'lancedb_errors_total',
    'Total number of errors',
    ['error_type']
)

TABLES_COUNT = Gauge(
    'lancedb_tables_total',
    'Number of tables in LanceDB'
)

SEARCH_OPS = Counter(
    'lancedb_search_operations_total',
    'Total number of search operations',
    ['table']
)

# ============================================================================
# 🗄️ Database Connection
# ============================================================================

app = Flask(__name__)

# Global database connection
db_connection = None


def get_db():
    """Get or create LanceDB connection to MinIO"""
    global db_connection
    
    if db_connection is None:
        try:
            logger.info(f"🔌 Connecting to LanceDB with MinIO endpoint: {AWS_ENDPOINT}")
            
            # Connect to LanceDB with S3-compatible storage (MinIO)
            db_connection = lancedb.connect(
                f"s3://{LANCEDB_BUCKET}/",
                storage_options={
                    "region": AWS_REGION,
                    "endpoint": AWS_ENDPOINT,
                    "aws_access_key_id": os.getenv("AWS_ACCESS_KEY_ID"),
                    "aws_secret_access_key": os.getenv("AWS_SECRET_ACCESS_KEY"),
                }
            )
            
            logger.info("✅ Successfully connected to LanceDB")
            
        except Exception as e:
            logger.error(f"❌ Failed to connect to LanceDB: {str(e)}", exc_info=True)
            ERROR_COUNT.labels(error_type='connection_error').inc()
            raise
    
    return db_connection


# ============================================================================
# 🔌 API Endpoints
# ============================================================================

@app.route('/health', methods=['GET'])
def health_check():
    """Health check endpoint"""
    try:
        # Test database connection
        get_db()
        return jsonify({
            'status': 'healthy',
            'timestamp': datetime.utcnow().isoformat(),
            'service': 'lancedb-serverless',
            'minio_endpoint': AWS_ENDPOINT
        }), 200
    except Exception as e:
        logger.error(f"❌ Health check failed: {str(e)}")
        return jsonify({
            'status': 'unhealthy',
            'error': str(e),
            'timestamp': datetime.utcnow().isoformat()
        }), 503


@app.route('/metrics', methods=['GET'])
def metrics():
    """Prometheus metrics endpoint"""
    return generate_latest(), 200, {'Content-Type': CONTENT_TYPE_LATEST}


@app.route('/tables', methods=['GET'])
def list_tables():
    """List all tables in LanceDB"""
    start_time = time.time()
    
    try:
        db = get_db()
        tables = db.table_names()
        
        TABLES_COUNT.set(len(tables))
        REQUEST_COUNT.labels(method='GET', endpoint='/tables', status='success').inc()
        REQUEST_DURATION.labels(method='GET', endpoint='/tables').observe(time.time() - start_time)
        
        logger.info(f"📋 Listed {len(tables)} tables")
        
        return jsonify({
            'tables': tables,
            'count': len(tables),
            'timestamp': datetime.utcnow().isoformat()
        }), 200
        
    except Exception as e:
        logger.error(f"❌ Error listing tables: {str(e)}", exc_info=True)
        ERROR_COUNT.labels(error_type='list_tables_error').inc()
        REQUEST_COUNT.labels(method='GET', endpoint='/tables', status='error').inc()
        
        return jsonify({
            'error': str(e),
            'timestamp': datetime.utcnow().isoformat()
        }), 500


@app.route('/tables', methods=['POST'])
def create_table():
    """Create a new table with data"""
    start_time = time.time()
    
    try:
        data = request.get_json()
        
        if not data or 'name' not in data or 'data' not in data:
            return jsonify({
                'error': 'Missing required fields: name and data'
            }), 400
        
        table_name = data['name']
        table_data = data['data']
        mode = data.get('mode', 'overwrite')  # overwrite or append
        
        db = get_db()
        
        # Create or append to table
        if mode == 'append' and table_name in db.table_names():
            table = db.open_table(table_name)
            table.add(table_data)
            logger.info(f"➕ Appended {len(table_data)} rows to table '{table_name}'")
        else:
            table = db.create_table(table_name, data=table_data, mode=mode)
            logger.info(f"✨ Created table '{table_name}' with {len(table_data)} rows")
        
        TABLES_COUNT.set(len(db.table_names()))
        REQUEST_COUNT.labels(method='POST', endpoint='/tables', status='success').inc()
        REQUEST_DURATION.labels(method='POST', endpoint='/tables').observe(time.time() - start_time)
        
        return jsonify({
            'message': f'Table {table_name} created successfully',
            'table': table_name,
            'rows': len(table_data),
            'mode': mode,
            'timestamp': datetime.utcnow().isoformat()
        }), 201
        
    except Exception as e:
        logger.error(f"❌ Error creating table: {str(e)}", exc_info=True)
        ERROR_COUNT.labels(error_type='create_table_error').inc()
        REQUEST_COUNT.labels(method='POST', endpoint='/tables', status='error').inc()
        
        return jsonify({
            'error': str(e),
            'timestamp': datetime.utcnow().isoformat()
        }), 500


@app.route('/search', methods=['POST'])
def search():
    """Perform vector search on a table"""
    start_time = time.time()
    
    try:
        data = request.get_json()
        
        if not data or 'table' not in data or 'query_vector' not in data:
            return jsonify({
                'error': 'Missing required fields: table and query_vector'
            }), 400
        
        table_name = data['table']
        query_vector = data['query_vector']
        limit = data.get('limit', 10)
        filter_expr = data.get('filter')
        
        db = get_db()
        
        if table_name not in db.table_names():
            return jsonify({
                'error': f'Table {table_name} not found'
            }), 404
        
        table = db.open_table(table_name)
        
        # Build search query
        search_query = table.search(query_vector).limit(limit)
        
        if filter_expr:
            search_query = search_query.where(filter_expr)
        
        # Execute search
        results = search_query.to_list()
        
        SEARCH_OPS.labels(table=table_name).inc()
        REQUEST_COUNT.labels(method='POST', endpoint='/search', status='success').inc()
        REQUEST_DURATION.labels(method='POST', endpoint='/search').observe(time.time() - start_time)
        
        logger.info(f"🔍 Searched table '{table_name}', found {len(results)} results")
        
        return jsonify({
            'results': results,
            'count': len(results),
            'table': table_name,
            'timestamp': datetime.utcnow().isoformat()
        }), 200
        
    except Exception as e:
        logger.error(f"❌ Error performing search: {str(e)}", exc_info=True)
        ERROR_COUNT.labels(error_type='search_error').inc()
        REQUEST_COUNT.labels(method='POST', endpoint='/search', status='error').inc()
        
        return jsonify({
            'error': str(e),
            'timestamp': datetime.utcnow().isoformat()
        }), 500


@app.route('/tables/<table_name>', methods=['DELETE'])
def delete_table(table_name: str):
    """Delete a table"""
    start_time = time.time()
    
    try:
        db = get_db()
        
        if table_name not in db.table_names():
            return jsonify({
                'error': f'Table {table_name} not found'
            }), 404
        
        db.drop_table(table_name)
        
        TABLES_COUNT.set(len(db.table_names()))
        REQUEST_COUNT.labels(method='DELETE', endpoint='/tables/<name>', status='success').inc()
        REQUEST_DURATION.labels(method='DELETE', endpoint='/tables/<name>').observe(time.time() - start_time)
        
        logger.info(f"🗑️ Deleted table '{table_name}'")
        
        return jsonify({
            'message': f'Table {table_name} deleted successfully',
            'timestamp': datetime.utcnow().isoformat()
        }), 200
        
    except Exception as e:
        logger.error(f"❌ Error deleting table: {str(e)}", exc_info=True)
        ERROR_COUNT.labels(error_type='delete_table_error').inc()
        REQUEST_COUNT.labels(method='DELETE', endpoint='/tables/<name>', status='error').inc()
        
        return jsonify({
            'error': str(e),
            'timestamp': datetime.utcnow().isoformat()
        }), 500


@app.route('/', methods=['GET'])
def index():
    """API documentation endpoint"""
    return jsonify({
        'service': 'lancedb-serverless',
        'version': '1.0.0',
        'description': 'Serverless LanceDB API with MinIO storage',
        'endpoints': {
            'GET /': 'This documentation',
            'GET /health': 'Health check',
            'GET /metrics': 'Prometheus metrics',
            'GET /tables': 'List all tables',
            'POST /tables': 'Create a new table',
            'POST /search': 'Perform vector search',
            'DELETE /tables/<name>': 'Delete a table'
        },
        'timestamp': datetime.utcnow().isoformat()
    }), 200


# ============================================================================
# 🚀 Application Entry Point
# ============================================================================

if __name__ == '__main__':
    port = int(os.getenv('PORT', '8080'))
    
    logger.info(f"🚀 Starting LanceDB Serverless API on port {port}")
    logger.info(f"🗄️ MinIO Endpoint: {AWS_ENDPOINT}")
    logger.info(f"📦 Bucket: {LANCEDB_BUCKET}")
    logger.info(f"🌍 Region: {AWS_REGION}")
    
    # Run Flask app
    app.run(
        host='0.0.0.0',
        port=port,
        debug=(LOG_LEVEL == 'DEBUG')
    )

