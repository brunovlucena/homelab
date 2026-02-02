# 🚀 Cloudflare Setup Job - Automated Performance Configuration

Kubernetes Job to automatically configure Cloudflare Page Rules and Speed settings for `lucena.cloud`.

## 📋 What It Does

This job automates the Cloudflare configuration described in `docs/CLOUDFLARE_SETUP.md`:

1. **Page Rules** (3 rules):
   - `lucena.cloud/*.js` - Cache Everything, Edge TTL: 1 month
   - `lucena.cloud/storage/*` - Cache Everything, Edge TTL: 1 year  
   - `lucena.cloud/` - Cache Everything, Edge TTL: 15 minutes

2. **Speed Optimization**:
   - Auto Minify: JavaScript ✅
   - Auto Minify: CSS ✅
   - Auto Minify: HTML ✅

## 🎯 Quick Start

### Prerequisites

1. **Cloudflare API Token**: The job requires a Cloudflare API token with the following permissions:
   - `Zone.Zone:Read` (to get zone ID)
   - `Zone.Page Rules:Edit` (to create/update page rules)
   - `Zone.Settings:Edit` (to update speed settings)

   The token should be stored in the secret:
   ```bash
   kubectl get secret cloudflare -n cloudflare-tunnel
   ```
   
   Expected keys:
   - `cloudflare-api-token` (preferred)
   - `cloudflare-api-key` (fallback)

2. **Secret Setup**: If the secret doesn't exist:
   ```bash
   kubectl create secret generic cloudflare -n cloudflare-tunnel \
     --from-literal=cloudflare-api-token='YOUR_API_TOKEN'
   ```

### Run the Job

```bash
# Apply the job
kubectl apply -f k8s/jobs/cloudflare-setup-job.yaml

# Watch the job logs
kubectl logs -f job/cloudflare-setup -n homepage

# Check job status
kubectl get job cloudflare-setup -n homepage
```

## 📊 How It Works

1. **Init Container**: Syncs Cloudflare API token from `cloudflare-tunnel` namespace to `homepage` namespace
2. **Main Container**: Uses curl to call Cloudflare API v4 to:
   - Get zone ID for `lucena.cloud`
   - Create or update Page Rules (updates if pattern matches)
   - Update Speed settings (Auto Minify)

## 🔧 Configuration

### Domain

The job is configured for `lucena.cloud`. To change the domain, edit the job manifest:

```yaml
DOMAIN="lucena.cloud"
```

### Page Rules

Page rules are defined in the `create_or_update_page_rule` function calls. Each rule specifies:
- URL pattern
- Cache level
- Edge Cache TTL (seconds)
- Browser Cache TTL (seconds or "respect_existing_headers")

### Speed Settings

Auto Minify settings are hardcoded in the job:
- JavaScript: ON
- CSS: ON  
- HTML: ON

## 🔍 Troubleshooting

### Job Fails: "Could not find zone ID"

- Verify the Cloudflare API token has `Zone.Zone:Read` permission
- Check that `lucena.cloud` domain exists in your Cloudflare account
- Verify the API token is correct

### Job Fails: "Could not extract Cloudflare API token"

- Check that the secret exists: `kubectl get secret cloudflare -n cloudflare-tunnel`
- Verify the secret has the correct key: `kubectl get secret cloudflare -n cloudflare-tunnel -o yaml`
- The job looks for keys: `cloudflare-api-token` or `cloudflare-api-key`

### Page Rules Not Created

- Check logs for API errors: `kubectl logs job/cloudflare-setup -n homepage`
- Verify API token has `Zone.Page Rules:Edit` permission
- Cloudflare free tier allows up to 3 page rules (this job creates 3)

### Auto Minify Not Enabled

- Check logs for API errors
- Verify API token has `Zone.Settings:Edit` permission
- Auto Minify is available on all Cloudflare plans (including free)

## 📝 API Token Setup

### Create Cloudflare API Token

1. Go to: https://dash.cloudflare.com/profile/api-tokens
2. Click "Create Token"
3. Use "Edit zone DNS" template or create custom:
   - **Permissions:**
     - `Zone` > `Zone` > `Read`
     - `Zone` > `Page Rules` > `Edit`
     - `Zone` > `Settings` > `Edit`
   - **Zone Resources:**
     - Include > Specific zone > `lucena.cloud`
4. Copy the token

### Update Kubernetes Secret

```bash
kubectl create secret generic cloudflare -n cloudflare-tunnel \
  --from-literal=cloudflare-api-token='YOUR_API_TOKEN' \
  --dry-run=client -o yaml | kubectl apply -f -
```

## 🔄 Re-running the Job

The job is idempotent - it can be run multiple times safely:
- Existing Page Rules with matching patterns will be updated
- Speed settings will be updated to the configured values

To re-run after changes:

```bash
# Delete the old job
kubectl delete job cloudflare-setup -n homepage

# Apply the job again
kubectl apply -f k8s/jobs/cloudflare-setup-job.yaml

# Watch logs
kubectl logs -f job/cloudflare-setup -n homepage
```

## 📚 Related Documentation

- [CLOUDFLARE_SETUP.md](../../docs/CLOUDFLARE_SETUP.md) - Manual setup guide
- [Cloudflare API Documentation](https://developers.cloudflare.com/api/)

## 🔒 Security

- The Cloudflare API token is synced via init container and stored temporarily in `homepage` namespace
- The secret is cleaned up automatically after job completion (TTL: 24 hours)
- RBAC limits the job to only read secrets (no write permissions)
- The API token should have minimal required permissions (zone-scoped, not account-scoped)
