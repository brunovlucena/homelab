# Session Management & Long-term Memory

**[← Back to README](../README.md)** | **[Architecture](ARCHITECTURE.md)** | **[Memory](MEMORY.md)** | **[Observability](OBSERVABILITY.md)**

---

## Table of Contents
1. [Core Concepts](#core-concepts)
2. [User Authentication & Identity](#user-authentication--identity)
3. [Stateless Architecture with Stateful Memory](#stateless-architecture-with-stateful-memory)
4. [Session Lifecycle](#session-lifecycle)
5. [Long-term Memory Integration](#long-term-memory-integration)
6. [Multi-User Concurrency](#multi-user-concurrency)
7. [Request Flow Deep Dive](#request-flow-deep-dive)
8. [Storage Architecture](#storage-architecture)
9. [Performance & Optimization](#performance--optimization)
10. [Failure Scenarios & Recovery](#failure-scenarios--recovery)

---

## Core Concepts

### Stateless vs. Stateful: Clarification

**Important**: Agent Bruno's architecture is **stateless compute** with **stateful storage**. This is a critical distinction:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    STATELESS COMPUTE LAYER                              │
│  ┌────────────────────────────────────────────────────────────────┐     │
│  │  Agent Pods (ephemeral, replaceable)                           │     │
│  │  - No local state persistence                                  │     │
│  │  - Can be killed/restarted anytime                             │     │
│  │  - Any pod can handle any request                              │     │
│  │  - Scales horizontally without coordination                    │     │
│  └────────────────────────────────────────────────────────────────┘     │
└───────────────────────────┬─────────────────────────────────────────────┘
                            │
                            │ Fetch state at request start
                            │ Store state at request end
                            ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    STATEFUL STORAGE LAYER                               │
│  ┌────────────────────────────────────────────────────────────────┐     │
│  │  Persistent Storage (survives pod restarts)                    │     │
│  │  ├─ Redis: Short-term session state (1 hour TTL)               │     │
│  │  ├─ LanceDB: Long-term memory (persistent)                     │     │
│  │  └─ Minio/S3: Backups & archives                               │     │
│  └────────────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────────────┘
```

### The Key Principle

**Stateless Compute** = The application logic has no memory of previous requests
**Stateful Storage** = All state is externalized to databases that outlive any single pod

This allows:
- ✅ **Horizontal scaling**: Add more pods without state synchronization
- ✅ **High availability**: Pods can crash without losing user data
- ✅ **Load balancing**: Any pod can serve any user's request
- ✅ **Rolling updates**: Deploy new versions without state migration

---

## User Authentication & Identity

### Current State: Anonymous Users (IP-based)

**⚠️ IMPORTANT**: As of v1.0, Agent Bruno does **NOT** implement user authentication. The current implementation uses **IP addresses** as temporary user identifiers.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Current Implementation (v1.0)                        │
│                                                                         │
│  👤 User visits homepage (bruno.dev)                                    │
│  │                                                                      │
│  ├─ Browser sends request → Homepage API                                │
│  │  ├─ Client IP: 192.168.1.100                                         │
│  │  └─ No authentication header                                         │
│  │                                                                      │
│  └─ Homepage proxies to Agent Bruno:                                    │
│     POST /chat                                                          │
│     Headers:                                                            │
│       X-Forwarded-For: 192.168.1.100                                    │
│       Content-Type: application/json                                    │
│     Body:                                                               │
│       {                                                                 │
│         "message": "How do I fix Loki crashes?"                         │
│       }                                                                 │
│                                                                         │
│  🤖 Agent Bruno:                                                        │
│     user_id = request.headers["X-Forwarded-For"]  # "192.168.1.100"     │
│     session_id = f"session-{user_id}-{timestamp}"                       │
└─────────────────────────────────────────────────────────────────────────┘
```

**Limitations of IP-based identification**:
- ❌ Multiple users behind same NAT/proxy share same IP
- ❌ User IP changes (mobile roaming, VPN, etc.)
- ❌ No persistent identity across sessions
- ❌ No cross-device continuity
- ❌ Privacy concerns (IP addresses are PII in GDPR)

---

### Planned Implementation: JWT Authentication (v2.0)

#### Architecture Overview

```
┌────────────────────────────────────────────────────────────────────────┐
│              Proposed Authentication Flow (v2.0)                       │
│                                                                        │
│  ┌────────────────┐        ┌──────────────────┐                        │
│  │  Person A      │        │  Person B        │                        │
│  │  (First Visit) │        │  (Returning User)│                        │
│  └────────┬───────┘        └────────┬─────────┘                        │
│           │                          │                                 │
│           │ 1. Visit homepage        │ 1. Visit homepage               │
│           │ (No auth cookie)         │ (Has auth cookie)               │
│           ▼                          ▼                                 │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │           Homepage Frontend (React/TypeScript)                  │   │
│  │                                                                 │   │
│  │  useAuth() hook checks for JWT cookie                           │   │
│  │  ├─ Person A: No JWT → Auto-create anonymous user               │   │
│  │  └─ Person B: Valid JWT → Extract user_id from claims           │   │
│  └────────────────────────┬────────────────────────────────────────┘   │
│                           │                                            │
│                           ▼                                            │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │              Homepage API (Go)                                  │   │
│  │                                                                 │   │
│  │  Middleware: JWT Validation                                     │   │
│  │  ────────────────────────                                       │   │
│  │  if no_jwt:                                                     │   │
│  │      user = create_anonymous_user()                             │   │
│  │      jwt = sign_jwt({                                           │   │
│  │          "sub": user.id,           # user_id                    │   │
│  │          "type": "anonymous",                                   │   │
│  │          "created_at": now(),                                   │   │
│  │          "exp": now() + 30_days    # Expiration                 │   │
│  │      })                                                         │   │
│  │      set_cookie("auth_token", jwt, httponly=True, secure=True)  │   │
│  │                                                                 │   │
│  │  else if jwt.expired:                                           │   │
│  │      refresh_jwt(user_id)                                       │   │
│  │                                                                 │   │
│  │  else:                                                          │   │
│  │      user_id = jwt.claims["sub"]                                │   │
│  │                                                                 │   │
│  │  Forward to Agent Bruno with:                                   │   │
│  │  ├─ Header: Authorization: Bearer {jwt}                         │   │
│  │  └─ Body includes user_id from JWT claims                       │   │
│  └────────────────────────┬────────────────────────────────────────┘   │
│                           │                                            │
│                           ▼                                            │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │              Agent Bruno API (Python)                           │   │
│  │                                                                 │   │
│  │  Middleware: JWT Verification                                   │   │
│  │  ───────────────────────────                                    │   │
│  │  jwt_token = request.headers["Authorization"]                   │   │
│  │  claims = verify_jwt(jwt_token, public_key)                     │   │
│  │                                                                 │   │
│  │  if not claims.valid:                                           │   │
│  │      return 401 Unauthorized                                    │   │
│  │                                                                 │   │
│  │  # Extract user identity                                        │   │
│  │  user_id = claims["sub"]        # "user-a1b2c3d4"               │   │
│  │  user_type = claims["type"]     # "anonymous" or "registered"   │   │
│  │  session_id = request.body.get("session_id") or generate_new()  │   │
│  │                                                                 │   │
│  │  # Process request with authenticated user context              │   │
│  │  memory_context = fetch_memory(user_id, session_id)             │   │
│  │  response = agent.process(message, memory_context)              │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────────────────────────┘
```

---

### User Types & Lifecycle

```
┌────────────────────────────────────────────────────────────────────────┐
│                        User Identity Lifecycle                         │
│                                                                        │
│  1️⃣  Anonymous User (Default)                                          │
│  ─────────────────────────────                                         │
│  • Created automatically on first visit                                │
│  • No email, password, or registration required                        │
│  • Gets persistent user_id stored in JWT cookie                        │
│  • Can use all Agent Bruno features                                    │
│  • Memory persists for 30 days (cookie lifetime)                       │
│                                                                        │
│  Flow:                                                                 │
│  ┌───────────────────────────────────────────────────────────────┐     │
│  │  First Visit:                                                 │     │
│  │  → user_id = uuid4()  # "user-a1b2c3d4-5e6f-7890..."          │     │
│  │  → INSERT INTO users (id, type, created_at)                   │     │
│  │       VALUES (user_id, 'anonymous', NOW())                    │     │
│  │  → jwt = sign({sub: user_id, type: 'anonymous'})              │     │
│  │  → Set-Cookie: auth_token={jwt}; HttpOnly; Secure; Max-Age=30d│     │
│  │                                                               │     │
│  │  Subsequent Visits:                                           │     │
│  │  → Read JWT from cookie                                       │     │
│  │  → Extract user_id from claims                                │     │
│  │  → Fetch user's conversation history from LanceDB             │     │
│  └───────────────────────────────────────────────────────────────┘     │
│                                                                        │
│  2️⃣  Registered User (Future Feature)                                  │
│  ───────────────────────────────────                                   │
│  • User provides email/password or OAuth (Google, GitHub)              │
│  • Persistent identity across devices                                  │
│  • Enhanced features (API access, longer retention, etc.)              │
│  • Can sync anonymous history on registration                          │
│                                                                        │
│  Flow:                                                                 │
│  ┌──────────────────────────────────────────────────────────────┐      │
│  │  Registration (converts anonymous → registered):             │      │
│  │  → existing_user_id = jwt.claims["sub"]                      │      │
│  │  → UPDATE users                                              │      │
│  │       SET type = 'registered',                               │      │
│  │           email = 'user@example.com',                        │      │
│  │           email_verified = false                             │      │
│  │       WHERE id = existing_user_id                            │      │
│  │  → Send verification email                                   │      │
│  │  → New JWT with additional claims:                           │      │
│  │     {                                                        │      │
│  │       "sub": user_id,                                        │      │
│  │       "type": "registered",                                  │      │
│  │       "email": "user@example.com",                           │      │
│  │       "email_verified": true,                                │      │
│  │       "exp": now() + 90_days  # Longer expiration            │      │
│  │     }                                                        │      │
│  └──────────────────────────────────────────────────────────────┘      │
└────────────────────────────────────────────────────────────────────────┘
```

---

### JWT Token Structure & Best Practices

#### Token Format

```json
{
  "header": {
    "alg": "RS256",          // Algorithm: RSA with SHA-256
    "typ": "JWT",            // Type: JSON Web Token
    "kid": "key-2025-01"     // Key ID for rotation
  },
  "payload": {
    // Standard claims (RFC 7519)
    "iss": "bruno.dev",                    // Issuer
    "sub": "user-a1b2c3d4-5e6f-7890",      // Subject (user_id)
    "aud": ["bruno.dev", "agent-bruno"],   // Audience
    "exp": 1735689600,                     // Expiration (Unix timestamp)
    "nbf": 1703066400,                     // Not Before
    "iat": 1703066400,                     // Issued At
    "jti": "token-unique-id",              // JWT ID (for revocation)
    
    // Custom claims
    "type": "anonymous",                   // "anonymous" or "registered"
    "email": null,                         // Only for registered users
    "email_verified": false,
    "created_at": "2025-10-22T10:00:00Z",
    "session_preferences": {
      "theme": "dark",
      "language": "en"
    }
  },
  "signature": "..."
}
```

#### Best Practices ✅

```python
# ✅ 1. Use RSA-256 (Asymmetric) for distributed systems
# Why: Homepage signs JWTs, Agent Bruno only verifies
# Homepage needs private key, Agent Bruno only needs public key
# No shared secret to leak!

from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import rsa
import jwt

# Homepage (Go) - Signs tokens with PRIVATE key
private_key = load_private_key("/secrets/jwt-private-key.pem")
token = jwt.encode(
    payload={"sub": user_id, "type": "anonymous"},
    key=private_key,
    algorithm="RS256",
    headers={"kid": "key-2025-01"}  # Key rotation support
)

# Agent Bruno (Python) - Verifies tokens with PUBLIC key
public_key = load_public_key("/secrets/jwt-public-key.pem")
claims = jwt.decode(
    token,
    key=public_key,
    algorithms=["RS256"],
    audience=["agent-bruno"],
    issuer="bruno.dev"
)

# ✅ 2. Short-lived tokens with refresh mechanism
ACCESS_TOKEN_TTL = 15 * 60      # 15 minutes
REFRESH_TOKEN_TTL = 30 * 24 * 3600  # 30 days

# Access token (short-lived, frequent renewal)
access_token = create_jwt(user_id, ttl=ACCESS_TOKEN_TTL)

# Refresh token (long-lived, stored securely)
refresh_token = create_jwt(user_id, ttl=REFRESH_TOKEN_TTL, type="refresh")

# ✅ 3. Secure cookie storage (HttpOnly + Secure + SameSite)
response.set_cookie(
    key="auth_token",
    value=access_token,
    httponly=True,      # Prevents JavaScript access (XSS protection)
    secure=True,        # HTTPS only
    samesite="Strict",  # CSRF protection
    max_age=ACCESS_TOKEN_TTL,
    domain="bruno.dev",
    path="/"
)

# ✅ 4. Token revocation support (JTI claim)
# Store revoked token IDs in Redis with expiration
async def revoke_token(jti: str, exp: int):
    ttl = exp - int(time.time())  # Remaining lifetime
    await redis.setex(f"revoked:jwt:{jti}", ttl, "1")

async def is_token_revoked(jti: str) -> bool:
    return await redis.exists(f"revoked:jwt:{jti}")

# ✅ 5. Key rotation (KID claim)
# Maintain multiple public keys, rotate every 90 days
PUBLIC_KEYS = {
    "key-2025-01": load_public_key("key-2025-01.pem"),
    "key-2024-10": load_public_key("key-2024-10.pem"),  # Old key, still valid
}

def verify_jwt(token: str):
    header = jwt.get_unverified_header(token)
    kid = header["kid"]
    
    if kid not in PUBLIC_KEYS:
        raise InvalidTokenError("Unknown key ID")
    
    return jwt.decode(token, key=PUBLIC_KEYS[kid], algorithms=["RS256"])
```

#### Anti-Patterns ❌

```python
# ❌ 1. NEVER use HS256 (HMAC) for distributed systems
# Why: Requires sharing secret key between services
# If Agent Bruno can verify, it can also CREATE tokens (security risk!)
token = jwt.encode(payload, "shared-secret", algorithm="HS256")  # BAD!

# ❌ 2. NEVER store tokens in localStorage (XSS vulnerable)
// JavaScript
localStorage.setItem("token", jwt)  // BAD! Accessible to XSS attacks

# ❌ 3. NEVER put sensitive data in JWT payload
# JWTs are base64-encoded, NOT encrypted!
token = jwt.encode({
    "sub": user_id,
    "password": "secret123",  # ❌ NEVER!
    "ssn": "123-45-6789"      # ❌ NEVER!
}, private_key)

# ❌ 4. NEVER skip expiration validation
claims = jwt.decode(token, public_key, options={"verify_exp": False})  # BAD!

# ❌ 5. NEVER use long-lived tokens without refresh mechanism
token = jwt.encode(
    {"sub": user_id, "exp": now() + 365 * 24 * 3600},  # 1 year - BAD!
    private_key
)
```

---

### User Database Schema

```sql
-- Users table (PostgreSQL - managed by Homepage)
CREATE TABLE users (
    -- Identity
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(20) NOT NULL DEFAULT 'anonymous',  -- 'anonymous' | 'registered'
    
    -- Registered user fields (NULL for anonymous)
    email VARCHAR(255) UNIQUE,
    email_verified BOOLEAN DEFAULT FALSE,
    password_hash TEXT,  -- bcrypt or argon2
    
    -- OAuth fields (future)
    oauth_provider VARCHAR(50),  -- 'google' | 'github' | null
    oauth_id VARCHAR(255),
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    
    -- Privacy
    gdpr_consent BOOLEAN DEFAULT FALSE,
    gdpr_consent_date TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT valid_type CHECK (type IN ('anonymous', 'registered')),
    CONSTRAINT email_required_for_registered 
        CHECK (type = 'anonymous' OR email IS NOT NULL)
);

CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_type ON users(type);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Example rows
INSERT INTO users (id, type) VALUES 
    ('a1b2c3d4-5e6f-7890-1234-567890abcdef', 'anonymous');

INSERT INTO users (id, type, email, email_verified) VALUES
    ('b2c3d4e5-6f78-9012-3456-7890abcdef12', 'registered', 'alice@example.com', true);
```

---

### Implementation Checklist

**Phase 1: Anonymous Users (MVP)**
- [ ] Generate RSA key pair (private/public keys)
- [ ] Store keys in Kubernetes secrets
- [ ] Implement JWT signing in Homepage API (Go)
- [ ] Implement JWT verification in Agent Bruno (Python)
- [ ] Create `users` table in PostgreSQL
- [ ] Auto-create anonymous user on first visit
- [ ] Set HttpOnly secure cookie with JWT
- [ ] Update Agent Bruno to accept `user_id` from JWT claims
- [ ] Migrate existing IP-based sessions to user_id-based
- [ ] Update LanceDB queries to use `user_id` instead of IP

**Phase 2: User Registration (Future)**
- [ ] Add email/password registration UI
- [ ] Implement email verification flow
- [ ] Add OAuth providers (Google, GitHub)
- [ ] Allow upgrading anonymous → registered
- [ ] Sync anonymous user history on registration
- [ ] Implement refresh token mechanism
- [ ] Add user profile management

**Phase 3: Advanced Features (Future)**
- [ ] Cross-device session sync
- [ ] Multiple device management
- [ ] Session termination (logout all devices)
- [ ] JWT token revocation list
- [ ] Key rotation automation
- [ ] Rate limiting per user_id
- [ ] User analytics & insights

---

### Security Considerations

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Security Checklist                               │
│                                                                         │
│  ✅ Transport Security                                                  │
│  ├─ HTTPS only (TLS 1.3)                                                │
│  ├─ HSTS headers (Strict-Transport-Security)                            │
│  └─ Certificate pinning (optional)                                      │
│                                                                         │
│  ✅ Token Security                                                      │
│  ├─ RS256 algorithm (asymmetric)                                        │
│  ├─ Short expiration (15 minutes for access token)                      │
│  ├─ Refresh token rotation                                              │
│  ├─ Token revocation support (JTI claim)                                │
│  └─ Key rotation every 90 days                                          │
│                                                                         │
│  ✅ Cookie Security                                                     │
│  ├─ HttpOnly flag (XSS protection)                                      │
│  ├─ Secure flag (HTTPS only)                                            │
│  ├─ SameSite=Strict (CSRF protection)                                   │
│  └─ Domain scoping                                                      │
│                                                                         │
│  ✅ Input Validation                                                    │
│  ├─ Validate JWT signature                                              │
│  ├─ Validate exp, nbf, iat claims                                       │
│  ├─ Validate iss, aud claims                                            │
│  ├─ Check revocation list                                               │
│  └─ Rate limiting per user_id                                           │
│                                                                         │
│  ✅ Privacy & Compliance                                                │
│  ├─ GDPR consent tracking                                               │
│  ├─ Data retention policies (90 days)                                   │
│  ├─ Right to be forgotten (user deletion)                               │
│  ├─ Data export capability                                              │
│  └─ Audit logging for access                                            │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Stateless Architecture with Stateful Memory

### How It Works: The Pull Pattern

Instead of keeping state in memory, **each request pulls its required state** from external storage:

```
┌────────────────────────────────────────────────────────────────────────┐
│                        User Request Arrives                            │
│  POST /api/chat                                                        │
│  {                                                                     │
│    "user_id": "user-456",                                              │
│    "session_id": "session-abc123",                                     │
│    "message": "How do I fix Loki crashes?",                            │
│    "trace_id": "trace-xyz789"                                          │
│  }                                                                     │
└────────────────────────────┬───────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────┐
│                   Step 1: Fetch Session Context                        │
│                                                                        │
│  pod.process_request():                                                │
│    # Pod has NO prior knowledge of this user/session                   │
│    # Must fetch everything needed                                      │
│                                                                        │
│    session_context = redis.get(f"session:{session_id}")                │
│    # Returns:                                                          │
│    # {                                                                 │
│    #   "conversation_history": [last 10 messages],                     │
│    #   "user_preferences": {...},                                      │
│    #   "active_context": {...},                                        │
│    #   "last_updated": "2025-10-22T10:15:00Z"                          │
│    # }                                                                 │
└────────────────────────────┬───────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────┐
│                   Step 2: Fetch Long-term Memory                       │
│                                                                        │
│  # Query LanceDB for user's historical context                         │
│  episodic_memory = lancedb.query(                                      │
│      table="episodic_memory",                                          │
│      filters=f"user_id = '{user_id}' AND session_id = '{session_id}'", │
│      limit=20  # Last 20 conversation turns                            │
│  )                                                                     │
│                                                                        │
│  semantic_memory = lancedb.search(                                     │
│      table="semantic_memory",                                          │
│      vector=embed(query),  # Semantic search for relevant facts        │
│      filters=f"user_id = '{user_id}'",                                 │
│      limit=5                                                           │
│  )                                                                     │
│                                                                        │
│  procedural_memory = lancedb.query(                                    │
│      table="procedural_memory",                                        │
│      filters=f"user_id = '{user_id}'",                                 │
│      order_by="frequency DESC",                                        │
│      limit=10  # Most frequent preferences/patterns                    │
│  )                                                                     │
└────────────────────────────┬───────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────┐
│                   Step 3: Assemble Request Context                     │
│                                                                        │
│  request_context = {                                                   │
│      "current_query": message,                                         │
│      "session_context": session_context,  # From Redis                 │
│      "episodic_memory": episodic_memory,  # From LanceDB               │
│      "semantic_memory": semantic_memory,  # From LanceDB               │
│      "procedural_memory": procedural_memory,  # From LanceDB           │
│      "trace_id": trace_id                                              │
│  }                                                                     │
│                                                                        │
│  # NOW the pod has all the context it needs                            │
│  # The pod itself had no prior state - it fetched everything           │
└────────────────────────────┬───────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────┐
│                   Step 4: Process Request                              │
│                                                                        │
│  response = agent.process(request_context)                             │
│  # - Uses Hybrid RAG to retrieve knowledge                             │
│  # - Applies user preferences from procedural memory                   │
│  # - References past conversations from episodic memory                │
│  # - Uses semantic facts from semantic memory                          │
│  # - Calls Ollama LLM with full context                                │
└────────────────────────────┬───────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────┐
│                   Step 5: Persist New State                            │
│                                                                        │
│  # Update short-term session state (Redis)                             │
│  session_context["conversation_history"].append({                      │
│      "role": "user",                                                   │
│      "content": message,                                               │
│      "timestamp": now()                                                │
│  })                                                                    │
│  session_context["conversation_history"].append({                      │
│      "role": "assistant",                                              │
│      "content": response,                                              │
│      "timestamp": now()                                                │
│  })                                                                    │
│  session_context["last_updated"] = now()                               │
│                                                                        │
│  redis.setex(                                                          │
│      key=f"session:{session_id}",                                      │
│      value=json.dumps(session_context),                                │
│      ttl=3600  # 1 hour                                                │
│  )                                                                     │
│                                                                        │
│  # Store in long-term memory (LanceDB) - async background task         │
│  lancedb.insert(                                                       │
│      table="episodic_memory",                                          │
│      records=[{                                                        │
│          "vector": embed(f"{message} {response}"),                     │
│          "user_id": user_id,                                           │
│          "session_id": session_id,                                     │
│          "timestamp": now(),                                           │
│          "query": message,                                             │
│          "response": response,                                         │
│          "trace_id": trace_id                                          │
│      }]                                                                │
│  )                                                                     │
│                                                                        │
│  # Extract and store semantic facts (background)                       │
│  facts = extract_facts(message, response)                              │
│  for fact in facts:                                                    │
│      lancedb.insert(table="semantic_memory", records=[fact])           │
│                                                                        │
│  # Update procedural patterns (background)                             │
│  patterns = analyze_preferences(user_id, message, response)            │
│  lancedb.upsert(table="procedural_memory", records=patterns)           │
└────────────────────────────┬───────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────┐
│                   Step 6: Return Response                              │
│                                                                        │
│  return {                                                              │
│      "response": response,                                             │
│      "session_id": session_id,                                         │
│      "trace_id": trace_id                                              │
│  }                                                                     │
│                                                                        │
│  # Pod forgets everything - ready for next request from any user       │
└────────────────────────────────────────────────────────────────────────┘
```

### Key Insight

**The pod doesn't "remember" anything between requests.** Every request is a fresh start:

1. Receive request with identifiers (`user_id`, `session_id`)
2. **Pull** all required state from external storage
3. Process with full context
4. **Push** updated state back to storage
5. Discard all local state
6. Ready for next request (could be from a completely different user)

This is the **stateless pattern**: The pod is a pure function transformer, not a state holder.

---

## Session Lifecycle

### Session Creation

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    First Request from User                              │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────┐
│  Check if session_id exists                                            │
│                                                                        │
│  session = redis.get(f"session:{session_id}")                          │
│                                                                        │
│  if not session:                                                       │
│      # New session - initialize                                        │
│      session = {                                                       │
│          "session_id": session_id,                                     │
│          "user_id": user_id,                                           │
│          "created_at": now(),                                          │
│          "last_updated": now(),                                        │
│          "conversation_history": [],                                   │
│          "user_preferences": load_preferences(user_id),  # From LanceDB│
│          "active_context": {},                                         │
│          "metadata": {                                                 │
│              "platform": "homepage",                                   │
│              "client_version": "1.0.0"                                 │
│          }                                                             │
│      }                                                                 │
│                                                                        │
│      redis.setex(f"session:{session_id}", 3600, json.dumps(session))   │
│                                                                        │
│      # Also create session record in LanceDB for long-term tracking    │
│      lancedb.insert(table="sessions", records=[{                       │
│          "session_id": session_id,                                     │
│          "user_id": user_id,                                           │
│          "started_at": now(),                                          │
│          "platform": "homepage"                                        │
│      }])                                                               │
└────────────────────────────────────────────────────────────────────────┘
```

### Session Continuation

```
┌─────────────────────────────────────────────────────────────────────────┐
│                 Subsequent Request (Same Session)                       │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌────────────────────────────────────────────────────────────────────────┐
│  # Load existing session from Redis (fast path)                        │
│  session = redis.get(f"session:{session_id}")                          │
│                                                                        │
│  if session:                                                           │
│      # Hot session - continue conversation                             │
│      # Redis cache hit - conversation_history is immediately available │
│      return session                                                    │
│                                                                        │
│  else:                                                                 │
│      # Session expired from Redis (cold path)                          │
│      # Reconstruct from LanceDB long-term memory                       │
│                                                                        │
│      # Fetch episodic memory for this session                          │
│      history = lancedb.query(                                          │
│          table="episodic_memory",                                      │
│          filters=f"session_id = '{session_id}'",                       │
│          order_by="timestamp DESC",                                    │
│          limit=10                                                      │
│      )                                                                 │
│                                                                        │
│      # Reconstruct session state                                       │
│      session = {                                                       │
│          "session_id": session_id,                                     │
│          "user_id": user_id,                                           │
│          "conversation_history": [                                     │
│              {"role": "user", "content": turn.query}                   │
│              {"role": "assistant", "content": turn.response}           │
│              for turn in history                                       │
│          ],                                                            │
│          "reconstructed_from": "lancedb",                              │
│          "last_updated": now()                                         │
│      }                                                                 │
│                                                                        │
│      # Warm up Redis again                                             │
│      redis.setex(f"session:{session_id}", 3600, json.dumps(session))   │
│                                                                        │
│      return session                                                    │
└────────────────────────────────────────────────────────────────────────┘
```

### Session Expiration & Archival

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     Session Lifecycle Timeline                          │
│                                                                         │
│  0 min                 60 min                    90 days                │
│  │                     │                         │                      │
│  ▼                     ▼                         ▼                      │
│  ┌───────────────┐    ┌────────────────┐       ┌────────────────┐      │
│  │  Redis Cache  │───▶│ Expires (TTL)  │       │  LanceDB       │      │
│  │  (Hot)        │    │ Falls back to  │       │  (Persistent)  │      │
│  │               │    │ LanceDB        │       │                │      │
│  └───────────────┘    └────────────────┘       └────────────────┘      │
│                                                                         │
│                                                 After 90 days:          │
│                                                 Archive to Minio/S3     │
│                                                 (GDPR compliance)       │
└─────────────────────────────────────────────────────────────────────────┘
```

**Expiration Strategy**:

```python
# Redis TTL Strategy
REDIS_SESSION_TTL = 3600  # 1 hour

# LanceDB Retention
LANCEDB_RETENTION = 90  # 90 days (GDPR compliant)

# Minio/S3 Archival
ARCHIVE_AFTER = 90  # Archive to cold storage after 90 days
ARCHIVE_RETENTION = 365 * 7  # Keep archives for 7 years
```

---

## Long-term Memory Integration

### Three Memory Types

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        LanceDB Memory Architecture                      │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  1. Episodic Memory (Conversation History)                       │   │
│  │  ─────────────────────────────────────────────────────────────── │   │
│  │  "What did we talk about?"                                       │   │
│  │                                                                  │   │
│  │  Schema:                                                         │   │
│  │  - vector: embedding([user_query, assistant_response])           │   │
│  │  - user_id: string                                               │   │
│  │  - session_id: string                                            │   │
│  │  - timestamp: datetime                                           │   │
│  │  - query: string                                                 │   │
│  │  - response: string                                              │   │
│  │  - trace_id: string                                              │   │
│  │  - sentiment: float (-1 to 1)                                    │   │
│  │  - topic: string                                                 │   │
│  │                                                                  │   │
│  │  Retrieval:                                                      │   │
│  │  - By session_id: Get conversation flow                          │   │
│  │  - By user_id: Get user's history across sessions                │   │
│  │  - By semantic similarity: Related past conversations            │   │
│  │  - By time range: Recent vs historical context                   │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  2. Semantic Memory (Facts & Knowledge)                          │   │
│  │  ─────────────────────────────────────────────────────────────── │   │
│  │  "What do I know?"                                               │   │
│  │                                                                  │   │
│  │  Schema:                                                         │   │
│  │  - vector: embedding(fact)                                       │   │
│  │  - user_id: string (global facts if null)                        │   │
│  │  - entity_type: string (person, place, concept, etc.)            │   │
│  │  - fact: string                                                  │   │
│  │  - confidence: float (0-1)                                       │   │
│  │  - source: string (conversation, document, etc.)                 │   │
│  │  - extracted_at: datetime                                        │   │
│  │  - verified: boolean                                             │   │
│  │                                                                  │   │
│  │  Examples:                                                       │   │
│  │  - "User prefers Python over JavaScript"                         │   │
│  │  - "Loki runs in namespace 'loki'"                               │   │
│  │  - "User's timezone is America/Los_Angeles"                      │   │
│  │                                                                  │   │
│  │  Retrieval:                                                      │   │
│  │  - By semantic similarity: Relevant facts for current query      │   │
│  │  - By entity_type: All facts about a specific topic              │   │
│  │  - By user_id: User-specific vs global knowledge                 │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  3. Procedural Memory (Preferences & Patterns)                   │   │
│  │  ─────────────────────────────────────────────────────────────── │   │
│  │  "How do I behave?"                                              │   │
│  │                                                                  │   │
│  │  Schema:                                                         │   │
│  │  - vector: embedding(pattern_description)                        │   │
│  │  - user_id: string                                               │   │
│  │  - preference_type: string (response_style, tool_choice, etc.)   │   │
│  │  - preference_value: json                                        │   │
│  │  - frequency: int (how often this pattern appears)               │   │
│  │  - confidence: float (0-1)                                       │   │
│  │  - last_observed: datetime                                       │   │
│  │  - first_observed: datetime                                      │   │
│  │                                                                  │   │
│  │  Examples:                                                       │   │
│  │  - "Always include code examples"                                │   │
│  │  - "Prefers concise over verbose explanations"                   │   │
│  │  - "Often asks about Kubernetes troubleshooting"                 │   │
│  │  - "Uses kubectl commands frequently"                            │   │
│  │                                                                  │   │
│  │  Retrieval:                                                      │   │
│  │  - By frequency: Most common patterns                            │   │
│  │  - By recency: Recently learned preferences                      │   │
│  │  - By type: Specific preference categories                       │   │
│  └──────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

### Memory Retrieval Strategy

**Multi-stage retrieval for each request:**

```python
class MemoryRetriever:
    """Retrieves relevant memories for current request"""
    
    def get_context_for_request(
        self,
        user_id: str,
        session_id: str,
        current_query: str,
        lancedb: LanceDB,
        redis: Redis
    ) -> MemoryContext:
        """
        Fetches all relevant context for the current request.
        This is called on EVERY request - the pod has no prior memory.
        """
        
        # === STAGE 1: Recent Session History (Redis - Fast) ===
        recent_history = redis.get(f"session:{session_id}")
        if recent_history:
            recent_turns = json.loads(recent_history)["conversation_history"][-10:]
        else:
            # Fallback to LanceDB if Redis expired
            recent_turns = self._reconstruct_from_lancedb(session_id)
        
        # === STAGE 2: Episodic Memory (LanceDB - Semantic) ===
        # Get semantically similar past conversations
        query_embedding = embed(current_query)
        similar_episodes = lancedb.search(
            table="episodic_memory",
            vector=query_embedding,
            filters=f"user_id = '{user_id}'",
            limit=5,
            metric="cosine"
        )
        
        # === STAGE 3: Semantic Facts (LanceDB - Factual) ===
        # Extract entities from current query
        entities = extract_entities(current_query)
        
        relevant_facts = []
        for entity in entities:
            facts = lancedb.search(
                table="semantic_memory",
                vector=embed(entity),
                filters=f"user_id = '{user_id}' OR user_id IS NULL",  # User + global
                limit=3
            )
            relevant_facts.extend(facts)
        
        # === STAGE 4: User Preferences (LanceDB - Behavioral) ===
        user_preferences = lancedb.query(
            table="procedural_memory",
            filters=f"user_id = '{user_id}'",
            order_by="frequency DESC, last_observed DESC",
            limit=10
        )
        
        # === STAGE 5: Assemble Full Context ===
        return MemoryContext(
            recent_history=recent_turns,           # Last 10 messages
            similar_episodes=similar_episodes,     # 5 related past conversations
            relevant_facts=relevant_facts,         # ~10 relevant facts
            user_preferences=user_preferences,     # Top 10 patterns
            total_tokens=self._estimate_tokens(...)
        )
    
    def _reconstruct_from_lancedb(self, session_id: str) -> List[Dict]:
        """Cold start: Rebuild session from long-term memory"""
        turns = lancedb.query(
            table="episodic_memory",
            filters=f"session_id = '{session_id}'",
            order_by="timestamp DESC",
            limit=10
        )
        
        return [
            {"role": "user", "content": turn.query},
            {"role": "assistant", "content": turn.response}
            for turn in reversed(turns)  # Chronological order
        ]
```

### Memory Updates (Async Pattern)

**Critical**: Memory updates happen **asynchronously** to avoid blocking the response:

```python
async def process_request_with_memory(
    user_id: str,
    session_id: str,
    message: str
) -> str:
    """Main request handler"""
    
    # === SYNCHRONOUS: Fetch context (blocking) ===
    memory_context = memory_retriever.get_context_for_request(
        user_id, session_id, message, lancedb, redis
    )
    
    # === SYNCHRONOUS: Process request (blocking) ===
    response = agent.process(message, memory_context)
    
    # === SYNCHRONOUS: Update Redis session (fast, blocking OK) ===
    await redis.setex(
        f"session:{session_id}",
        3600,
        json.dumps({
            "conversation_history": memory_context.recent_history + [
                {"role": "user", "content": message},
                {"role": "assistant", "content": response}
            ],
            "last_updated": now()
        })
    )
    
    # === ASYNCHRONOUS: Update long-term memory (non-blocking) ===
    asyncio.create_task(
        update_long_term_memory(user_id, session_id, message, response)
    )
    
    # Return immediately - don't wait for LanceDB writes
    return response


async def update_long_term_memory(
    user_id: str,
    session_id: str,
    message: str,
    response: str
):
    """Background task: Update all three memory types"""
    
    try:
        # 1. Episodic Memory
        await lancedb.insert(
            table="episodic_memory",
            records=[{
                "vector": embed(f"{message} {response}"),
                "user_id": user_id,
                "session_id": session_id,
                "timestamp": now(),
                "query": message,
                "response": response,
                "sentiment": analyze_sentiment(response),
                "topic": classify_topic(message)
            }]
        )
        
        # 2. Semantic Memory (extract new facts)
        facts = await extract_facts_llm(message, response)
        if facts:
            await lancedb.insert(
                table="semantic_memory",
                records=[{
                    "vector": embed(fact),
                    "user_id": user_id,
                    "entity_type": fact.entity_type,
                    "fact": fact.text,
                    "confidence": fact.confidence,
                    "source": f"conversation:{session_id}",
                    "extracted_at": now()
                } for fact in facts]
            )
        
        # 3. Procedural Memory (update patterns)
        patterns = await analyze_interaction_patterns(
            user_id, message, response
        )
        for pattern in patterns:
            await lancedb.upsert(  # Upsert to increment frequency
                table="procedural_memory",
                records=[{
                    "vector": embed(pattern.description),
                    "user_id": user_id,
                    "preference_type": pattern.type,
                    "preference_value": pattern.value,
                    "frequency": pattern.frequency + 1,  # Increment
                    "last_observed": now()
                }]
            )
        
    except Exception as e:
        logger.error(f"Failed to update long-term memory: {e}")
        # Don't fail the user request - this is background processing
```

---

## Multi-User Concurrency

### Concurrent Request Handling

**Scenario**: 3 users send requests simultaneously to Agent Bruno

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Time: T0 (Simultaneous Requests)                     │
│                                                                         │
│  User A (session-123) ────┐                                             │
│  User B (session-456) ────┼──→  Kubernetes Load Balancer                │
│  User C (session-789) ────┘                                             │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│            Load Balancer Distributes to Available Pods                  │
│                                                                         │
│  Request A → Pod 1 (CPU: 30%, Memory: 512MB, Concurrency: 5/100)        │
│  Request B → Pod 1 (CPU: 31%, Memory: 520MB, Concurrency: 6/100)        │
│  Request C → Pod 2 (CPU: 25%, Memory: 480MB, Concurrency: 3/100)        │
│                                                                         │
│  Note: Same pod can handle multiple users concurrently                  │
│  Each request is completely isolated via user_id/session_id             │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                         Pod 1 Processing                                │
│                                                                         │
│  Thread 1: User A (session-123)                                         │
│  ├─ Fetch session:{session-123} from Redis → isolated                   │
│  ├─ Fetch user_id='user-a' memory from LanceDB → isolated               │
│  ├─ Process with User A's context only                                  │
│  └─ Write back to session:{session-123} → isolated                      │
│                                                                         │
│  Thread 2: User B (session-456)                                         │
│  ├─ Fetch session:{session-456} from Redis → isolated                   │
│  ├─ Fetch user_id='user-b' memory from LanceDB → isolated               │
│  ├─ Process with User B's context only                                  │
│  └─ Write back to session:{session-456} → isolated                      │
│                                                                         │
│  NO SHARED STATE between threads!                                       │
│  Each request has its own:                                              │
│  - Memory context (fetched independently)                               │
│  - Processing pipeline (isolated)                                       │
│  - Response generation (independent)                                    │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                    Storage Layer (Concurrent Access)                    │
│                                                                         │
│  Redis (Fast):                                                          │
│  ├─ session:session-123 ← Read/Write by User A's request                │
│  ├─ session:session-456 ← Read/Write by User B's request                │
│  └─ session:session-789 ← Read/Write by User C's request                │
│                                                                         │
│  LanceDB (Concurrent Reads):                                            │
│  ├─ SELECT * FROM episodic_memory WHERE user_id='user-a' (Thread 1)     │
│  ├─ SELECT * FROM episodic_memory WHERE user_id='user-b' (Thread 2)     │
│  └─ SELECT * FROM episodic_memory WHERE user_id='user-c' (Thread 3)     │
│                                                                         │
│  Key Isolation Mechanism:                                               │
│  - Different Redis keys (session:{id})                                  │
│  - Different LanceDB filters (user_id, session_id)                      │
│  - No cross-user data leakage                                           │
└─────────────────────────────────────────────────────────────────────────┘
```

### Isolation Guarantees

```python
# Every query MUST include user/session filters
# This prevents data leakage between users

# ✅ CORRECT: Filtered by user_id
episodic = lancedb.query(
    table="episodic_memory",
    filters=f"user_id = '{user_id}' AND session_id = '{session_id}'"
)

# ❌ WRONG: No filter - would return ALL users' data
episodic = lancedb.query(
    table="episodic_memory",
    filters=None  # DANGEROUS!
)

# ✅ CORRECT: Redis keys include session_id
session = redis.get(f"session:{session_id}")

# ❌ WRONG: Global key - shared state
session = redis.get("current_session")  # DANGEROUS!
```

### Race Condition Handling

**Scenario**: Same user sends 2 requests from different tabs

```
┌─────────────────────────────────────────────────────────────────────────┐
│     User A sends Request 1 and Request 2 simultaneously                 │
│     (Same session_id: session-123)                                      │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                 ┌───────────┴────────────┐
                 ▼                        ▼
          ┌─────────────┐          ┌─────────────┐
          │   Pod 1     │          │   Pod 2     │
          │   Req 1     │          │   Req 2     │
          └─────────────┘          └─────────────┘
                 │                        │
                 │ T0: Read session       │ T0: Read session
                 │ {"history": [A, B]}    │ {"history": [A, B]}
                 │                        │
                 │ T1: Add message C      │ T1: Add message D
                 │ {"history": [A,B,C]}   │ {"history": [A,B,D]}
                 │                        │
                 │ T2: Write session      │ T2: Write session
                 │ WINS (last write)      │ OVERWRITES Pod 1
                 └────────────────────────┴───────────────────▶
                                                    Result: [A, B, D]
                                                    Lost: C

Problem: Lost update (C is overwritten by D)
```

**Solution: Optimistic Locking with Versioning**

```python
class SessionManager:
    """Manages session state with concurrency control"""
    
    async def update_session(
        self,
        session_id: str,
        new_turn: Dict,
        max_retries: int = 3
    ) -> bool:
        """Update session with optimistic locking"""
        
        for attempt in range(max_retries):
            # 1. Read current session with version
            session_data = await redis.get(f"session:{session_id}")
            if not session_data:
                # No existing session - create new
                session = {"version": 0, "history": []}
            else:
                session = json.loads(session_data)
            
            current_version = session.get("version", 0)
            
            # 2. Modify session
            session["history"].append(new_turn)
            session["version"] = current_version + 1
            session["last_updated"] = now()
            
            # 3. Atomic write with version check (Lua script)
            success = await redis.eval("""
                local key = KEYS[1]
                local expected_version = ARGV[1]
                local new_data = ARGV[2]
                
                local current = redis.call('GET', key)
                if current == false then
                    -- No existing key - create
                    redis.call('SETEX', key, 3600, new_data)
                    return 1
                end
                
                local current_version = cjson.decode(current)['version']
                if current_version == tonumber(expected_version) then
                    -- Version matches - safe to update
                    redis.call('SETEX', key, 3600, new_data)
                    return 1
                else
                    -- Version mismatch - retry
                    return 0
                end
            """, 1, f"session:{session_id}", current_version, json.dumps(session))
            
            if success:
                return True
            
            # Version conflict - retry with backoff
            await asyncio.sleep(0.1 * (2 ** attempt))
        
        raise ConcurrencyError(f"Failed to update session after {max_retries} retries")
```

---

## Request Flow Deep Dive

### Complete Request Lifecycle

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    1. Request Arrives at API Gateway                    │
│  POST https://agent-api.bruno.dev/api/chat                              │
│  Authorization: Bearer <jwt_token>                                      │
│  {                                                                      │
│    "user_id": "user-456",                                               │
│    "session_id": "session-abc123",                                      │
│    "message": "How do I fix Loki crashes?",                             │
│    "context": {...}                                                     │
│  }                                                                      │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    2. Authentication & Rate Limiting                    │
│                                                                         │
│  # Validate JWT token                                                   │
│  claims = jwt.verify(token)                                             │
│  user_id = claims["sub"]                                                │
│                                                                         │
│  # Check rate limit (Redis)                                             │
│  rate_limit_key = f"ratelimit:{user_id}:minute"                         │
│  requests = redis.incr(rate_limit_key)                                  │
│  redis.expire(rate_limit_key, 60)                                       │
│                                                                         │
│  if requests > 100:  # Max 100 requests/minute                          │
│      return HTTPException(429, "Rate limit exceeded")                   │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    3. Initialize Tracing Context                        │
│                                                                         │
│  # Create trace for observability                                       │
│  trace_id = generate_trace_id()                                         │
│  span = tracer.start_span("agent.process_request")                      │
│  span.set_attribute("user_id", user_id)                                 │
│  span.set_attribute("session_id", session_id)                           │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    4. Fetch Session Context (L1 Cache)                  │
│                                                                         │
│  with span.child_span("memory.fetch_session"):                          │
│      # Try pod-local L1 cache first (LRU)                               │
│      session = l1_cache.get(f"session:{session_id}")                    │
│                                                                         │
│      if not session:                                                    │
│          # L1 miss - fetch from L2 (Redis)                              │
│          session = await redis.get(f"session:{session_id}")             │
│                                                                         │
│          if not session:                                                │
│              # L2 miss - reconstruct from LanceDB (cold start)          │
│              session = await reconstruct_session_from_lancedb(          │
│                  session_id, user_id                                    │
│              )                                                          │
│                                                                         │
│          # Warm up L1 cache                                             │
│          l1_cache.set(f"session:{session_id}", session, ttl=300)        │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    5. Fetch Long-term Memory (LanceDB)                  │
│                                                                         │
│  with span.child_span("memory.fetch_episodic"):                         │
│      # Semantic search for related past conversations                   │
│      query_vector = embed(message)                                      │
│      episodic = await lancedb.search(                                   │
│          table="episodic_memory",                                       │
│          vector=query_vector,                                           │
│          filters=f"user_id = '{user_id}'",                              │
│          limit=5,                                                       │
│          metric="cosine"                                                │
│      )                                                                  │
│                                                                         │
│  with span.child_span("memory.fetch_semantic"):                         │
│      # Get relevant facts                                               │
│      semantic = await lancedb.search(                                   │
│          table="semantic_memory",                                       │
│          vector=query_vector,                                           │
│          filters=f"user_id = '{user_id}' OR user_id IS NULL",           │
│          limit=5                                                        │
│      )                                                                  │
│                                                                         │
│  with span.child_span("memory.fetch_procedural"):                       │
│      # Get user preferences                                             │
│      procedural = await lancedb.query(                                  │
│          table="procedural_memory",                                     │
│          filters=f"user_id = '{user_id}'",                              │
│          order_by="frequency DESC",                                     │
│          limit=10                                                       │
│      )                                                                  │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    6. Assemble Full Context                             │
│                                                                         │
│  context = {                                                            │
│      "current_query": message,                                          │
│      "recent_history": session["conversation_history"][-10:],           │
│      "episodic_memory": episodic,  # Related past conversations         │
│      "semantic_facts": semantic,   # Relevant knowledge                 │
│      "user_preferences": procedural,  # Behavioral patterns             │
│      "trace_id": trace_id                                               │
│  }                                                                      │
│                                                                         │
│  # Token budget management                                              │
│  context_tokens = estimate_tokens(context)                              │
│  if context_tokens > MAX_CONTEXT_TOKENS:                                │
│      context = prioritize_and_truncate(context)                         │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    7. Hybrid RAG Retrieval                              │
│                                                                         │
│  with span.child_span("rag.search"):                                    │
│      # Search knowledge base for relevant docs                          │
│      rag_results = await hybrid_rag.search(                             │
│          query=message,                                                 │
│          user_context=context,                                          │
│          top_k=5                                                        │
│      )                                                                  │
│                                                                         │
│      # Combine with context                                             │
│      context["rag_results"] = rag_results                               │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    8. LLM Generation (Ollama)                           │
│                                                                         │
│  with span.child_span("llm.generate"):                                  │
│      # Build prompt with full context                                   │
│      prompt = build_prompt(                                             │
│          system_message=AGENT_SYSTEM_PROMPT,                            │
│          context=context,                                               │
│          user_query=message                                             │
│      )                                                                  │
│                                                                         │
│      # Call Ollama                                                      │
│      response = await ollama.generate(                                  │
│          model="llama3.2",                                              │
│          prompt=prompt,                                                 │
│          max_tokens=2048,                                               │
│          temperature=0.7                                                │
│      )                                                                  │
│                                                                         │
│      span.set_attribute("llm.tokens_in", response.tokens_in)            │
│      span.set_attribute("llm.tokens_out", response.tokens_out)          │
│      span.set_attribute("llm.latency_ms", response.latency_ms)          │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    9. Update Session State (Synchronous)                │
│                                                                         │
│  with span.child_span("session.update"):                                │
│      # Update conversation history                                      │
│      session["conversation_history"].extend([                           │
│          {"role": "user", "content": message, "timestamp": now()},      │
│          {"role": "assistant", "content": response, "timestamp": now()} │
│      ])                                                                 │
│                                                                         │
│      # Keep only last 20 messages (memory management)                   │
│      if len(session["conversation_history"]) > 20:                      │
│          session["conversation_history"] =                              │
│              session["conversation_history"][-20:]                      │
│                                                                         │
│      session["last_updated"] = now()                                    │
│      session["version"] += 1                                            │
│                                                                         │
│      # Write to Redis (atomic with version check)                       │
│      await session_manager.update_session(session_id, session)          │
│                                                                         │
│      # Update L1 cache                                                  │
│      l1_cache.set(f"session:{session_id}", session, ttl=300)            │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    10. Update Long-term Memory (Async)                  │
│                                                                         │
│  # Don't block response - fire and forget                               │
│  asyncio.create_task(                                                   │
│      update_long_term_memory(                                           │
│          user_id=user_id,                                               │
│          session_id=session_id,                                         │
│          message=message,                                               │
│          response=response,                                             │
│          trace_id=trace_id                                              │
│      )                                                                  │
│  )                                                                      │
│                                                                         │
│  # Background task will:                                                │
│  # - Insert into episodic_memory (conversation turn)                    │
│  # - Extract and insert semantic_memory (new facts)                     │
│  # - Update procedural_memory (pattern reinforcement)                   │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    11. Return Response to User                          │
│                                                                         │
│  span.end()                                                             │
│                                                                         │
│  return {                                                               │
│      "response": response,                                              │
│      "session_id": session_id,                                          │
│      "trace_id": trace_id,                                              │
│      "metadata": {                                                      │
│          "model": "llama3.2",                                           │
│          "tokens": response.tokens_out,                                 │
│          "latency_ms": total_latency                                    │
│      }                                                                  │
│  }                                                                      │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                    12. Pod State After Request                          │
│                                                                         │
│  # Pod has NO memory of this request                                    │
│  # All state was:                                                       │
│  #   - Fetched from Redis/LanceDB                                       │
│  #   - Processed                                                        │
│  #   - Written back to Redis/LanceDB                                    │
│  #   - Discarded from local memory                                      │
│  #                                                                      │
│  # Pod is now ready to handle ANY user's request                        │
│  # (Could be same user, could be different user)                        │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Storage Architecture

### Data Partitioning Strategy

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Storage Layer Architecture                           │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Redis (Hot Data - TTL: 1 hour)                                  │   │
│  │  ──────────────────────────────────────────────────────────────  │   │
│  │  Purpose: Active session state, fast reads                       │   │
│  │                                                                  │   │
│  │  Data:                                                           │   │
│  │  - session:{session_id}                                          │   │
│  │    * conversation_history (last 20 messages)                     │   │
│  │    * active_context                                              │   │
│  │    * last_updated                                                │   │
│  │                                                                  │   │
│  │  - ratelimit:{user_id}:minute                                    │   │
│  │    * Request counter for rate limiting                           │   │
│  │                                                                  │   │
│  │  Performance:                                                    │   │
│  │  - Read: <1ms                                                    │   │
│  │  - Write: <5ms                                                   │   │
│  │  - Size: ~100KB per session                                      │   │
│  │  - Max sessions: ~10,000 concurrent                              │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  LanceDB (Warm/Cold Data - Persistent)                           │   │
│  │  ──────────────────────────────────────────────────────────────  │   │
│  │  Purpose: Long-term memory, semantic search                      │   │
│  │                                                                  │   │
│  │  Tables:                                                         │   │
│  │                                                                  │   │
│  │  1. episodic_memory (Conversation History)                       │   │
│  │     - Size: ~100M records (1 million users × 100 conversations)  │   │
│  │     - Vector dimension: 768                                      │   │
│  │     - Index: IVF_PQ (Inverted File Product Quantization)         │   │
│  │     - Query time: 10-50ms (ANN search)                           │   │
│  │     - Filters: user_id, session_id, timestamp, topic             │   │
│  │                                                                  │   │
│  │  2. semantic_memory (Facts & Knowledge)                          │   │
│  │     - Size: ~10M records                                         │   │
│  │     - Vector dimension: 768                                      │   │
│  │     - Index: IVF_PQ                                              │   │
│  │     - Query time: 5-20ms                                         │   │
│  │     - Filters: user_id, entity_type, confidence                  │   │
│  │                                                                  │   │
│  │  3. procedural_memory (Preferences & Patterns)                   │   │
│  │     - Size: ~1M records                                          │   │
│  │     - Vector dimension: 768                                      │   │
│  │     - Index: BTree on (user_id, frequency)                       │   │
│  │     - Query time: <5ms (indexed lookup)                          │   │
│  │                                                                  │   │
│  │  Performance:                                                    │   │
│  │  - Vector search: 10-50ms                                        │   │
│  │  - Filtered query: 5-20ms                                        │   │
│  │  - Batch insert: 100-500ms (background)                          │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  Minio/S3 (Cold Storage - Archives)                              │   │
│  │  ──────────────────────────────────────────────────────────────  │   │
│  │  Purpose: Backups, compliance, disaster recovery                 │   │
│  │                                                                  │   │
│  │  Buckets:                                                        │   │
│  │  - agent-bruno-backups                                           │   │
│  │    * LanceDB snapshots (hourly)                                  │   │
│  │    * Redis snapshots (daily)                                     │   │
│  │    * Retention: 30 days                                          │   │
│  │                                                                  │   │
│  │  - agent-bruno-archives                                          │   │
│  │    * Conversation history (>90 days)                             │   │
│  │    * GDPR-compliant retention                                    │   │
│  │    * Retention: 7 years                                          │   │
│  └──────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

### Data Consistency Model

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Consistency Guarantees                               │
│                                                                         │
│  Redis (Session State):                                                 │
│  ──────────────────────────                                             │
│  - Consistency: Strong (single primary)                                 │
│  - Availability: HA with replication                                    │
│  - Partition tolerance: Sentinel for failover                           │
│  - CAP: CP (Consistency + Partition tolerance)                          │
│                                                                         │
│  Conflict Resolution:                                                   │
│  - Optimistic locking with version numbers                              │
│  - Last-write-wins on version match                                     │
│  - Retry on version mismatch                                            │
│                                                                         │
│  LanceDB (Long-term Memory):                                            │
│  ─────────────────────────────                                          │
│  - Consistency: Eventual (async writes)                                 │
│  - Availability: High (embedded DB per pod)                             │
│  - Partition tolerance: N/A (embedded)                                  │
│  - CAP: AP (Availability + Partition tolerance)                         │
│                                                                         │
│  Conflict Resolution:                                                   │
│  - Append-only episodic memory (no conflicts)                           │
│  - Upsert with increment for procedural (frequency counts)              │
│  - Merge with highest confidence for semantic (facts)                   │
│                                                                         │
│  Acceptable Tradeoffs:                                                  │
│  ────────────────────                                                   │
│  ✅ Session state: Must be immediately consistent (user expects it)      │
│  ✅ Long-term memory: Can be eventually consistent (historical data)     │
│  ✅ Lost memory write: Better than blocking user response                │
│  ✅ Duplicate memory: Deduplicated during retrieval (vector similarity)  │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Performance & Optimization

### Caching Strategy

```python
class MemoryCacheManager:
    """Multi-level cache for memory retrieval"""
    
    def __init__(self):
        # L1: In-memory LRU (per pod)
        self.l1_cache = LRUCache(maxsize=1000, ttl=300)  # 5 min TTL
        
        # L2: Redis (shared)
        self.redis = Redis(host="redis-cluster")
        
        # L3: LanceDB (persistent)
        self.lancedb = LanceDB(path="/data/lancedb")
    
    async def get_session_context(
        self,
        session_id: str,
        user_id: str
    ) -> Dict:
        """Fetch session with multi-level caching"""
        
        # L1 Cache (fastest - local memory)
        cache_key = f"session:{session_id}"
        session = self.l1_cache.get(cache_key)
        if session:
            logger.info("L1 cache hit", cache_key=cache_key)
            metrics.increment("cache.l1.hit")
            return session
        
        # L2 Cache (Redis - shared, still fast)
        session = await self.redis.get(cache_key)
        if session:
            logger.info("L2 cache hit", cache_key=cache_key)
            metrics.increment("cache.l2.hit")
            
            # Warm up L1
            self.l1_cache.set(cache_key, session)
            return session
        
        # L3 (LanceDB - persistent, slower)
        logger.info("Cache miss - reconstructing from LanceDB", session_id=session_id)
        metrics.increment("cache.miss")
        
        session = await self._reconstruct_from_lancedb(session_id, user_id)
        
        # Warm up L2 and L1
        await self.redis.setex(cache_key, 3600, json.dumps(session))
        self.l1_cache.set(cache_key, session)
        
        return session
    
    async def invalidate(self, session_id: str):
        """Invalidate cache on session update"""
        cache_key = f"session:{session_id}"
        
        # Clear L1
        self.l1_cache.delete(cache_key)
        
        # L2 will auto-expire or be overwritten on next write
```

### Query Optimization

```python
# ❌ BAD: Multiple sequential queries
async def get_memory_context_slow(user_id: str, query: str):
    episodic = await lancedb.search("episodic_memory", ...)
    semantic = await lancedb.search("semantic_memory", ...)
    procedural = await lancedb.query("procedural_memory", ...)
    
    return {
        "episodic": episodic,
        "semantic": semantic,
        "procedural": procedural
    }
    # Total time: 50ms + 20ms + 10ms = 80ms

# ✅ GOOD: Parallel queries
async def get_memory_context_fast(user_id: str, query: str):
    # Execute all queries in parallel
    episodic, semantic, procedural = await asyncio.gather(
        lancedb.search("episodic_memory", ...),
        lancedb.search("semantic_memory", ...),
        lancedb.query("procedural_memory", ...)
    )
    
    return {
        "episodic": episodic,
        "semantic": semantic,
        "procedural": procedural
    }
    # Total time: max(50ms, 20ms, 10ms) = 50ms (40% faster!)
```

### Token Budget Management

```python
class ContextManager:
    """Manages context size to fit within LLM token limits"""
    
    MAX_CONTEXT_TOKENS = 8192  # llama3.2 context window
    RESERVED_TOKENS = 2048     # For response generation
    
    def build_context(
        self,
        query: str,
        session: Dict,
        episodic: List,
        semantic: List,
        procedural: List,
        rag: List
    ) -> Dict:
        """Build context with token budget management"""
        
        available_tokens = self.MAX_CONTEXT_TOKENS - self.RESERVED_TOKENS
        
        # Priority order (most important first)
        priorities = [
            ("query", query, 1.0),                    # Always include
            ("recent_history", session["conversation_history"][-5:], 0.8),
            ("rag_results", rag[:3], 0.7),
            ("semantic_facts", semantic[:5], 0.6),
            ("episodic_memory", episodic[:3], 0.5),
            ("user_preferences", procedural[:5], 0.4)
        ]
        
        context = {}
        used_tokens = 0
        
        for key, value, priority in priorities:
            value_tokens = estimate_tokens(value)
            
            if used_tokens + value_tokens <= available_tokens:
                # Fits - include all
                context[key] = value
                used_tokens += value_tokens
            else:
                # Doesn't fit - truncate proportionally
                remaining = available_tokens - used_tokens
                if remaining > 100:  # Only include if meaningful
                    truncated = truncate_to_tokens(value, remaining)
                    context[key] = truncated
                    used_tokens += remaining
                break
        
        logger.info(
            "Context assembled",
            total_tokens=used_tokens,
            budget=available_tokens,
            utilization=f"{(used_tokens/available_tokens)*100:.1f}%"
        )
        
        return context
```

---

## Failure Scenarios & Recovery

### Pod Crash Recovery

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Scenario: Pod Crashes Mid-Request                    │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────────────┐
│  T0: User sends request → Pod 1                                         │
│  T1: Pod 1 processes request (fetches memory from Redis/LanceDB)        │
│  T2: Pod 1 calls Ollama LLM                                             │
│  T3: 💥 Pod 1 CRASHES (OOMKilled, node failure, etc.)                   │
│  T4: Kubernetes detects crash, starts new Pod 1'                        │
│  T5: User's client times out (30s), retries request                     │
│  T6: Load balancer routes to Pod 2 (or new Pod 1')                      │
│  T7: Pod 2 fetches session from Redis (unchanged - no write happened)   │
│  T8: Pod 2 processes request successfully                               │
│  T9: Pod 2 writes response to Redis                                     │
│  T10: User receives response                                            │
└─────────────────────────────────────────────────────────────────────────┘

Key Points:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ No data loss: Session state in Redis (external to pod)
✅ No corruption: Write only happened at the end (crashed before)
✅ Automatic recovery: Another pod handles retry
✅ User transparency: Client just sees higher latency
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Redis Failure

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Scenario: Redis Primary Fails                        │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────────────┐
│  T0: Redis primary crashes                                              │
│  T1: Redis Sentinel detects failure (5s health check)                   │
│  T2: Sentinel promotes replica to primary (automatic)                   │
│  T3: Agent pods reconnect to new primary (retry logic)                  │
│  T4: Total downtime: ~10-15 seconds                                     │
│                                                                         │
│  During failover (10-15s):                                              │
│  ─────────────────────────                                              │
│  - New requests: Fallback to LanceDB (slower but works)                 │
│  - Session reads: Reconstruct from episodic memory                      │
│  - Session writes: Buffered locally, flushed when Redis recovers        │
│                                                                         │
│  After failover:                                                        │
│  ──────────────                                                         │
│  - Redis available again (new primary)                                  │
│  - Sessions gradually repopulated (on-demand)                           │
│  - No data loss (replicas were in sync)                                 │
└─────────────────────────────────────────────────────────────────────────┘
```

### LanceDB Corruption

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Scenario: LanceDB Data Corruption                    │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────────────┐
│  Detection:                                                             │
│  - Checksum validation fails on read                                    │
│  - Query returns unexpected null results                                │
│  - Vector index errors                                                  │
│                                                                         │
│  Recovery Process:                                                      │
│  ────────────────                                                       │
│  1. Alert SRE team (PagerDuty)                                          │
│  2. Automatically stop writes to corrupted table                        │
│  3. Switch to read-only mode (graceful degradation)                     │
│  4. Restore from last hourly snapshot (Minio/S3)                        │
│     - Download snapshot: s3://backups/lancedb-2025-10-22-10:00.tar      │
│     - Extract to /data/lancedb-recovery                                 │
│     - Rebuild vector indexes (5-10 minutes)                             │
│  5. Switch traffic to recovered instance                                │
│  6. Replay missing data from Redis/audit logs                           │
│                                                                         │
│  RPO (Recovery Point Objective): 1 hour (snapshot frequency)            │
│  RTO (Recovery Time Objective): 15 minutes (restore + rebuild)          │
│                                                                         │
│  During recovery:                                                       │
│  - Agent continues to work with Redis sessions only                     │
│  - No long-term memory (degraded but functional)                        │
│  - User experience: Slightly less contextual responses                  │
└─────────────────────────────────────────────────────────────────────────┘
```

### Network Partition (Split Brain)

```
┌─────────────────────────────────────────────────────────────────────────┐
│            Scenario: Network Partition Between Pods & Redis             │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────────────┐
│  Partition:                                                             │
│  ┌────────────────┐                          ┌──────────────────┐       │
│  │  Pods 1-2      │  ❌ Network Split ❌     │  Redis + Pods 3-4│       │
│  │  (isolated)    │                          │  (can communicate)│       │
│  └────────────────┘                          └──────────────────┘       │
│                                                                         │
│  Behavior:                                                              │
│  ─────────                                                              │
│  Pods 1-2:                                                              │
│  - Cannot reach Redis                                                   │
│  - Fallback to LanceDB-only mode (read)                                 │
│  - Buffer writes locally (in-memory queue)                              │
│  - Return responses with warning: "degraded mode"                       │
│                                                                         │
│  Pods 3-4:                                                              │
│  - Normal operation (Redis accessible)                                  │
│  - Serve majority of traffic                                            │
│                                                                         │
│  Load Balancer:                                                         │
│  - Health checks detect Pods 1-2 degraded                               │
│  - Route all traffic to Pods 3-4                                        │
│                                                                         │
│  When partition heals:                                                  │
│  - Pods 1-2 reconnect to Redis                                          │
│  - Flush buffered writes (with version checks)                          │
│  - Resume normal operation                                              │
│  - Discard conflicts (last-write-wins in Redis)                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Summary: Stateless + Stateful Working Together

### The Full Picture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    Agent Bruno Memory Architecture                      │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                  STATELESS COMPUTE                               │   │
│  │  ────────────────────────────────────────────────────────────    │   │
│  │  - Pods can crash anytime (Kubernetes auto-restarts)             │   │
│  │  - Any pod can handle any user's request                         │   │
│  │  - No local state persistence                                    │   │
│  │  - Horizontal scaling without coordination                       │   │
│  │  - Load balancing across all pods                                │   │
│  │                                                                  │   │
│  │  Each Request:                                                   │   │
│  │  1. Pull state (user_id, session_id) from storage                │   │
│  │  2. Process with full context                                    │   │
│  │  3. Push updated state back to storage                           │   │
│  │  4. Discard all local state                                      │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                               ▲                                         │
│                               │                                         │
│                     Pull state │ Push state                             │
│                               │                                         │
│                               ▼                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                  STATEFUL STORAGE                                │   │
│  │  ────────────────────────────────────────────────────────────    │   │
│  │  Redis (Hot):                                                    │   │
│  │  - Active sessions (1 hour TTL)                                  │   │
│  │  - Recent conversation history                                   │   │
│  │  - Fast reads (<1ms), writes (<5ms)                              │   │
│  │                                                                  │   │
│  │  LanceDB (Persistent):                                           │   │
│  │  - Episodic memory (all conversations, 90 days)                  │   │
│  │  - Semantic memory (extracted facts)                             │   │
│  │  - Procedural memory (learned preferences)                       │   │
│  │  - Vector search (10-50ms)                                       │   │
│  │                                                                  │   │
│  │  Minio/S3 (Archive):                                             │   │
│  │  - Backups (hourly snapshots)                                    │   │
│  │  - Long-term archives (>90 days, GDPR)                           │   │
│  └──────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘

Benefits:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Scalability: Add pods without state migration
✅ Reliability: Pod failures don't lose user data
✅ Performance: Multi-level caching (L1 → L2 → L3)
✅ Consistency: Strong for sessions, eventual for memory
✅ Observability: Every request fully traced
✅ Multi-tenancy: User isolation via filters (user_id, session_id)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

**Document Version**: 1.0  
**Last Updated**: 2025-10-22  
**Owner**: SRE Team / Bruno

---

## 📋 Document Review

**Review Completed By**: 
- [AI Senior SRE (Pending)]
- [AI Senior Pentester (Pending)]
- [AI Senior Cloud Architect (Pending)]
- [AI Senior Mobile iOS and Android Engineer (Pending)]
- [AI Senior DevOps Engineer (Pending)]
- [AI ML Engineer (Pending)]
- [Bruno (Pending)]

**Review Date**: October 22, 2025  
**Document Status**: Under Review  
**Next Review**: TBD

---


