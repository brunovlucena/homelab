"""
TRM (Tiny Recursive Model) client with built-in reflection.

This client provides direct access to TRM models for all homelab agents.
It implements the recursive self-refinement mechanism that makes TRM powerful.
"""
import os
import time
from typing import Optional, Dict, Any
import structlog
import httpx

try:
    from transformers import AutoModel, AutoTokenizer
    import torch
    TRANSFORMERS_AVAILABLE = True
except ImportError:
    TRANSFORMERS_AVAILABLE = False

from .types import TRMRequest, TRMResponse, ReflectionStep, ReflectionMode

logger = structlog.get_logger()


class TRMClient:
    """
    Client for TRM (Tiny Recursive Model) with built-in reflection.
    
    This client can work in two modes:
    1. Direct model loading (if transformers available)
    2. Hugging Face Inference API (fallback)
    
    The reflection mechanism is built-in and automatic.
    """
    
    def __init__(
        self,
        model_name: Optional[str] = None,
        model_path: Optional[str] = None,
        device: Optional[str] = None,
        use_hf_api: bool = False,
        hf_api_token: Optional[str] = None,
        h_cycles: int = 3,
        l_cycles: int = 6,
    ):
        """
        Initialize TRM client.
        
        Args:
            model_name: Hugging Face model name (e.g., "ainz/tiny-recursive-model")
            model_path: Local path to model checkpoint
            device: Device to use ("cuda", "cpu", "mps")
            use_hf_api: Use Hugging Face Inference API instead of local model
            hf_api_token: Hugging Face API token
            h_cycles: High-level reflection cycles
            l_cycles: Low-level latent update cycles
        """
        self.model_name = model_name or os.getenv(
            "TRM_MODEL_NAME",
            "ainz/tiny-recursive-model"
        )
        self.model_path = model_path or os.getenv("TRM_MODEL_PATH")
        self.device = device or os.getenv("DEVICE", "auto")
        self.use_hf_api = use_hf_api or os.getenv("TRM_USE_HF_API", "false").lower() == "true"
        self.hf_api_token = hf_api_token or os.getenv("HF_API_TOKEN")
        self.h_cycles = h_cycles
        self.l_cycles = l_cycles
        
        # Model state
        self.model = None
        self.tokenizer = None
        self._model_loaded = False
        
        # Auto-detect device
        if self.device == "auto":
            if TRANSFORMERS_AVAILABLE and torch.cuda.is_available():
                self.device = "cuda"
            elif TRANSFORMERS_AVAILABLE and hasattr(torch.backends, "mps") and torch.backends.mps.is_available():
                self.device = "mps"
            else:
                self.device = "cpu"
        
        logger.info(
            "trm_client_initialized",
            model_name=self.model_name,
            device=self.device,
            use_hf_api=self.use_hf_api,
            h_cycles=h_cycles,
            l_cycles=l_cycles,
        )
    
    async def load_model(self):
        """Load TRM model (lazy loading)."""
        if self._model_loaded:
            return
        
        if self.use_hf_api:
            # Using Hugging Face Inference API - no local model needed
            self._model_loaded = True
            logger.info("trm_using_hf_api", model_name=self.model_name)
            return
        
        if not TRANSFORMERS_AVAILABLE:
            logger.warning("transformers_not_available", fallback="hf_api")
            self.use_hf_api = True
            self._model_loaded = True
            return
        
        try:
            logger.info("trm_loading_model", model_name=self.model_name, device=self.device)
            
            # Load model with trust_remote_code=True for custom architectures
            self.tokenizer = AutoTokenizer.from_pretrained(
                self.model_name,
                trust_remote_code=True,
            )
            
            self.model = AutoModel.from_pretrained(
                self.model_name,
                trust_remote_code=True,
                torch_dtype=torch.float16 if self.device == "cuda" else torch.float32,
            )
            
            if self.device != "cpu":
                self.model = self.model.to(self.device)
            
            self.model.eval()
            self._model_loaded = True
            
            logger.info("trm_model_loaded", model_name=self.model_name, device=self.device)
            
        except Exception as e:
            logger.error("trm_model_load_failed", error=str(e), fallback="hf_api")
            # Fallback to HF API
            self.use_hf_api = True
            self._model_loaded = True
    
    async def generate(
        self,
        request: TRMRequest,
    ) -> TRMResponse:
        """
        Generate response using TRM with built-in reflection.
        
        This implements the recursive self-refinement mechanism:
        1. Generate initial answer
        2. For each reflection step:
           a. Reflect on the answer (identify issues, improvements)
           b. Refine the answer based on reflection
        3. Return final refined answer
        """
        start_time = time.time()
        
        # Ensure model is loaded
        if not self._model_loaded:
            await self.load_model()
        
        log = logger.bind(
            prompt=request.prompt[:100],
            reflection_mode=request.reflection_mode.value,
            max_steps=request.max_reflection_steps,
        )
        
        log.info("trm_generation_started")
        
        try:
            # Generate initial answer
            initial_answer = await self._generate_initial(request)
            reflection_trace: list[ReflectionStep] = []
            
            current_answer = initial_answer
            current_confidence = 0.5
            
            # Reflection loop
            reflection_steps = 0
            if request.reflection_mode != ReflectionMode.NEVER:
                for step in range(1, request.max_reflection_steps + 1):
                    # Check if we should continue reflecting
                    if request.reflection_mode == ReflectionMode.AUTO and current_confidence > 0.85:
                        log.info("trm_confidence_threshold_reached", confidence=current_confidence)
                        break
                    
                    # Reflect on current answer
                    reflection = await self._reflect_on_answer(
                        request.prompt,
                        current_answer,
                        step,
                        request.context,
                    )
                    
                    # Refine answer based on reflection
                    refined_answer = await self._refine_answer(
                        request.prompt,
                        current_answer,
                        reflection,
                        request.context,
                    )
                    
                    # Calculate improvement
                    improvement = self._calculate_improvement(
                        current_answer,
                        refined_answer,
                    )
                    
                    # Update confidence
                    current_confidence = min(
                        current_confidence + (improvement * 0.2),
                        0.95
                    )
                    
                    reflection_trace.append(ReflectionStep(
                        step=step,
                        initial_answer=current_answer,
                        reflection=reflection,
                        refined_answer=refined_answer,
                        confidence=current_confidence,
                        improvement_score=improvement,
                    ))
                    
                    current_answer = refined_answer
                    reflection_steps = step
                    
                    # Early stopping if no improvement
                    if improvement < 0.05:
                        log.info("trm_no_improvement", step=step)
                        break
            
            duration_ms = (time.time() - start_time) * 1000
            
            # Estimate tokens (rough approximation)
            tokens_used = len(request.prompt.split()) + len(current_answer.split())
            
            log.info(
                "trm_generation_completed",
                steps=reflection_steps,
                confidence=current_confidence,
                duration_ms=duration_ms,
            )
            
            return TRMResponse(
                answer=current_answer,
                reflection_steps=reflection_steps,
                confidence=current_confidence,
                reflection_trace=reflection_trace,
                duration_ms=duration_ms,
                tokens_used=tokens_used,
                model_name=self.model_name,
                conversation_id=request.conversation_id,
            )
            
        except Exception as e:
            log.error("trm_generation_failed", error=str(e))
            raise
    
    async def _generate_initial(self, request: TRMRequest) -> str:
        """Generate initial answer."""
        if self.use_hf_api:
            return await self._generate_via_hf_api(request.prompt, request.max_tokens)
        else:
            return await self._generate_via_model(request.prompt, request.max_tokens)
    
    async def _generate_via_model(self, prompt: str, max_tokens: int) -> str:
        """Generate using local model."""
        if not self.model or not self.tokenizer:
            raise RuntimeError("Model not loaded")
        
        inputs = self.tokenizer(prompt, return_tensors="pt")
        if self.device != "cpu":
            inputs = {k: v.to(self.device) for k, v in inputs.items()}
        
        with torch.no_grad():
            outputs = self.model.generate(
                **inputs,
                max_new_tokens=max_tokens,
                do_sample=True,
                temperature=0.7,
                pad_token_id=self.tokenizer.eos_token_id,
            )
        
        generated_text = self.tokenizer.decode(outputs[0], skip_special_tokens=True)
        # Remove the prompt from the generated text
        if generated_text.startswith(prompt):
            generated_text = generated_text[len(prompt):].strip()
        
        return generated_text
    
    async def _generate_via_hf_api(self, prompt: str, max_tokens: int) -> str:
        """Generate using Hugging Face Inference API."""
        api_url = f"https://api-inference.huggingface.co/models/{self.model_name}"
        headers = {}
        if self.hf_api_token:
            headers["Authorization"] = f"Bearer {self.hf_api_token}"
        
        async with httpx.AsyncClient(timeout=120.0) as client:
            response = await client.post(
                api_url,
                headers=headers,
                json={
                    "inputs": prompt,
                    "parameters": {
                        "max_new_tokens": max_tokens,
                        "temperature": 0.7,
                        "return_full_text": False,
                    },
                },
            )
            response.raise_for_status()
            result = response.json()
            
            if isinstance(result, list) and len(result) > 0:
                return result[0].get("generated_text", "")
            elif isinstance(result, dict):
                return result.get("generated_text", "")
            else:
                return str(result)
    
    async def _reflect_on_answer(
        self,
        prompt: str,
        current_answer: str,
        step: int,
        context: Dict[str, Any],
    ) -> str:
        """Reflect on the current answer and identify improvements."""
        reflection_prompt = f"""You are a reflection model. Analyze this answer and identify:
1. Any errors or inaccuracies
2. Areas that could be improved
3. Missing information
4. Logical inconsistencies

Original Question: {prompt}

Current Answer: {current_answer}

Provide a brief reflection on what could be improved:"""
        
        if self.use_hf_api:
            reflection = await self._generate_via_hf_api(reflection_prompt, 256)
        else:
            reflection = await self._generate_via_model(reflection_prompt, 256)
        
        return reflection.strip()
    
    async def _refine_answer(
        self,
        prompt: str,
        current_answer: str,
        reflection: str,
        context: Dict[str, Any],
    ) -> str:
        """Refine the answer based on reflection."""
        refinement_prompt = f"""Original Question: {prompt}

Previous Answer: {current_answer}

Reflection: {reflection}

Provide an improved answer that addresses the reflection:"""
        
        if self.use_hf_api:
            refined = await self._generate_via_hf_api(refinement_prompt, 1024)
        else:
            refined = await self._generate_via_model(refinement_prompt, 1024)
        
        return refined.strip()
    
    def _calculate_improvement(self, old_answer: str, new_answer: str) -> float:
        """Calculate improvement score between old and new answers."""
        # Simple heuristic: longer and more detailed = better
        # In production, you'd use a more sophisticated metric
        old_len = len(old_answer.split())
        new_len = len(new_answer.split())
        
        if old_len == 0:
            return 1.0
        
        improvement = min((new_len - old_len) / old_len, 0.5)
        return max(improvement, 0.0)
    
    async def health_check(self) -> bool:
        """Check if TRM client is ready."""
        if not self._model_loaded:
            try:
                await self.load_model()
            except Exception:
                return False
        
        if self.use_hf_api:
            # Check HF API availability
            try:
                async with httpx.AsyncClient(timeout=5.0) as client:
                    response = await client.get(
                        f"https://api-inference.huggingface.co/models/{self.model_name}",
                    )
                    return response.status_code in [200, 404]  # 404 means model exists but not loaded yet
            except Exception:
                return False
        else:
            return self.model is not None and self.tokenizer is not None
