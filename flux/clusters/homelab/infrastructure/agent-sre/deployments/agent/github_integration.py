#!/usr/bin/env python3
"""
GitHub Integration for Agent-SRE
Handles automated issue creation, updates, and investigation management
"""

import logging
import os
from datetime import datetime
from typing import Any, Dict, List, Optional

import httpx
from pydantic import BaseModel

logger = logging.getLogger(__name__)


class GitHubIssue(BaseModel):
    """GitHub issue model"""

    number: int
    title: str
    body: str
    state: str
    labels: List[str]
    assignees: List[str]
    html_url: str
    created_at: str


class GitHubClient:
    """🐙 GitHub API client for issue management"""

    def __init__(self):
        self.token = os.getenv("GITHUB_TOKEN")
        self.owner = os.getenv("GITHUB_OWNER", "brunovlucena")
        self.repo = os.getenv("GITHUB_REPO", "homelab")
        self.base_url = "https://api.github.com"

        if not self.token:
            logger.warning("⚠️  GITHUB_TOKEN not set - GitHub integration will be limited")

        self.headers = {
            "Authorization": f"Bearer {self.token}",
            "Accept": "application/vnd.github.v3+json",
            "X-GitHub-Api-Version": "2022-11-28",
        }

    async def create_issue(
        self,
        title: str,
        body: str,
        labels: Optional[List[str]] = None,
        assignees: Optional[List[str]] = None,
    ) -> Optional[GitHubIssue]:
        """📝 Create a new GitHub issue"""
        if not self.token:
            logger.error("❌ Cannot create issue - GITHUB_TOKEN not set")
            return None

        url = f"{self.base_url}/repos/{self.owner}/{self.repo}/issues"

        payload = {
            "title": title,
            "body": body,
            "labels": labels or [],
            "assignees": assignees or [],
        }

        try:
            logger.info(f"📝 Creating GitHub issue: {title}")
            async with httpx.AsyncClient() as client:
                response = await client.post(url, headers=self.headers, json=payload, timeout=30)
                response.raise_for_status()
                data = response.json()

                issue = GitHubIssue(
                    number=data["number"],
                    title=data["title"],
                    body=data["body"],
                    state=data["state"],
                    labels=[label["name"] for label in data.get("labels", [])],
                    assignees=[assignee["login"] for assignee in data.get("assignees", [])],
                    html_url=data["html_url"],
                    created_at=data["created_at"],
                )

                logger.info(f"✅ Created issue #{issue.number}: {issue.html_url}")
                return issue

        except Exception as e:
            logger.error(f"❌ Failed to create GitHub issue: {e}", exc_info=True)
            return None

    async def update_issue(
        self,
        issue_number: int,
        title: Optional[str] = None,
        body: Optional[str] = None,
        state: Optional[str] = None,
        labels: Optional[List[str]] = None,
    ) -> Optional[GitHubIssue]:
        """🔄 Update an existing GitHub issue"""
        if not self.token:
            logger.error("❌ Cannot update issue - GITHUB_TOKEN not set")
            return None

        url = f"{self.base_url}/repos/{self.owner}/{self.repo}/issues/{issue_number}"

        payload = {}
        if title:
            payload["title"] = title
        if body:
            payload["body"] = body
        if state:
            payload["state"] = state
        if labels is not None:
            payload["labels"] = labels

        try:
            logger.info(f"🔄 Updating GitHub issue #{issue_number}")
            async with httpx.AsyncClient() as client:
                response = await client.patch(url, headers=self.headers, json=payload, timeout=30)
                response.raise_for_status()
                data = response.json()

                issue = GitHubIssue(
                    number=data["number"],
                    title=data["title"],
                    body=data["body"],
                    state=data["state"],
                    labels=[label["name"] for label in data.get("labels", [])],
                    assignees=[assignee["login"] for assignee in data.get("assignees", [])],
                    html_url=data["html_url"],
                    created_at=data["created_at"],
                )

                logger.info(f"✅ Updated issue #{issue.number}")
                return issue

        except Exception as e:
            logger.error(f"❌ Failed to update GitHub issue: {e}", exc_info=True)
            return None

    async def add_comment(self, issue_number: int, comment: str) -> bool:
        """💬 Add a comment to a GitHub issue"""
        if not self.token:
            logger.error("❌ Cannot add comment - GITHUB_TOKEN not set")
            return False

        url = f"{self.base_url}/repos/{self.owner}/{self.repo}/issues/{issue_number}/comments"

        payload = {"body": comment}

        try:
            logger.info(f"💬 Adding comment to issue #{issue_number}")
            async with httpx.AsyncClient() as client:
                response = await client.post(url, headers=self.headers, json=payload, timeout=30)
                response.raise_for_status()
                logger.info(f"✅ Comment added to issue #{issue_number}")
                return True

        except Exception as e:
            logger.error(f"❌ Failed to add comment: {e}", exc_info=True)
            return False

    async def get_issue(self, issue_number: int) -> Optional[GitHubIssue]:
        """🔍 Get a GitHub issue"""
        if not self.token:
            logger.error("❌ Cannot get issue - GITHUB_TOKEN not set")
            return None

        url = f"{self.base_url}/repos/{self.owner}/{self.repo}/issues/{issue_number}"

        try:
            async with httpx.AsyncClient() as client:
                response = await client.get(url, headers=self.headers, timeout=30)
                response.raise_for_status()
                data = response.json()

                return GitHubIssue(
                    number=data["number"],
                    title=data["title"],
                    body=data["body"],
                    state=data["state"],
                    labels=[label["name"] for label in data.get("labels", [])],
                    assignees=[assignee["login"] for assignee in data.get("assignees", [])],
                    html_url=data["html_url"],
                    created_at=data["created_at"],
                )

        except Exception as e:
            logger.error(f"❌ Failed to get GitHub issue: {e}", exc_info=True)
            return None

    async def search_issues(
        self, query: str, state: str = "open", labels: Optional[List[str]] = None
    ) -> List[GitHubIssue]:
        """🔎 Search for GitHub issues"""
        if not self.token:
            logger.error("❌ Cannot search issues - GITHUB_TOKEN not set")
            return []

        # Build search query
        search_query = f"repo:{self.owner}/{self.repo} {query} is:issue state:{state}"
        if labels:
            for label in labels:
                search_query += f" label:{label}"

        url = f"{self.base_url}/search/issues"
        params = {"q": search_query, "per_page": 10}

        try:
            async with httpx.AsyncClient() as client:
                response = await client.get(url, headers=self.headers, params=params, timeout=30)
                response.raise_for_status()
                data = response.json()

                issues = []
                for item in data.get("items", []):
                    issues.append(
                        GitHubIssue(
                            number=item["number"],
                            title=item["title"],
                            body=item.get("body", ""),
                            state=item["state"],
                            labels=[label["name"] for label in item.get("labels", [])],
                            assignees=[assignee["login"] for assignee in item.get("assignees", [])],
                            html_url=item["html_url"],
                            created_at=item["created_at"],
                        )
                    )

                logger.info(f"🔎 Found {len(issues)} issues matching query: {query}")
                return issues

        except Exception as e:
            logger.error(f"❌ Failed to search GitHub issues: {e}", exc_info=True)
            return []

    async def close_issue(self, issue_number: int, comment: Optional[str] = None) -> bool:
        """🔒 Close a GitHub issue"""
        if comment:
            await self.add_comment(issue_number, comment)

        issue = await self.update_issue(issue_number, state="closed")
        return issue is not None


# Global GitHub client instance
github_client = GitHubClient()

__all__ = ["GitHubClient", "GitHubIssue", "github_client"]



