# 📱 WhatsApp Business API Setup Guide

This guide walks you through setting up WhatsApp Business API integration for the Multi-Brand Store.

## Prerequisites

- Meta Business Account
- Meta Developer Account
- Phone number for WhatsApp Business
- HTTPS endpoint accessible from the internet (for webhook)

## Step 1: Create Meta App

1. Go to [Meta for Developers](https://developers.facebook.com/)
2. Click **My Apps** → **Create App**
3. Select **Business** as the app type
4. Fill in the app details:
   - App name: `MultiBrand Store WhatsApp`
   - Contact email: Your email
   - Business Account: Select your business

## Step 2: Add WhatsApp Product

1. In your app dashboard, click **Add Product**
2. Find **WhatsApp** and click **Set up**
3. You'll see the **WhatsApp Business** configuration page

## Step 3: Get API Credentials

### Access Token (Temporary)
1. Go to **WhatsApp** → **API Setup**
2. Copy the **Temporary access token** (expires in 24h)

### Permanent Token
1. Go to **Business Settings** → **System Users**
2. Create a new System User with `Admin` role
3. Add the WhatsApp app to this user
4. Generate a token with permissions:
   - `whatsapp_business_management`
   - `whatsapp_business_messaging`

### Phone Number ID
1. In **WhatsApp** → **API Setup**
2. Copy the **Phone number ID** (not the phone number itself)

### WhatsApp Business Account ID
1. In **WhatsApp** → **API Setup**  
2. Copy the **WhatsApp Business Account ID**

## Step 4: Configure Webhook

### Get Your Webhook URL
Your webhook URL will be:
```
https://store-whatsapp.homelab.local/webhook
```

For local development, use [ngrok](https://ngrok.com/) or similar:
```bash
ngrok http 8080
# Use the HTTPS URL provided by ngrok
```

### Configure in Meta Developer Console

1. Go to **WhatsApp** → **Configuration**
2. Click **Edit** on the Webhook section
3. Enter:
   - **Callback URL**: `https://your-domain/webhook`
   - **Verify Token**: A secret string of your choice (e.g., `multibrands-store-verify-2024`)
4. Click **Verify and Save**

### Subscribe to Webhook Fields

After verification, subscribe to these fields:
- ✅ `messages` - Receive incoming messages
- ✅ `message_status` - Delivery/read receipts
- ✅ `message_template_status_update` - Template approvals

## Step 5: Update Kubernetes Secrets

Create a file with your credentials:

```bash
cat > /tmp/whatsapp-secrets.yaml << 'EOF'
apiVersion: v1
kind: Secret
metadata:
  name: whatsapp-credentials
  namespace: agent-store-multibrands
type: Opaque
stringData:
  # Your permanent access token
  WHATSAPP_ACCESS_TOKEN: "EAAxxxxxxxxxxxxxxxx"
  
  # Your phone number ID (not the phone number)
  WHATSAPP_PHONE_NUMBER_ID: "123456789012345"
  
  # Your WhatsApp Business Account ID
  WHATSAPP_BUSINESS_ACCOUNT_ID: "987654321098765"
  
  # The verify token you set in Step 4
  WHATSAPP_VERIFY_TOKEN: "multibrands-store-verify-2024"
  
  # Optional: App Secret for signature verification
  WHATSAPP_APP_SECRET: "your-app-secret"
EOF
```

Apply the secret:
```bash
kubectl apply -f /tmp/whatsapp-secrets.yaml
# Clean up
rm /tmp/whatsapp-secrets.yaml
```

## Step 6: Test the Integration

### Verify Webhook is Working

```bash
# Check the gateway logs
kubectl logs -n agent-store-multibrands -l app.kubernetes.io/name=whatsapp-gateway -f

# Test health endpoint
curl -s http://whatsapp-gateway.agent-store-multibrands.svc.cluster.local/health
```

### Send a Test Message

1. Open WhatsApp on your phone
2. Send a message to your WhatsApp Business number
3. Check the logs for the incoming message
4. You should receive a response from the AI seller

## Step 7: Register Message Templates (Optional)

For proactive messages (outside 24h window), create message templates:

1. Go to **WhatsApp** → **Message Templates**
2. Create templates for:
   - Order confirmation
   - Shipping updates
   - Promotional messages

Example template:
```
Name: order_confirmation
Category: UTILITY
Language: pt_BR

Body:
Olá {{1}}! 🛍️

Seu pedido #{{2}} foi confirmado!

📦 Total: R$ {{3}}
📍 Entrega em até {{4}} dias úteis

Obrigado por comprar conosco! ✨
```

## Troubleshooting

### Webhook Verification Fails

1. Ensure your endpoint returns the `hub.challenge` parameter
2. Check if the verify token matches exactly
3. Verify HTTPS is properly configured

### Messages Not Being Received

1. Check webhook subscriptions are active
2. Verify the phone number is properly connected
3. Check firewall rules allow Meta IPs

### Response Not Being Sent

1. Check access token hasn't expired
2. Verify phone number ID is correct
3. Check API rate limits (1000 messages/phone/day for test)

### Common Error Codes

| Code | Description | Solution |
|------|-------------|----------|
| 130429 | Rate limit hit | Wait and retry, or request higher limits |
| 131031 | Business account restricted | Contact Meta support |
| 132000 | Invalid parameters | Check message format |
| 368 | Temporarily blocked | Wait 24-72 hours |

## Production Checklist

- [ ] Permanent access token configured
- [ ] Webhook URL using HTTPS with valid certificate
- [ ] Rate limiting configured on ingress
- [ ] Monitoring and alerting set up
- [ ] Message templates approved
- [ ] Business verification completed
- [ ] Privacy policy URL configured
- [ ] Customer support contact configured

## Useful Links

- [WhatsApp Business API Documentation](https://developers.facebook.com/docs/whatsapp/cloud-api)
- [Webhook Reference](https://developers.facebook.com/docs/whatsapp/cloud-api/webhooks)
- [Message Templates](https://developers.facebook.com/docs/whatsapp/message-templates)
- [Error Codes](https://developers.facebook.com/docs/whatsapp/cloud-api/support/error-codes)

## Security Recommendations

1. **Validate Webhook Signatures**: Enable and verify the `X-Hub-Signature-256` header
2. **Use IP Allowlisting**: Restrict webhook access to Meta's IP ranges
3. **Rotate Tokens**: Regularly rotate access tokens
4. **Audit Logs**: Enable and monitor API access logs
5. **Rate Limiting**: Configure appropriate rate limits to prevent abuse
