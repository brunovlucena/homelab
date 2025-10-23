"""
Tests for Agent Bruno core functionality
"""

import pytest


@pytest.mark.asyncio
async def test_agent_chat():
    """Test basic chat functionality"""
    # This is a placeholder test
    assert True


@pytest.mark.asyncio
async def test_agent_knowledge_search():
    """Test knowledge base search"""
    # This is a placeholder test
    assert True


@pytest.mark.asyncio
async def test_agent_with_context():
    """Test chat with conversation context"""
    # This is a placeholder test
    assert True


def test_knowledge_base():
    """Test knowledge base initialization"""
    from src.knowledge.homepage import HomepageKnowledge
    
    kb = HomepageKnowledge()
    
    # Test getting architecture info
    arch = kb.get_info("architecture")
    assert arch is not None
    assert "layers" in arch
    
    # Test getting API endpoints
    api = kb.get_info("api")
    assert api is not None
    assert "projects" in api
    
    # Test search
    results = kb.search("deploy")
    assert len(results) > 0

