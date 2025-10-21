#!/usr/bin/env python3
"""
🔍 Automated Investigation Workflow with LangGraph
Handles end-to-end investigation from detection to resolution
"""

import logging
import os
from datetime import datetime
from typing import Annotated, Any, Dict, List, Optional, Sequence, TypedDict

import httpx
from langchain_core.messages import AIMessage, BaseMessage, HumanMessage, SystemMessage
from langchain_ollama import ChatOllama
from langgraph.checkpoint.memory import MemorySaver
from langgraph.graph import END, START, StateGraph
from langgraph.graph.message import add_messages
from langsmith import traceable

from github_integration import GitHubIssue, github_client

logger = logging.getLogger(__name__)


# Configuration
OLLAMA_URL = os.environ.get("OLLAMA_URL", "http://192.168.0.16:11434")
MODEL_NAME = os.environ.get("MODEL_NAME", "llama3.2:3b")
MCP_SERVER_URL = os.getenv("MCP_SERVER_URL", "http://sre-agent-mcp-server-service:30120")


class InvestigationState(TypedDict):
    """🔧 State for Investigation Workflow"""

    # Investigation metadata
    investigation_id: Optional[str]
    issue_number: Optional[int]
    title: str
    description: str
    severity: str  # critical, high, medium, low
    component: str  # e.g., homepage, agent-sre, etc.

    # Investigation data
    messages: Annotated[Sequence[BaseMessage], add_messages]
    context: Dict[str, Any]

    # Sift investigation results
    sift_investigation: Optional[Dict[str, Any]]
    error_patterns: Optional[List[Dict[str, Any]]]
    slow_requests: Optional[List[Dict[str, Any]]]

    # Analysis results
    root_cause: Optional[str]
    recommendations: List[str]
    resolution_steps: List[str]

    # GitHub issue
    github_issue: Optional[GitHubIssue]

    # Workflow control
    next_action: Optional[str]
    completed: bool


