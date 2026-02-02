"""Tests for report generator."""
import pytest
from unittest.mock import AsyncMock, MagicMock, patch
import json

from src.report_generator import ReportGenerator, HealthReport


def test_health_report_to_markdown():
    """Test HealthReport markdown conversion."""
    report = HealthReport(
        component="loki",
        status="Healthy",
        score=0.95,
        metrics={"test": 1.0},
        summary="All good",
        recommendations=["Keep monitoring"]
    )
    
    md = report.to_markdown()
    assert "LOKI Health Report" in md
    assert "Healthy" in md
    assert "95.00%" in md
    assert "All good" in md
    assert "Keep monitoring" in md


def test_health_report_to_json():
    """Test HealthReport JSON conversion."""
    report = HealthReport(
        component="loki",
        status="Healthy",
        score=0.95,
        metrics={"test": 1.0},
        summary="All good",
        recommendations=["Keep monitoring"]
    )
    
    json_str = report.to_json()
    data = json.loads(json_str)
    
    assert data["component"] == "loki"
    assert data["status"] == "Healthy"
    assert data["score"] == 0.95
    assert data["summary"] == "All good"
    assert len(data["recommendations"]) == 1


def test_health_report_to_html():
    """Test HealthReport HTML conversion."""
    report = HealthReport(
        component="loki",
        status="Healthy",
        score=0.95,
        metrics={"test": 1.0},
        summary="All good",
        recommendations=["Keep monitoring"]
    )
    
    html = report.to_html()
    assert "<h1>" in html
    assert "Healthy" in html
    assert "95.00%" in html


@pytest.mark.asyncio
async def test_parse_json_response_valid():
    """Test parsing valid JSON response."""
    generator = ReportGenerator("test-model", "ollama")
    
    text = '{"status": "Healthy", "score": 0.95, "summary": "Good", "recommendations": ["Monitor"]}'
    result = generator._parse_json_response(text)
    
    assert result["status"] == "Healthy"
    assert result["score"] == 0.95
    assert result["summary"] == "Good"
    assert len(result["recommendations"]) == 1


@pytest.mark.asyncio
async def test_parse_json_response_markdown_code_block():
    """Test parsing JSON from markdown code block."""
    generator = ReportGenerator("test-model", "ollama")
    
    text = '```json\n{"status": "Healthy", "score": 0.95, "summary": "Good", "recommendations": []}\n```'
    result = generator._parse_json_response(text)
    
    assert result["status"] == "Healthy"
    assert result["score"] == 0.95


@pytest.mark.asyncio
async def test_parse_json_response_invalid_fallback():
    """Test fallback parsing for invalid JSON."""
    generator = ReportGenerator("test-model", "ollama")
    
    text = "Status: Healthy\nScore: 95\nSummary: All systems operational"
    result = generator._parse_json_response(text)
    
    # Should extract what it can
    assert "status" in result
    assert "score" in result
    assert "summary" in result


@pytest.mark.asyncio
async def test_extract_from_text():
    """Test extracting data from unstructured text."""
    generator = ReportGenerator("test-model", "ollama")
    
    text = """
    Status: Healthy
    Health Score: 0.95
    Summary: All systems are operational and performing well.
    Recommendations:
    - Continue monitoring
    - Review metrics weekly
    """
    
    result = generator._extract_from_text(text)
    
    assert result["status"] == "Healthy"
    assert result["score"] == 0.95
    assert len(result["recommendations"]) > 0


@pytest.mark.asyncio
@patch('report_generator.generator.load')
async def test_load_mlx_model_success(mock_load):
    """Test loading MLX model successfully."""
    mock_model = MagicMock()
    mock_tokenizer = MagicMock()
    mock_generate = MagicMock()
    mock_load.return_value = (mock_model, mock_tokenizer)
    
    generator = ReportGenerator("functiongemma-270m-it", "mlx", mlx_enabled=True)
    
    with patch('report_generator.generator.generate', mock_generate):
        model = await generator._load_mlx_model()
    
    assert model["type"] == "mlx"
    assert model["model"] == mock_model
    assert model["tokenizer"] == mock_tokenizer


