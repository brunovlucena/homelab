"""
Jira API Client for Homelab Agents

A simple client for agents to interact with Jira's REST API.
"""
import os
import base64
import httpx
from typing import Optional, Dict, Any, List
from datetime import datetime


class JiraClient:
    """Jira API client for agents."""
    
    def __init__(
        self,
        url: Optional[str] = None,
        email: Optional[str] = None,
        api_token: Optional[str] = None
    ):
        """
        Initialize Jira client.
        
        Args:
            url: Jira instance URL (e.g., "https://your-domain.atlassian.net")
            email: Jira user email
            api_token: Jira API token. If not provided, reads from JIRA_API_TOKEN env var.
        """
        jira_url = url or os.getenv("JIRA_URL")
        self.url = jira_url.rstrip("/") if jira_url else None
        self.email = email or os.getenv("JIRA_EMAIL")
        self.api_token = api_token or os.getenv("JIRA_API_TOKEN")
        
        if not self.url:
            raise ValueError("Jira URL required. Set JIRA_URL env var or pass url parameter.")
        if not self.email:
            raise ValueError("Jira email required. Set JIRA_EMAIL env var or pass email parameter.")
        if not self.api_token:
            raise ValueError("Jira API token required. Set JIRA_API_TOKEN env var or pass api_token parameter.")
        
        # Create Basic Auth header (email:api_token base64 encoded)
        credentials = f"{self.email}:{self.api_token}"
        encoded_credentials = base64.b64encode(credentials.encode()).decode()
        
        self.base_url = f"{self.url}/rest/api/3"
        self.headers = {
            "Authorization": f"Basic {encoded_credentials}",
            "Content-Type": "application/json",
            "Accept": "application/json",
        }
    
    async def _request(
        self,
        method: str,
        endpoint: str,
        data: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Execute a REST API request."""
        url = f"{self.base_url}/{endpoint.lstrip('/')}"
        
        async with httpx.AsyncClient(timeout=30.0) as client:
            response = await client.request(
                method=method,
                url=url,
                headers=self.headers,
                json=data,
                params=params,
            )
            
            response.raise_for_status()
            
            # Jira returns empty body for some operations (like DELETE)
            if response.status_code == 204:
                return {}
            
            return response.json()
    
    async def get_issue(self, issue_key: str) -> Dict[str, Any]:
        """
        Get a specific issue by key.
        
        Args:
            issue_key: Jira issue key (e.g., "PROJ-123")
            
        Returns:
            Issue dictionary
        """
        return await self._request("GET", f"issue/{issue_key}")
    
    async def create_issue(
        self,
        project_key: str,
        summary: str,
        description: Optional[str] = None,
        issue_type: str = "Task",
        priority: Optional[str] = None,
        labels: Optional[List[str]] = None,
        assignee: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Create a Jira issue.
        
        Args:
            project_key: Project key (e.g., "PROJ")
            summary: Issue summary/title
            description: Issue description
            issue_type: Issue type (e.g., "Task", "Bug", "Story")
            priority: Priority name (e.g., "Highest", "High", "Medium", "Low", "Lowest")
            labels: List of label names
            assignee: Assignee account ID (optional)
            
        Returns:
            Created issue dictionary
        """
        fields: Dict[str, Any] = {
            "project": {"key": project_key},
            "summary": summary,
            "issuetype": {"name": issue_type},
        }
        
        if description:
            fields["description"] = {
                "type": "doc",
                "version": 1,
                "content": [
                    {
                        "type": "paragraph",
                        "content": [
                            {
                                "type": "text",
                                "text": description
                            }
                        ]
                    }
                ]
            }
        
        if priority:
            fields["priority"] = {"name": priority}
        
        if labels:
            fields["labels"] = labels
        
        if assignee:
            fields["assignee"] = {"accountId": assignee}
        
        data = {"fields": fields}
        
        return await self._request("POST", "issue", data=data)
    
    async def update_issue(
        self,
        issue_key: str,
        summary: Optional[str] = None,
        description: Optional[str] = None,
        priority: Optional[str] = None,
        assignee: Optional[str] = None,
        labels: Optional[List[str]] = None
    ) -> Dict[str, Any]:
        """
        Update a Jira issue.
        
        Args:
            issue_key: Issue key to update
            summary: New summary
            description: New description
            priority: New priority name
            assignee: New assignee account ID
            labels: New labels list
            
        Returns:
            Updated issue dictionary
        """
        fields: Dict[str, Any] = {}
        
        if summary:
            fields["summary"] = summary
        
        if description:
            fields["description"] = {
                "type": "doc",
                "version": 1,
                "content": [
                    {
                        "type": "paragraph",
                        "content": [
                            {
                                "type": "text",
                                "text": description
                            }
                        ]
                    }
                ]
            }
        
        if priority:
            fields["priority"] = {"name": priority}
        
        if assignee:
            fields["assignee"] = {"accountId": assignee}
        
        if labels:
            fields["labels"] = labels
        
        data = {"fields": fields}
        
        return await self._request("PUT", f"issue/{issue_key}", data=data)
    
    async def add_comment(self, issue_key: str, body: str) -> Dict[str, Any]:
        """
        Add a comment to an issue.
        
        Args:
            issue_key: Issue key
            body: Comment body (plain text or markdown)
            
        Returns:
            Created comment dictionary
        """
        data = {
            "body": {
                "type": "doc",
                "version": 1,
                "content": [
                    {
                        "type": "paragraph",
                        "content": [
                            {
                                "type": "text",
                                "text": body
                            }
                        ]
                    }
                ]
            }
        }
        
        return await self._request("POST", f"issue/{issue_key}/comment", data=data)
    
    async def search_issues(
        self,
        jql: str,
        fields: Optional[List[str]] = None,
        max_results: int = 50
    ) -> List[Dict[str, Any]]:
        """
        Search for issues using JQL (Jira Query Language).
        
        Args:
            jql: JQL query string
            fields: List of fields to return (default: key, summary, status)
            max_results: Maximum number of results
            
        Returns:
            List of issue dictionaries
        """
        if fields is None:
            fields = ["key", "summary", "status", "priority", "assignee", "created", "updated"]
        
        params = {
            "jql": jql,
            "fields": ",".join(fields),
            "maxResults": max_results,
        }
        
        result = await self._request("GET", "search", params=params)
        return result.get("issues", [])
    
    async def get_project(self, project_key: str) -> Dict[str, Any]:
        """
        Get project information.
        
        Args:
            project_key: Project key
            
        Returns:
            Project dictionary
        """
        return await self._request("GET", f"project/{project_key}")
    
    async def list_projects(self) -> List[Dict[str, Any]]:
        """List all accessible projects."""
        result = await self._request("GET", "project")
        return result if isinstance(result, list) else []


# Convenience function for quick usage
async def create_agent_issue(
    project_key: str,
    summary: str,
    description: str,
    agent_name: str,
    issue_type: str = "Task",
    priority: str = "Medium"
) -> str:
    """
    Quick helper to create an issue from an agent.
    
    Args:
        project_key: Jira project key
        summary: Issue summary
        description: Issue description
        agent_name: Name of the agent creating the issue
        issue_type: Issue type (default: "Task")
        priority: Priority level (default: "Medium")
        
    Returns:
        Issue key (e.g., "PROJ-123")
    """
    client = JiraClient()
    
    # Add agent attribution to description
    full_description = f"{description}\n\n---\n*Created by {agent_name}*"
    
    issue = await client.create_issue(
        project_key=project_key,
        summary=summary,
        description=full_description,
        issue_type=issue_type,
        priority=priority
    )
    
    return issue["key"]

