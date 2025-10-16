# K6 Load Testing

This directory contains k6 load tests for the Homepage application.

## Overview

The k6 load tests simulate realistic user traffic to test the application's performance and reliability. The tests cover key endpoints including health checks, projects, metrics, and analytics tracking.

## Test Configuration

### Load Test Stages

The default load test (`load-test.js`) includes the following stages:

1. **Ramp-up to 10 users** (2 minutes) - Gradually increase load
2. **Sustain 10 users** (5 minutes) - Maintain steady load
3. **Ramp-up to 20 users** (2 minutes) - Increase to moderate load
4. **Sustain 20 users** (5 minutes) - Maintain moderate load
5. **Ramp-down** (2 minutes) - Gradually decrease to 0 users

**Total duration:** ~16 minutes

### Performance Thresholds

The tests enforce the following SLOs (Service Level Objectives):

- **Response Time (P95):** < 500ms - 95% of requests must complete within 500ms
- **Error Rate:** < 10% - Less than 10% of requests can fail
- **Availability:** > 90% - The service must be available for at least 90% of requests

## Running Tests

### Local Development

Test against local development environment (requires `make up` first):

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/homepage
make test-k6-local
```

### Production Testing

Test against production environment (k8s-api.lucena.cloud):

```bash
cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/clusters/homelab/infrastructure/homepage
make test-k6-prod
```

### Custom URL Testing

Test against a custom URL:

```bash
cd tests/k6
BASE_URL=https://your-custom-url.com k6 run --out json=k6-results.json load-test.js
```

### GitHub Actions

The k6 tests can also be run via GitHub Actions:

1. **Manual Trigger:** Go to Actions → "Homepage K6 Load Tests" → Run workflow
2. **Scheduled:** Tests run automatically daily at 2 AM UTC
3. **On Push:** Tests run on pushes to `main` that affect the homepage infrastructure

#### Workflow Options

- **Duration:** Choose between `short` (4 min), `medium` (16 min), or `long` (35 min)
- **Target URL:** Specify a custom target URL (defaults to k8s-api.lucena.cloud)

## Test Results

### Output Files

- `k6-results.json` - Detailed test results in JSON format
- Console output - Real-time summary during test execution

### Key Metrics

The tests report on:

- **HTTP Request Duration:** Response time statistics (min, max, avg, p90, p95, p99)
- **HTTP Request Failed:** Percentage of failed requests
- **Checks:** Success rate of assertion checks
- **Virtual Users:** Number of concurrent users
- **Iterations:** Total number of test iterations completed

## Endpoints Tested

The load test covers the following endpoints:

1. **Health Check** - `GET /health`
   - Expected: 200 OK
   - Threshold: < 200ms response time

2. **Projects List** - `GET /api/v1/projects`
   - Expected: 200 OK with JSON array
   - Threshold: < 300ms response time

3. **Metrics** - `GET /metrics`
   - Expected: 200 OK
   - Threshold: < 500ms response time

4. **Analytics Tracking** - `POST /api/v1/analytics/track`
   - Expected: 200 OK
   - Threshold: < 400ms response time

## Customizing Tests

### Adjusting Load

Edit the `options.stages` array in `load-test.js`:

```javascript
export const options = {
  stages: [
    { duration: '1m', target: 5 },   // Ramp up to 5 users over 1 minute
    { duration: '3m', target: 5 },   // Stay at 5 users for 3 minutes
    { duration: '1m', target: 0 },   // Ramp down to 0 users
  ],
  // ...
};
```

### Adjusting Thresholds

Modify the `options.thresholds` object:

```javascript
export const options = {
  // ...
  thresholds: {
    http_req_duration: ['p(95)<300'],  // Stricter: 300ms
    http_req_failed: ['rate<0.05'],    // Stricter: 5% error rate
    errors: ['rate<0.05'],
  },
};
```

### Adding New Endpoints

Add new test scenarios in the `default` function:

```javascript
export default function () {
  // Existing tests...
  
  // New endpoint test
  const newEndpoint = http.get(`${BASE_URL}/api/v1/new-endpoint`, params);
  check(newEndpoint, {
    'new endpoint status is 200': (r) => r.status === 200,
    'new endpoint response time < 500ms': (r) => r.timings.duration < 500,
  });
}
```

## CI/CD Integration

The k6 tests are integrated into the CI/CD pipeline via GitHub Actions (`.github/workflows/homepage-k6.yaml`).

### Workflow Features

- ✅ Multiple test duration profiles (short/medium/long)
- ✅ Configurable target URLs
- ✅ Automatic test result artifacts
- ✅ GitHub Actions summary with key metrics
- ✅ Scheduled daily runs
- ✅ Manual trigger with custom parameters

### Viewing Results

1. Go to GitHub Actions → Select a workflow run
2. Check the "Summary" tab for high-level metrics
3. Download artifacts for detailed JSON results

## Troubleshooting

### Test Failures

If tests fail, check:

1. **Target availability:** Is the target URL accessible?
2. **Response times:** Are response times exceeding thresholds?
3. **Error rates:** Are there server errors (5xx) or client errors (4xx)?
4. **Resource limits:** Is the target under-provisioned?

### Common Issues

**Issue:** Test hangs or times out
- **Solution:** Reduce the number of virtual users or test duration

**Issue:** High error rate
- **Solution:** Check server logs, database connections, and resource utilization

**Issue:** k6 not found
- **Solution:** Install k6 following instructions at https://k6.io/docs/get-started/installation/

## Resources

- [k6 Documentation](https://k6.io/docs/)
- [k6 Best Practices](https://k6.io/docs/testing-guides/test-types/)
- [Homepage API Documentation](../../api/README.md)
- [Grafana Dashboard](../../chart/templates/monitoring/dashboard.yaml)

## Support

For issues or questions about the k6 tests, please:
1. Check the test output for specific error messages
2. Review the Grafana dashboards for system metrics
3. Consult the runbooks in `docs/runbooks/homepage/`

