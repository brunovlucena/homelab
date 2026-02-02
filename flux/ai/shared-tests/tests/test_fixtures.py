"""Tests for shared fixtures."""
import pytest


class TestCloudEventFactory:
    """Tests for CloudEvent factory fixture."""
    
    def test_create_basic_event(self, cloudevent_factory):
        """Test creating a basic CloudEvent."""
        event = cloudevent_factory.create(
            type="io.homelab.test.event",
            data={"key": "value"}
        )
        
        assert event.type == "io.homelab.test.event"
        assert event.data == {"key": "value"}
        assert event.specversion == "1.0"
        assert event.id is not None
    
    def test_create_chat_message(self, cloudevent_factory):
        """Test creating a chat message event."""
        event = cloudevent_factory.create_chat_message(
            message="Hello, world!",
            user_id="test-user"
        )
        
        assert event.type == "io.homelab.chat.message"
        assert event.data["message"] == "Hello, world!"
        assert event.data["user_id"] == "test-user"
    
    def test_create_exploit_event(self, cloudevent_factory):
        """Test creating an exploit event."""
        event = cloudevent_factory.create_exploit_event(
            exploit_id="vuln-001",
            status="success",
            severity="critical"
        )
        
        assert event.type == "io.homelab.exploit.success"
        assert event.data["exploit_id"] == "vuln-001"
        assert event.data["severity"] == "critical"
    
    def test_create_batch(self, cloudevent_factory):
        """Test creating a batch of events."""
        events = cloudevent_factory.create_batch(5, type="health.check")
        
        assert len(events) == 5
        for event in events:
            assert event.type == "io.homelab.health.check"
    
    def test_event_to_dict(self, cloudevent_factory):
        """Test converting event to dictionary."""
        event = cloudevent_factory.create(type="test.event", data={"x": 1})
        event_dict = event.to_dict()
        
        assert "specversion" in event_dict
        assert "type" in event_dict
        assert "source" in event_dict
        assert "id" in event_dict
        assert "data" in event_dict
    
    def test_event_to_headers(self, cloudevent_factory):
        """Test converting event to HTTP headers."""
        event = cloudevent_factory.create(type="test.event")
        headers = event.to_headers()
        
        assert "ce-specversion" in headers
        assert "ce-type" in headers
        assert "ce-source" in headers
        assert "ce-id" in headers


class TestMockKubernetesClient:
    """Tests for mock Kubernetes client fixture."""
    
    @pytest.mark.asyncio
    async def test_apply_manifest(self, mock_k8s_client):
        """Test applying a manifest."""
        success, message = await mock_k8s_client.apply_manifest({
            "apiVersion": "v1",
            "kind": "Pod",
            "metadata": {"name": "test-pod"}
        })
        
        assert success is True
        assert "created" in message
    
    @pytest.mark.asyncio
    async def test_delete_resource(self, mock_k8s_client):
        """Test deleting a resource."""
        success, message = await mock_k8s_client.delete_resource(
            "Pod", "test-pod", "default"
        )
        
        assert success is True
    
    @pytest.mark.asyncio
    async def test_get_logs(self, mock_k8s_client):
        """Test getting pod logs."""
        success, logs = await mock_k8s_client.get_logs("test-pod", "default")
        
        assert success is True
        assert isinstance(logs, str)


class TestK8sResourceFactory:
    """Tests for Kubernetes resource factory."""
    
    def test_create_pod(self, k8s_resource_factory):
        """Test creating a Pod resource."""
        pod = k8s_resource_factory.create_pod(
            name="test-pod",
            image="nginx:latest",
            phase="Running"
        )
        
        assert pod.kind == "Pod"
        assert pod.name == "test-pod"
        assert pod.container_image == "nginx:latest"
        assert pod.phase == "Running"
    
    def test_create_deployment(self, k8s_resource_factory):
        """Test creating a Deployment resource."""
        deployment = k8s_resource_factory.create_deployment(
            name="test-deployment",
            replicas=3,
            image="nginx:latest"
        )
        
        assert deployment.kind == "Deployment"
        assert deployment.replicas == 3
    
    def test_resource_to_dict(self, k8s_resource_factory):
        """Test converting resource to dictionary."""
        pod = k8s_resource_factory.create_pod(name="test-pod")
        pod_dict = pod.to_dict()
        
        assert pod_dict["apiVersion"] == "v1"
        assert pod_dict["kind"] == "Pod"
        assert pod_dict["metadata"]["name"] == "test-pod"


class TestMockHttpClient:
    """Tests for mock HTTP client fixture."""
    
    @pytest.mark.asyncio
    async def test_get_request(self, mock_httpx_client):
        """Test making a GET request."""
        response = await mock_httpx_client.get("http://example.com/api")
        
        assert response.status_code == 200
        assert response.json() == {"status": "ok"}
    
    @pytest.mark.asyncio
    async def test_post_request(self, mock_httpx_client):
        """Test making a POST request."""
        response = await mock_httpx_client.post(
            "http://example.com/api",
            json={"data": "test"}
        )
        
        assert response.status_code == 200


