# 🐛 BUG-002: Fix Grafana Google Analytics Dashboard Query Error

**Linear ID**: BVL-318  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-318/bug-fix-grafana-google-analytics-dashboard-query-error

---

## 📋 User Story

**As a** Site Reliability Engineer  
**I want** the Grafana Google Analytics dashboard to display data correctly  
**So that** I can monitor website analytics and user engagement metrics

---

## 🐛 Bug Description

The Grafana Google Analytics dashboard at `https://grafana.lucena.cloud/d/google-analytics-homepage/google-analytics` is failing to display data across multiple panels. All panels show "No data" with a critical error.

**Error Message:**
```
failed to read query: error reading query: json: cannot unmarshal string into Go struct field QueryModel.metrics of type []string
```

**Affected Panels:**
- Active Users
- Page Views
- Sessions
- Avg Session Duration
- Active Users Over Time

**Root Cause:**
The query configuration has the `metrics` field set as a string value, but the Grafana Google Analytics data source expects it to be an array of strings (`[]string`). This JSON unmarshalling error prevents the dashboard from loading any data.

**Dashboard URL:**
https://grafana.lucena.cloud/d/google-analytics-homepage/google-analytics?orgId=1&from=now-7d&to=now&timezone=browser&refresh=5m

---

## 🎯 Acceptance Criteria

- [ ] Dashboard loads without JSON unmarshalling errors
- [ ] All panels display data correctly:
  - [ ] Active Users panel shows user count
  - [ ] Page Views panel shows page view metrics
  - [ ] Sessions panel shows session data
  - [ ] Avg Session Duration panel shows duration metrics
  - [ ] Active Users Over Time panel displays time series data
- [ ] Query configuration uses correct data types (metrics as array)
- [ ] Dashboard works across different time ranges
- [ ] No console errors in browser developer tools
- [ ] Dashboard refresh functionality works correctly

---

## 🔧 Technical Details

**Current Configuration (Incorrect):**
- Metrics field: `ga:sessions` (string)

**Expected Configuration (Correct):**
- Metrics field: `["ga:sessions"]` (array of strings)

**Investigation Steps:**
1. Access Grafana dashboard in edit mode
2. Review query configuration for all affected panels
3. Verify metrics field format in panel JSON
4. Update metrics field from string to array format
5. Test dashboard with various time ranges
6. Verify all panels load data correctly

---

## 📊 Impact

- **Severity**: High - Dashboard is completely non-functional
- **Affected Users**: SRE team, analytics monitoring
- **Business Impact**: Unable to monitor website analytics and user engagement

---

## 🔗 Related Links

- Dashboard: https://grafana.lucena.cloud/d/google-analytics-homepage/google-analytics?orgId=1&from=now-7d&to=now&timezone=browser&refresh=5m

---

**Last Updated**: January 13, 2026  
**Status**: Backlog  
**Priority**: High  
**Labels**: Bug, SRE
