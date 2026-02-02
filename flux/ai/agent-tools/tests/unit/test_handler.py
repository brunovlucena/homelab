"""Unit tests for K8s handler."""
import pytest
from unittest.mock import patch, MagicMock

import sys
sys.path.insert(0, "src")


class TestParseEventType:
    """Tests for parse_event_type function."""
    
    def test_parse_k8s_event(self):
        """Test parsing K8s event type."""
        from k8s_tools.handler import parse_event_type
        
        domain, resource, operation = parse_event_type("io.homelab.k8s.pods.list")
        
        assert domain == "k8s"
        assert resource == "pods"
        assert operation == "list"
    
    def test_parse_flux_event(self):
        """Test parsing Flux event type."""
        from k8s_tools.handler import parse_event_type
        
        domain, resource, operation = parse_event_type("io.homelab.flux.kustomizations.reconcile")
        
        assert domain == "flux"
        assert resource == "kustomizations"
        assert operation == "reconcile"
    
    def test_parse_knative_event(self):
        """Test parsing Knative event type."""
        from k8s_tools.handler import parse_event_type
        
        domain, resource, operation = parse_event_type("io.homelab.knative.services.create")
        
        assert domain == "knative"
        assert resource == "services"
        assert operation == "create"
    
    def test_parse_lambda_event(self):
        """Test parsing Lambda event type."""
        from k8s_tools.handler import parse_event_type
        
        domain, resource, operation = parse_event_type("io.homelab.lambda.functions.get")
        
        assert domain == "lambda"
        assert resource == "functions"
        assert operation == "get"
    
    def test_parse_invalid_event(self):
        """Test parsing invalid event type."""
        from k8s_tools.handler import parse_event_type
        
        with pytest.raises(ValueError, match="Invalid event type"):
            parse_event_type("invalid")


class TestAPIGroups:
    """Tests for API_GROUPS mapping."""
    
    def test_core_resources(self):
        """Test core resource mappings."""
        from k8s_tools.handler import API_GROUPS
        
        assert "pods" in API_GROUPS
        assert "services" in API_GROUPS
        assert "configmaps" in API_GROUPS
        assert "secrets" in API_GROUPS
    
    def test_apps_resources(self):
        """Test apps resource mappings."""
        from k8s_tools.handler import API_GROUPS
        
        assert "deployments" in API_GROUPS
        assert "statefulsets" in API_GROUPS
        assert "daemonsets" in API_GROUPS
    
    def test_knative_resources(self):
        """Test Knative resource mappings."""
        from k8s_tools.handler import API_GROUPS
        
        assert "ksvc" in API_GROUPS
        assert "brokers" in API_GROUPS
        assert "triggers" in API_GROUPS
    
    def test_lambda_resources(self):
        """Test Lambda resource mappings."""
        from k8s_tools.handler import API_GROUPS
        
        assert "lambdafunctions" in API_GROUPS
        assert "functions" in API_GROUPS
        assert "lambdaagents" in API_GROUPS
    
    def test_flux_resources(self):
        """Test Flux resource mappings."""
        from k8s_tools.handler import API_GROUPS
        
        assert "kustomizations" in API_GROUPS
        assert "helmreleases" in API_GROUPS
        assert "gitrepositories" in API_GROUPS