class TestMockOllamaClient:
    """Tests for mock Ollama client fixture."""
    
    @pytest.mark.asyncio
    async def test_generate(self, mock_ollama_client):
        """Test generating text."""
        response = await mock_ollama_client.generate(
            model="llama3.2:3b",
            prompt="Hello"
        )
        
        assert response.model == "llama3.2:3b"
        assert response.response is not None
    
    @pytest.mark.asyncio
    async def test_chat(self, mock_ollama_client):
        """Test chat completion."""
        response = await mock_ollama_client.chat(
            model="llama3.2:3b",
            messages=[{"role": "user", "content": "Hello"}]
        )
        
        assert response.message is not None
    
    @pytest.mark.asyncio
    async def test_embeddings(self, mock_ollama_client):
        """Test generating embeddings."""
        response = await mock_ollama_client.embeddings(
            model="nomic-embed-text",
            prompt="Test text"
        )
        
        assert response.embedding is not None
        assert len(response.embedding) == 768
    
    def test_set_response(self, mock_ollama_client):
        """Test setting custom response."""
        mock_ollama_client.set_response("Custom response")
        
        # Check the response was updated
        assert mock_ollama_client.generate.return_value.response == "Custom response"


class TestMockRedisClient:
    """Tests for mock Redis client fixture."""
    
    @pytest.mark.asyncio
    async def test_set_get(self, mock_redis_client):
        """Test setting and getting values."""
        await mock_redis_client.set("key", "value")
        result = await mock_redis_client.get("key")
        
        assert result == "value"
    
    @pytest.mark.asyncio
    async def test_setex(self, mock_redis_client):
        """Test setting value with expiry."""
        await mock_redis_client.setex("key", 60, "value")
        result = await mock_redis_client.get("key")
        ttl = await mock_redis_client.ttl("key")
        
        assert result == "value"
        assert ttl == 60
    
    @pytest.mark.asyncio
    async def test_incr_decr(self, mock_redis_client):
        """Test increment and decrement."""
        await mock_redis_client.set("counter", "10")
        
        result = await mock_redis_client.incr("counter")
        assert result == 11
        
        result = await mock_redis_client.decr("counter")
        assert result == 10
    
    @pytest.mark.asyncio
    async def test_list_operations(self, mock_redis_client):
        """Test list operations."""
        await mock_redis_client.rpush("list", "a", "b", "c")
        length = await mock_redis_client.llen("list")
        items = await mock_redis_client.lrange("list", 0, -1)
        
        assert length == 3
        assert items == ["a", "b", "c"]
    
    @pytest.mark.asyncio
    async def test_hash_operations(self, mock_redis_client):
        """Test hash operations."""
        await mock_redis_client.hset("hash", "field1", "value1")
        await mock_redis_client.hset("hash", "field2", "value2")
        
        result = await mock_redis_client.hgetall("hash")
        
        assert result["field1"] == "value1"
        assert result["field2"] == "value2"


class TestMetricsRegistry:
    """Tests for mock metrics registry fixture."""
    
    def test_create_counter(self, metrics_registry):
        """Test creating a counter."""
        counter = metrics_registry.counter(
            "test_counter",
            "Test counter",
            labelnames=["status"]
        )
        
        counter.labels(status="success").inc()
        counter.labels(status="success").inc()
        
        assert counter.get_value({"status": "success"}) == 2.0
    
    def test_create_gauge(self, metrics_registry):
        """Test creating a gauge."""
        gauge = metrics_registry.gauge(
            "test_gauge",
            "Test gauge",
            labelnames=["server"]
        )
        
        gauge.labels(server="a").set(100)
        gauge.labels(server="b").set(50)
        
        assert gauge.get_value({"server": "a"}) == 100
        assert gauge.get_value({"server": "b"}) == 50
    
    def test_create_histogram(self, metrics_registry):
        """Test creating a histogram."""
        histogram = metrics_registry.histogram(
            "test_histogram",
            "Test histogram",
            labelnames=["endpoint"]
        )
        
        histogram.labels(endpoint="/api").observe(0.1)
        histogram.labels(endpoint="/api").observe(0.5)
        histogram.labels(endpoint="/api").observe(1.0)
        
        observations = histogram.get_observations({"endpoint": "/api"})
        assert len(observations) == 3
        assert 0.1 in observations


# Import fixtures from the library
@pytest.fixture
def cloudevent_factory():
    from shared_tests.fixtures.cloudevents import CloudEventFactory
    return CloudEventFactory()


@pytest.fixture
def mock_k8s_client():
    from shared_tests.fixtures.kubernetes import MockKubernetesClient
    return MockKubernetesClient()


@pytest.fixture
def k8s_resource_factory():
    from shared_tests.fixtures.kubernetes import K8sResourceFactory
    return K8sResourceFactory()


@pytest.fixture
def mock_httpx_client():
    from shared_tests.fixtures.http import mock_httpx_client as create_client
    from unittest.mock import AsyncMock
    from shared_tests.fixtures.http import MockHTTPResponse
    
    mock_response = MockHTTPResponse(
        status_code=200,
        json_data={"status": "ok"}
    )
    
    client = AsyncMock()
    client.get = AsyncMock(return_value=mock_response)
    client.post = AsyncMock(return_value=mock_response)
    client.put = AsyncMock(return_value=mock_response)
    client.delete = AsyncMock(return_value=mock_response)
    
    client.__aenter__ = AsyncMock(return_value=client)
    client.__aexit__ = AsyncMock(return_value=None)
    
    return client


@pytest.fixture
def mock_ollama_client():
    from shared_tests.fixtures.ollama import MockOllamaClient
    return MockOllamaClient()


@pytest.fixture
def mock_redis_client():
    from shared_tests.fixtures.redis import MockRedisClient
    return MockRedisClient()


@pytest.fixture
def metrics_registry():
    from shared_tests.fixtures.metrics import MockMetricsRegistry
    return MockMetricsRegistry()
