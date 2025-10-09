#!/bin/bash

# 🚨 Bruno Site - Slack Emoji Setup Script
# This script helps set up the custom Bruno Slack emoji for Alertmanager notifications

set -e

echo "🎨 Setting up Bruno's custom Slack emoji for Alertmanager..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if required files exist
ICON_FILE="frontend/public/assets/logos/bruno-slack-logo.png"
ICON_32_FILE="frontend/public/assets/logos/bruno-slack-icon-32.png"

if [[ ! -f "$ICON_FILE" ]]; then
    echo -e "${RED}❌ Error: $ICON_FILE not found!${NC}"
    exit 1
fi

if [[ ! -f "$ICON_32_FILE" ]]; then
    echo -e "${RED}❌ Error: $ICON_32_FILE not found!${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Found Bruno's Slack icon files${NC}"

# Display instructions
echo -e "${BLUE}📋 Setup Instructions:${NC}"
echo ""
echo -e "${YELLOW}1. Upload the custom emoji to your Slack workspace:${NC}"
echo "   - Go to your Slack workspace settings"
echo "   - Navigate to 'Customize' → 'Emoji'"
echo "   - Click 'Add Custom Emoji'"
echo "   - Upload: $ICON_FILE"
echo "   - Set emoji name as: bruno-slack"
echo "   - Click 'Save'"
echo ""
echo -e "${YELLOW}2. Update your Alertmanager configuration:${NC}"
echo "   - Set the webhook URL in values.yaml:"
echo "     monitoring.alertmanager.slack.webhook_url"
echo "   - The icon will automatically be used via :bruno-slack: emoji"
echo ""
echo -e "${YELLOW}3. Test the integration:${NC}"
echo "   - Deploy the updated configuration"
echo "   - Trigger a test alert"
echo "   - Verify the custom icon appears in Slack"
echo ""

# Show file paths
echo -e "${BLUE}📁 Icon Files:${NC}"
echo "   Main icon (64x64): $ICON_FILE"
echo "   Small icon (32x32): $ICON_32_FILE"
echo ""

# Show current configuration
echo -e "${BLUE}⚙️ Current Configuration:${NC}"
echo "   Emoji name: bruno-slack"
echo "   Channel: #bruno-site-alerts"
echo "   Icon URL: https://lucena.cloud/assets/logos/bruno-slack-logo.png"
echo ""

# Check if webhook URL is set
WEBHOOK_URL=$(grep -E "webhook_url:" chart/values.yaml | head -1 | sed 's/.*webhook_url: *//' | tr -d '"' | tr -d "'")
if [[ -z "$WEBHOOK_URL" || "$WEBHOOK_URL" == "" ]]; then
    echo -e "${YELLOW}⚠️  Warning: Slack webhook URL not configured in values.yaml${NC}"
    echo "   Please set monitoring.alertmanager.slack.webhook_url"
else
    echo -e "${GREEN}✅ Slack webhook URL is configured${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Setup complete! Your personalized Bruno Slack icon is ready for Alertmanager.${NC}"
echo ""
echo -e "${BLUE}💡 Pro tip: The icon includes your personal 'B' monogram and alert notification dot${NC}"
echo -e "${BLUE}   to make it uniquely yours while maintaining Slack's professional appearance.${NC}"
