"""
Test suite for Linear Client

Run tests:
    pytest test_linear_client.py -v

Run with coverage:
    pytest test_linear_client.py --cov=linear_client --cov-report=html
"""
import pytest
import httpx
from unittest.mock import AsyncMock, patch, MagicMock
from linear_client import LinearClient, create_agent_issue


class TestLinearClient:
    """Test LinearClient class."""
    
    @pytest.fixture
    def client(self):
        """Create a LinearClient instance with test API key."""
        return LinearClient(api_key="test_api_key_123")
    
    @pytest.fixture
    def mock_response(self):
        """Create a mock HTTP response."""
        response = MagicMock()
        response.json.return_value = {
            "data": {
                "issues": {
                    "nodes": [
                        {
                            "id": "issue-1",
                            "identifier": "LIN-123",
                            "title": "Test Issue",
                            "state": {"name": "In Progress"},
                            "url": "https://linear.app/issue/LIN-123"
                        }
                    ]
                }
            }
        }
        response.raise_for_status = MagicMock()
        return response
    
    @pytest.mark.asyncio
    async def test_list_issues_success(self, client, mock_response):
        """Test successful issue listing."""
        with patch("httpx.AsyncClient") as mock_client:
            mock_client.return_value.__aenter__.return_value.post = AsyncMock(
                return_value=mock_response
            )
            
            issues = await client.list_issues(team="SRE", limit=10)
            
            assert len(issues) == 1
            assert issues[0]["identifier"] == "LIN-123"
            assert issues[0]["title"] == "Test Issue"
    
    @pytest.mark.asyncio
    async def test_list_issues_with_filters(self, client, mock_response):
        """Test issue listing with filters."""
        with patch("httpx.AsyncClient") as mock_client:
            mock_client.return_value.__aenter__.return_value.post = AsyncMock(
                return_value=mock_response
            )
            
            issues = await client.list_issues(
                team="SRE",
                assignee="me",
                state="In Progress"
            )
            
            assert len(issues) == 1
    
    @pytest.mark.asyncio
    async def test_get_issue_success(self, client):
        """Test getting a specific issue."""
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "data": {
                "issue": {
                    "id": "issue-1",
                    "identifier": "LIN-123",
                    "title": "Test Issue",
                    "description": "Test description",
                    "state": {"name": "In Progress", "type": "started"},
                    "url": "https://linear.app/issue/LIN-123"
                }
            }
        }
        mock_response.raise_for_status = MagicMock()
        
        with patch("httpx.AsyncClient") as mock_client:
            mock_client.return_value.__aenter__.return_value.post = AsyncMock(
                return_value=mock_response
            )
            
            issue = await client.get_issue("LIN-123")
            
            assert issue["identifier"] == "LIN-123"
            assert issue["title"] == "Test Issue"
    
    @pytest.mark.asyncio
    async def test_create_issue_success(self, client):
        """Test creating an issue."""
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "data": {
                "issueCreate": {
                    "success": True,
                    "issue": {
                        "id": "issue-1",
                        "identifier": "LIN-124",
                        "title": "New Issue",
                        "url": "https://linear.app/issue/LIN-124",
                        "state": {"name": "Todo"},
                        "team": {"name": "SRE"}
                    }
                }
            }
        }
        mock_response.raise_for_status = MagicMock()
        
        with patch("httpx.AsyncClient") as mock_client:
            mock_client.return_value.__aenter__.return_value.post = AsyncMock(
                return_value=mock_response
            )
            
            issue = await client.create_issue(
                title="New Issue",
                description="Test description",
                team_id="team-1",
                priority=2
            )
            
            assert issue["identifier"] == "LIN-124"
            assert issue["title"] == "New Issue"
            assert issue["url"] == "https://linear.app/issue/LIN-124"
    
    @pytest.mark.asyncio
    async def test_create_issue_failure(self, client):
        """Test issue creation failure."""
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "data": {
                "issueCreate": {
                    "success": False
                }
            }
        }
        mock_response.raise_for_status = MagicMock()
        
        with patch("httpx.AsyncClient") as mock_client:
            mock_client.return_value.__aenter__.return_value.post = AsyncMock(
                return_value=mock_response
            )
            
            with pytest.raises(Exception, match="Failed to create issue"):
                await client.create_issue(
                    title="New Issue",
                    team_id="team-1"
                )
    
    @pytest.mark.asyncio
    async def test_update_issue_success(self, client):
        """Test updating an issue."""
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "data": {
                "issueUpdate": {
                    "success": True,
                    "issue": {
                        "id": "issue-1",
                        "identifier": "LIN-123",
                        "title": "Updated Issue",
                        "url": "https://linear.app/issue/LIN-123",
                        "state": {"name": "In Progress"}
                    }
                }
            }
        }
        mock_response.raise_for_status = MagicMock()
        
        with patch("httpx.AsyncClient") as mock_client:
            mock_client.return_value.__aenter__.return_value.post = AsyncMock(
                return_value=mock_response
            )
            
            issue = await client.update_issue(
                issue_id="LIN-123",
                title="Updated Issue",
                state_id="state-1"
            )
            
            assert issue["title"] == "Updated Issue"
    
    @pytest.mark.asyncio
    async def test_create_comment_success(self, client):
        """Test creating a comment."""
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "data": {
                "commentCreate": {
                    "success": True,
                    "comment": {
                        "id": "comment-1",
                        "body": "Test comment",
                        "createdAt": "2024-01-01T00:00:00Z",
                        "user": {"name": "Test User"}
                    }
                }
            }
        }
        mock_response.raise_for_status = MagicMock()
        
        with patch("httpx.AsyncClient") as mock_client:
            mock_client.return_value.__aenter__.return_value.post = AsyncMock(
                return_value=mock_response
            )
            
            comment = await client.create_comment(
                issue_id="LIN-123",
                body="Test comment"
            )
            
            assert comment["body"] == "Test comment"
            assert comment["user"]["name"] == "Test User"
    
    @pytest.mark.asyncio
    async def test_list_teams_success(self, client):
        """Test listing teams."""
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "data": {
                "teams": {
                    "nodes": [
                        {
                            "id": "team-1",
                            "name": "SRE",
                            "key": "SRE",
                            "description": "Site Reliability Engineering"
                        },
                        {
                            "id": "team-2",
                            "name": "Engineering",
                            "key": "ENG",
                            "description": "Engineering Team"
                        }
                    ]
                }
            }
        }
        mock_response.raise_for_status = MagicMock()
        
        with patch("httpx.AsyncClient") as mock_client:
            mock_client.return_value.__aenter__.return_value.post = AsyncMock(
                return_value=mock_response
            )
            
            teams = await client.list_teams()
            
            assert len(teams) == 2
            assert teams[0]["name"] == "SRE"
            assert teams[1]["name"] == "Engineering"
    
    @pytest.mark.asyncio
    async def test_get_team_success(self, client):
        """Test getting a team by key."""
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "data": {
                "team": {
                    "id": "team-1",
                    "name": "SRE",
                    "key": "SRE",
                    "description": "Site Reliability Engineering"
                }
            }
        }
        mock_response.raise_for_status = MagicMock()
        
        with patch("httpx.AsyncClient") as mock_client:
            mock_client.return_value.__aenter__.return_value.post = AsyncMock(
                return_value=mock_response
            )
            
            team = await client.get_team("SRE")
            
            assert team["name"] == "SRE"
            assert team["key"] == "SRE"
    
    @pytest.mark.asyncio
    async def test_api_error_handling(self, client):
        """Test API error handling."""
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "errors": [
                {"message": "Invalid API key"}
            ]
        }
        mock_response.raise_for_status = MagicMock()
        
        with patch("httpx.AsyncClient") as mock_client:
            mock_client.return_value.__aenter__.return_value.post = AsyncMock(
                return_value=mock_response
            )
            
            with pytest.raises(Exception, match="Linear API error"):
                await client.list_issues()
    
    def test_missing_api_key(self):
        """Test that missing API key raises error."""
        with patch.dict("os.environ", {}, clear=True):
            with pytest.raises(ValueError, match="Linear API key required"):
                LinearClient()


