"""
Report Generator - Uses TRM (Tiny Recursive Model) with built-in reflection to generate reports.
"""
import os
from typing import Dict, Any, Optional
import structlog
import json
import re
import httpx
import asyncio

from src.sre_agent.config import AgentConfig

# Import TRM client
try:
    from agent_trm import TRMClient, TRMRequest, ReflectionMode
    TRM_AVAILABLE = True
except ImportError:
    TRM_AVAILABLE = False

logger = structlog.get_logger()


class HealthReport:
    """Health report data structure."""
    
    def __init__(
        self,
        component: str,
        status: str,
        score: float,
        metrics: Dict[str, Any],
        summary: str,
        recommendations: list
    ):
        self.component = component
        self.status = status
        self.score = score
        self.metrics = metrics
        self.summary = summary
        self.recommendations = recommendations
    
    def to_markdown(self) -> str:
        """Convert report to Markdown format."""
        md = f"# {self.component.upper()} Health Report\n\n"
        md += f"**Status**: {self.status}\n"
        md += f"**Health Score**: {self.score:.2%}\n\n"
        md += f"## Summary\n\n{self.summary}\n\n"
        md += "## Metrics\n\n"
        for key, value in self.metrics.items():
            md += f"- **{key}**: {value}\n"
        md += "\n## Recommendations\n\n"
        for rec in self.recommendations:
            md += f"- {rec}\n"
        return md
    
    def to_json(self) -> str:
        """Convert report to JSON format."""
        return json.dumps({
            "component": self.component,
            "status": self.status,
            "score": self.score,
            "metrics": self.metrics,
            "summary": self.summary,
            "recommendations": self.recommendations
        }, indent=2)
    
    def to_html(self) -> str:
        """Convert report to HTML format."""
        html = f"<h1>{self.component.upper()} Health Report</h1>"
        html += f"<p><strong>Status</strong>: {self.status}</p>"
        html += f"<p><strong>Health Score</strong>: {self.score:.2%}</p>"
        html += f"<h2>Summary</h2><p>{self.summary}</p>"
        html += "<h2>Metrics</h2><ul>"
        for key, value in self.metrics.items():
            html += f"<li><strong>{key}</strong>: {value}</li>"
        html += "</ul><h2>Recommendations</h2><ul>"
        for rec in self.recommendations:
            html += f"<li>{rec}</li>"
        html += "</ul>"
        return html