@pytest.mark.asyncio
async def test_load_mlx_model_import_error():
    """Test MLX model loading with import error."""
    generator = ReportGenerator("functiongemma-270m-it", "mlx", mlx_enabled=True)
    
    with patch('builtins.__import__', side_effect=ImportError("No module named 'mlx_lm'")):
        with patch.object(generator, '_load_ollama_model', new_callable=AsyncMock) as mock_ollama:
            mock_ollama.return_value = {"type": "ollama", "model_name": "test"}
            model = await generator._load_mlx_model()
    
    assert model["type"] == "ollama"


@pytest.mark.asyncio
async def test_load_ollama_model():
    """Test loading Ollama model."""
    generator = ReportGenerator("test-model", "ollama", ollama_url="http://test:11434")
    
    model = await generator._load_ollama_model()
    
    assert model["type"] == "ollama"
    assert model["model_name"] == "test-model"
    assert model["client"] is not None


@pytest.mark.asyncio
async def test_load_anthropic_model():
    """Test loading Anthropic model."""
    generator = ReportGenerator("test", "anthropic", anthropic_api_key="test-key")
    
    model = await generator._load_anthropic_model()
    
    assert model["type"] == "anthropic"
    assert model["api_key"] == "test-key"


@pytest.mark.asyncio
async def test_load_anthropic_model_no_key():
    """Test Anthropic model loading without API key."""
    generator = ReportGenerator("test", "anthropic", anthropic_api_key=None)
    
    with pytest.raises(ValueError, match="ANTHROPIC_API_KEY"):
        await generator._load_anthropic_model()


@pytest.mark.asyncio
@patch('httpx.AsyncClient')
async def test_generate_ollama(mock_client_class, sample_metrics):
    """Test generating report with Ollama."""
    mock_client = AsyncMock()
    mock_response = MagicMock()
    mock_response.json.return_value = {
        "response": '{"status": "Healthy", "score": 0.95, "summary": "Good", "recommendations": ["Monitor"]}'
    }
    mock_response.raise_for_status = MagicMock()
    mock_client.post = AsyncMock(return_value=mock_response)
    mock_client_class.return_value = mock_client
    
    generator = ReportGenerator("test-model", "ollama", ollama_url="http://test:11434")
    
    model = await generator._load_ollama_model()
    prompt = generator._create_prompt("loki", sample_metrics, "1h")
    
    result = await generator._generate_ollama(model, prompt)
    
    assert result["status"] == "Healthy"
    assert result["score"] == 0.95


@pytest.mark.asyncio
async def test_create_prompt(sample_metrics):
    """Test prompt creation."""
    generator = ReportGenerator("test-model", "ollama")
    
    prompt = generator._create_prompt("loki", sample_metrics, "1h")
    
    assert "loki" in prompt.lower()
    assert "1h" in prompt
    assert "metrics" in prompt.lower()
    assert "recommendations" in prompt.lower()


@pytest.mark.asyncio
async def test_generate_report(sample_metrics):
    """Test generating full report."""
    generator = ReportGenerator("test-model", "ollama", ollama_url="http://test:11434")
    
    # Mock the model loading and generation
    with patch.object(generator, '_load_ollama_model', new_callable=AsyncMock) as mock_load:
        mock_load.return_value = {
            "type": "ollama",
            "model_name": "test-model",
            "client": AsyncMock()
        }
        
        with patch.object(generator, '_generate_ollama', new_callable=AsyncMock) as mock_gen:
            mock_gen.return_value = {
                "status": "Healthy",
                "score": 0.95,
                "summary": "All good",
                "recommendations": ["Monitor"]
            }
            
            report = await generator.generate_report("loki", sample_metrics, "1h")
    
    assert isinstance(report, HealthReport)
    assert report.component == "loki"
    assert report.status == "Healthy"
    assert report.score == 0.95


@pytest.mark.asyncio
async def test_close():
    """Test closing generator resources."""
    generator = ReportGenerator("test-model", "ollama")
    generator._ollama_client = AsyncMock()
    
    await generator.close()
    
    generator._ollama_client.aclose.assert_called_once()
    assert generator._ollama_client is None

