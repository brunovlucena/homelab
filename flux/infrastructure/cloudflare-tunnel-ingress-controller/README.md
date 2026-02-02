# STRRL Cloudflare Tunnel Ingress Controller

This controller replaces the custom `cloudflare-tunnel-operator` with a standard Kubernetes Ingress controller.

## Prerequisites

### Cloudflare API Token

The STRRL controller requires a **Cloudflare API Token** (not an API Key) with the following permissions:

- `Account.Cloudflare Tunnel:Edit`
- `Zone.DNS:Edit`
- `Zone.Zone:Read`

**Create API Token:**

1. Go to: https://dash.cloudflare.com/profile/api-tokens
2. Click "Create Token"
3. Use this template URL to pre-fill permissions:
   ```
   https://dash.cloudflare.com/profile/api-tokens?permissionGroupKeys=[{"key":"zone","type":"read"},{"key":"dns","type":"edit"},{"key":"argotunnel","type":"edit"}]&name=Cloudflare%20Tunnel%20Ingress%20Controller&accountId=*&zoneId=all
   ```
4. Set:
   - **Account Resources**: Include - All accounts
   - **Zone Resources**: Include - All zones
   - **Permissions**:
     - Zone - Zone:Read
     - Zone - DNS:Edit
     - Account - Cloudflare Tunnel:Edit
5. Copy the generated token

### Update Secret

The API Token must be stored in the `cloudflare-api` secret in the `cloudflare-tunnel-ingress-controller` namespace:

```bash
kubectl create secret generic cloudflare-api \
  --from-literal=api-token="YOUR_API_TOKEN" \
  --from-literal=account-id="a2862058e1cc276aa01de068d23f6e1f" \
  --from-literal=tunnel-name="studio" \
  -n cloudflare-tunnel-ingress-controller \
  --dry-run=client -o yaml | kubectl apply -f -
```

Or update the existing secret:

```bash
kubectl patch secret cloudflare-api -n cloudflare-tunnel-ingress-controller \
  --type='json' \
  -p='[{"op": "replace", "path": "/data/api-token", "value": "'$(echo -n "YOUR_API_TOKEN" | base64)'"}]'
```

## Troubleshooting

### Authentication Error (10001)

If you see `Unable to authenticate request (10001)`, the API Token is either:
- Missing or incorrect
- Doesn't have the required permissions
- Is an API Key instead of an API Token

**Solution:** Create a new API Token with the permissions listed above and update the secret.

### Check Controller Logs

```bash
kubectl logs -n cloudflare-tunnel-ingress-controller -l app.kubernetes.io/name=cloudflare-tunnel-ingress-controller
```

### Verify Secret

```bash
kubectl get secret -n cloudflare-tunnel-ingress-controller cloudflare-api -o yaml
```
