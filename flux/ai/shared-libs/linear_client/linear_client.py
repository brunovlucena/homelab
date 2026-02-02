"""
Linear API Client for Homelab Agents

A simple client for agents to interact with Linear's GraphQL API.
"""
import os
import httpx
from typing import Optional, Dict, Any, List
from datetime import datetime


class LinearClient:
    """Linear API client for agents."""
    
    def __init__(self, api_key: Optional[str] = None):
        """
        Initialize Linear client.
        
        Args:
            api_key: Linear API key. If not provided, reads from LINEAR_API_KEY env var.
        """
        self.api_key = api_key or os.getenv("LINEAR_API_KEY")
        if not self.api_key:
            raise ValueError("Linear API key required. Set LINEAR_API_KEY env var or pass api_key parameter.")
        
        self.base_url = "https://api.linear.app/graphql"
        self.headers = {
            "Authorization": self.api_key,
            "Content-Type": "application/json",
        }
    
    async def _query(self, query: str, variables: Optional[Dict] = None) -> Dict[str, Any]:
        """Execute a GraphQL query."""
        async with httpx.AsyncClient(timeout=30.0) as client:
            response = await client.post(
                self.base_url,
                headers=self.headers,
                json={"query": query, "variables": variables or {}},
            )
            data = response.json()
            
            if "errors" in data:
                raise Exception(f"Linear API error: {data['errors']}")
            
            response.raise_for_status()
            
            return data
    
    async def list_issues(
        self,
        team: Optional[str] = None,
        assignee: Optional[str] = None,
        state: Optional[str] = None,
        limit: int = 50
    ) -> List[Dict[str, Any]]:
        """
        List Linear issues.
        
        Args:
            team: Team name to filter by
            assignee: Assignee name or "me" for current user
            state: State name (e.g., "In Progress", "Done")
            limit: Maximum number of issues to return
            
        Returns:
            List of issue dictionaries
        """
        query = """
        query($team: String, $assignee: String, $state: String, $first: Int) {
          issues(
            filter: {
              team: { name: { eq: $team } }
              assignee: { name: { eq: $assignee } }
              state: { name: { eq: $state } }
            }
            first: $first
          ) {
            nodes {
              id
              identifier
              title
              description
              state {
                name
                type
              }
              assignee {
                name
                email
              }
              team {
                name
                key
              }
              priority
              labels {
                nodes {
                  name
                  color
                }
              }
              createdAt
              updatedAt
              url
            }
          }
        }
        """
        variables = {
            "team": team,
            "assignee": assignee,
            "state": state,
            "first": limit,
        }
        result = await self._query(query, variables)
        return result["data"]["issues"]["nodes"]
    
    async def get_issue(self, issue_id: str) -> Dict[str, Any]:
        """
        Get a specific issue by ID.
        
        Args:
            issue_id: Linear issue ID (e.g., "LIN-123" or UUID)
            
        Returns:
            Issue dictionary
        """
        query = """
        query($id: String!) {
          issue(id: $id) {
            id
            identifier
            title
            description
            state {
              name
              type
            }
            assignee {
              name
              email
            }
            team {
              name
              key
            }
            priority
            labels {
              nodes {
                name
                color
              }
            }
            createdAt
            updatedAt
            url
            comments {
              nodes {
                body
                createdAt
                user {
                  name
                }
              }
            }
          }
        }
        """
        result = await self._query(query, {"id": issue_id})
        return result["data"]["issue"]
    
    async def get_project(self, project_name: str) -> Dict[str, Any]:
        """
        Get project by name.
        
        Args:
            project_name: Project name
            
        Returns:
            Project dictionary
        """
        query = """
        query($name: String!) {
          projects(filter: { name: { eq: $name } }) {
            nodes {
              id
              name
              description
              state
            }
          }
        }
        """
        result = await self._query(query, {"name": project_name})
        projects = result["data"]["projects"]["nodes"]
        if not projects:
            raise ValueError(f"Project with name '{project_name}' not found")
        return projects[0]
    
    async def create_issue(
        self,
        title: str,
        description: Optional[str] = None,
        team_id: Optional[str] = None,
        assignee_id: Optional[str] = None,
        priority: Optional[int] = None,
        state_id: Optional[str] = None,
        label_ids: Optional[List[str]] = None,
        project_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Create a Linear issue.
        
        Args:
            title: Issue title
            description: Issue description (markdown supported)
            team_id: Team ID (required if team_key not provided)
            assignee_id: Assignee user ID
            priority: Priority (0=No priority, 1=Urgent, 2=High, 3=Normal, 4=Low)
            state_id: Initial state ID
            label_ids: List of label IDs
            project_id: Project ID to add issue to
            
        Returns:
            Created issue dictionary
        """
        mutation = """
        mutation($input: IssueCreateInput!) {
          issueCreate(input: $input) {
            success
            issue {
              id
              identifier
              title
              description
              url
              state {
                name
              }
              team {
                name
              }
            }
          }
        }
        """
        input_data = {"title": title}
        
        if description:
            input_data["description"] = description
        if team_id:
            input_data["teamId"] = team_id
        if assignee_id:
            input_data["assigneeId"] = assignee_id
        if priority is not None:
            input_data["priority"] = priority
        if state_id:
            input_data["stateId"] = state_id
        if label_ids:
            input_data["labelIds"] = label_ids
        if project_id:
            input_data["projectId"] = project_id
        
        variables = {"input": input_data}
        result = await self._query(mutation, variables)
        
        if not result["data"]["issueCreate"]["success"]:
            raise Exception("Failed to create issue")
        
        return result["data"]["issueCreate"]["issue"]
    
    async def update_issue(
        self,
        issue_id: str,
        title: Optional[str] = None,
        description: Optional[str] = None,
        assignee_id: Optional[str] = None,
        state_id: Optional[str] = None,
        priority: Optional[int] = None
    ) -> Dict[str, Any]:
        """
        Update a Linear issue.
        
        Args:
            issue_id: Issue ID to update
            title: New title
            description: New description
            assignee_id: New assignee ID
            state_id: New state ID
            priority: New priority
            
        Returns:
            Updated issue dictionary
        """
        mutation = """
        mutation($id: String!, $input: IssueUpdateInput!) {
          issueUpdate(id: $id, input: $input) {
            success
            issue {
              id
              identifier
              title
              description
              url
              state {
                name
              }
            }
          }
        }
        """
        input_data = {}
        
        if title:
            input_data["title"] = title
        if description:
            input_data["description"] = description
        if assignee_id:
            input_data["assigneeId"] = assignee_id
        if state_id:
            input_data["stateId"] = state_id
        if priority is not None:
            input_data["priority"] = priority
        
        variables = {"id": issue_id, "input": input_data}
        result = await self._query(mutation, variables)
        
        if not result["data"]["issueUpdate"]["success"]:
            raise Exception("Failed to update issue")
        
        return result["data"]["issueUpdate"]["issue"]
    
    async def create_comment(self, issue_id: str, body: str) -> Dict[str, Any]:
        """
        Create a comment on an issue.
        
        Args:
            issue_id: Issue ID
            body: Comment body (markdown supported)
            
        Returns:
            Created comment dictionary
        """
        mutation = """
        mutation($issueId: String!, $body: String!) {
          commentCreate(input: { issueId: $issueId, body: $body }) {
            success
            comment {
              id
              body
              createdAt
              user {
                name
              }
            }
          }
        }
        """
        result = await self._query(mutation, {"issueId": issue_id, "body": body})
        
        if not result["data"]["commentCreate"]["success"]:
            raise Exception("Failed to create comment")
        
        return result["data"]["commentCreate"]["comment"]
    
    async def list_teams(self) -> List[Dict[str, Any]]:
        """List all teams in the workspace."""
        query = """
        query {
          teams {
            nodes {
              id
              name
              key
              description
            }
          }
        }
        """
        result = await self._query(query)
        return result["data"]["teams"]["nodes"]
    
    async def get_team(self, team_key: str) -> Dict[str, Any]:
        """
        Get team by key (e.g., "ENG", "SRE").
        
        Args:
            team_key: Team key identifier
            
        Returns:
            Team dictionary
        """
        # Linear API requires team ID, not key, so we filter teams by key
        query = """
        query($key: String!) {
          teams(filter: { key: { eq: $key } }) {
            nodes {
              id
              name
              key
              description
            }
          }
        }
        """
        result = await self._query(query, {"key": team_key})
        teams = result["data"]["teams"]["nodes"]
        if not teams:
            raise ValueError(f"Team with key '{team_key}' not found")
        return teams[0]


# Convenience function for quick usage
async def create_agent_issue(
    title: str,
    description: str,
    team_key: str,
    agent_name: str,
    priority: int = 3
) -> str:
    """
    Quick helper to create an issue from an agent.
    
    Args:
        title: Issue title
        description: Issue description
        team_key: Team key (e.g., "SRE", "ENG")
        agent_name: Name of the agent creating the issue
        priority: Priority level (default: 3 = Normal)
        
    Returns:
        Issue URL
    """
    client = LinearClient()
    
    # Get team ID
    team = await client.get_team(team_key)
    team_id = team["id"]
    
    # Create issue with agent attribution
    full_description = f"""
{description}

---
*Created by {agent_name}*
"""
    issue = await client.create_issue(
        title=title,
        description=full_description,
        team_id=team_id,
        priority=priority
    )
    
    return issue["url"]

