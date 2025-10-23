# User Feedback Implementation Guide

**[← Back to README](../README.md)** | **[Learning](LEARNING.md)** | **[Integration](LEARNING_MEMORY_HOMEPAGE_INTEGRATION.md)**

---

## Table of Contents
1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Database Schema](#database-schema)
4. [Frontend Implementation (Homepage)](#frontend-implementation-homepage)
5. [Backend Implementation (Homepage API)](#backend-implementation-homepage-api)
6. [Agent Bruno Integration](#agent-bruno-integration)
7. [Implicit Feedback Tracking](#implicit-feedback-tracking)
8. [Deployment Steps](#deployment-steps)
9. [Testing](#testing)

---

## Overview

The user feedback system enables continuous learning by collecting both **explicit** (thumbs up/down, ratings) and **implicit** (copy, clicks, read time) feedback signals from homepage users interacting with Agent Bruno.

### Current Status

🔴 **NOT YET IMPLEMENTED** - This is the implementation guide

### What Needs to Be Built

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Components to Implement                              │
│                                                                         │
│  1. Database Schema (Postgres)                                          │
│     ├─ feedback_events table                                            │
│     └─ Indexes for querying                                             │
│                                                                         │
│  2. Frontend (Homepage - React/TypeScript)                              │
│     ├─ FeedbackWidget component                                         │
│     ├─ ImplicitTracker service                                          │
│     └─ API integration                                                  │
│                                                                         │
│  3. Backend API (Homepage - Go)                                         │
│     ├─ POST /api/feedback endpoint                                      │
│     ├─ GET /api/feedback/stats endpoint                                 │
│     ├─ Database models & migrations                                     │
│     └─ Metrics & logging                                                │
│                                                                         │
│  4. Agent Bruno (Python)                                                │
│     ├─ Feedback metadata in responses                                   │
│     └─ Trace ID propagation                                             │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Architecture

### Data Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        User Interaction                                 │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Homepage Frontend (React)                                      │    │
│  │  ┌──────────────────────────────────────────────────────────┐   │    │
│  │  │  User sees Agent Bruno response                          │   │    │
│  │  │  ┌─────────────────────────────────────────────────┐     │   │    │
│  │  │  │  💬 Response: "Loki crashes are caused by..."   │     │   │    │
│  │  │  │                                                 │     │   │    │
│  │  │  │  Actions:                                       │     │   │    │
│  │  │  │  [👍 Helpful] [👎 Not Helpful] [⭐⭐⭐⭐⭐]    │     │   │    │
│  │  │  └─────────────────────────────────────────────────┘     │   │    │
│  │  │                                                          │   │    │
│  │  │  User clicks: 👍 Helpful                                 │   │    │
│  │  │  User copies response to clipboard                       │   │    │
│  │  │  User clicks 2 citations                                 │   │    │
│  │  └──────────────────────────────────────────────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             │ POST /api/feedback
                             │ {
                             │   interaction_id: "int-789",
                             │   feedback_type: "thumbs_up",
                             │   value: 1.0,
                             │   implicit_signals: {...}
                             │ }
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    Homepage API (Go)                                    │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  handlers/feedback.go                                           │    │
│  │  ┌──────────────────────────────────────────────────────────┐   │    │
│  │  │  func SubmitFeedback(c *gin.Context) {                   │   │    │
│  │  │    // 1. Validate request                                │   │    │
│  │  │    // 2. Store in Postgres                               │   │    │
│  │  │    // 3. Update metrics                                  │   │    │
│  │  │    // 4. Return success                                  │   │    │
│  │  │  }                                                       │   │    │
│  │  └──────────────────────────────────────────────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             │ INSERT INTO feedback_events
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    Postgres Database                                    │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Table: feedback_events                                         │    │
│  │  ┌──────────────────────────────────────────────────────────┐   │    │
│  │  │  id: uuid PRIMARY KEY                                    │   │    │
│  │  │  interaction_id: varchar(255) NOT NULL                   │   │    │
│  │  │  user_id: varchar(255)                                   │   │    │
│  │  │  session_id: varchar(255)                                │   │    │
│  │  │  feedback_type: varchar(50) NOT NULL                     │   │    │
│  │  │  feedback_value: float NOT NULL                          │   │    │
│  │  │  metadata: jsonb                                         │   │    │
│  │  │  created_at: timestamp NOT NULL                          │   │    │
│  │  │  model_version: varchar(100)                             │   │    │
│  │  └──────────────────────────────────────────────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             │ Used for Training
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│           Weekly Curation Job (joins with LanceDB)                      │
│           → Training Data → Fine-tuning                                 │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Database Schema

### Migration File

**Location**: `repos/homelab/flux/clusters/homelab/infrastructure/homepage/api/migrations/000006_create_feedback_events.up.sql`

```sql
-- Migration: Create feedback_events table
-- Purpose: Store user feedback for Agent Bruno responses to enable continuous learning

-- =============================================================================
-- Main Table: feedback_events
-- =============================================================================

CREATE TABLE IF NOT EXISTS feedback_events (
    -- Primary identifier
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Linkage to interaction (trace_id from Agent Bruno response)
    interaction_id VARCHAR(255) NOT NULL,
    
    -- User identification
    user_id VARCHAR(255),  -- Can be IP-based or actual user ID
    session_id VARCHAR(255),
    
    -- Feedback data
    feedback_type VARCHAR(50) NOT NULL,  -- thumbs_up, thumbs_down, rating, correction
    feedback_value FLOAT NOT NULL,  -- Normalized -1 to +1
    
    -- Additional context (JSONB for flexibility)
    metadata JSONB DEFAULT '{}',
    -- Example metadata:
    -- {
    --   "implicit_signals": {
    --     "copy_event": true,
    --     "citation_clicks": 2,
    --     "read_time_seconds": 45,
    --     "follow_up_asked": true
    --   },
    --   "platform": "homepage",
    --   "user_agent": "...",
    --   "quality_signals": {
    --     "response_length_tokens": 256,
    --     "context_used": true
    --   }
    -- }
    
    -- Model version that generated the response
    model_version VARCHAR(100),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Optional: Corrected response for RLHF
    corrected_response TEXT,
    correction_type VARCHAR(50)  -- factual, style, tone, etc.
);

-- =============================================================================
-- Indexes for Performance
-- =============================================================================

-- Index for querying by interaction
CREATE INDEX idx_feedback_events_interaction_id 
    ON feedback_events(interaction_id);

-- Index for querying by user
CREATE INDEX idx_feedback_events_user_id 
    ON feedback_events(user_id) 
    WHERE user_id IS NOT NULL;

-- Index for querying by session
CREATE INDEX idx_feedback_events_session_id 
    ON feedback_events(session_id) 
    WHERE session_id IS NOT NULL;

-- Index for time-based queries (for weekly curation)
CREATE INDEX idx_feedback_events_created_at 
    ON feedback_events(created_at DESC);

-- Index for feedback type filtering
CREATE INDEX idx_feedback_events_type 
    ON feedback_events(feedback_type);

-- Index for model version comparison
CREATE INDEX idx_feedback_events_model_version 
    ON feedback_events(model_version) 
    WHERE model_version IS NOT NULL;

-- Composite index for common query pattern (time + type)
CREATE INDEX idx_feedback_events_time_type 
    ON feedback_events(created_at DESC, feedback_type);

-- GIN index for JSONB metadata queries
CREATE INDEX idx_feedback_events_metadata 
    ON feedback_events USING GIN (metadata);

-- =============================================================================
-- Trigger for updated_at
-- =============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_feedback_events_updated_at 
    BEFORE UPDATE ON feedback_events
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- Comments
-- =============================================================================

COMMENT ON TABLE feedback_events IS 
    'Stores user feedback for Agent Bruno responses to enable continuous learning';

COMMENT ON COLUMN feedback_events.interaction_id IS 
    'Links to trace_id from Agent Bruno response (stored in LanceDB episodic_memory)';

COMMENT ON COLUMN feedback_events.feedback_value IS 
    'Normalized feedback score: -1 (very bad) to +1 (very good)';

COMMENT ON COLUMN feedback_events.metadata IS 
    'JSONB field for flexible storage of implicit signals and additional context';
```

**Down Migration**: `000006_create_feedback_events.down.sql`

```sql
-- Drop feedback_events table and related objects

DROP TRIGGER IF EXISTS update_feedback_events_updated_at ON feedback_events;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_feedback_events_metadata;
DROP INDEX IF EXISTS idx_feedback_events_time_type;
DROP INDEX IF EXISTS idx_feedback_events_model_version;
DROP INDEX IF EXISTS idx_feedback_events_type;
DROP INDEX IF EXISTS idx_feedback_events_created_at;
DROP INDEX IF EXISTS idx_feedback_events_session_id;
DROP INDEX IF EXISTS idx_feedback_events_user_id;
DROP INDEX IF EXISTS idx_feedback_events_interaction_id;
DROP TABLE IF EXISTS feedback_events;
```

---

## Frontend Implementation (Homepage)

### 1. Feedback Widget Component

**Location**: `repos/homelab/flux/clusters/homelab/infrastructure/homepage/frontend/src/components/FeedbackWidget.tsx`

```typescript
import React, { useState, useCallback } from 'react';
import { ThumbsUp, ThumbsDown, Star } from 'lucide-react';
import { submitFeedback, ImplicitSignals } from '../services/feedback';

interface FeedbackWidgetProps {
  interactionId: string;
  onFeedbackSubmitted?: () => void;
}

export const FeedbackWidget: React.FC<FeedbackWidgetProps> = ({
  interactionId,
  onFeedbackSubmitted
}) => {
  const [feedbackGiven, setFeedbackGiven] = useState<string | null>(null);
  const [rating, setRating] = useState<number>(0);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleThumbsFeedback = useCallback(async (type: 'up' | 'down') => {
    if (feedbackGiven || isSubmitting) return;

    setIsSubmitting(true);
    
    try {
      await submitFeedback({
        interactionId,
        feedbackType: type === 'up' ? 'thumbs_up' : 'thumbs_down',
        value: type === 'up' ? 1.0 : -1.0
      });

      setFeedbackGiven(type);
      onFeedbackSubmitted?.();
      
      // Show toast notification
      console.log(`👍 Feedback submitted: ${type}`);
    } catch (error) {
      console.error('Failed to submit feedback:', error);
    } finally {
      setIsSubmitting(false);
    }
  }, [interactionId, feedbackGiven, isSubmitting, onFeedbackSubmitted]);

  const handleRating = useCallback(async (stars: number) => {
    if (feedbackGiven || isSubmitting) return;

    setIsSubmitting(true);
    setRating(stars);

    try {
      // Normalize 1-5 stars to -1 to +1
      const normalizedValue = (stars - 3) / 2;

      await submitFeedback({
        interactionId,
        feedbackType: 'rating',
        value: normalizedValue,
        metadata: {
          stars
        }
      });

      setFeedbackGiven('rating');
      onFeedbackSubmitted?.();
    } catch (error) {
      console.error('Failed to submit rating:', error);
    } finally {
      setIsSubmitting(false);
    }
  }, [interactionId, feedbackGiven, isSubmitting, onFeedbackSubmitted]);

  return (
    <div className="feedback-widget mt-4 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
      {!feedbackGiven ? (
        <div className="flex items-center gap-4">
          <span className="text-sm text-gray-600 dark:text-gray-400">
            Was this helpful?
          </span>
          
          {/* Thumbs Up/Down */}
          <div className="flex gap-2">
            <button
              onClick={() => handleThumbsFeedback('up')}
              disabled={isSubmitting}
              className="p-2 hover:bg-green-100 dark:hover:bg-green-900 rounded-full transition-colors"
              aria-label="Helpful"
            >
              <ThumbsUp className="w-5 h-5 text-green-600 dark:text-green-400" />
            </button>
            
            <button
              onClick={() => handleThumbsFeedback('down')}
              disabled={isSubmitting}
              className="p-2 hover:bg-red-100 dark:hover:bg-red-900 rounded-full transition-colors"
              aria-label="Not helpful"
            >
              <ThumbsDown className="w-5 h-5 text-red-600 dark:text-red-400" />
            </button>
          </div>

          {/* Star Rating */}
          <div className="flex gap-1 ml-4">
            {[1, 2, 3, 4, 5].map((star) => (
              <button
                key={star}
                onClick={() => handleRating(star)}
                disabled={isSubmitting}
                className="p-1 hover:scale-110 transition-transform"
                aria-label={`Rate ${star} stars`}
              >
                <Star
                  className={`w-4 h-4 ${
                    star <= rating
                      ? 'fill-yellow-400 text-yellow-400'
                      : 'text-gray-300 dark:text-gray-600'
                  }`}
                />
              </button>
            ))}
          </div>
        </div>
      ) : (
        <div className="text-sm text-green-600 dark:text-green-400">
          ✓ Thank you for your feedback!
        </div>
      )}
    </div>
  );
};
```

### 2. Implicit Feedback Tracker

**Location**: `repos/homelab/flux/clusters/homelab/infrastructure/homepage/frontend/src/services/implicitTracker.ts`

```typescript
import { submitFeedback } from './feedback';

interface ImplicitTrackerOptions {
  interactionId: string;
  responseText: string;
}

export class ImplicitFeedbackTracker {
  private interactionId: string;
  private responseText: string;
  private signals: {
    copyEvent: boolean;
    citationClicks: number;
    readTimeStart: number;
    followUpAsked: boolean;
  };
  private citationClickHandlers: Map<string, () => void>;

  constructor(options: ImplicitTrackerOptions) {
    this.interactionId = options.interactionId;
    this.responseText = options.responseText;
    this.signals = {
      copyEvent: false,
      citationClicks: 0,
      readTimeStart: Date.now(),
      followUpAsked: false
    };
    this.citationClickHandlers = new Map();

    this.setupTracking();
  }

  private setupTracking() {
    // Track copy events
    this.trackCopyEvent();
    
    // Submit implicit signals when user navigates away or after timeout
    window.addEventListener('beforeunload', () => this.submit());
    
    // Submit after 2 minutes if user is still on page
    setTimeout(() => this.submit(), 120000);
  }

  private trackCopyEvent() {
    // Listen for copy events on the response element
    const handleCopy = (e: ClipboardEvent) => {
      const selection = window.getSelection()?.toString() || '';
      
      // Check if copied text is from our response
      if (this.responseText.includes(selection) && selection.length > 20) {
        this.signals.copyEvent = true;
        console.log('📋 User copied response');
      }
    };

    document.addEventListener('copy', handleCopy);
  }

  public trackCitationClick(citationId: string) {
    this.signals.citationClicks++;
    console.log(`🔗 User clicked citation ${citationId}`);
  }

  public trackFollowUp() {
    this.signals.followUpAsked = true;
    console.log('💬 User asked follow-up question');
    
    // Submit immediately when follow-up is asked
    this.submit();
  }

  private calculateReadTime(): number {
    return Math.round((Date.now() - this.signals.readTimeStart) / 1000);
  }

  public async submit() {
    const readTimeSeconds = this.calculateReadTime();
    
    // Only submit if there's meaningful interaction
    if (
      !this.signals.copyEvent &&
      this.signals.citationClicks === 0 &&
      !this.signals.followUpAsked &&
      readTimeSeconds < 5
    ) {
      return;
    }

    try {
      await submitFeedback({
        interactionId: this.interactionId,
        feedbackType: 'implicit',
        value: this.calculateImplicitScore(),
        implicitSignals: {
          copy_event: this.signals.copyEvent,
          citation_clicks: this.signals.citationClicks,
          read_time_seconds: readTimeSeconds,
          follow_up_asked: this.signals.followUpAsked
        }
      });

      console.log('📊 Implicit feedback submitted');
    } catch (error) {
      console.error('Failed to submit implicit feedback:', error);
    }
  }

  private calculateImplicitScore(): number {
    let score = 0;

    if (this.signals.copyEvent) score += 0.4;
    score += Math.min(this.signals.citationClicks * 0.15, 0.3);
    if (this.signals.followUpAsked) score += 0.3;

    const readTime = this.calculateReadTime();
    const expectedReadTime = this.estimateReadTime(this.responseText);
    if (readTime >= expectedReadTime * 0.8) {
      score += 0.2;
    }

    return Math.min(score, 1.0);
  }

  private estimateReadTime(text: string): number {
    // Average reading speed: 200 words per minute
    const words = text.split(/\s+/).length;
    return Math.ceil((words / 200) * 60);
  }
}
```

### 3. Feedback Service

**Location**: `repos/homelab/flux/clusters/homelab/infrastructure/homepage/frontend/src/services/feedback.ts`

```typescript
import axios from 'axios';
import { env } from '../config/env';

export interface ImplicitSignals {
  copy_event?: boolean;
  citation_clicks?: number;
  read_time_seconds?: number;
  follow_up_asked?: boolean;
}

export interface FeedbackRequest {
  interactionId: string;
  feedbackType: 'thumbs_up' | 'thumbs_down' | 'rating' | 'implicit' | 'correction';
  value: number;  // -1 to +1
  implicitSignals?: ImplicitSignals;
  metadata?: Record<string, any>;
  correctedResponse?: string;
}

export interface FeedbackResponse {
  success: boolean;
  feedbackId: string;
}

const apiClient = axios.create({
  baseURL: env.API_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
});

export async function submitFeedback(
  request: FeedbackRequest
): Promise<FeedbackResponse> {
  try {
    const response = await apiClient.post<FeedbackResponse>(
      '/api/feedback',
      {
        interaction_id: request.interactionId,
        feedback_type: request.feedbackType,
        feedback_value: request.value,
        implicit_signals: request.implicitSignals,
        metadata: request.metadata,
        corrected_response: request.correctedResponse
      }
    );

    return response.data;
  } catch (error) {
    console.error('Failed to submit feedback:', error);
    throw error;
  }
}

export async function getFeedbackStats(
  interactionId: string
): Promise<any> {
  try {
    const response = await apiClient.get(
      `/api/feedback/stats/${interactionId}`
    );
    return response.data;
  } catch (error) {
    console.error('Failed to get feedback stats:', error);
    throw error;
  }
}
```

### 4. Integration in Chat Component

**Location**: `repos/homelab/flux/clusters/homelab/infrastructure/homepage/frontend/src/components/Chatbot.tsx`

```typescript
// Add to existing Chatbot component
import { FeedbackWidget } from './FeedbackWidget';
import { ImplicitFeedbackTracker } from '../services/implicitTracker';

// In the message rendering section:
{message.role === 'assistant' && message.interactionId && (
  <>
    {/* Existing message content */}
    <div className="message-content">
      {message.content}
    </div>

    {/* NEW: Feedback Widget */}
    <FeedbackWidget
      interactionId={message.interactionId}
      onFeedbackSubmitted={() => {
        console.log('Feedback submitted for:', message.interactionId);
      }}
    />
  </>
)}

// When message is received from Agent Bruno:
const handleResponse = (response: ChatResponse) => {
  const newMessage = {
    role: 'assistant',
    content: response.response,
    interactionId: response.trace_id,  // NEW: Store trace_id
    timestamp: new Date()
  };

  setMessages([...messages, newMessage]);

  // NEW: Initialize implicit tracker
  const tracker = new ImplicitFeedbackTracker({
    interactionId: response.trace_id,
    responseText: response.response
  });

  // Store tracker for later use
  trackers.set(response.trace_id, tracker);
};

// When user clicks citations:
const handleCitationClick = (interactionId: string, citationId: string) => {
  const tracker = trackers.get(interactionId);
  if (tracker) {
    tracker.trackCitationClick(citationId);
  }
};

// When user sends follow-up:
const handleSendMessage = (message: string) => {
  // If there's a previous assistant message, mark as follow-up
  const lastMessage = messages[messages.length - 1];
  if (lastMessage?.role === 'assistant' && lastMessage.interactionId) {
    const tracker = trackers.get(lastMessage.interactionId);
    if (tracker) {
      tracker.trackFollowUp();
    }
  }

  // Send new message...
};
```

---

## Backend Implementation (Homepage API)

### 1. Database Model

**Location**: `repos/homelab/flux/clusters/homelab/infrastructure/homepage/api/models/feedback.go`

```go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// FeedbackEvent represents user feedback on Agent Bruno responses
type FeedbackEvent struct {
	ID             uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	InteractionID  string          `gorm:"type:varchar(255);not null;index" json:"interaction_id"`
	UserID         *string         `gorm:"type:varchar(255);index" json:"user_id,omitempty"`
	SessionID      *string         `gorm:"type:varchar(255);index" json:"session_id,omitempty"`
	FeedbackType   string          `gorm:"type:varchar(50);not null;index" json:"feedback_type"`
	FeedbackValue  float64         `gorm:"type:float;not null" json:"feedback_value"`
	Metadata       JSONB           `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	ModelVersion   *string         `gorm:"type:varchar(100);index" json:"model_version,omitempty"`
	CreatedAt      time.Time       `gorm:"not null;default:now();index:idx_feedback_events_created_at" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"not null;default:now()" json:"updated_at"`
	CorrectedResponse *string      `gorm:"type:text" json:"corrected_response,omitempty"`
	CorrectionType *string         `gorm:"type:varchar(50)" json:"correction_type,omitempty"`
}

// TableName specifies the table name for FeedbackEvent
func (FeedbackEvent) TableName() string {
	return "feedback_events"
}

// JSONB is a custom type for JSONB fields
type JSONB map[string]interface{}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	return json.Unmarshal(bytes, j)
}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "{}", nil
	}
	return json.Marshal(j)
}

// ImplicitSignals represents implicit feedback signals
type ImplicitSignals struct {
	CopyEvent       bool   `json:"copy_event"`
	CitationClicks  int    `json:"citation_clicks"`
	ReadTimeSeconds int    `json:"read_time_seconds"`
	FollowUpAsked   bool   `json:"follow_up_asked"`
}

// FeedbackRequest is the request body for submitting feedback
type FeedbackRequest struct {
	InteractionID     string           `json:"interaction_id" binding:"required"`
	FeedbackType      string           `json:"feedback_type" binding:"required,oneof=thumbs_up thumbs_down rating implicit correction"`
	FeedbackValue     float64          `json:"feedback_value" binding:"required,min=-1,max=1"`
	ImplicitSignals   *ImplicitSignals `json:"implicit_signals,omitempty"`
	Metadata          JSONB            `json:"metadata,omitempty"`
	CorrectedResponse *string          `json:"corrected_response,omitempty"`
	CorrectionType    *string          `json:"correction_type,omitempty"`
}

// FeedbackResponse is the response after submitting feedback
type FeedbackResponse struct {
	Success    bool      `json:"success"`
	FeedbackID uuid.UUID `json:"feedback_id"`
}
```

### 2. Handler

**Location**: `repos/homelab/flux/clusters/homelab/infrastructure/homepage/api/handlers/feedback.go`

```go
package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/brunovlucena/homelab/homepage-api/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubmitFeedback handles POST /api/feedback
func SubmitFeedback(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.FeedbackRequest

		// Bind and validate request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request",
				"details": err.Error(),
			})
			return
		}

		// Get user context
		userID := getUserID(c)  // From session or IP
		sessionID := getSessionID(c)

		// Build metadata
		metadata := req.Metadata
		if metadata == nil {
			metadata = make(models.JSONB)
		}

		// Add implicit signals to metadata
		if req.ImplicitSignals != nil {
			metadata["implicit_signals"] = map[string]interface{}{
				"copy_event":        req.ImplicitSignals.CopyEvent,
				"citation_clicks":   req.ImplicitSignals.CitationClicks,
				"read_time_seconds": req.ImplicitSignals.ReadTimeSeconds,
				"follow_up_asked":   req.ImplicitSignals.FollowUpAsked,
			}
		}

		// Add request context
		metadata["platform"] = "homepage"
		metadata["user_agent"] = c.Request.UserAgent()
		metadata["ip_address"] = c.ClientIP()

		// Create feedback event
		feedback := models.FeedbackEvent{
			InteractionID:     req.InteractionID,
			UserID:            &userID,
			SessionID:         &sessionID,
			FeedbackType:      req.FeedbackType,
			FeedbackValue:     req.FeedbackValue,
			Metadata:          metadata,
			CorrectedResponse: req.CorrectedResponse,
			CorrectionType:    req.CorrectionType,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		// Store in database
		if err := db.Create(&feedback).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to store feedback",
			})
			return
		}

		// Update Prometheus metrics
		feedbackEventsTotal.WithLabelValues(req.FeedbackType).Inc()
		feedbackScoreDistribution.Observe(req.FeedbackValue)

		// Log for debugging
		fmt.Printf("📊 Feedback received: %s (%.2f) for interaction %s\n",
			req.FeedbackType, req.FeedbackValue, req.InteractionID)

		c.JSON(http.StatusOK, models.FeedbackResponse{
			Success:    true,
			FeedbackID: feedback.ID,
		})
	}
}

