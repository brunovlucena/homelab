# 🎨 Homepage Logos & Icons

This directory contains all the logos and icons used throughout the Homepage application.

## 📁 Files

### Company Logos
- `crealytics-logo.svg` - Crealytics company logo
- `deutsche-bahn-logo.svg` - Deutsche Bahn company logo  
- `mobimeo-logo.svg` - Mobimeo company logo
- `notifi-logo.svg` - Notifi company logo
- `tempest-logo.svg` - Tempest company logo

### Personal Icons
- `bruno-slack-logo.svg` - **Personalized Slack API icon for Alertmanager** (64x64px SVG)
- `bruno-slack-logo.png` - **Personalized Slack API icon for Alertmanager** (64x64px PNG)
- `bruno-slack-icon-32.svg` - **Small version of Bruno's Slack icon** (32x32px SVG)
- `bruno-slack-icon-32.png` - **Small version of Bruno's Slack icon** (32x32px PNG)

## 🚨 Bruno's Slack Alert Icon

The `bruno-slack-logo.png` and `bruno-slack-icon-32.png` files are personalized icons created specifically for the Alertmanager Slack integration. These icons feature:

- **Slack's signature hash (#) symbol** with gradient colors
- **Personal "B" monogram** in the corner representing Bruno
- **Alert notification dot** indicating this is for monitoring/alerting
- **Professional gradient background** matching Slack's brand colors

### Usage in Alertmanager

The icon is referenced in the Alertmanager configuration as `:bruno-slack:` emoji. To use this custom emoji in Slack:

1. Upload the `bruno-slack-logo.png` file to your Slack workspace
2. Set the emoji name as `bruno-slack`
3. The Alertmanager will automatically use this icon for Homepage alerts

### Design Features

- **64x64px version**: Full-size logo for main usage (PNG format)
- **32x32px version**: Compact version for notifications (PNG format)
- **PNG format**: High-quality raster graphics optimized for Slack compatibility
- **Gradient colors**: Uses Slack's official color palette
- **Personal touch**: Includes Bruno's "B" monogram for personalization

## 🎯 Integration

These icons are automatically included in the frontend build process and can be referenced in:

- Alertmanager Slack notifications
- Grafana dashboards
- Frontend components
- Documentation

## 🔧 Customization

To modify the icons:

1. Edit the SVG files directly, then convert to PNG using rsvg-convert
2. Maintain the same dimensions (64x64 or 32x32)
3. Keep the "B" monogram for personal branding
4. Test in Slack to ensure proper display
5. Use PNG format for best Slack compatibility

---

*These icons were created specifically for Bruno's personal monitoring setup and Alertmanager integration.*