class TestHelperFunctions:
    """Test helper functions."""
    
    @pytest.mark.asyncio
    async def test_create_agent_issue(self):
        """Test create_agent_issue helper function."""
        # Mock team lookup
        team_response = MagicMock()
        team_response.json.return_value = {
            "data": {
                "team": {
                    "id": "team-1",
                    "name": "SRE",
                    "key": "SRE"
                }
            }
        }
        team_response.raise_for_status = MagicMock()
        
        # Mock issue creation
        issue_response = MagicMock()
        issue_response.json.return_value = {
            "data": {
                "issueCreate": {
                    "success": True,
                    "issue": {
                        "id": "issue-1",
                        "identifier": "LIN-125",
                        "title": "Test Issue",
                        "url": "https://linear.app/issue/LIN-125"
                    }
                }
            }
        }
        issue_response.raise_for_status = MagicMock()
        
        with patch("httpx.AsyncClient") as mock_client, \
             patch.dict("os.environ", {"LINEAR_API_KEY": "test_api_key"}):
            mock_post = AsyncMock(side_effect=[team_response, issue_response])
            mock_client.return_value.__aenter__.return_value.post = mock_post
            
            url = await create_agent_issue(
                title="Test Issue",
                description="Test description",
                team_key="SRE",
                agent_name="agent-sre",
                priority=2
            )
            
            assert url == "https://linear.app/issue/LIN-125"
            assert mock_post.call_count == 2


if __name__ == "__main__":
    pytest.main([__file__, "-v"])