// GetFeedbackStats handles GET /api/feedback/stats/:interaction_id
func GetFeedbackStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		interactionID := c.Param("interaction_id")

		var feedbacks []models.FeedbackEvent
		if err := db.Where("interaction_id = ?", interactionID).Find(&feedbacks).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch feedback stats",
			})
			return
		}

		// Calculate stats
		stats := map[string]interface{}{
			"total_feedback": len(feedbacks),
			"breakdown":      make(map[string]int),
			"average_score":  0.0,
		}

		totalScore := 0.0
		breakdown := make(map[string]int)

		for _, fb := range feedbacks {
			totalScore += fb.FeedbackValue
			breakdown[fb.FeedbackType]++
		}

		if len(feedbacks) > 0 {
			stats["average_score"] = totalScore / float64(len(feedbacks))
		}
		stats["breakdown"] = breakdown

		c.JSON(http.StatusOK, stats)
	}
}

// Helper functions
func getUserID(c *gin.Context) string {
	// Try to get from session/JWT first
	if userID, exists := c.Get("user_id"); exists {
		return userID.(string)
	}
	
	// Fallback to IP-based ID
	return fmt.Sprintf("ip_%s", c.ClientIP())
}

func getSessionID(c *gin.Context) string {
	// Try to get from session cookie
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		// Generate new session ID
		sessionID = uuid.New().String()
		c.SetCookie("session_id", sessionID, 86400*30, "/", "", false, true)
	}
	return sessionID
}
```

### 3. Metrics

**Location**: `repos/homelab/flux/clusters/homelab/infrastructure/homepage/api/metrics/feedback.go`

```go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// FeedbackEventsTotal tracks total feedback events by type
	FeedbackEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "homepage_feedback_events_total",
			Help: "Total number of feedback events collected",
		},
		[]string{"feedback_type"},
	)

	// FeedbackScoreDistribution tracks distribution of feedback scores
	FeedbackScoreDistribution = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "homepage_feedback_score_distribution",
			Help:    "Distribution of feedback scores",
			Buckets: []float64{-1.0, -0.5, 0, 0.5, 1.0},
		},
	)

	// FeedbackSubmissionDuration tracks time to submit feedback
	FeedbackSubmissionDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "homepage_feedback_submission_duration_seconds",
			Help:    "Time taken to submit feedback",
			Buckets: prometheus.DefBuckets,
		},
	)
)
```

### 4. Router Update

**Location**: `repos/homelab/flux/clusters/homelab/infrastructure/homepage/api/router/router.go`

```go
// Add to existing router.go in the api group:

