# iOS DevOps Best Practices - Principal Engineer Guide

## Table of Contents
1. [CI/CD Architecture](#cicd-architecture)
2. [Testing Strategy](#testing-strategy)
3. [Code Quality](#code-quality)
4. [Security](#security)
5. [Performance](#performance)
6. [Monitoring](#monitoring)

## CI/CD Architecture

### GitHub Actions Workflow

**Key Components:**
- **Lint Job**: SwiftLint validation
- **Test Job**: Unit + UI tests with coverage
- **Build Job**: IPA creation
- **Release Job**: TestFlight/App Store deployment

**Best Practices:**
```yaml
# Use matrix builds for multiple iOS versions
strategy:
  matrix:
    ios-version: ['17.0', '17.5', '18.0']
    device: ['iPhone 15', 'iPhone 15 Pro']
```

### Fastlane Configuration

**Essential Lanes:**
- `test` - Run all tests
- `build` - Create IPA
- `beta` - TestFlight deployment
- `release` - App Store submission
- `screenshots` - Generate screenshots

**Match for Certificates:**
```ruby
match(
  type: "appstore",
  readonly: true,  # Use readonly in CI
  app_identifier: ["com.yourcompany.app"]
)
```

## Testing Strategy

### Unit Tests
- **Coverage Target**: >80%
- **Focus Areas**: Business logic, ViewModels, Services
- **Tools**: XCTest, Quick/Nimble

### UI Tests
- **Critical Paths**: Login, Core workflows
- **Tools**: XCUITest
- **Run Frequency**: Pre-release only

### Integration Tests
- **API Mocking**: Use URLProtocol
- **Database**: In-memory CoreData
- **Network**: Mock responses

### Test Organization
```
Tests/
├── Unit/
│   ├── ViewModels/
│   ├── Services/
│   └── Models/
├── UI/
│   └── Flows/
└── Integration/
    └── API/
```

## Code Quality

### SwiftLint Rules
```yaml
# Enforce strict rules
opt_in_rules:
  - empty_count
  - first_where
  - sorted_first_last

# Custom rules
custom_rules:
  no_print_statements:
    name: "No print statements"
    regex: 'print\('
    message: "Use logger instead of print()"
```

### Pre-commit Hooks
```bash
#!/bin/bash
# .git/hooks/pre-commit

# Run SwiftLint
swiftlint lint --strict || exit 1

# Run tests
fastlane test || exit 1

# Validate code
./scripts/validate.sh || exit 1
```

### Code Review Checklist
- [ ] Tests written and passing
- [ ] SwiftLint passes
- [ ] No force unwraps
- [ ] Error handling implemented
- [ ] Documentation updated
- [ ] Performance considered

## Security

### Secrets Management
1. **GitHub Secrets**: Store API keys, passwords
2. **Match**: Certificate management
3. **Keychain**: Local development secrets
4. **Environment Variables**: CI/CD only

### Code Signing
```ruby
# Use Match for all certificates
match(
  type: "appstore",
  git_url: "https://github.com/org/certificates",
  app_identifier: ["com.company.app"]
)
```

### API Security
- Use App Store Connect API keys (not passwords)
- Rotate credentials quarterly
- Use app-specific passwords
- Enable 2FA everywhere

### Dependency Management
```ruby
# Gemfile.lock should be committed
# Podfile.lock should be committed
# Use exact versions in production
```

## Performance

### Build Optimization
```yaml
# Parallel builds
xcodebuild -jobs $(sysctl -n hw.ncpu)

# DerivedData caching
- uses: actions/cache@v3
  with:
    path: ~/Library/Developer/Xcode/DerivedData
    key: derived-data-${{ hashFiles('**/*.pbxproj') }}
```

### Test Optimization
- Run tests in parallel
- Use test sharding
- Cache test results
- Skip UI tests in PRs

### App Performance
- Monitor build times
- Track app size
- Profile memory usage
- Monitor network calls

## Monitoring

### Crash Reporting
- **Sentry** or **Firebase Crashlytics**
- Track crash-free rate
- Monitor ANR (App Not Responding)

### Analytics
- User engagement metrics
- Feature usage
- Performance metrics
- Error rates

### CI/CD Metrics
- Build success rate
- Test pass rate
- Deployment frequency
- Mean time to recovery

## Automation Scripts

### Build Script
```bash
#!/bin/bash
# Automated build with error handling

set -e
set -o pipefail

# Clean
xcodebuild clean

# Test
xcodebuild test | xcpretty

# Build
xcodebuild archive | xcpretty

# Upload
fastlane beta
```

### Release Script
```bash
#!/bin/bash
# Automated release process

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

# Update version
agvtool new-marketing-version $VERSION

# Tag release
git tag -a "v$VERSION" -m "Release $VERSION"
git push origin "v$VERSION"

# Deploy
fastlane release
```

## Troubleshooting

### Common Issues

**Build Fails:**
1. Check Xcode version
2. Verify certificates
3. Check provisioning profiles
4. Clear DerivedData

**Tests Fail:**
1. Check simulator availability
2. Verify test data
3. Check test target configuration

**Fastlane Issues:**
1. Run `fastlane env`
2. Check match certificates
3. Verify API access

## Resources

- [Fastlane Docs](https://docs.fastlane.tools/)
- [GitHub Actions for iOS](https://docs.github.com/en/actions)
- [SwiftLint Rules](https://realm.github.io/SwiftLint/)
- [Apple Developer Resources](https://developer.apple.com/)
