# 🏗️ Architecture Documentation

## Overview

Agent Store MultiBrands is a microservices-based e-commerce platform built on Knative and powered by AI sellers. The system uses event-driven architecture for loose coupling and scalability.

## System Components

### 1. WhatsApp Gateway (`whatsapp-gateway`)

**Purpose**: Interface between WhatsApp Business API and internal agents.

**Responsibilities**:
- Webhook verification for Meta
- Message parsing and routing
- Media handling (images, audio, video)
- Interactive message construction (buttons, lists)
- Delivery status tracking

**Event Flow**:
```
WhatsApp API → Gateway → store.whatsapp.message.received
                      → store.chat.message.new
                      
store.chat.response → Gateway → WhatsApp API
```

### 2. AI Sellers (`ai-seller-{brand}`)

**Purpose**: Intelligent sales agents for each brand vertical.

**Brands**:
| Brand | Agent Name | Personality |
|-------|------------|-------------|
| Fashion | Luna | Stylish, trend-aware |
| Tech | Max | Technical, precise |
| Home | Sofia | Warm, design-focused |
| Beauty | Bella | Caring, expert |
| Gaming | Pixel | Energetic, gamer |

**Responsibilities**:
- Customer conversation handling
- Product recommendations
- Cart management
- Escalation detection
- Sentiment analysis

**Event Flow**:
```
store.chat.message.new → AI Seller
                              ↓
                      ┌───────┴───────┐
                      ↓               ↓
            store.chat.response   store.sales.escalate
```

### 3. Sales Assistant (`sales-assistant`)

**Purpose**: AI-powered support for human sales representatives.

**Responsibilities**:
- Escalation queue management
- Real-time response suggestions
- Objection handling strategies
- Sales insight generation
- Performance analytics

### 4. Product Catalog (`product-catalog`)

**Purpose**: Central product inventory and search service.

**Responsibilities**:
- Product search and filtering
- Stock management
- Recommendation generation
- Price calculations

**Event Flow**:
```
store.product.query → Catalog → store.product.query.result
store.product.recommend → Catalog → store.product.recommend.result
```

### 5. Order Processor (`order-processor`)

**Purpose**: Order lifecycle management.

**Responsibilities**:
- Order creation and validation
- Status transitions
- Payment integration hooks
- Shipping tracking

## Event Architecture

### CloudEvents Standard

All inter-agent communication uses CloudEvents 1.0 specification:

```json
{
  "specversion": "1.0",
  "type": "store.chat.message.new",
  "source": "/agent-store-multibrands/whatsapp-gateway",
  "id": "uuid-here",
  "data": { ... }
}
```

### Event Types

#### WhatsApp Events
- `store.whatsapp.message.received` - Incoming message
- `store.whatsapp.message.send` - Outgoing message request
- `store.whatsapp.message.status` - Delivery status

#### Chat Events
- `store.chat.message.new` - New customer message
- `store.chat.response` - AI/human response
- `store.chat.intent.detected` - Detected customer intent

#### Product Events
- `store.product.query` - Search request
- `store.product.query.result` - Search results
- `store.product.recommend` - Recommendation request
- `store.product.recommend.result` - Recommendations

#### Order Events
- `store.order.create` - Order creation request
- `store.order.created` - Order created
- `store.order.confirmed` - Payment confirmed
- `store.order.shipped` - Order shipped
- `store.order.delivered` - Order delivered

#### Sales Events
- `store.sales.escalate` - Escalation to human
- `store.sales.insight` - Sales insight generated

## Data Flow

### Customer Purchase Journey

```
1. Customer sends WhatsApp message
   │
   ▼
2. Gateway receives and emits store.chat.message.new
   │
   ▼
3. AI Seller processes message with LLM
   │
   ├─► Queries products (store.product.query)
   │
   ├─► Gets recommendations (store.product.recommend)
   │
   └─► Generates response (store.chat.response)
   │
   ▼
4. Gateway sends response via WhatsApp
   │
   ▼
5. Customer adds to cart → AI Seller emits store.cart.item.added
   │
   ▼
6. Customer confirms purchase → AI Seller emits store.order.create
   │
   ▼
7. Order Processor creates order → Emits store.order.created
   │
   ▼
8. Gateway notifies customer of order confirmation
```

### Escalation Flow

```
1. AI Seller detects escalation trigger
   │
   ├─► Customer frustration (sentiment < 0.3)
   ├─► Complex query
   └─► Customer request for human
   │
   ▼
2. Emits store.sales.escalate
   │
   ▼
3. Sales Assistant receives and queues
   │
   ▼
4. Generates insights and suggestions
   │
   ▼
5. Human seller accepts via API
   │
   ▼
6. Sales Assistant provides real-time suggestions
```

## Scalability Design

### Horizontal Scaling

All agents are stateless and can scale horizontally:

| Agent | Min | Max | Target Concurrency |
|-------|-----|-----|-------------------|
| Gateway | 1 | 5 | 50 |
| AI Sellers | 0 | 10 | 5 |
| Sales Assistant | 1 | 5 | 10 |
| Product Catalog | 1 | 5 | 50 |
| Order Processor | 1 | 5 | 20 |

### Event Buffering

RabbitMQ handles event buffering with:
- Persistent message delivery
- Dead letter queues for failed events
- At-least-once delivery guarantee

## Security Considerations

### Authentication
- WhatsApp webhook signature verification
- Service account for K8s API access
- Secret management for credentials

### Data Privacy
- Customer phone numbers hashed in logs
- Conversation data not persisted long-term
- PII redaction in analytics events

## Monitoring & Observability

### Metrics (Prometheus)
- Message counts by brand/type
- Response latencies
- Order conversion rates
- AI token usage
- Escalation rates

### Tracing (OpenTelemetry)
- Distributed traces across all agents
- LLM inference spans
- Event processing spans

### Logging (Structured JSON)
- Correlation IDs
- Customer context (hashed)
- Event metadata