class AutomatedInvestigationWorkflow:
    """🤖 Automated Investigation Workflow using LangGraph"""

    def __init__(self):
        # Initialize LLM
        self.llm = ChatOllama(
            model=MODEL_NAME,
            base_url=OLLAMA_URL,
            temperature=0.3,  # Lower temperature for more focused analysis
            num_ctx=8192,
        )

        # Initialize checkpointer for state persistence
        self.checkpointer = MemorySaver()

        # Build the workflow graph
        self.graph = self._build_graph()

        logger.info("🤖 Automated Investigation Workflow initialized")

    def _build_graph(self) -> StateGraph:
        """🏗️ Build the LangGraph investigation workflow"""

        workflow = StateGraph(InvestigationState)

        # Add nodes
        workflow.add_node("detect_issue", self._detect_issue_node)
        workflow.add_node("create_github_issue", self._create_github_issue_node)
        workflow.add_node("run_sift_investigation", self._run_sift_investigation_node)
        workflow.add_node("analyze_findings", self._analyze_findings_node)
        workflow.add_node("generate_recommendations", self._generate_recommendations_node)
        workflow.add_node("update_github_issue", self._update_github_issue_node)

        # Define the flow
        workflow.add_edge(START, "detect_issue")
        workflow.add_edge("detect_issue", "create_github_issue")
        workflow.add_edge("create_github_issue", "run_sift_investigation")
        workflow.add_edge("run_sift_investigation", "analyze_findings")
        workflow.add_edge("analyze_findings", "generate_recommendations")
        workflow.add_edge("generate_recommendations", "update_github_issue")
        workflow.add_edge("update_github_issue", END)

        return workflow.compile(checkpointer=self.checkpointer)

    @traceable(name="detect_issue")
    async def _detect_issue_node(self, state: InvestigationState) -> InvestigationState:
        """🔍 Detect and classify the issue"""
        logger.info(f"🔍 Detecting issue: {state['title']}")

        # Extract issue details using LLM
        system_prompt = """You are an expert SRE analyzing incidents and issues.
        Extract key information from the issue description and classify severity.
        Identify the affected component and potential investigation scope."""

        user_message = f"""
        Title: {state['title']}
        Description: {state['description']}

        Please analyze and extract:
        1. Severity level (critical/high/medium/low)
        2. Affected component
        3. Time scope for investigation (last 30m, 1h, 24h)
        4. Key labels/tags that should be extracted from the description
        """

        try:
            response = await self.llm.ainvoke(
                [SystemMessage(content=system_prompt), HumanMessage(content=user_message)]
            )

            state["messages"] = [HumanMessage(content=user_message), response]
            state["context"]["detection_analysis"] = response.content

            logger.info(f"✅ Issue detected and classified: {state.get('severity', 'unknown')}")

        except Exception as e:
            logger.error(f"❌ Error in detect_issue node: {e}", exc_info=True)
            state["context"]["detection_error"] = str(e)

        return state

    @traceable(name="create_github_issue")
    async def _create_github_issue_node(self, state: InvestigationState) -> InvestigationState:
        """📝 Create a GitHub issue for tracking"""
        logger.info(f"📝 Creating GitHub issue: {state['title']}")

        # Prepare issue body
        issue_body = f"""## Issue Description

{state['description']}

## Investigation Details

* **Severity**: {state.get('severity', 'unknown')}
* **Component**: {state.get('component', 'unknown')}
* **Detected**: {datetime.now().isoformat()}
* **Status**: 🔍 Investigation in progress

---

## Investigation Timeline

### 1. Detection
- Issue detected and workflow initiated
- Severity classified as: {state.get('severity', 'unknown')}

### 2. Investigation
- Running Grafana Sift analysis...

---

_This issue was automatically created by Agent-SRE_
"""

        # Determine labels based on severity and component
        labels = ["automated", "investigation"]
        if state.get("severity"):
            labels.append(f"severity:{state['severity']}")
        if state.get("component"):
            labels.append(f"component:{state['component']}")

        # Create the issue
        issue = await github_client.create_issue(
            title=state["title"],
            body=issue_body,
            labels=labels,
            assignees=[os.getenv("GITHUB_DEFAULT_ASSIGNEE", "brunovlucena")],
        )

        if issue:
            state["github_issue"] = issue
            state["issue_number"] = issue.number
            logger.info(f"✅ GitHub issue created: {issue.html_url}")
        else:
            logger.error("❌ Failed to create GitHub issue")

        return state

    @traceable(name="run_sift_investigation")
    async def _run_sift_investigation_node(self, state: InvestigationState) -> InvestigationState:
        """🔬 Run Grafana Sift investigation"""
        logger.info(f"🔬 Running Sift investigation for: {state['title']}")

        # Build labels for investigation scope
        labels = {
            "cluster": "homelab",
        }

        # Add component-specific labels
        if state.get("component"):
            labels["namespace"] = state["component"]
            labels["app"] = state["component"]

        try:
            # Create Sift investigation via MCP server
            async with httpx.AsyncClient() as client:
                # Create investigation
                create_payload = {
                    "tool": "sift_create_investigation",
                    "arguments": {
                        "name": f"Investigation: {state['title']}",
                        "labels": labels,
                    },
                }

                logger.info(f"🔬 Creating Sift investigation with labels: {labels}")
                response = await client.post(
                    f"{MCP_SERVER_URL}/mcp/tool", json=create_payload, timeout=60
                )

                if response.status_code == 200:
                    result = response.json()
                    investigation_id = result.get("investigation_id")
                    state["investigation_id"] = investigation_id
                    logger.info(f"✅ Sift investigation created: {investigation_id}")

                    # Run error pattern analysis
                    logger.info("🔍 Running error pattern analysis...")
                    error_payload = {
                        "tool": "sift_run_error_pattern_analysis",
                        "arguments": {"investigation_id": investigation_id},
                    }
                    error_response = await client.post(
                        f"{MCP_SERVER_URL}/mcp/tool", json=error_payload, timeout=120
                    )

                    if error_response.status_code == 200:
                        error_result = error_response.json()
                        state["error_patterns"] = error_result.get("elevated_patterns", [])
                        logger.info(f"✅ Found {len(state['error_patterns'])} elevated error patterns")
                    else:
                        logger.warning(f"⚠️  Error pattern analysis failed: {error_response.status_code}")

                    # Run slow request analysis
                    logger.info("🔍 Running slow request analysis...")
                    slow_payload = {
                        "tool": "sift_run_slow_request_analysis",
                        "arguments": {"investigation_id": investigation_id},
                    }
                    slow_response = await client.post(
                        f"{MCP_SERVER_URL}/mcp/tool", json=slow_payload, timeout=120
                    )

                    if slow_response.status_code == 200:
                        slow_result = slow_response.json()
                        state["slow_requests"] = slow_result.get("slow_operations", [])
                        logger.info(f"✅ Found {len(state['slow_requests'])} slow operations")
                    else:
                        logger.warning(f"⚠️  Slow request analysis failed: {slow_response.status_code}")

                    state["sift_investigation"] = {
                        "id": investigation_id,
                        "error_patterns": state.get("error_patterns", []),
                        "slow_requests": state.get("slow_requests", []),
                    }

                else:
                    logger.error(f"❌ Failed to create Sift investigation: {response.status_code}")

        except Exception as e:
            logger.error(f"❌ Error running Sift investigation: {e}", exc_info=True)
            state["context"]["sift_error"] = str(e)

        return state

    @traceable(name="analyze_findings")
    async def _analyze_findings_node(self, state: InvestigationState) -> InvestigationState:
        """🧠 Analyze investigation findings using LLM"""
        logger.info(f"🧠 Analyzing findings for: {state['title']}")

        # Prepare analysis context
        error_patterns = state.get("error_patterns", [])
        slow_requests = state.get("slow_requests", [])

        analysis_prompt = f"""You are an expert SRE analyzing investigation results.

Issue: {state['title']}
Description: {state['description']}

## Investigation Findings

### Error Patterns ({len(error_patterns)} found)
{self._format_error_patterns(error_patterns)}

### Slow Operations ({len(slow_requests)} found)
{self._format_slow_requests(slow_requests)}

## Your Task

Analyze the above findings and provide:
1. **Root Cause Analysis**: What is the most likely root cause?
2. **Impact Assessment**: How severe is this issue?
3. **Evidence**: What evidence supports your analysis?
4. **Related Issues**: Are there any related or cascading issues?

Be specific and technical. Reference the actual error patterns and metrics.
"""

        try:
            response = await self.llm.ainvoke([HumanMessage(content=analysis_prompt)])

            state["root_cause"] = response.content
            state["messages"] = state["messages"] + [HumanMessage(content=analysis_prompt), response]

            logger.info(f"✅ Root cause analysis completed")

        except Exception as e:
            logger.error(f"❌ Error analyzing findings: {e}", exc_info=True)
            state["root_cause"] = f"Analysis failed: {str(e)}"

        return state

    @traceable(name="generate_recommendations")
    async def _generate_recommendations_node(self, state: InvestigationState) -> InvestigationState:
        """💡 Generate actionable recommendations"""
        logger.info(f"💡 Generating recommendations for: {state['title']}")

        recommendations_prompt = f"""Based on the root cause analysis:

{state.get('root_cause', 'No analysis available')}

Generate specific, actionable recommendations:

1. **Immediate Actions** (next 5 minutes):
   - What should be done RIGHT NOW to mitigate the issue?

2. **Short-term Fixes** (next 1 hour):
   - What code/config changes are needed?
   - Provide specific file paths and changes

3. **Long-term Prevention** (next sprint):
   - What monitoring/alerting should be added?
   - What architectural changes would prevent this?

4. **Investigation Queries**:
   - Provide specific PromQL/LogQL queries for further investigation

Be extremely specific with file paths, configuration keys, and actual code snippets.
"""

        try:
            response = await self.llm.ainvoke([HumanMessage(content=recommendations_prompt)])

            # Parse recommendations
            recommendations = response.content.split("\n")
            state["recommendations"] = [rec.strip() for rec in recommendations if rec.strip()]
            state["messages"] = state["messages"] + [HumanMessage(content=recommendations_prompt), response]

            logger.info(f"✅ Generated {len(state['recommendations'])} recommendations")

        except Exception as e:
            logger.error(f"❌ Error generating recommendations: {e}", exc_info=True)
            state["recommendations"] = [f"Failed to generate recommendations: {str(e)}"]

        return state

    @traceable(name="update_github_issue")
    async def _update_github_issue_node(self, state: InvestigationState) -> InvestigationState:
        """🔄 Update GitHub issue with findings"""
        logger.info(f"🔄 Updating GitHub issue with findings")

        if not state.get("issue_number"):
            logger.warning("⚠️  No issue number - skipping GitHub update")
            return state

        # Build investigation comment
        error_patterns = state.get("error_patterns", [])
        slow_requests = state.get("slow_requests", [])

        comment = f"""## 🔍 Investigation Results

**Investigation ID**: `{state.get('investigation_id', 'N/A')}`  
**Completed**: {datetime.now().isoformat()}

### 📊 Findings

#### Error Patterns
{self._format_error_patterns_markdown(error_patterns)}

#### Slow Operations
{self._format_slow_requests_markdown(slow_requests)}

### 🧠 Root Cause Analysis

{state.get('root_cause', 'No root cause identified')}

### 💡 Recommendations

{chr(10).join(state.get('recommendations', ['No recommendations available']))}

---

### 🔗 Investigation Links

* **Sift Investigation**: `{state.get('investigation_id', 'N/A')}`
* **Grafana**: [View in Grafana]({os.getenv('GRAFANA_URL', 'http://grafana.grafana.svc.cluster.local:3000')})
* **Prometheus**: [View Metrics]({os.getenv('PROMETHEUS_URL', 'http://prometheus-k8s.prometheus.svc.cluster.local:9090')})

---

_Investigation performed by Agent-SRE using Grafana Sift and LLM analysis_
"""

        # Add comment to issue
        success = await github_client.add_comment(state["issue_number"], comment)

        if success:
            logger.info(f"✅ GitHub issue #{state['issue_number']} updated with findings")
            state["completed"] = True
        else:
            logger.error(f"❌ Failed to update GitHub issue")

        return state

    def _format_error_patterns(self, patterns: List[Dict[str, Any]]) -> str:
        """Format error patterns for LLM analysis"""
        if not patterns:
            return "No elevated error patterns detected"

        formatted = []
        for idx, pattern in enumerate(patterns[:5], 1):  # Top 5
            formatted.append(
                f"{idx}. Pattern: {pattern.get('pattern', 'Unknown')}\n"
                f"   Current Count: {pattern.get('current_count', 0)}\n"
                f"   Baseline Count: {pattern.get('baseline_count', 0)}\n"
                f"   Elevation Factor: {pattern.get('elevation_factor', 0):.2f}x\n"
                f"   Severity: {pattern.get('severity', 'unknown')}"
            )

        return "\n\n".join(formatted)

    def _format_slow_requests(self, requests: List[Dict[str, Any]]) -> str:
        """Format slow requests for LLM analysis"""
        if not requests:
            return "No slow operations detected"

        formatted = []
        for idx, req in enumerate(requests[:5], 1):  # Top 5
            formatted.append(
                f"{idx}. Operation: {req.get('operation', 'Unknown')}\n"
                f"   Current P95: {req.get('current_p95_ms', 0)}ms\n"
                f"   Baseline P95: {req.get('baseline_p95_ms', 0)}ms\n"
                f"   Slowdown Factor: {req.get('slowdown_factor', 0):.2f}x\n"
                f"   Severity: {req.get('severity', 'unknown')}"
            )

        return "\n\n".join(formatted)

    def _format_error_patterns_markdown(self, patterns: List[Dict[str, Any]]) -> str:
        """Format error patterns for GitHub markdown"""
        if not patterns:
            return "✅ No elevated error patterns detected"

        formatted = ["| Pattern | Current | Baseline | Factor | Severity |", "| --- | --- | --- | --- | --- |"]

        for pattern in patterns[:10]:  # Top 10
            formatted.append(
                f"| `{pattern.get('pattern', 'Unknown')[:50]}` | "
                f"{pattern.get('current_count', 0)} | "
                f"{pattern.get('baseline_count', 0)} | "
                f"{pattern.get('elevation_factor', 0):.2f}x | "
                f"{pattern.get('severity', 'unknown')} |"
            )

        return "\n".join(formatted)

    def _format_slow_requests_markdown(self, requests: List[Dict[str, Any]]) -> str:
        """Format slow requests for GitHub markdown"""
        if not requests:
            return "✅ No slow operations detected"

        formatted = [
            "| Operation | Current P95 | Baseline P95 | Factor | Severity |",
            "| --- | --- | --- | --- | --- |",
        ]

        for req in requests[:10]:  # Top 10
            formatted.append(
                f"| `{req.get('operation', 'Unknown')[:50]}` | "
                f"{req.get('current_p95_ms', 0)}ms | "
                f"{req.get('baseline_p95_ms', 0)}ms | "
                f"{req.get('slowdown_factor', 0):.2f}x | "
                f"{req.get('severity', 'unknown')} |"
            )

        return "\n".join(formatted)

    async def investigate(
        self,
        title: str,
        description: str,
        severity: str = "medium",
        component: str = "unknown",
        thread_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        """🚀 Run the complete investigation workflow"""

        if not thread_id:
            thread_id = f"investigation-{datetime.now().timestamp()}"

        logger.info(f"🚀 Starting investigation workflow: {title}")

        # Initialize state
        initial_state: InvestigationState = {
            "investigation_id": None,
            "issue_number": None,
            "title": title,
            "description": description,
            "severity": severity,
            "component": component,
            "messages": [],
            "context": {},
            "sift_investigation": None,
            "error_patterns": None,
            "slow_requests": None,
            "root_cause": None,
            "recommendations": [],
            "resolution_steps": [],
            "github_issue": None,
            "next_action": None,
            "completed": False,
        }

        try:
            # Execute the workflow
            config = {"configurable": {"thread_id": thread_id}}
            final_state = await self.graph.ainvoke(initial_state, config)

            logger.info(f"✅ Investigation workflow completed: {title}")

            return {
                "investigation_id": final_state.get("investigation_id"),
                "issue_number": final_state.get("issue_number"),
                "issue_url": final_state.get("github_issue").html_url
                if final_state.get("github_issue")
                else None,
                "root_cause": final_state.get("root_cause"),
                "recommendations": final_state.get("recommendations"),
                "error_patterns": len(final_state.get("error_patterns", [])),
                "slow_requests": len(final_state.get("slow_requests", [])),
                "completed": final_state.get("completed", False),
            }

        except Exception as e:
            logger.error(f"❌ Investigation workflow failed: {e}", exc_info=True)
            return {
                "error": str(e),
                "completed": False,
            }


# Global workflow instance
investigation_workflow = AutomatedInvestigationWorkflow()

__all__ = ["AutomatedInvestigationWorkflow", "InvestigationState", "investigation_workflow"]