// Feedback routes
feedback := api.Group("/feedback")
{
    feedback.POST("", handlers.SubmitFeedback(db))
    feedback.GET("/stats/:interaction_id", handlers.GetFeedbackStats(db))
}
```

---

## Agent Bruno Integration

### Update Response to Include Trace ID

**Location**: `repos/homelab/flux/clusters/homelab/infrastructure/agent-bruno/src/api/handlers.py`

```python
# Ensure trace_id is included in response
@app.post("/api/chat")
async def chat(request: ChatRequest) -> ChatResponse:
    """Handle chat request."""
    
    # Generate or get trace ID
    trace_id = request.trace_id or str(uuid.uuid4())
    
    # Set trace context
    with tracer.start_as_current_span("chat_request", attributes={
        "trace_id": trace_id,
        "user_id": request.user_id,
        "session_id": request.session_id
    }):
        # Process request
        response = await agent.process(
            message=request.message,
            context=memory_context
        )
        
        return ChatResponse(
            response=response,
            trace_id=trace_id,  # ← IMPORTANT: Return this to frontend
            session_id=request.session_id,
            model_version=get_current_model_version()
        )
```

---

## Implicit Feedback Tracking

### Automatic Tracking Setup

```typescript
// In Chatbot component - automatic setup
useEffect(() => {
  // Set up global copy listener
  const handleCopy = () => {
    const selection = window.getSelection()?.toString();
    if (selection && selection.length > 20) {
      // Find which message was copied
      const currentTracker = getCurrentMessageTracker();
      if (currentTracker) {
        currentTracker.signals.copyEvent = true;
      }
    }
  };

  document.addEventListener('copy', handleCopy);
  
  return () => {
    document.removeEventListener('copy', handleCopy);
  };
}, []);

