#!/usr/bin/env python3
"""
🤖 Jamie MCP Server

Este servidor MCP fornece acesso às funcionalidades de SRE do Jamie.
Funcionalidades:
- Chat com Jamie AI (powered by Ollama + Agent-SRE)
- Query Prometheus metrics
- Verificar Golden Signals (latency, traffic, errors, saturation)
- Obter logs de pods do Kubernetes
- Análise de logs com AI
- Queries Grafana
"""

import asyncio
import json
import logging
import os
from typing import Any, Dict, Optional

import aiohttp
from mcp.server import Server
from mcp.server.models import InitializationOptions
from mcp.server.stdio import stdio_server
from mcp.types import CallToolResult, TextContent, Tool

from core import logger, logfire

# Configuração
JAMIE_SLACK_BOT_URL = os.getenv("JAMIE_SLACK_BOT_URL", "http://jamie-slack-bot-service.jamie.svc.cluster.local:8080")


class JamieAPIClient:
    """Cliente para a API REST do Jamie Slack Bot"""

    def __init__(self, base_url: str = None):
        self.base_url = base_url or JAMIE_SLACK_BOT_URL
        logger.info(f"🔌 Jamie API Client initialized - base URL: {self.base_url}")

    @logfire.instrument("jamie_api_chat")
    async def chat(self, message: str) -> Dict[str, Any]:
        """Envia mensagem para o Jamie via API de chat"""
        try:
            async with aiohttp.ClientSession() as session:
                async with session.post(
                    f"{self.base_url}/api/chat", json={"message": message}, timeout=aiohttp.ClientTimeout(total=60)
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return {
                            "success": True,
                            "response": data.get("response", "No response"),
                            "timestamp": data.get("timestamp"),
                        }
                    else:
                        error_text = await response.text()
                        logger.error(f"❌ Jamie chat error: {response.status} - {error_text}")
                        return {"success": False, "error": f"HTTP {response.status}: {error_text}"}
        except Exception as e:
            logger.error(f"❌ Error calling Jamie chat: {e}")
            return {"success": False, "error": str(e)}

    @logfire.instrument("jamie_api_prometheus_query")
    async def query_prometheus(self, query: str, time: Optional[str] = None) -> Dict[str, Any]:
        """Query Prometheus via Jamie"""
        try:
            async with aiohttp.ClientSession() as session:
                payload = {"query": query}
                if time:
                    payload["time"] = time

                async with session.post(
                    f"{self.base_url}/api/prometheus/query", json=payload, timeout=aiohttp.ClientTimeout(total=30)
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return {"success": True, "result": data}
                    else:
                        error_text = await response.text()
                        return {"success": False, "error": f"HTTP {response.status}: {error_text}"}
        except Exception as e:
            logger.error(f"❌ Error querying Prometheus: {e}")
            return {"success": False, "error": str(e)}

    @logfire.instrument("jamie_api_golden_signals")
    async def check_golden_signals(self, service: str, namespace: str = "default") -> Dict[str, Any]:
        """Verifica os Golden Signals de um serviço"""
        try:
            async with aiohttp.ClientSession() as session:
                async with session.post(
                    f"{self.base_url}/api/golden-signals",
                    json={"service": service, "namespace": namespace},
                    timeout=aiohttp.ClientTimeout(total=30),
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return {"success": True, "signals": data}
                    else:
                        error_text = await response.text()
                        return {"success": False, "error": f"HTTP {response.status}: {error_text}"}
        except Exception as e:
            logger.error(f"❌ Error checking golden signals: {e}")
            return {"success": False, "error": str(e)}

    @logfire.instrument("jamie_api_pod_logs")
    async def get_pod_logs(
        self, pod_name: str, namespace: str = "default", container: Optional[str] = None, lines: int = 100
    ) -> Dict[str, Any]:
        """Obtém logs de um pod do Kubernetes"""
        try:
            async with aiohttp.ClientSession() as session:
                payload = {"pod_name": pod_name, "namespace": namespace, "lines": lines}
                if container:
                    payload["container"] = container

                async with session.post(
                    f"{self.base_url}/api/pod-logs", json=payload, timeout=aiohttp.ClientTimeout(total=30)
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return {"success": True, "logs": data.get("logs", [])}
                    else:
                        error_text = await response.text()
                        return {"success": False, "error": f"HTTP {response.status}: {error_text}"}
        except Exception as e:
            logger.error(f"❌ Error getting pod logs: {e}")
            return {"success": False, "error": str(e)}

    @logfire.instrument("jamie_api_analyze_logs")
    async def analyze_logs(self, logs: str, context: Optional[str] = None) -> Dict[str, Any]:
        """Analisa logs com AI"""
        try:
            async with aiohttp.ClientSession() as session:
                payload = {"logs": logs}
                if context:
                    payload["context"] = context

                async with session.post(
                    f"{self.base_url}/api/analyze-logs", json=payload, timeout=aiohttp.ClientTimeout(total=60)
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return {"success": True, "analysis": data.get("analysis", "")}
                    else:
                        error_text = await response.text()
                        return {"success": False, "error": f"HTTP {response.status}: {error_text}"}
        except Exception as e:
            logger.error(f"❌ Error analyzing logs: {e}")
            return {"success": False, "error": str(e)}

    @logfire.instrument("jamie_api_health")
    async def health_check(self) -> Dict[str, Any]:
        """Verifica o status de saúde do Jamie"""
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(f"{self.base_url}/health", timeout=aiohttp.ClientTimeout(total=5)) as response:
                    if response.status == 200:
                        data = await response.json()
                        return {"success": True, "status": data}
                    else:
                        return {"success": False, "error": f"HTTP {response.status}"}
        except Exception as e:
            return {"success": False, "error": str(e)}


class JamieMCPServer:
    """Servidor MCP para Jamie - SRE Assistant"""

    def __init__(self):
        self.server = Server("jamie-mcp")
        self.api_client = JamieAPIClient()
        self._setup_handlers()

    def _setup_handlers(self):
        """Configura os handlers do servidor MCP"""

        @self.server.list_tools()
        async def handle_list_tools():
            """Lista as ferramentas disponíveis"""
            return [
                Tool(
                    name="chat",
                    description=(
                        "Conversa com Jamie, o assistente de SRE. Jamie pode responder perguntas sobre "
                        "infraestrutura, monitoramento, Kubernetes, etc."
                    ),
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "message": {"type": "string", "description": "Mensagem ou pergunta para o Jamie"}
                        },
                        "required": ["message"],
                    },
                ),
                Tool(
                    name="query_prometheus",
                    description="Executa query PromQL no Prometheus para obter métricas de infraestrutura",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "query": {
                                "type": "string",
                                "description": "Query PromQL (ex: 'up{job=\"homepage\"}')",
                            },
                            "time": {
                                "type": "string",
                                "description": "Timestamp opcional para query instantânea (formato RFC3339)",
                            },
                        },
                        "required": ["query"],
                    },
                ),
                Tool(
                    name="check_golden_signals",
                    description="Verifica os Golden Signals (latency, traffic, errors, saturation) de um serviço",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "service": {"type": "string", "description": "Nome do serviço para verificar"},
                            "namespace": {
                                "type": "string",
                                "description": "Namespace do Kubernetes (padrão: 'default')",
                                "default": "default",
                            },
                        },
                        "required": ["service"],
                    },
                ),
                Tool(
                    name="get_pod_logs",
                    description="Obtém logs de um pod do Kubernetes",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "pod_name": {"type": "string", "description": "Nome do pod"},
                            "namespace": {
                                "type": "string",
                                "description": "Namespace do Kubernetes (padrão: 'default')",
                                "default": "default",
                            },
                            "container": {
                                "type": "string",
                                "description": "Nome do container (opcional, usa o primeiro se não especificado)",
                            },
                            "lines": {
                                "type": "integer",
                                "description": "Número de linhas de log para retornar (padrão: 100)",
                                "default": 100,
                            },
                        },
                        "required": ["pod_name"],
                    },
                ),
                Tool(
                    name="analyze_logs",
                    description="Analisa logs com AI para identificar erros, padrões, e problemas",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "logs": {"type": "string", "description": "Logs para analisar (texto ou JSON)"},
                            "context": {
                                "type": "string",
                                "description": "Contexto adicional para a análise (opcional)",
                            },
                        },
                        "required": ["logs"],
                    },
                ),
                Tool(
                    name="health_check",
                    description="Verifica o status de saúde do Jamie e seus serviços",
                    inputSchema={"type": "object", "properties": {}},
                ),
            ]

        @self.server.call_tool()
        async def handle_call_tool(name: str, arguments: Dict[str, Any]) -> CallToolResult:
            """Executa uma ferramenta"""

            if name == "chat":
                message = arguments.get("message", "")

                if not message:
                    return CallToolResult(content=[TextContent(type="text", text="❌ Erro: Mensagem não fornecida")])

                result = await self.api_client.chat(message)

                if result["success"]:
                    text = f"🤖 Jamie: {result['response']}"
                    return CallToolResult(content=[TextContent(type="text", text=text)])
                else:
                    return CallToolResult(
                        content=[
                            TextContent(
                                type="text",
                                text=f"❌ Erro ao conversar com Jamie: {result.get('error', 'Erro desconhecido')}",
                            )
                        ]
                    )

            elif name == "query_prometheus":
                query = arguments.get("query", "")
                time = arguments.get("time")

                if not query:
                    return CallToolResult(
                        content=[TextContent(type="text", text="❌ Erro: Query PromQL não fornecida")]
                    )

                result = await self.api_client.query_prometheus(query, time)

                if result["success"]:
                    data = result.get("result", {})
                    text = "📊 **Prometheus Query Result**\n\n"
                    text += f"**Query**: `{query}`\n\n"
                    text += f"**Result**:\n```json\n{json.dumps(data, indent=2)}\n```"

                    return CallToolResult(content=[TextContent(type="text", text=text)])
                else:
                    return CallToolResult(
                        content=[
                            TextContent(
                                type="text",
                                text=f"❌ Erro ao consultar Prometheus: {result.get('error', 'Erro desconhecido')}",
                            )
                        ]
                    )

            elif name == "check_golden_signals":
                service = arguments.get("service", "")
                namespace = arguments.get("namespace", "default")

                if not service:
                    return CallToolResult(
                        content=[TextContent(type="text", text="❌ Erro: Nome do serviço não fornecido")]
                    )

                result = await self.api_client.check_golden_signals(service, namespace)

                if result["success"]:
                    signals = result.get("signals", {})
                    text = f"📊 **Golden Signals - {service}**\n"
                    text += f"Namespace: `{namespace}`\n\n"

                    for signal_name, signal_data in signals.items():
                        text += f"**{signal_name.title()}**: {signal_data}\n"

                    return CallToolResult(content=[TextContent(type="text", text=text)])
                else:
                    return CallToolResult(
                        content=[
                            TextContent(
                                type="text",
                                text=f"❌ Erro ao verificar golden signals: {result.get('error', 'Erro desconhecido')}",
                            )
                        ]
                    )

            elif name == "get_pod_logs":
                pod_name = arguments.get("pod_name", "")
                namespace = arguments.get("namespace", "default")
                container = arguments.get("container")
                lines = arguments.get("lines", 100)

                if not pod_name:
                    return CallToolResult(content=[TextContent(type="text", text="❌ Erro: Nome do pod não fornecido")])

                result = await self.api_client.get_pod_logs(pod_name, namespace, container, lines)

                if result["success"]:
                    logs = result.get("logs", [])
                    text = f"📝 **Logs do Pod: {pod_name}**\n"
                    text += f"Namespace: `{namespace}`\n"
                    if container:
                        text += f"Container: `{container}`\n"
                    text += "\n```\n"

                    if isinstance(logs, list):
                        text += "\n".join(logs)
                    else:
                        text += str(logs)

                    text += "\n```"

                    return CallToolResult(content=[TextContent(type="text", text=text)])
                else:
                    return CallToolResult(
                        content=[
                            TextContent(
                                type="text", text=f"❌ Erro ao obter logs: {result.get('error', 'Erro desconhecido')}"
                            )
                        ]
                    )

            elif name == "analyze_logs":
                logs = arguments.get("logs", "")
                context = arguments.get("context")

                if not logs:
                    return CallToolResult(content=[TextContent(type="text", text="❌ Erro: Logs não fornecidos")])

                result = await self.api_client.analyze_logs(logs, context)

                if result["success"]:
                    analysis = result.get("analysis", "")
                    text = f"🔍 **Análise de Logs**\n\n{analysis}"

                    return CallToolResult(content=[TextContent(type="text", text=text)])
                else:
                    return CallToolResult(
                        content=[
                            TextContent(
                                type="text",
                                text=f"❌ Erro ao analisar logs: {result.get('error', 'Erro desconhecido')}",
                            )
                        ]
                    )

            elif name == "health_check":
                result = await self.api_client.health_check()

                if result["success"]:
                    status = result.get("status", {})
                    text = "❤️ **Jamie Status**\n\n"
                    text += f"**Status**: {status.get('status', 'unknown')}\n"
                    text += f"**Service**: {status.get('service', 'jamie-slack-bot')}\n"
                    text += f"**Version**: {status.get('version', 'unknown')}\n"
                    text += f"**Agent-SRE URL**: {status.get('agent_sre_url', 'N/A')}\n"

                    return CallToolResult(content=[TextContent(type="text", text=text)])
                else:
                    return CallToolResult(
                        content=[
                            TextContent(
                                type="text",
                                text=f"❌ Jamie não está disponível: {result.get('error', 'Erro desconhecido')}",
                            )
                        ]
                    )

            else:
                return CallToolResult(content=[TextContent(type="text", text=f"❌ Ferramenta desconhecida: {name}")])

    async def run(self):
        """Executa o servidor MCP"""
        async with stdio_server() as (read_stream, write_stream):
            await self.server.run(
                read_stream,
                write_stream,
                InitializationOptions(
                    server_name="jamie-mcp",
                    server_version="1.0.0",
                    capabilities=self.server.get_capabilities(
                        notification_options=None,
                        experimental_capabilities=None,
                    ),
                ),
            )


async def main():
    """Função principal"""
    logger.info("🚀 Starting Jamie MCP Server...")
    server = JamieMCPServer()
    await server.run()


if __name__ == "__main__":
    asyncio.run(main())
