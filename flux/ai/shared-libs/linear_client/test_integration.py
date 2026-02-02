#!/usr/bin/env python3
"""
Integration test for Linear Client

This script tests the Linear client against the real Linear API.
Requires LINEAR_API_KEY environment variable to be set.

Usage:
    export LINEAR_API_KEY=lin_api_xxxxxxxxxxxxx
    python test_integration.py
"""
import os
import asyncio
import sys
from linear_client import LinearClient, create_agent_issue


async def test_integration():
    """Run integration tests against real Linear API."""
    
    # Check for API key
    api_key = os.getenv("LINEAR_API_KEY")
    if not api_key:
        print("❌ ERROR: LINEAR_API_KEY environment variable not set")
        print("   Set it with: export LINEAR_API_KEY=lin_api_xxxxxxxxxxxxx")
        sys.exit(1)
    
    print("🧪 Starting Linear Client Integration Tests\n")
    print(f"✅ API Key found: {api_key[:10]}...\n")
    
    client = LinearClient(api_key=api_key)
    
    try:
        # Test 1: List teams
        print("📋 Test 1: Listing teams...")
        teams = await client.list_teams()
        print(f"   ✅ Found {len(teams)} teams")
        for team in teams[:3]:  # Show first 3
            print(f"      - {team['name']} ({team['key']})")
        print()
        
        # Test 2: Get specific team
        if teams:
            team_key = teams[0]["key"]
            print(f"📋 Test 2: Getting team '{team_key}'...")
            team = await client.get_team(team_key)
            print(f"   ✅ Team: {team['name']} - {team.get('description', 'No description')}")
            print()
            
            # Test 3: List issues for this team
            print(f"📋 Test 3: Listing issues for team '{team_key}'...")
            issues = await client.list_issues(team=team_key, limit=5)
            print(f"   ✅ Found {len(issues)} issues")
            for issue in issues[:3]:  # Show first 3
                print(f"      - {issue['identifier']}: {issue['title']} ({issue['state']['name']})")
            print()
            
            # Test 4: Create a test issue
            print(f"📋 Test 4: Creating test issue in team '{team_key}'...")
            test_issue = await client.create_issue(
                title="[TEST] Linear Client Integration Test",
                description="""
This is a test issue created by the Linear client integration test.

**Test Details:**
- Created via Python client
- Testing API integration
- Can be safely deleted

---
*This issue was automatically created by test_integration.py*
                """.strip(),
                team_id=team["id"],
                priority=4  # Low priority
            )
            print(f"   ✅ Issue created: {test_issue['identifier']}")
            print(f"      URL: {test_issue['url']}")
            print()
            
            # Test 5: Get the created issue
            print(f"📋 Test 5: Retrieving issue {test_issue['identifier']}...")
            retrieved_issue = await client.get_issue(test_issue["id"])
            print(f"   ✅ Retrieved: {retrieved_issue['title']}")
            print(f"      State: {retrieved_issue['state']['name']}")
            print()
            
            # Test 6: Add a comment
            print(f"📋 Test 6: Adding comment to {test_issue['identifier']}...")
            comment = await client.create_comment(
                issue_id=test_issue["id"],
                body="This is a test comment from the integration test."
            )
            print(f"   ✅ Comment added by {comment['user']['name']}")
            print()
            
            # Test 7: Update the issue
            print(f"📋 Test 7: Updating issue {test_issue['identifier']}...")
            updated_issue = await client.update_issue(
                issue_id=test_issue["id"],
                description=test_issue.get("description", "") + "\n\n**Updated:** Issue was updated via integration test."
            )
            print(f"   ✅ Issue updated")
            print()
            
            # Test 8: Test helper function
            print(f"📋 Test 8: Testing create_agent_issue helper...")
            helper_url = await create_agent_issue(
                title="[TEST] Helper Function Test",
                description="This issue was created using the create_agent_issue helper function.",
                team_key=team_key,
                agent_name="test-integration-script",
                priority=4
            )
            print(f"   ✅ Helper function worked: {helper_url}")
            print()
            
            print("=" * 60)
            print("✅ All integration tests passed!")
            print("=" * 60)
            print(f"\n📝 Test issues created:")
            print(f"   - {test_issue['url']}")
            print(f"   - {helper_url}")
            print(f"\n💡 You can delete these test issues in Linear.")
            
        else:
            print("⚠️  No teams found. Skipping team-specific tests.")
            print("   Make sure your Linear workspace has at least one team.")
        
    except Exception as e:
        print(f"\n❌ Test failed with error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


async def test_quick():
    """Quick test - just verify API key works."""
    api_key = os.getenv("LINEAR_API_KEY")
    if not api_key:
        print("❌ ERROR: LINEAR_API_KEY environment variable not set")
        sys.exit(1)
    
    client = LinearClient(api_key=api_key)
    
    try:
        teams = await client.list_teams()
        print(f"✅ Linear API connection successful!")
        print(f"   Found {len(teams)} teams")
        return True
    except Exception as e:
        print(f"❌ Linear API connection failed: {e}")
        return False


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "--quick":
        success = asyncio.run(test_quick())
        sys.exit(0 if success else 1)
    else:
        asyncio.run(test_integration())

