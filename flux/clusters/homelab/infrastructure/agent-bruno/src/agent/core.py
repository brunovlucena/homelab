"""
🤖 Agent Bruno Core

Core agent logic with homepage knowledge and memory integration.
"""

import logging
import httpx
from typing import Dict, Any, Optional
from ..knowledge.homepage import HomepageKnowledge
from ..memory.manager import MemoryManager

logger = logging.getLogger(__name__)


class AgentBruno:
    """Agent Bruno with homepage knowledge and memory"""

    def __init__(
        self,
        memory_manager: MemoryManager,
        ollama_url: str = "http://ollama.homepage.svc.cluster.local:11434",
        model: str = "llama3.2:3b"
    ):
        """
        Initialize Agent Bruno
        
        Args:
            memory_manager: Memory manager instance
            ollama_url: Ollama server URL
            model: LLM model to use
        """
        self.memory = memory_manager
        self.knowledge = HomepageKnowledge()
        self.ollama_url = ollama_url
        self.model = model
        self.system_prompt = self._build_system_prompt()

    def _build_system_prompt(self) -> str:
        """Build system prompt with homepage knowledge"""
        return f"""You are Agent Bruno, an AI assistant with deep knowledge of Bruno's homepage application.

{self.knowledge.get_summary()}

CAPABILITIES:
- Answer questions about the homepage architecture, APIs, deployment, and components
- Help with troubleshooting and debugging
- Provide code examples and best practices
- Remember previous conversations per user (IP-based memory)
- Search the knowledge base for specific information

GUIDELINES:
- Be helpful, friendly, and professional
- Use emojis 🎵 when appropriate to add personality
- Provide accurate information from the knowledge base
- If you don't know something, say so - don't make up information
- Reference specific files, endpoints, or components when relevant
- Keep responses concise but complete

KNOWLEDGE AREAS:
- Architecture: Frontend (React), API (Go), Database (PostgreSQL), Cache (Redis), Storage (MinIO)
- API Endpoints: Projects, Skills, Experiences, Content, Agents, Assets, Cloudflare
- Deployment: Docker Compose (local), Kubernetes/Helm (production), GitHub Actions (CI/CD)
- Components: Frontend files, API handlers, Database schema
- Tech Stack: Go 1.23, React 18, PostgreSQL 15, Redis 7, OpenTelemetry

Remember: You have access to the user's conversation history through memory. Use it for context!
"""

    async def chat(
        self,
        message: str,
        ip: str,
        context: Dict[str, Any] = None
    ) -> Dict[str, Any]:
        """
        Process a chat message

        Args:
            message: User message
            ip: User IP address
            context: Additional context

        Returns:
            Response dictionary
        """
        try:
            # Get recent conversation context
            recent_messages = await self.memory.get_recent_context(ip, limit=5)
            conversation_context = self.memory.format_context_for_prompt(recent_messages)

            # Search knowledge base if needed
            knowledge_context = self._get_knowledge_context(message)

            # Build prompt
            full_prompt = self._build_prompt(message, conversation_context, knowledge_context)

            # Get LLM response
            response_text = await self._query_ollama(full_prompt)

            # Save to memory
            await self.memory.save(ip, message, response_text, context)

            return {
                "success": True,
                "response": response_text,
                "model": self.model,
                "context_used": {
                    "recent_messages": len(recent_messages),
                    "knowledge_context": bool(knowledge_context)
                }
            }

        except Exception as e:
            logger.error(f"❌ Chat error: {e}")
            return {
                "success": False,
                "error": str(e),
                "response": "I'm sorry, I encountered an error processing your message. Please try again."
            }

    def _get_knowledge_context(self, message: str) -> Optional[str]:
        """Get relevant knowledge context for message"""
        # Keywords that trigger knowledge search
        keywords = [
            "how", "what", "where", "deploy", "api", "endpoint", "database",
            "architecture", "component", "tech", "stack", "configuration"
        ]

        message_lower = message.lower()
        if any(keyword in message_lower for keyword in keywords):
            # Search knowledge base
            results = self.knowledge.search(message_lower)
            if results:
                # Return top result
                return str(results[0]["data"])

        return None

    def _build_prompt(
        self,
        message: str,
        conversation_context: str,
        knowledge_context: Optional[str]
    ) -> str:
        """Build complete prompt for LLM"""
        parts = [self.system_prompt]

        if conversation_context:
            parts.append(f"\n{conversation_context}")

        if knowledge_context:
            parts.append(f"\nRelevant Knowledge:\n{knowledge_context}")

        parts.append(f"\nCurrent User Message: {message}")
        parts.append("\nAgent Bruno:")

        return "\n".join(parts)

    async def _query_ollama(self, prompt: str) -> str:
        """Query Ollama LLM"""
        url = f"{self.ollama_url}/api/generate"

        payload = {
            "model": self.model,
            "prompt": prompt,
            "stream": False,
            "options": {
                "temperature": 0.7,
                "top_p": 0.9,
                "top_k": 40
            }
        }

        try:
            async with httpx.AsyncClient(timeout=60.0) as client:
                response = await client.post(url, json=payload)
                response.raise_for_status()

                data = response.json()
                return data.get("response", "").strip()

        except httpx.TimeoutException:
            logger.error("⏱️ Ollama request timed out")
            raise Exception("LLM request timed out")
        except httpx.HTTPStatusError as e:
            logger.error(f"❌ Ollama HTTP error: {e}")
            raise Exception(f"LLM service error: {e.response.status_code}")
        except Exception as e:
            logger.error(f"❌ Ollama error: {e}")
            raise Exception(f"Failed to query LLM: {str(e)}")

    async def get_memory_stats(self, ip: str) -> Dict[str, Any]:
        """Get memory statistics for IP"""
        recent = await self.memory.get_recent_context(ip)
        total = await self.memory.mongo_store.get_total_conversations(ip)

        return {
            "ip": ip,
            "recent_messages": len(recent),
            "total_conversations": total,
            "has_history": total > 0
        }

    async def clear_memory(self, ip: str):
        """Clear memory for IP"""
        await self.memory.clear_memory(ip)
        logger.info(f"🗑️ Cleared memory for IP: {ip}")

    def get_knowledge_summary(self) -> str:
        """Get knowledge base summary"""
        return self.knowledge.get_summary()

    def search_knowledge(self, query: str) -> list:
        """Search knowledge base"""
        return self.knowledge.search(query)