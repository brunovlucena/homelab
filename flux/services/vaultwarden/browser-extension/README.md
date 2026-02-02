# Browser Extensions for Vaultwarden

Browser extensions for Chrome, Firefox, and Safari that provide password auto-fill and management capabilities.

## Features

- ✅ Auto-fill passwords on websites
- ✅ Save new passwords
- ✅ Password generator
- ✅ Secure password storage
- ✅ Quick access to vault
- ✅ Biometric unlock (where supported)

## Architecture

All extensions share the same core logic:
- **Content Scripts**: Inject into web pages for auto-fill
- **Background Scripts**: Handle API communication and state
- **Popup UI**: Quick access to passwords
- **Options Page**: Settings and configuration

## Project Structure

```
browser-extension/
├── chrome/              # Chrome extension
│   ├── manifest.json
│   ├── background.js
│   ├── content.js
│   ├── popup.html
│   ├── popup.js
│   ├── options.html
│   ├── options.js
│   └── icons/
├── firefox/             # Firefox extension
│   ├── manifest.json
│   └── [same structure as Chrome]
└── safari/              # Safari Web Extension
    ├── VaultwardenExtension/
    └── VaultwardenExtension.xcodeproj
```

## Shared Code

The extensions use a shared codebase with platform-specific manifests. Core functionality includes:

- API client for server communication
- Encryption/decryption (using Web Crypto API)
- Keychain/secure storage integration
- Auto-fill detection and injection

## Development

### Chrome Extension

1. Load unpacked extension:
   - Open Chrome → Extensions → Developer mode
   - Click "Load unpacked"
   - Select the `chrome/` directory

2. Build for production:
   ```bash
   cd chrome
   npm run build
   ```

### Firefox Extension

1. Load temporary extension:
   - Open Firefox → about:debugging
   - Click "This Firefox"
   - Click "Load Temporary Add-on"
   - Select `manifest.json` in `firefox/` directory

2. Build for production:
   ```bash
   cd firefox
   npm run build
   ```

### Safari Extension

1. Open `safari/VaultwardenExtension.xcodeproj` in Xcode
2. Build and run
3. Enable extension in Safari → Preferences → Extensions

## Configuration

Set your server URL in the extension options:

```javascript
const API_BASE_URL = "https://vaultwarden.lucena.cloud";
```

## Security

- ✅ Passwords encrypted client-side before storage
- ✅ Secure storage using browser's secure storage APIs
- ✅ HTTPS only communication
- ✅ No plaintext passwords in extension storage
- ✅ Auto-lock after inactivity

## Manifest Versions

- **Chrome**: Manifest V3
- **Firefox**: Manifest V3 (compatible)
- **Safari**: Web Extension (compatible with MV2/MV3)
