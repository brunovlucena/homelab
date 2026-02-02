#!/usr/bin/env python3
"""
Example usage of Linear Client

This script demonstrates how to use the Linear client in your agents.
"""
import asyncio
import os
from linear_client import LinearClient, create_agent_issue


async def example_basic_usage():
    """Basic usage examples."""
    print("=" * 60)
    print("Linear Client - Basic Usage Examples")
    print("=" * 60)
    print()
    
    # Initialize client (reads LINEAR_API_KEY from environment)
    api_key = os.getenv("LINEAR_API_KEY")
    if not api_key:
        print("⚠️  LINEAR_API_KEY not set. Using example code only.\n")
        print("To run with real API:")
        print("  export LINEAR_API_KEY=lin_api_xxxxxxxxxxxxx")
        print("  python example_usage.py\n")
        return
    
    client = LinearClient(api_key=api_key)
    
    # Example 1: List teams
    print("📋 Example 1: Listing teams...")
    teams = await client.list_teams()
    print(f"   Found {len(teams)} teams")
    for team in teams[:3]:
        print(f"   - {team['name']} ({team['key']})")
    print()
    
    if not teams:
        print("   No teams found. Exiting.")
        return
    
    # Example 2: List issues for a team
    team_key = teams[0]["key"]
    print(f"📋 Example 2: Listing issues for team '{team_key}'...")
    issues = await client.list_issues(team=team_key, limit=5)
    print(f"   Found {len(issues)} issues")
    for issue in issues[:3]:
        print(f"   - {issue['identifier']}: {issue['title']}")
        print(f"     State: {issue['state']['name']}")
    print()
    
    # Example 3: Get a specific issue
    if issues:
        issue_id = issues[0]["id"]
        print(f"📋 Example 3: Getting issue details...")
        issue = await client.get_issue(issue_id)
        print(f"   {issue['identifier']}: {issue['title']}")
        print(f"   URL: {issue['url']}")
        print()
    
    # Example 4: Create an issue (commented out to avoid spam)
    print("📋 Example 4: Creating an issue (commented out)...")
    print("   Uncomment the code below to create a real issue:")
    print()
    print("   issue = await client.create_issue(")
    print("       title='Example Issue',")
    print("       description='This is an example issue',")
    print(f"       team_id=teams[0]['id'],")
    print("       priority=4  # Low priority")
    print("   )")
    print("   print(f\"Created: {issue['url']}\")")
    print()
    
    # Example 5: Using the helper function
    print("📋 Example 5: Using create_agent_issue helper...")
    print("   url = await create_agent_issue(")
    print("       title='Agent Example Issue',")
    print("       description='Created by example script',")
    print(f"       team_key='{team_key}',")
    print("       agent_name='example-script',")
    print("       priority=4")
    print("   )")
    print()


async def example_agent_integration():
    """Example of how an agent would use the client."""
    print("=" * 60)
    print("Linear Client - Agent Integration Example")
    print("=" * 60)
    print()
    
    api_key = os.getenv("LINEAR_API_KEY")
    if not api_key:
        print("⚠️  LINEAR_API_KEY not set. Showing example code only.\n")
        return
    
    client = LinearClient(api_key=api_key)
    
    # Simulate an agent detecting an issue
    print("🤖 Agent detects high CPU usage...")
    print("   Creating Linear issue to track...")
    print()
    
    # Get SRE team
    try:
        sre_team = await client.get_team("SRE")
        team_id = sre_team["id"]
    except:
        # Fallback to first team
        teams = await client.list_teams()
        if teams:
            team_id = teams[0]["id"]
        else:
            print("   ❌ No teams found")
            return
    
    # Create issue
    issue = await client.create_issue(
        title="[Alert] High CPU Usage Detected",
        description="""
**Alert Details:**
- CPU usage: 95%
- Duration: 5 minutes
- Node: worker-1

**Action Required:**
Investigate high CPU usage on worker node.

---
*Created by agent-sre*
        """.strip(),
        team_id=team_id,
        priority=2  # High priority
    )
    
    print(f"   ✅ Issue created: {issue['identifier']}")
    print(f"   🔗 URL: {issue['url']}")
    print()
    
    # Add a comment with more details
    print("   Adding follow-up comment...")
    comment = await client.create_comment(
        issue_id=issue["id"],
        body="Additional context: CPU spike occurred during batch job execution."
    )
    print(f"   ✅ Comment added")
    print()
    
    # Update issue when resolved
    print("   Updating issue status...")
    # Note: You'd need to get the state ID first, this is just an example
    # updated = await client.update_issue(
    #     issue_id=issue["id"],
    #     description=issue.get("description", "") + "\n\n**Resolved:** CPU returned to normal."
    # )
    print("   ✅ Issue updated (example)")
    print()


if __name__ == "__main__":
    print()
    asyncio.run(example_basic_usage())
    print()
    # Uncomment to run agent integration example
    # asyncio.run(example_agent_integration())

