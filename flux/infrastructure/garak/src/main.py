"""
🛡️ Garak API Wrapper
Exposes Garak LLM vulnerability scanning via HTTP API
"""

import asyncio
import subprocess
import json
import os
from typing import Optional, Dict, Any
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

app = FastAPI(
    title="Garak API",
    description="LLM Vulnerability Scanner API",
    version="1.0.0"
)


class ScanRequest(BaseModel):
    """Request model for running a Garak scan"""
    model_type: str = "openai"  # openai, anthropic, etc.
    model_name: str  # e.g., "gpt-3.5-turbo"
    probes: Optional[str] = None  # Comma-separated probe names
    url: Optional[str] = None  # For HTTP-based LLM endpoints
    report_format: Optional[str] = "json"  # json, txt, html
    output_file: Optional[str] = None


class ScanResponse(BaseModel):
    """Response model for scan results"""
    status: str
    scan_id: str
    command: str
    output: Optional[str] = None
    error: Optional[str] = None


@app.get("/health")
async def health():
    """Health check endpoint"""
    return {"status": "healthy", "service": "garak-api"}


@app.get("/")
async def root():
    """Root endpoint with API information"""
    return {
        "service": "Garak API",
        "description": "LLM Vulnerability Scanner by NVIDIA",
        "version": "1.0.0",
        "endpoints": {
            "health": "/health",
            "scan": "/scan",
            "scan_agent_evil": "/scan/agent-evil",
            "scan_agent_sre": "/scan/agent-sre",
            "list_probes": "/probes",
            "docs": "/docs"
        }
    }


@app.get("/probes")
async def list_probes():
    """List available Garak probes"""
    try:
        result = subprocess.run(
            ["garak", "--list_probes"],
            capture_output=True,
            text=True,
            timeout=30
        )
        return {
            "status": "success",
            "probes": result.stdout.split("\n") if result.returncode == 0 else []
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error listing probes: {str(e)}")


@app.post("/scan", response_model=ScanResponse)
async def run_scan(request: ScanRequest) -> ScanResponse:
    """
    Run a Garak vulnerability scan
    
    Example:
    ```json
    {
        "model_type": "openai",
        "model_name": "gpt-3.5-turbo",
        "probes": "prompt_injection,jailbreak",
        "report_format": "json"
    }
    ```
    """
    import uuid
    scan_id = str(uuid.uuid4())[:8]
    
    # Build Garak command
    cmd = ["garak"]
    
    if request.model_type:
        cmd.extend(["--model_type", request.model_type])
    
    if request.model_name:
        cmd.extend(["--model_name", request.model_name])
    
    if request.url:
        cmd.extend(["--url", request.url])
    
    if request.probes:
        cmd.extend(["--probes", request.probes])
    
    if request.report_format:
        cmd.extend(["--report_format", request.report_format])
    
    if request.output_file:
        cmd.extend(["--output", request.output_file])
    else:
        # Default output location
        output_file = f"/tmp/garak_scan_{scan_id}.{request.report_format}"
        cmd.extend(["--output", output_file])
    
    try:
        # Run Garak scan
        process = await asyncio.create_subprocess_exec(
            *cmd,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        
        stdout, stderr = await asyncio.wait_for(
            process.communicate(),
            timeout=300  # 5 minute timeout
        )
        
        # Read output file if it exists
        output_content = None
        if request.output_file and os.path.exists(request.output_file):
            with open(request.output_file, 'r') as f:
                output_content = f.read()
        elif os.path.exists(output_file):
            with open(output_file, 'r') as f:
                output_content = f.read()
        
        if process.returncode != 0:
            return ScanResponse(
                status="error",
                scan_id=scan_id,
                command=" ".join(cmd),
                error=stderr.decode() if stderr else "Scan failed",
                output=output_content
            )
        
        return ScanResponse(
            status="success",
            scan_id=scan_id,
            command=" ".join(cmd),
            output=output_content or stdout.decode(),
            error=stderr.decode() if stderr else None
        )
    
    except asyncio.TimeoutError:
        raise HTTPException(
            status_code=504,
            detail=f"Scan timeout after 5 minutes. Scan ID: {scan_id}"
        )
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Error running scan: {str(e)}"
        )


@app.post("/scan/agent-evil")
async def scan_agent_evil():
    """
    Quick scan endpoint for agent-evil (pre-configured)
    """
    request = ScanRequest(
        model_type="http",
        model_name="agent-evil",
        url="http://agent-evil-service.agent-evil.svc.cluster.local:8080/chat",
        probes="prompt_injection,jailbreak",
        report_format="json"
    )
    return await run_scan(request)


@app.post("/scan/agent-sre")
async def scan_agent_sre():
    """
    Quick scan endpoint for agent-sre (pre-configured)
    
    Note: Agent-SRE currently doesn't have a chat endpoint for LLM vulnerability testing.
    This endpoint tests the health endpoint to verify connectivity.
    For proper LLM vulnerability testing, agent-sre would need a chat endpoint.
    """
    request = ScanRequest(
        model_type="http",
        model_name="agent-sre",
        url="http://agent-sre.ai.svc.cluster.local/health",
        probes="prompt_injection,jailbreak",
        report_format="json"
    )
    return await run_scan(request)


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)

