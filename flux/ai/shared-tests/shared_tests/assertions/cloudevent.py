"""
CloudEvent assertions for testing.

Provides semantic assertions for validating CloudEvents
in agent tests.
"""

from typing import Any, Optional


def assert_cloudevent_valid(event: dict, strict: bool = True):
    """
    Assert that a CloudEvent has all required fields and is valid.
    
    Args:
        event: The CloudEvent dictionary to validate
        strict: If True, also validates optional fields format
    
    Raises:
        AssertionError: If the event is invalid
    """
    # Required CloudEvents 1.0 fields
    required_fields = ["specversion", "type", "source", "id"]
    
    for field in required_fields:
        assert field in event, f"CloudEvent missing required field: {field}"
    
    # Validate specversion
    assert event["specversion"] == "1.0", (
        f"Invalid specversion: {event['specversion']}, expected 1.0"
    )
    
    # Validate type format (should be reverse-DNS)
    event_type = event["type"]
    assert "." in event_type, (
        f"Event type '{event_type}' should use reverse-DNS notation"
    )
    
    # Validate source format (should be URI-reference)
    source = event["source"]
    assert source.startswith("/") or "://" in source, (
        f"Event source '{source}' should be URI or path reference"
    )
    
    # Validate id is not empty
    assert event["id"], "Event id cannot be empty"
    
    if strict:
        # Validate time format if present (ISO 8601)
        if "time" in event:
            time_str = event["time"]
            assert "T" in time_str, (
                f"Event time '{time_str}' should be ISO 8601 format"
            )
        
        # Validate datacontenttype if present
        if "datacontenttype" in event:
            content_type = event["datacontenttype"]
            assert "/" in content_type, (
                f"Invalid datacontenttype: {content_type}"
            )


def assert_cloudevent_type(
    event: dict,
    expected_type: str,
    prefix: str = "io.homelab.",
):
    """
    Assert CloudEvent has expected type.
    
    Args:
        event: The CloudEvent dictionary
        expected_type: Expected type (can be shorthand or full)
        prefix: Prefix to add if expected_type doesn't include it
    """
    actual_type = event.get("type", "")
    
    # Allow shorthand without prefix
    if not expected_type.startswith(prefix):
        expected_type = f"{prefix}{expected_type}"
    
    assert actual_type == expected_type, (
        f"CloudEvent type mismatch: expected '{expected_type}', got '{actual_type}'"
    )


def assert_cloudevent_source(
    event: dict,
    expected_source: str,
    partial_match: bool = False,
):
    """
    Assert CloudEvent has expected source.
    
    Args:
        event: The CloudEvent dictionary
        expected_source: Expected source URI
        partial_match: If True, check if source contains expected_source
    """
    actual_source = event.get("source", "")
    
    if partial_match:
        assert expected_source in actual_source, (
            f"CloudEvent source '{actual_source}' does not contain '{expected_source}'"
        )
    else:
        assert actual_source == expected_source, (
            f"CloudEvent source mismatch: expected '{expected_source}', got '{actual_source}'"
        )


def assert_cloudevent_data_contains(
    event: dict,
    expected_data: dict,
):
    """
    Assert CloudEvent data contains expected key-value pairs.
    
    Args:
        event: The CloudEvent dictionary
        expected_data: Dictionary of expected key-value pairs in data
    """
    actual_data = event.get("data", {})
    
    for key, expected_value in expected_data.items():
        assert key in actual_data, (
            f"CloudEvent data missing key: '{key}'"
        )
        
        actual_value = actual_data[key]
        assert actual_value == expected_value, (
            f"CloudEvent data['{key}'] mismatch: "
            f"expected '{expected_value}', got '{actual_value}'"
        )


def assert_cloudevent_data_type(
    event: dict,
    key: str,
    expected_type: type,
):
    """
    Assert CloudEvent data field has expected type.
    
    Args:
        event: The CloudEvent dictionary
        key: Key in data to check
        expected_type: Expected Python type
    """
    actual_data = event.get("data", {})
    
    assert key in actual_data, f"CloudEvent data missing key: '{key}'"
    
    actual_value = actual_data[key]
    assert isinstance(actual_value, expected_type), (
        f"CloudEvent data['{key}'] type mismatch: "
        f"expected {expected_type.__name__}, got {type(actual_value).__name__}"
    )


def assert_cloudevent_sequence(
    events: list[dict],
    expected_types: list[str],
    same_source: bool = True,
):
    """
    Assert a sequence of CloudEvents has expected types in order.
    
    Args:
        events: List of CloudEvent dictionaries
        expected_types: Expected event types in order
        same_source: If True, assert all events have same source
    """
    assert len(events) == len(expected_types), (
        f"Event sequence length mismatch: "
        f"expected {len(expected_types)}, got {len(events)}"
    )
    
    first_source = events[0].get("source") if events else None
    
    for i, (event, expected_type) in enumerate(zip(events, expected_types)):
        assert_cloudevent_type(event, expected_type)
        
        if same_source and first_source:
            actual_source = event.get("source")
            assert actual_source == first_source, (
                f"Event {i} source mismatch: "
                f"expected '{first_source}', got '{actual_source}'"
            )


def assert_cloudevent_headers_valid(
    headers: dict,
):
    """
    Assert HTTP headers contain valid CloudEvent headers.
    
    Args:
        headers: HTTP headers dictionary
    """
    required_headers = ["ce-specversion", "ce-type", "ce-source", "ce-id"]
    
    # Normalize header names to lowercase
    normalized = {k.lower(): v for k, v in headers.items()}
    
    for header in required_headers:
        assert header in normalized, (
            f"Missing CloudEvent header: {header}"
        )
    
    assert normalized["ce-specversion"] == "1.0", (
        f"Invalid ce-specversion header: {normalized['ce-specversion']}"
    )