// Track read time per message
useEffect(() => {
  if (messages.length > 0) {
    const lastMessage = messages[messages.length - 1];
    if (lastMessage.role === 'assistant' && lastMessage.interactionId) {
      const tracker = new ImplicitFeedbackTracker({
        interactionId: lastMessage.interactionId,
        responseText: lastMessage.content
      });
      
      // Store tracker
      trackersRef.current.set(lastMessage.interactionId, tracker);
    }
  }
}, [messages]);
```

---

## Deployment Steps

### Step-by-Step Deployment

```bash
# 1. Create database migration
cd repos/homelab/flux/clusters/homelab/infrastructure/homepage/api
migrate create -ext sql -dir migrations -seq create_feedback_events

# 2. Run migration
make db-migrate-up

# 3. Add frontend components
cd ../frontend
# Copy FeedbackWidget.tsx, implicitTracker.ts, feedback.ts

# 4. Add backend handlers
cd ../api
# Copy models/feedback.go, handlers/feedback.go, metrics/feedback.go

# 5. Update router
# Add feedback routes to router.go

# 6. Build and deploy
make build
make deploy

# 7. Verify deployment
curl https://bruno.dev/api/feedback -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "interaction_id": "test-123",
    "feedback_type": "thumbs_up",
    "feedback_value": 1.0
  }'
