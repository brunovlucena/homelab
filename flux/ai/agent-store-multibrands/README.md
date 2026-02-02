# 🏪 Agent Store MultiBrands

AI-powered multi-brand online store with WhatsApp integration using LambdaAgents.

## 🎯 Overview

This project implements an intelligent e-commerce platform where AI sellers assist customers via WhatsApp, while also providing tools for human sales representatives. The system uses event-driven architecture with Knative and LambdaAgents.

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Agent Store MultiBrands                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌──────────────┐    CloudEvents     ┌──────────────────────────┐         │
│   │   WhatsApp   │◄──────────────────►│    AI Seller Agents      │         │
│   │   Gateway    │                    │  ┌────────────────────┐  │         │
│   │              │                    │  │ 👗 Fashion Seller  │  │         │
│   │  📱 Handles  │                    │  │ 📱 Tech Seller     │  │         │
│   │  WhatsApp    │                    │  │ 🏠 Home Seller     │  │         │
│   │  Business    │                    │  │ 💄 Beauty Seller   │  │         │
│   │  API         │                    │  │ 🎮 Gaming Seller   │  │         │
│   └──────────────┘                    │  └────────────────────┘  │         │
│          │                            └──────────────────────────┘         │
│          │                                        │                         │
│          ▼                                        ▼                         │
│   ┌──────────────┐                    ┌──────────────────────────┐         │
│   │    Sales     │◄──────────────────►│   Product Catalog       │         │
│   │  Assistant   │                    │                          │         │
│   │              │                    │  📦 Manages inventory    │         │
│   │  🤝 Helps    │                    │  🏷️ Pricing engine       │         │
│   │  Human       │                    │  🔍 Search & recommend   │         │
│   │  Sellers     │                    └──────────────────────────┘         │
│   └──────────────┘                                │                         │
│          │                                        │                         │
│          └────────────────────┬───────────────────┘                         │
│                               ▼                                             │
│                    ┌──────────────────────────┐                             │
│                    │    Order Processor       │                             │
│                    │                          │                             │
│                    │  📋 Order management     │                             │
│                    │  💳 Payment integration  │                             │
│                    │  🚚 Shipping tracking    │                             │
│                    └──────────────────────────┘                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 🤖 AI Seller Agents

Each brand has a specialized AI seller with:

| Brand | Agent | Personality | Focus |
|-------|-------|-------------|-------|
| 👗 Fashion | `ai-seller-fashion` | Stylish, trend-aware | Clothing, accessories |
| 📱 Tech | `ai-seller-tech` | Knowledgeable, precise | Electronics, gadgets |
| 🏠 Home | `ai-seller-home` | Warm, helpful | Furniture, decor |
| 💄 Beauty | `ai-seller-beauty` | Glamorous, caring | Cosmetics, skincare |
| 🎮 Gaming | `ai-seller-gaming` | Enthusiastic, gamer | Games, consoles |

## 📱 WhatsApp Integration

The WhatsApp Gateway handles:
- Incoming messages from WhatsApp Business API
- Message routing to appropriate brand seller
- Media handling (product images, voice messages)
- Quick replies and interactive buttons
- Order confirmations and shipping updates

## 🤝 Human Seller Assistant

The Sales Assistant agent helps human sellers by:
- Providing real-time product information
- Suggesting upsell/cross-sell opportunities
- Handling complex customer queries
- Managing escalations from AI sellers
- Generating sales reports and insights

## 🔄 CloudEvents Flow

```
Customer (WhatsApp)
       │
       ▼ store.whatsapp.message.received
┌──────────────────┐
│ WhatsApp Gateway │
└──────────────────┘
       │
       ▼ store.chat.message.new
┌──────────────────┐
│   AI Seller      │ ◄──► store.product.query
└──────────────────┘      store.product.recommend
       │
       ├─► store.order.create
       │         │
       │         ▼
       │   ┌──────────────────┐
       │   │ Order Processor  │
       │   └──────────────────┘
       │
       ├─► store.sales.escalate
       │         │
       │         ▼
       │   ┌──────────────────┐
       │   │ Sales Assistant  │
       │   └──────────────────┘
       │
       └─► store.chat.response
                  │
                  ▼ store.whatsapp.message.send
           ┌──────────────────┐
           │ WhatsApp Gateway │
           └──────────────────┘
                  │
                  ▼
           Customer (WhatsApp)
```

## 🚀 Deployment

### Prerequisites
- Kubernetes cluster with Knative installed
- knative-lambda-operator deployed
- RabbitMQ cluster for eventing
- WhatsApp Business API credentials
- Ollama or OpenAI for LLM inference

### Deploy to Homelab

```bash
# Apply base configuration
kubectl apply -k k8s/kustomize/studio/

# Verify deployment
kubectl get lambdaagents -n agent-store-multibrands
```

### Configuration

1. **WhatsApp Credentials**: Create secret with Meta Business API tokens
2. **AI Configuration**: Configure Ollama endpoint or OpenAI API key
3. **Product Catalog**: Initialize with product data via ConfigMap or API

## 📊 Observability

All agents export:
- **Metrics**: Prometheus metrics for sales, conversations, response times
- **Traces**: OpenTelemetry traces via Alloy
- **Logs**: Structured JSON logs with trace context

### Key Dashboards
- Sales Performance Dashboard
- AI Seller Response Times
- Customer Satisfaction Metrics
- Product Recommendation Accuracy

## 🛠️ Development

### Local Development

```bash
# Install dependencies
pip install -r src/requirements.txt

# Run tests
pytest tests/ -v

# Run single agent locally
cd src/ai_seller && python main.py
```

### Adding New Brand

1. Create new seller prompt in `configmap-prompts.yaml`
2. Add LambdaAgent definition
3. Configure event subscriptions
4. Update routing in WhatsApp Gateway

## 📝 License

Copyright 2024 Bruno Lucena. Licensed under Apache License 2.0.