class TestK8sHandler:
    """Tests for K8sHandler class."""
    
    @patch('k8s_tools.handler.config')
    @patch('k8s_tools.handler.client')
    @patch('k8s_tools.handler.DynamicClient')
    def test_handler_init_in_cluster(self, mock_dynamic, mock_client, mock_config):
        """Test handler initialization in cluster."""
        from k8s_tools.handler import K8sHandler
        
        handler = K8sHandler()
        
        mock_config.load_incluster_config.assert_called_once()
        assert handler.dynamic is not None
    
    @patch('k8s_tools.handler.config')
    @patch('k8s_tools.handler.client')
    @patch('k8s_tools.handler.DynamicClient')
    def test_handler_init_out_of_cluster(self, mock_dynamic, mock_client, mock_config):
        """Test handler initialization out of cluster."""
        from k8s_tools.handler import K8sHandler
        
        mock_config.load_incluster_config.side_effect = Exception("Not in cluster")
        
        handler = K8sHandler()
        
        mock_config.load_kube_config.assert_called_once()
    
    @patch('k8s_tools.handler.config')
    @patch('k8s_tools.handler.client')
    @patch('k8s_tools.handler.DynamicClient')
    def test_get_resource(self, mock_dynamic, mock_client, mock_config):
        """Test getting resource client."""
        from k8s_tools.handler import K8sHandler
        
        handler = K8sHandler()
        mock_dynamic.return_value.resources.get.return_value = MagicMock()
        
        resource = handler._get_resource("pods")
        
        mock_dynamic.return_value.resources.get.assert_called_once()
    
    @patch('k8s_tools.handler.config')
    @patch('k8s_tools.handler.client')
    @patch('k8s_tools.handler.DynamicClient')
    def test_get_resource_unknown(self, mock_dynamic, mock_client, mock_config):
        """Test getting unknown resource type."""
        from k8s_tools.handler import K8sHandler
        
        handler = K8sHandler()
        
        with pytest.raises(ValueError, match="Unknown resource type"):
            handler._get_resource("unknown")


class TestHandleFunction:
    """Tests for handle() function."""
    
    @patch('k8s_tools.handler.get_handler')
    def test_handle_list_operation(self, mock_get_handler):
        """Test handling list operation."""
        from k8s_tools.handler import handle
        
        mock_handler = MagicMock()
        mock_handler.list.return_value = {"items": [], "count": 0}
        mock_get_handler.return_value = mock_handler
        
        event = {
            "type": "io.homelab.k8s.pods.list",
            "data": {"namespace": "default"}
        }
        
        result = handle(event)
        
        assert result["success"] is True
        assert result["operation"] == "list"
        mock_handler.list.assert_called_once()
    
    @patch('k8s_tools.handler.get_handler')
    def test_handle_get_operation(self, mock_get_handler):
        """Test handling get operation."""
        from k8s_tools.handler import handle
        
        mock_handler = MagicMock()
        mock_handler.get.return_value = {"metadata": {"name": "test"}}
        mock_get_handler.return_value = mock_handler
        
        event = {
            "type": "io.homelab.k8s.pods.get",
            "data": {"namespace": "default", "name": "test-pod"}
        }
        
        result = handle(event)
        
        assert result["success"] is True
        assert result["operation"] == "get"
    
    @patch('k8s_tools.handler.get_handler')
    def test_handle_scale_operation(self, mock_get_handler):
        """Test handling scale operation."""
        from k8s_tools.handler import handle
        
        mock_handler = MagicMock()
        mock_handler.scale.return_value = {"spec": {"replicas": 3}}
        mock_get_handler.return_value = mock_handler
        
        event = {
            "type": "io.homelab.k8s.deployments.scale",
            "data": {"namespace": "default", "name": "test", "replicas": 3}
        }
        
        result = handle(event)
        
        assert result["success"] is True
        assert result["operation"] == "scale"
    
    @patch('k8s_tools.handler.get_handler')
    def test_handle_unknown_operation(self, mock_get_handler):
        """Test handling unknown operation."""
        from k8s_tools.handler import handle
        
        mock_get_handler.return_value = MagicMock()
        
        event = {
            "type": "io.homelab.k8s.pods.unknown",
            "data": {}
        }
        
        result = handle(event)
        
        assert result["success"] is False
        assert "Unknown operation" in result["error"]
    
    @patch('k8s_tools.handler.get_handler')
    def test_handle_api_exception(self, mock_get_handler):
        """Test handling API exception."""
        from k8s_tools.handler import handle
        from kubernetes.client.rest import ApiException
        
        mock_handler = MagicMock()
        mock_handler.list.side_effect = ApiException(status=404, reason="Not Found")
        mock_get_handler.return_value = mock_handler
        
        event = {
            "type": "io.homelab.k8s.pods.list",
            "data": {"namespace": "nonexistent"}
        }
        
        result = handle(event)
        
        assert result["success"] is False
        assert result["errorCode"] == 404