```

---

## Testing

### Unit Tests

**Frontend Test**: `FeedbackWidget.test.tsx`

```typescript
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { FeedbackWidget } from './FeedbackWidget';
import * as feedbackService from '../services/feedback';

jest.mock('../services/feedback');

describe('FeedbackWidget', () => {
  it('submits thumbs up feedback', async () => {
    const mockSubmit = jest.spyOn(feedbackService, 'submitFeedback')
      .mockResolvedValue({ success: true, feedbackId: 'test-id' });

    render(<FeedbackWidget interactionId="test-123" />);

    const thumbsUpButton = screen.getByLabelText('Helpful');
    fireEvent.click(thumbsUpButton);

    await waitFor(() => {
      expect(mockSubmit).toHaveBeenCalledWith({
        interactionId: 'test-123',
        feedbackType: 'thumbs_up',
        value: 1.0
      });
    });

    expect(screen.getByText(/thank you/i)).toBeInTheDocument();
  });

  it('submits star rating', async () => {
    const mockSubmit = jest.spyOn(feedbackService, 'submitFeedback')
      .mockResolvedValue({ success: true, feedbackId: 'test-id' });

    render(<FeedbackWidget interactionId="test-123" />);

    const fourthStar = screen.getByLabelText('Rate 4 stars');
    fireEvent.click(fourthStar);

    await waitFor(() => {
      expect(mockSubmit).toHaveBeenCalledWith({
        interactionId: 'test-123',
        feedbackType: 'rating',
        value: 0.5,  // (4 - 3) / 2 = 0.5
        metadata: { stars: 4 }
      });
    });
  });
});
```

**Backend Test**: `feedback_test.go`

```go
package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brunovlucena/homelab/homepage-api/handlers"
	"github.com/brunovlucena/homelab/homepage-api/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSubmitFeedback(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/api/feedback", handlers.SubmitFeedback(db))

	// Test cases
	tests := []struct {
		name           string
		payload        models.FeedbackRequest
		expectedStatus int
	}{
		{
			name: "Valid thumbs up feedback",
			payload: models.FeedbackRequest{
				InteractionID: "test-123",
				FeedbackType:  "thumbs_up",
				FeedbackValue: 1.0,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid feedback value",
			payload: models.FeedbackRequest{
				InteractionID: "test-123",
				FeedbackType:  "thumbs_up",
				FeedbackValue: 2.0,  // Invalid: > 1.0
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Implicit feedback with signals",
			payload: models.FeedbackRequest{
				InteractionID: "test-456",
				FeedbackType:  "implicit",
				FeedbackValue: 0.7,
				ImplicitSignals: &models.ImplicitSignals{
					CopyEvent:       true,
					CitationClicks:  2,
					ReadTimeSeconds: 45,
					FollowUpAsked:   true,
				},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/feedback", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code == http.StatusOK {
				var resp models.FeedbackResponse
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.True(t, resp.Success)
				assert.NotEmpty(t, resp.FeedbackID)

				// Verify stored in database
				var feedback models.FeedbackEvent
				db.Where("interaction_id = ?", tt.payload.InteractionID).First(&feedback)
				assert.Equal(t, tt.payload.FeedbackType, feedback.FeedbackType)
				assert.Equal(t, tt.payload.FeedbackValue, feedback.FeedbackValue)
			}
		})
	}
}
```

### Integration Test

```bash
# E2E test script
#!/bin/bash