class ReportGenerator:
    """Generates health reports using TRM (Tiny Recursive Model) with built-in reflection."""
    
    def __init__(
        self,
        model_name: str,
        model_backend: str,
        mlx_enabled: bool = True,
        ollama_url: Optional[str] = None,
        anthropic_api_key: Optional[str] = None,
        # TRM configuration
        trm_model_name: Optional[str] = None,
        trm_use_hf_api: bool = False,
    ):
        self.model_name = model_name
        self.model_backend = model_backend
        self.mlx_enabled = mlx_enabled
        self.ollama_url = ollama_url or "http://ollama-native.ollama.svc.cluster.local:11434"
        self.anthropic_api_key = anthropic_api_key
        
        # TRM configuration (primary)
        self.trm_model_name = trm_model_name or os.getenv(
            "TRM_MODEL_NAME",
            "ainz/tiny-recursive-model"
        )
        self.trm_use_hf_api = trm_use_hf_api or os.getenv("TRM_USE_HF_API", "false").lower() == "true"
        
        # Legacy model state (fallback)
        self._model = None
        self._ollama_client = None
        
        # TRM client
        self.trm_client: Optional[TRMClient] = None
        if TRM_AVAILABLE:
            try:
                self.trm_client = TRMClient(
                    model_name=self.trm_model_name,
                    use_hf_api=self.trm_use_hf_api,
                )
                logger.info("trm_client_initialized_for_reports", model_name=self.trm_model_name)
            except Exception as e:
                logger.error("trm_client_init_failed", error=str(e))
                self.trm_client = None
    
    async def _get_model(self):
        """Get model instance (lazy loading)."""
        if self._model is None:
            if self.model_backend == "mlx" and self.mlx_enabled:
                self._model = await self._load_mlx_model()
            elif self.model_backend == "ollama":
                self._model = await self._load_ollama_model()
            else:
                self._model = await self._load_anthropic_model()
        return self._model
    
    async def _load_mlx_model(self):
        """Load model using MLX-LM framework."""
        try:
            from mlx_lm import load, generate
            logger.info("Loading MLX model", model=self.model_name)
            
            # Map model name to MLX model path
            # MLX community models are typically prefixed with mlx-community/
            mlx_model_name = self.model_name
            if not self.model_name.startswith("mlx-community/"):
                mlx_model_name = f"mlx-community/{self.model_name}"
            
            # Load model and tokenizer
            model, tokenizer = load(mlx_model_name)
            
            return {
                "type": "mlx",
                "model_name": self.model_name,
                "mlx_model": model,
                "tokenizer": tokenizer,
                "generate_fn": generate
            }
        except ImportError:
            logger.warning("MLX not available, falling back to Ollama")
            return await self._load_ollama_model()
        except Exception as e:
            logger.error("Failed to load MLX model", error=str(e))
            logger.warning("Falling back to Ollama")
            return await self._load_ollama_model()
    
    async def _load_ollama_model(self):
        """Load model using Ollama."""
        logger.info("Using Ollama backend", model=self.model_name, url=self.ollama_url)
        if self._ollama_client is None:
            self._ollama_client = httpx.AsyncClient(
                base_url=self.ollama_url.rstrip("/"),
                timeout=300.0  # 5 minutes for model inference
            )
        return {"type": "ollama", "model_name": self.model_name, "client": self._ollama_client}
    
    async def _load_anthropic_model(self):
        """Load Anthropic Claude as fallback."""
        if not self.anthropic_api_key:
            raise ValueError("ANTHROPIC_API_KEY is required for Anthropic backend")
        logger.info("Using Anthropic backend", model="claude-3-haiku")
        return {"type": "anthropic", "model_name": "claude-3-haiku", "api_key": self.anthropic_api_key}
    
    async def generate_report(
        self,
        component: str,
        metrics: Dict[str, Any],
        time_range: str
    ) -> HealthReport:
        """Generate health report for component using TRM (or fallback)."""
        # Prepare prompt
        prompt = self._create_prompt(component, metrics, time_range)
        
        # Use TRM if available, otherwise fallback to legacy
        if self.trm_client:
            report_data = await self._generate_with_trm(prompt)
        else:
            model = await self._get_model()
            report_data = await self._generate_with_model(model, prompt)
        
        # Parse and create report
        return self._parse_report(component, metrics, report_data)
    
    async def generate_full_report(
        self,
        metrics: Dict[str, Any],
        time_range: str
    ) -> HealthReport:
        """Generate comprehensive health report."""
        # Aggregate all components
        all_metrics = {}
        for component, comp_metrics in metrics.items():
            all_metrics.update(comp_metrics)
        
        return await self.generate_report("observability", all_metrics, time_range)
    
    def _create_prompt(
        self,
        component: str,
        metrics: Dict[str, Any],
        time_range: str
    ) -> str:
        """Create prompt for LLM."""
        prompt = f"""Generate a comprehensive SRE health report for {component}.

Metrics (from Prometheus record rules):
{json.dumps(metrics, indent=2)}

Time Range: {time_range}

Please provide:
1. Overall status (Healthy/Warning/Critical)
2. Health score (0-1)
3. Executive summary
4. Key findings from metrics
5. Actionable recommendations

Format the response as JSON with keys: status, score, summary, findings, recommendations.
"""
        return prompt
    
    async def _generate_with_trm(self, prompt: str) -> Dict[str, Any]:
        """Generate report using TRM with built-in reflection."""
        if not self.trm_client:
            raise RuntimeError("TRM client not initialized")
        
        try:
            request = TRMRequest(
                prompt=prompt,
                max_reflection_steps=3,
                reflection_mode=ReflectionMode.AUTO,
                max_tokens=2048,
            )
            
            response = await self.trm_client.generate(request)
            
            logger.info(
                "trm_report_generation_completed",
                reflection_steps=response.reflection_steps,
                confidence=response.confidence,
                duration_ms=response.duration_ms,
            )
            
            # Parse JSON from TRM response
            return self._parse_json_response(response.answer)
            
        except Exception as e:
            logger.error("trm_report_generation_failed", error=str(e))
            # Fallback to legacy
            logger.warning("falling_back_to_legacy_model")
            model = await self._get_model()
            return await self._generate_with_model(model, prompt)
    
    async def _generate_with_model(
        self,
        model: Dict[str, Any],
        prompt: str
    ) -> Dict[str, Any]:
        """Generate report using the model (legacy fallback)."""
        backend_type = model.get("type")
        logger.info("Generating report", backend=backend_type)
        
        if backend_type == "mlx":
            return await self._generate_mlx(model, prompt)
        elif backend_type == "ollama":
            return await self._generate_ollama(model, prompt)
        elif backend_type == "anthropic":
            return await self._generate_anthropic(model, prompt)
        else:
            raise ValueError(f"Unknown backend type: {backend_type}")
    
    async def _generate_mlx(self, model: Dict[str, Any], prompt: str) -> Dict[str, Any]:
        """Generate report using MLX-LM."""
        try:
            mlx_model = model["mlx_model"]
            tokenizer = model["tokenizer"]
            generate_fn = model["generate_fn"]
            
            # Generate response
            # MLX-LM generate function signature: generate(model, tokenizer, prompt, ...)
            response = generate_fn(
                mlx_model,
                tokenizer,
                prompt=prompt,
                max_tokens=1024,
                temp=0.7
            )
            
            # Parse JSON from response
            return self._parse_json_response(response)
        except Exception as e:
            logger.error("MLX generation failed", error=str(e))
            raise
    
    async def _generate_ollama(self, model: Dict[str, Any], prompt: str) -> Dict[str, Any]:
        """Generate report using Ollama with retry logic."""
        import asyncio
        client = model["client"]
        model_name = model["model_name"]
        
        max_retries = 3
        backoff = 1.0
        
        for attempt in range(max_retries):
            try:
                # Call Ollama API
                response = await client.post(
                    "/api/generate",
                    json={
                        "model": model_name,
                        "prompt": prompt,
                        "stream": False,
                        "options": {
                            "temperature": 0.7,
                            "num_predict": 1024
                        }
                    },
                    timeout=300.0
                )
                response.raise_for_status()
                result = response.json()
                
                # Extract response text
                response_text = result.get("response", "")
                
                # Parse JSON from response
                return self._parse_json_response(response_text)
            except (httpx.ConnectError, httpx.TimeoutException) as e:
                if attempt < max_retries - 1:
                    logger.warning(
                        "ollama_connection_failed_retrying",
                        attempt=attempt + 1,
                        max_retries=max_retries,
                        error=str(e),
                        backoff=backoff,
                        ollama_url=self.ollama_url
                    )
                    await asyncio.sleep(backoff)
                    backoff *= 2
                else:
                    logger.error(
                        "ollama_connection_failed_final",
                        error=str(e),
                        ollama_url=self.ollama_url
                    )
                    raise
            except Exception as e:
                logger.error("ollama_generation_failed", error=str(e), ollama_url=self.ollama_url)
                raise
    
    async def _generate_anthropic(self, model: Dict[str, Any], prompt: str) -> Dict[str, Any]:
        """Generate report using Anthropic Claude."""
        try:
            import anthropic
            import asyncio
            
            client = anthropic.Anthropic(api_key=model["api_key"])
            
            # Anthropic SDK is synchronous, run in executor
            def call_anthropic():
                message = client.messages.create(
                    model="claude-3-haiku-20240307",
                    max_tokens=1024,
                    temperature=0.7,
                    messages=[
                        {
                            "role": "user",
                            "content": prompt
                        }
                    ]
                )
                return message.content[0].text
            
            # Run synchronous call in executor
            loop = asyncio.get_event_loop()
            response_text = await loop.run_in_executor(None, call_anthropic)
            
            # Parse JSON from response
            return self._parse_json_response(response_text)
        except ImportError:
            logger.error("Anthropic library not installed")
            raise
        except Exception as e:
            logger.error("Anthropic generation failed", error=str(e))
            raise
    
    def _parse_json_response(self, text: str) -> Dict[str, Any]:
        """Parse JSON from LLM response, handling markdown code blocks."""
        # Try to extract JSON from markdown code blocks
        json_match = re.search(r'```(?:json)?\s*(\{.*?\})\s*```', text, re.DOTALL)
        if json_match:
            text = json_match.group(1)
        
        # Try to find JSON object in text
        json_match = re.search(r'\{.*\}', text, re.DOTALL)
        if json_match:
            text = json_match.group(0)
        
        try:
            data = json.loads(text)
            
            # Normalize response format
            normalized = {
                "status": data.get("status", "Unknown"),
                "score": float(data.get("score", 0.0)),
                "summary": data.get("summary", ""),
                "recommendations": data.get("recommendations", [])
            }
            
            # Handle findings if present
            if "findings" in data:
                normalized["recommendations"].extend(data["findings"])
            
            return normalized
        except json.JSONDecodeError as e:
            logger.warning("Failed to parse JSON, using fallback", error=str(e), text=text[:200])
            # Fallback: try to extract key information from text
            return self._extract_from_text(text)
    
    def _extract_from_text(self, text: str) -> Dict[str, Any]:
        """Extract report data from unstructured text as fallback."""
        # Try to extract status
        status_match = re.search(r'(?:status|state):\s*(\w+)', text, re.IGNORECASE)
        status = status_match.group(1) if status_match else "Unknown"
        
        # Try to extract score
        score_match = re.search(r'(?:score|health):\s*([0-9.]+)', text, re.IGNORECASE)
        score = float(score_match.group(1)) if score_match else 0.0
        if score > 1.0:
            score = score / 100.0  # Normalize percentage
        
        # Extract summary (first paragraph)
        summary_match = re.search(r'(?:summary|overview):\s*(.+?)(?:\n\n|\n[A-Z]|$)', text, re.IGNORECASE | re.DOTALL)
        summary = summary_match.group(1).strip() if summary_match else "Unable to generate summary"
        
        # Extract recommendations (bullet points)
        rec_matches = re.findall(r'(?:recommendation|action|suggestion):\s*(.+?)(?:\n|$)', text, re.IGNORECASE)
        recommendations = [r.strip() for r in rec_matches if r.strip()]
        
        if not recommendations:
            recommendations = ["Review metrics manually"]
        
        return {
            "status": status,
            "score": score,
            "summary": summary,
            "recommendations": recommendations
        }
    
    async def close(self):
        """Clean up resources."""
        if self._ollama_client:
            await self._ollama_client.aclose()
            self._ollama_client = None
    
    def _parse_report(
        self,
        component: str,
        metrics: Dict[str, Any],
        report_data: Dict[str, Any]
    ) -> HealthReport:
        """Parse model output into HealthReport."""
        return HealthReport(
            component=component,
            status=report_data.get("status", "Unknown"),
            score=report_data.get("score", 0.0),
            metrics=metrics,
            summary=report_data.get("summary", ""),
            recommendations=report_data.get("recommendations", [])
        )

