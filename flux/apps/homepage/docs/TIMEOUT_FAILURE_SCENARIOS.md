# Database Query Timeout Failure Scenarios

## 🔴 Scenario 1: Database Connection Hang (Network Partition)

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Client    │         │   API Pod    │         │  PostgreSQL │
│  Browser    │         │  (Go App)    │         │   Database  │
└──────┬──────┘         └──────┬───────┘         └──────┬──────┘
       │                       │                        │
       │  GET /api/projects    │                        │
       │──────────────────────>│                        │
       │                       │                        │
       │                       │ db.Query("SELECT...") │
       │                       │───────────────────────>│
       │                       │                        │
       │                       │  ⚠️ NETWORK PARTITION │
       │                       │  (connection hangs)   │
       │                       │                        │
       │                       │  ❌ NO TIMEOUT!        │
       │                       │  Query waits forever  │
       │                       │                        │
       │  ⏳ Waiting...        │  ⏳ Waiting...         │  ⏳ Waiting...
       │  (30 seconds)         │  (forever)            │  (forever)
       │                       │                        │
       │  ⏳ Still waiting...  │  ⏳ Still waiting...   │  ⏳ Still waiting...
       │  (60 seconds)         │  (forever)            │  (forever)
       │                       │                        │
       │  ❌ TIMEOUT           │  ⏳ STILL WAITING!     │  ⏳ STILL WAITING!
       │  (browser gives up)   │  (connection stuck)   │  (connection stuck)
       │                       │                        │
       │                       │  🔒 CONNECTION POOL   │
       │                       │     EXHAUSTED!        │
       │                       │  (can't serve others) │
       │                       │                        │
       │                       │  💥 ALL REQUESTS     │
       │                       │     START FAILING     │
       │                       │                        │
```

**Impact:**
- Client waits 30-60s then times out
- API connection stuck forever
- Connection pool exhausted
- All subsequent requests fail
- **NO RECOVERY** until pod restart

---

## 🔴 Scenario 2: Slow Query (Database Under Load)

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Client    │         │   API Pod    │         │  PostgreSQL │
│  Browser    │         │  (Go App)    │         │   Database  │
└──────┬──────┘         └──────┬───────┘         └──────┬──────┘
       │                       │                        │
       │  GET /api/projects    │                        │
       │──────────────────────>│                        │
       │                       │                        │
       │                       │ db.Query("SELECT...") │
       │                       │───────────────────────>│
       │                       │                        │
       │                       │  ⚠️ DB IS SLOW         │
       │                       │  (high CPU/IO wait)    │
       │                       │                        │
       │                       │  ❌ NO TIMEOUT!        │
       │                       │  Query waits...        │
       │                       │                        │
       │  ⏳ Waiting...        │  ⏳ Waiting...         │  🔄 Processing...
       │  (5 seconds)          │  (5 seconds)           │  (slow query)
       │                       │                        │
       │  ⏳ Still waiting...  │  ⏳ Still waiting...   │  🔄 Still processing...
       │  (10 seconds)         │  (10 seconds)          │  (still slow)
       │                       │                        │
       │  ⏳ Still waiting...  │  ⏳ Still waiting...   │  🔄 Still processing...
       │  (30 seconds)         │  (30 seconds)          │  (very slow)
       │                       │                        │
       │  ❌ TIMEOUT           │  ⏳ STILL WAITING!     │  🔄 Still processing...
       │  (browser gives up)   │  (query still running) │  (query still running)
       │                       │                        │
       │                       │  🔒 CONNECTION HELD    │
       │                       │     FOR 2+ MINUTES!    │
       │                       │                        │
       │                       │  💥 POOL EXHAUSTED    │
       │                       │     (25 connections)  │
       │                       │                        │
       │                       │  ❌ NEW REQUESTS      │
       │                       │     CAN'T GET CONN    │
       │                       │                        │
```

**Impact:**
- Client times out after 30-60s
- Query continues running for minutes
- Connection pool exhausted
- New requests can't get connections
- **CASCADING FAILURE**

---

## 🔴 Scenario 3: Database Deadlock

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Client    │         │   API Pod    │         │  PostgreSQL │
│  Browser    │         │  (Go App)    │         │   Database  │
└──────┬──────┘         └──────┬───────┘         └──────┬──────┘
       │                       │                        │
       │  GET /api/projects    │                        │
       │──────────────────────>│                        │
       │                       │                        │
       │                       │ db.Query("SELECT...") │
       │                       │───────────────────────>│
       │                       │                        │
       │                       │  ⚠️ DEADLOCK!          │
       │                       │  (waiting for lock)   │
       │                       │                        │
       │                       │  ❌ NO TIMEOUT!        │
       │                       │  Query waits forever  │
       │                       │                        │
       │  ⏳ Waiting...        │  ⏳ Waiting...         │  🔒 Locked
       │  (10 seconds)         │  (10 seconds)          │  (deadlock)
       │                       │                        │
       │  ⏳ Still waiting...  │  ⏳ Still waiting...   │  🔒 Still locked
       │  (30 seconds)         │  (30 seconds)          │  (deadlock)
       │                       │                        │
       │  ⏳ Still waiting...  │  ⏳ Still waiting...   │  🔒 Still locked
       │  (60 seconds)         │  (60 seconds)          │  (deadlock)
       │                       │                        │
       │  ❌ TIMEOUT           │  ⏳ STILL WAITING!     │  🔒 STILL LOCKED!
       │  (browser gives up)   │  (query stuck)         │  (deadlock)
       │                       │                        │
       │                       │  🔒 CONNECTION HELD    │
       │                       │     FOREVER!           │
       │                       │                        │
       │                       │  💥 POOL EXHAUSTED    │
       │                       │                        │
```

**Impact:**
- Query waits forever for lock
- Connection never released
- Pool exhausted
- **REQUIRES MANUAL INTERVENTION** (kill query or restart)

---

## 🔴 Scenario 4: Multiple Concurrent Requests (Cascading Failure)

```
Time: 0s
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Client    │         │   API Pod    │         │  PostgreSQL │
│  Browser    │         │  (Go App)    │         │   Database  │
└──────┬──────┘         └──────┬───────┘         └──────┬──────┘
       │                       │                        │
       │  Request 1           │                        │
       │──────────────────────>│                        │
       │  Request 2            │                        │
       │──────────────────────>│                        │
       │  Request 3            │                        │
       │──────────────────────>│                        │
       │  ... (25 requests)    │                        │
       │──────────────────────>│                        │
       │                       │                        │
       │                       │  ❌ ALL QUERIES        │
       │                       │     WITHOUT TIMEOUT!   │
       │                       │                        │
       │                       │  🔒 ALL 25 CONNECTIONS │
       │                       │     IN USE             │
       │                       │                        │
       │                       │  ⚠️ DB IS SLOW         │
       │                       │  (all queries slow)    │
       │                       │                        │

Time: 30s
       │  ⏳ All waiting...     │  ⏳ All waiting...     │  🔄 All processing...
       │  (30 seconds)         │  (30 seconds)          │  (all slow)
       │                       │                        │
       │  ❌ Request 1 timeout │  ⏳ Still waiting...   │  🔄 Still processing...
       │  ❌ Request 2 timeout │  ⏳ Still waiting...   │  🔄 Still processing...
       │  ❌ Request 3 timeout │  ⏳ Still waiting...   │  🔄 Still processing...
       │  ...                  │  ...                   │  ...
       │                       │                        │
       │  Request 26 (NEW)      │  ❌ NO CONNECTION!    │
       │──────────────────────>│  (pool exhausted)     │
       │                       │                        │
       │  Request 27 (NEW)      │  ❌ NO CONNECTION!    │
       │──────────────────────>│  (pool exhausted)     │
       │                       │                        │
       │  💥 ALL NEW REQUESTS  │  💥 ALL FAILING!      │
       │     FAILING!          │                        │
       │                       │                        │

Time: 2 minutes
       │                       │  ⏳ STILL WAITING!     │  🔄 STILL PROCESSING!
       │                       │  (2 minutes later)     │  (2 minutes later)
       │                       │                        │
       │                       │  🔒 ALL 25 CONNECTIONS │
       │                       │     STILL HELD!        │
       │                       │                        │
       │                       │  💥 COMPLETE OUTAGE!   │
       │                       │                        │
```

**Impact:**
- All connections exhausted
- New requests fail immediately
- **COMPLETE SERVICE OUTAGE**
- No recovery until queries finish or pod restarts

---

## ✅ Scenario 5: WITH TIMEOUT (Correct Behavior)

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Client    │         │   API Pod    │         │  PostgreSQL │
│  Browser    │         │  (Go App)    │         │   Database  │
└──────┬──────┘         └──────┬───────┘         └──────┬──────┘
       │                       │                        │
       │  GET /api/projects    │                        │
       │──────────────────────>│                        │
       │                       │                        │
       │                       │ ctx, cancel :=         │
       │                       │   WithTimeout(5s)      │
       │                       │                        │
       │                       │ QueryRowContext(ctx)   │
       │                       │───────────────────────>│
       │                       │                        │
       │                       │  ⚠️ DB IS SLOW         │
       │                       │  (high CPU/IO wait)   │
       │                       │                        │
       │  ⏳ Waiting...        │  ⏳ Waiting...         │  🔄 Processing...
       │  (1 second)           │  (1 second)            │  (slow query)
       │                       │                        │
       │  ⏳ Still waiting...  │  ⏳ Still waiting...   │  🔄 Still processing...
       │  (3 seconds)           │  (3 seconds)           │  (still slow)
       │                       │                        │
       │  ⏳ Still waiting...  │  ⏳ Still waiting...   │  🔄 Still processing...
       │  (5 seconds)          │  (5 seconds)           │  (still slow)
       │                       │                        │
       │                       │  ⏰ TIMEOUT!            │
       │                       │  ctx.Err() ==          │
       │                       │    DeadlineExceeded    │
       │                       │                        │
       │                       │  ✅ CANCEL QUERY       │
       │                       │  (connection released) │
       │                       │                        │
       │                       │  📊 Record metrics     │
       │                       │     (timeout=true)     │
       │                       │                        │
       │  504 Gateway Timeout  │                        │
       │<──────────────────────│                        │
       │                       │                        │
       │  ✅ CLIENT GETS       │  ✅ CONNECTION        │
       │     RESPONSE          │     RELEASED           │
       │  (can retry)          │  (available for next)  │
       │                       │                        │
```

**Benefits:**
- Client gets response in 5 seconds
- Connection released immediately
- Pool stays healthy
- Client can retry
- **SERVICE STAYS AVAILABLE**

---

## 📊 Comparison Table

| Scenario | Without Timeout | With Timeout (5s) |
|----------|----------------|-------------------|
| **Network Partition** | ❌ Hangs forever | ✅ Fails after 5s |
| **Slow Query** | ❌ Hangs for minutes | ✅ Fails after 5s |
| **Deadlock** | ❌ Hangs forever | ✅ Fails after 5s |
| **Connection Pool** | ❌ Exhausted | ✅ Stays healthy |
| **New Requests** | ❌ All fail | ✅ Can proceed |
| **Recovery** | ❌ Requires restart | ✅ Automatic |
| **User Experience** | ❌ 30-60s wait | ✅ 5s max wait |

---

## 🎯 Key Takeaways

1. **Without timeout**: One slow query can kill the entire service
2. **With timeout**: Slow queries fail fast, service stays available
3. **Connection pool**: Without timeout, pool gets exhausted quickly
4. **Cascading failure**: One problem becomes many problems
5. **Recovery**: With timeout, automatic recovery; without, manual intervention needed

---

## 🔧 Current Status

**Fixed (with timeout):**
- ✅ `getSiteConfig()` - has 5s timeout
- ✅ `getAbout()` - has 5s timeout  
- ✅ `updateSiteConfig()` - has 5s timeout

**Broken (no timeout):**
- ❌ `getProjects()` - NO timeout (25+ queries)
- ❌ `getProject()` - NO timeout
- ❌ `createProject()` - NO timeout
- ❌ `updateProject()` - NO timeout
- ❌ `deleteProject()` - NO timeout
- ❌ All skills endpoints - NO timeout
- ❌ All experience endpoints - NO timeout
- ❌ All content endpoints - NO timeout
- ❌ ContextBuilder queries - NO timeout (affects LLM chat)

**Total: 30+ queries without timeout = 30+ ways to crash the service**