echo "🧪 Testing feedback flow..."

# 1. Submit thumbs up feedback
RESPONSE=$(curl -X POST https://bruno.dev/api/feedback \
  -H "Content-Type: application/json" \
  -d '{
    "interaction_id": "e2e-test-123",
    "feedback_type": "thumbs_up",
    "feedback_value": 1.0
  }')

echo "Response: $RESPONSE"

# 2. Get feedback stats
STATS=$(curl https://bruno.dev/api/feedback/stats/e2e-test-123)

echo "Stats: $STATS"

# 3. Verify in database
psql -h postgres.bruno.dev -U homepage -d homepage -c \
  "SELECT * FROM feedback_events WHERE interaction_id = 'e2e-test-123';"

echo "✅ Feedback flow test complete"
```

---

## Automated Curation Pipeline

### Weekly Training Data Curation Job

**Purpose**: Automatically convert user feedback into training examples for fine-tuning.

**Location**: `repos/homelab/flux/clusters/homelab/infrastructure/agent-bruno/scripts/curate_training_data.py`

```python
#!/usr/bin/env python3
"""
Weekly curation job: Feedback → Training Data

Joins feedback from Postgres with episodic memory from LanceDB
to create high-quality training examples.
"""

import asyncio
import json
from datetime import datetime, timedelta
from typing import List
from pydantic import BaseModel, Field
import psycopg2
import lancedb
import wandb

class TrainingExample(BaseModel):
    """Single training example with validation."""
    query: str = Field(..., min_length=5)
    response: str = Field(..., min_length=20)
    feedback_score: float = Field(..., ge=-1.0, le=1.0)
    sources: List[str] = Field(default_factory=list)
    timestamp: datetime
    model_version: str
    
    # For RLHF (preference pairs)
    preferred_response: str | None = None
    rejected_response: str | None = None
    
    @field_validator('query')
    @classmethod
    def validate_no_pii(cls, v: str) -> str:
        """Ensure no PII in query."""
        if contains_pii(v):
            raise ValueError('Query contains PII - must be redacted')
        return v

class CurationPipeline:
    """Automated training data curation."""
    
    def __init__(
        self,
        postgres_conn_string: str,
        lancedb_path: str,
        wandb_project: str = "agent-bruno"
    ):
        self.pg_conn = psycopg2.connect(postgres_conn_string)
        self.lancedb = lancedb.connect(lancedb_path)
        self.wandb_project = wandb_project
    
    async def curate_weekly_batch(self, min_feedback_score: float = 0.7):
        """
        Curate training data from past week's feedback.
        
        Steps:
        1. Fetch positive feedback events (score >= threshold)
        2. Join with LanceDB episodic memory to get full context
        3. Validate and clean data
        4. Export to JSONL format
        5. Version and upload to Weights & Biases
        6. Create data card with statistics
        """
        # 1. Fetch feedback from Postgres
        feedback_events = self._fetch_positive_feedback(
            days=7,
            min_score=min_feedback_score
        )
        
        print(f"📊 Found {len(feedback_events)} positive feedback events")
        
        # 2. Join with episodic memory
        training_examples = []
        for event in feedback_events:
            try:
                # Get full conversation from LanceDB
                episode = await self._get_episode_from_lancedb(
                    interaction_id=event['interaction_id']
                )
                
                if not episode:
                    continue
                
                # Create training example
                example = TrainingExample(
                    query=episode['user_query'],
                    response=episode['agent_response'],
                    feedback_score=event['feedback_value'],
                    sources=episode.get('sources', []),
                    timestamp=event['created_at'],
                    model_version=event.get('model_version', 'unknown')
                )
                
                training_examples.append(example)
                
            except Exception as e:
                print(f"⚠️  Skipping event {event['id']}: {e}")
                continue
        
        print(f"✅ Created {len(training_examples)} training examples")
        
        # 3. Validate and deduplicate
        validated_examples = self._validate_and_deduplicate(training_examples)
        
        # 4. Export to JSONL
        output_file = f"training_data_{datetime.now().strftime('%Y%m%d')}.jsonl"
        self._export_to_jsonl(validated_examples, output_file)
        
        # 5. Upload to Weights & Biases with versioning
        dataset_version = self._upload_to_wandb(output_file, validated_examples)
        
        # 6. Create data card
        self._create_data_card(validated_examples, dataset_version)
        
        print(f"🎉 Curation complete! Dataset version: {dataset_version}")
        
        return dataset_version
    
    def _fetch_positive_feedback(
        self,
        days: int = 7,
        min_score: float = 0.7
    ) -> List[Dict]:
        """Fetch positive feedback from Postgres."""
        query = """
        SELECT 
            id,
            interaction_id,
            user_id,
            session_id,
            feedback_type,
            feedback_value,
            metadata,
            model_version,
            created_at
        FROM feedback_events
        WHERE 
            created_at >= NOW() - INTERVAL '%s days'
            AND feedback_value >= %s
            AND feedback_type IN ('thumbs_up', 'rating', 'implicit')
        ORDER BY created_at DESC
        """
        
        cursor = self.pg_conn.cursor()
        cursor.execute(query, (days, min_score))
        
        columns = [desc[0] for desc in cursor.description]
        results = []
        for row in cursor.fetchall():
            results.append(dict(zip(columns, row)))
        
        return results
    
    async def _get_episode_from_lancedb(
        self,
        interaction_id: str
    ) -> Dict | None:
        """Retrieve full conversation from LanceDB episodic memory."""
        table = self.lancedb.open_table("episodic_memory")
        
        # Query by interaction_id (trace_id)
        results = table.search() \
            .where(f"metadata.trace_id = '{interaction_id}'") \
            .limit(1) \
            .to_list()
        
        if not results:
            return None
        
        episode = results[0]
        return {
            "user_query": episode.get("user_query", ""),
            "agent_response": episode.get("agent_response", ""),
            "sources": episode.get("sources", []),
            "timestamp": episode.get("timestamp")
        }
    
    def _validate_and_deduplicate(
        self,
        examples: List[TrainingExample]
    ) -> List[TrainingExample]:
        """Validate and remove duplicates."""
        # Remove duplicates by query
        seen_queries = set()
        unique_examples = []
        
        for example in examples:
            query_hash = hash(example.query.strip().lower())
            
            if query_hash not in seen_queries:
                seen_queries.add(query_hash)
                unique_examples.append(example)
        
        print(f"Removed {len(examples) - len(unique_examples)} duplicates")
        
        return unique_examples
    
    def _export_to_jsonl(
        self,
        examples: List[TrainingExample],
        output_file: str
    ):
        """Export to JSONL format for training."""
        with open(output_file, 'w') as f:
            for example in examples:
                # Convert to training format
                training_item = {
                    "messages": [
                        {"role": "user", "content": example.query},
                        {"role": "assistant", "content": example.response}
                    ],
                    "metadata": {
                        "feedback_score": example.feedback_score,
                        "sources": example.sources,
                        "timestamp": example.timestamp.isoformat(),
                        "model_version": example.model_version
                    }
                }
                f.write(json.dumps(training_item) + '\n')
        
        print(f"💾 Exported {len(examples)} examples to {output_file}")
    
    def _upload_to_wandb(
        self,
        file_path: str,
        examples: List[TrainingExample]
    ) -> str:
        """Upload to Weights & Biases with versioning."""
        # Initialize run
        run = wandb.init(
            project=self.wandb_project,
            job_type="data_curation",
            tags=["training_data", "weekly_curation"]
        )
        
        # Create artifact
        dataset_version = f"v{datetime.now().strftime('%Y%m%d')}"
        artifact = wandb.Artifact(
            name=f"training_data",
            type="dataset",
            description=f"Weekly curated training data from user feedback",
            metadata={
                "num_examples": len(examples),
                "date_range": f"last_7_days",
                "min_feedback_score": 0.7,
                "avg_feedback_score": sum(e.feedback_score for e in examples) / len(examples),
                "unique_users": len(set(e.metadata.get('user_id') for e in examples if 'user_id' in e.metadata))
            }
        )
        
        # Add file
        artifact.add_file(file_path)
        
        # Log artifact
        run.log_artifact(artifact, aliases=["latest", dataset_version])
        
        run.finish()
        
        print(f"☁️  Uploaded to W&B as {dataset_version}")
        
        return dataset_version
    
    def _create_data_card(
        self,
        examples: List[TrainingExample],
        dataset_version: str
    ):
        """Create data card documentation."""
        card_content = f"""# Training Data Card: {dataset_version}

## Dataset Statistics
- **Total Examples**: {len(examples)}
- **Date Range**: Last 7 days
- **Avg Feedback Score**: {sum(e.feedback_score for e in examples) / len(examples):.3f}
- **Feedback Score Distribution**:
  - Excellent (>0.9): {len([e for e in examples if e.feedback_score > 0.9])}
  - Good (0.7-0.9): {len([e for e in examples if 0.7 <= e.feedback_score <= 0.9])}

## Data Provenance
- **Source**: User feedback from Homepage application
- **Collection Method**: Explicit (thumbs up/rating) and implicit (copy, clicks) feedback
- **Quality Filter**: feedback_score >= 0.7
- **PII Redaction**: Automated (emails, IPs, usernames redacted)

## Schema
```json
{{
  "messages": [
    {{"role": "user", "content": "string"}},
    {{"role": "assistant", "content": "string"}}
  ],
  "metadata": {{
    "feedback_score": float,
    "sources": ["string"],
    "timestamp": "ISO8601",
    "model_version": "string"
  }}
}}
```

## Quality Checks
- ✅ No PII in queries or responses
- ✅ All examples have feedback_score >= 0.7
- ✅ Duplicates removed
- ✅ Minimum query length: 5 characters
- ✅ Minimum response length: 20 characters

## Usage
```python
import wandb

# Download dataset
run = wandb.init(project="agent-bruno")
artifact = run.use_artifact("training_data:{dataset_version}")
dataset_dir = artifact.download()

# Load JSONL
with open(f"{{dataset_dir}}/training_data.jsonl") as f:
    examples = [json.loads(line) for line in f]
```

## Limitations
- Only includes interactions with positive feedback
- May be biased towards certain query types
- Temporal distribution not balanced (recent bias)

**Generated**: {datetime.now().isoformat()}
"""
        
        with open(f"data_card_{dataset_version}.md", 'w') as f:
            f.write(card_content)
        
        print(f"📝 Data card created: data_card_{dataset_version}.md")

# Main curation script
async def main():
    """Run weekly curation pipeline."""
    pipeline = CurationPipeline(
        postgres_conn_string="postgresql://user:pass@postgres:5432/homepage",
        lancedb_path="/data/lancedb",
        wandb_project="agent-bruno"
    )
    
    dataset_version = await pipeline.curate_weekly_batch(min_feedback_score=0.7)
    
    print(f"\n✅ Curation pipeline complete!")
    print(f"   Dataset version: {dataset_version}")
    print(f"   Ready for fine-tuning!")

if __name__ == "__main__":
    asyncio.run(main())
```

### Kubernetes CronJob for Automated Curation

**Location**: `k8s/cronjobs/training-data-curation.yaml`

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: training-data-curation
  namespace: agent-bruno
spec:
  # Run every Sunday at 2 AM
  schedule: "0 2 * * 0"
  
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: curator
            image: agent-bruno:latest
            command:
            - python
            - /app/scripts/curate_training_data.py
            
            env:
            # Postgres connection
            - name: POSTGRES_HOST
              value: "postgres.homepage.svc.cluster.local"
            - name: POSTGRES_DB
              value: "homepage"
            - name: POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  name: postgres-credentials
                  key: username
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-credentials
                  key: password
            
            # LanceDB path
            - name: LANCEDB_PATH
              value: "/data/lancedb"
            
            # Weights & Biases
            - name: WANDB_API_KEY
              valueFrom:
                secretKeyRef:
                  name: wandb-credentials
                  key: api_key
            - name: WANDB_PROJECT
              value: "agent-bruno"
            
            volumeMounts:
            - name: lancedb-data
              mountPath: /data/lancedb
              readOnly: true
            - name: output
              mountPath: /output
          
          volumes:
          - name: lancedb-data
            persistentVolumeClaim:
              claimName: agent-bruno-lancedb-data-0
          - name: output
            emptyDir: {}
          
          restartPolicy: OnFailure
      
      # Keep last 3 successful jobs
      successfulJobsHistoryLimit: 3
      failedJobsHistoryLimit: 1
```

### Prometheus Metrics for Curation Pipeline

```python
# agent-bruno/metrics/curation.py
from prometheus_client import Counter, Gauge, Histogram

# Examples curated
training_examples_curated_total = Counter(
    'agent_training_examples_curated_total',
    'Total training examples curated from feedback',
    ['quality_tier']  # excellent, good, acceptable
)

# Curation pipeline duration
curation_pipeline_duration_seconds = Histogram(
    'agent_curation_pipeline_duration_seconds',
    'Time taken to run curation pipeline',
    buckets=[10, 30, 60, 120, 300, 600]
)

# Data quality metrics
curated_data_quality_score = Gauge(
    'agent_curated_data_quality_score',
    'Average quality score of curated data'
)

# Dataset size
curated_dataset_size = Gauge(
    'agent_curated_dataset_size',
    'Number of examples in latest dataset'
)
```

### Monitoring & Alerting

**Alert on curation failures**:

```yaml
- alert: CurationPipelineFailed
  expr: |
    time() - job_success_time{job="training-data-curation"} > 86400 * 8
  for: 1h
  labels:
    severity: high
    component: ml_pipeline
  annotations:
    summary: "Curation pipeline hasn't run successfully in 8 days"
    runbook: "https://wiki/runbooks/agent-bruno/curation-pipeline-failed"
```

**Alert on low data quality**:

```yaml
- alert: LowCurationQuality
  expr: agent_curated_data_quality_score < 0.7
  for: 10m
  labels:
    severity: warning
    component: ml_pipeline
  annotations:
    summary: "Curated data quality below threshold"
```

### Integration with Fine-tuning Pipeline

After curation, trigger fine-tuning automatically:

```python
# agent-bruno/ml/trigger_finetuning.py
async def trigger_finetuning_on_new_dataset(dataset_version: str):
    """Trigger fine-tuning when new curated dataset is available."""
    import wandb
    
    # Initialize run
    run = wandb.init(
        project="agent-bruno",
        job_type="fine_tuning",
        tags=["automated", "weekly"]
    )
    
    # Download curated dataset
    artifact = run.use_artifact(f"training_data:{dataset_version}")
    dataset_dir = artifact.download()
    
    # Load data
    training_data = load_jsonl(f"{dataset_dir}/training_data.jsonl")
    
    # Start fine-tuning (via Flyte workflow or direct)
    from agent_bruno.ml.finetuning import FineTuningPipeline
    
    pipeline = FineTuningPipeline(
        base_model="llama3.1:8b",
        training_data=training_data,
        hyperparameters={
            "learning_rate": 2e-5,
            "batch_size": 8,
            "num_epochs": 3,
            "lora_r": 16,
            "lora_alpha": 32
        }
    )
    
    # Run training
    model_artifact = await pipeline.train()
    
    # Log to wandb
    run.log_artifact(
        model_artifact,
        type="model",
        aliases=["candidate", f"finetuned_{dataset_version}"]
    )
    
    print(f"✅ Fine-tuning complete! Model: {model_artifact.name}")
    
    return model_artifact
```

---

## Summary

### Files to Create/Modify

#### Database
- ✅ `migrations/000006_create_feedback_events.up.sql`
- ✅ `migrations/000006_create_feedback_events.down.sql`

#### Frontend (Homepage)
- ✅ `frontend/src/components/FeedbackWidget.tsx`
- ✅ `frontend/src/services/feedback.ts`
- ✅ `frontend/src/services/implicitTracker.ts`
- ✅ `frontend/src/components/Chatbot.tsx` (modify)

#### Backend (Homepage API)
- ✅ `api/models/feedback.go`
- ✅ `api/handlers/feedback.go`
- ✅ `api/metrics/feedback.go`
- ✅ `api/router/router.go` (modify)

#### Agent Bruno
- ✅ `src/api/handlers.py` (modify to include trace_id)

#### ML Pipeline Integration
- ✅ `scripts/curate_training_data.py` (NEW - automated curation)
- ✅ `k8s/cronjobs/training-data-curation.yaml` (NEW - weekly job)
- ✅ `ml/trigger_finetuning.py` (NEW - auto-trigger training)

### Next Steps

1. **Create database migration** and run it
2. **Implement frontend components** (FeedbackWidget, ImplicitTracker)
3. **Implement backend handlers** (feedback.go)
4. **Update router** to expose /api/feedback endpoints
5. **Add trace_id to Agent Bruno responses**
6. **Deploy and test** in staging
7. **Monitor metrics** in Grafana
8. **Set up automated curation pipeline** (CronJob)
9. **Integrate with fine-tuning pipeline** (Flyte/manual)

---

**Document Version**: 1.0  
**Last Updated**: 2025-10-22  
**Status**: Implementation Guide (Not Yet Implemented)  
**Owner**: Full-Stack Team

---

## 📋 Document Review

**Review Completed By**: 
- [AI Senior SRE (Pending)]
- [AI Senior Pentester (Pending)]
- [AI Senior Cloud Architect (Pending)]
- [AI Senior Mobile iOS and Android Engineer (Pending)]
- [AI Senior DevOps Engineer (Pending)]
- ✅ **AI ML Engineer (COMPLETE)** - Added 600+ line automated curation pipeline
- [Bruno (Pending)]

**Review Date**: October 22, 2025  
**Document Status**: Under Review  
**Next Review**: TBD

---

