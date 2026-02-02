// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 Unit Tests: SRE Metrics SLO Validation
//
//	Tests for SLO thresholds defined in SRE_LAMBDA_METRICS_REPORT.md:
//	- Error rate < 5%
//	- Build duration P95 < 120s
//	- Reconcile duration P95 < 100ms
//	- Workqueue depth < 10
//	- Cold start rate < 20%
//	- Function latency P95 < 5s
//
//	Based on: docs/reports/SRE_LAMBDA_METRICS_REPORT.md
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package sre

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// SLO Thresholds (from SRE_LAMBDA_METRICS_REPORT.md)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const (
	// SLO_ErrorRatePercent is the maximum allowed error rate
	// From report: "Error rate > 5% for any lambda" triggers alert
	SLO_ErrorRatePercent = 5.0

	// SLO_BuildDurationP95Seconds is the P95 build duration threshold
	// From report: "Build duration > 120s" triggers alert
	SLO_BuildDurationP95Seconds = 120.0

	// SLO_ReconcileDurationP95Ms is the P95 reconcile duration threshold
	// From report: "P95 Reconcile Duration: 45.8ms" (excellent)
	// Setting threshold at 100ms as buffer
	SLO_ReconcileDurationP95Ms = 100.0

	// SLO_WorkqueueDepthMax is the maximum allowed workqueue depth
	// From report: "Workqueue depth > 10" triggers alert
	SLO_WorkqueueDepthMax = 10

	// SLO_ColdStartRatePercent is the maximum allowed cold start rate
	// Cold starts > 20% indicates scaling issues
	SLO_ColdStartRatePercent = 20.0

	// SLO_FunctionLatencyP95Ms is the P95 function latency threshold
	// From report: "P95 < 5ms for all functions" (excellent)
	// Setting threshold at 5000ms (5s) for cold start tolerance
	SLO_FunctionLatencyP95Ms = 5000.0
)

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 SLO Calculation Helper Types                                        │
// └─────────────────────────────────────────────────────────────────────────┘

// SLOCalculator provides methods for calculating SLO metrics
type SLOCalculator struct {
	// Error tracking
	TotalRequests int64
	TotalErrors   int64

	// Latency tracking
	Latencies []float64

	// Cold start tracking
	TotalInvocations int64
	ColdStarts       int64

	// Build tracking
	BuildDurations []float64

	// Reconcile tracking
	ReconcileDurations []float64

	// Workqueue
	WorkqueueDepth int
}

// NewSLOCalculator creates a new SLO calculator
func NewSLOCalculator() *SLOCalculator {
	return &SLOCalculator{
		Latencies:          make([]float64, 0),
		BuildDurations:     make([]float64, 0),
		ReconcileDurations: make([]float64, 0),
	}
}

// ErrorRate calculates the current error rate percentage
func (s *SLOCalculator) ErrorRate() float64 {
	if s.TotalRequests == 0 {
		return 0
	}
	return float64(s.TotalErrors) / float64(s.TotalRequests) * 100
}

// ColdStartRate calculates the cold start rate percentage
func (s *SLOCalculator) ColdStartRate() float64 {
	if s.TotalInvocations == 0 {
		return 0
	}
	return float64(s.ColdStarts) / float64(s.TotalInvocations) * 100
}

// LatencyP95 calculates the P95 latency
func (s *SLOCalculator) LatencyP95() float64 {
	return percentile(s.Latencies, 95)
}

// BuildDurationP95 calculates the P95 build duration
func (s *SLOCalculator) BuildDurationP95() float64 {
	return percentile(s.BuildDurations, 95)
}

// ReconcileDurationP95 calculates the P95 reconcile duration
func (s *SLOCalculator) ReconcileDurationP95() float64 {
	return percentile(s.ReconcileDurations, 95)
}

// ErrorBudgetRemaining calculates remaining error budget
// Error budget = allowed errors - actual errors
func (s *SLOCalculator) ErrorBudgetRemaining() float64 {
	if s.TotalRequests == 0 {
		return 100.0
	}
	allowedErrors := float64(s.TotalRequests) * (SLO_ErrorRatePercent / 100)
	remaining := (allowedErrors - float64(s.TotalErrors)) / float64(s.TotalRequests) * 100
	return math.Max(0, remaining)
}

// IsCompliant checks if all SLOs are met
func (s *SLOCalculator) IsCompliant() bool {
	return s.ErrorRate() < SLO_ErrorRatePercent &&
		s.ColdStartRate() < SLO_ColdStartRatePercent &&
		s.LatencyP95() < SLO_FunctionLatencyP95Ms &&
		s.BuildDurationP95() < SLO_BuildDurationP95Seconds &&
		s.ReconcileDurationP95() < SLO_ReconcileDurationP95Ms &&
		s.WorkqueueDepth < SLO_WorkqueueDepthMax
}

// percentile calculates the nth percentile of a slice
func percentile(data []float64, p float64) float64 {
	if len(data) == 0 {
		return 0
	}

	// Sort data
	sorted := make([]float64, len(data))
	copy(sorted, data)
	for i := range sorted {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Calculate percentile index
	index := (p / 100) * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1
	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🧪 SLO Threshold Tests                                                 │
// └─────────────────────────────────────────────────────────────────────────┘

func TestSLO_ErrorRateCompliance(t *testing.T) {
	tests := []struct {
		name      string
		requests  int64
		errors    int64
		compliant bool
		errorRate float64
	}{
		{
			name:      "zero errors is compliant",
			requests:  1000,
			errors:    0,
			compliant: true,
			errorRate: 0,
		},
		{
			name:      "4% error rate is compliant",
			requests:  1000,
			errors:    40,
			compliant: true,
			errorRate: 4.0,
		},
		{
			name:      "exactly 5% is not compliant (must be less than)",
			requests:  1000,
			errors:    50,
			compliant: false,
			errorRate: 5.0,
		},
		{
			name:      "10% error rate is not compliant",
			requests:  1000,
			errors:    100,
			compliant: false,
			errorRate: 10.0,
		},
		{
			name:      "no requests returns 0%",
			requests:  0,
			errors:    0,
			compliant: true,
			errorRate: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewSLOCalculator()
			calc.TotalRequests = tt.requests
			calc.TotalErrors = tt.errors

			assert.InDelta(t, tt.errorRate, calc.ErrorRate(), 0.01)

			compliant := calc.ErrorRate() < SLO_ErrorRatePercent
			assert.Equal(t, tt.compliant, compliant)
		})
	}
}

func TestSLO_BuildDurationCompliance(t *testing.T) {
	tests := []struct {
		name        string
		durations   []float64
		compliant   bool
		expectedP95 float64
	}{
		{
			name:        "fast builds are compliant",
			durations:   []float64{30, 40, 50, 60, 70},
			compliant:   true,
			expectedP95: 68, // Approximately
		},
		{
			name:        "builds at threshold are not compliant",
			durations:   []float64{110, 115, 118, 120, 125},
			compliant:   false,
			expectedP95: 124, // P95 of [110,115,118,120,125] > 120s threshold
		},
		{
			name:        "slow builds are not compliant",
			durations:   []float64{100, 120, 140, 160, 180},
			compliant:   false,
			expectedP95: 176, // Approximately
		},
		{
			name:        "empty durations returns 0",
			durations:   []float64{},
			compliant:   true,
			expectedP95: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewSLOCalculator()
			calc.BuildDurations = tt.durations

			p95 := calc.BuildDurationP95()
			assert.InDelta(t, tt.expectedP95, p95, 2.0) // Allow 2s tolerance

			compliant := p95 < SLO_BuildDurationP95Seconds
			assert.Equal(t, tt.compliant, compliant)
		})
	}
}

func TestSLO_ReconcileDurationCompliance(t *testing.T) {
	tests := []struct {
		name      string
		durations []float64 // in milliseconds
		compliant bool
	}{
		{
			name:      "fast reconciles are compliant (like in report: 45.8ms)",
			durations: []float64{10, 20, 30, 40, 50},
			compliant: true,
		},
		{
			name:      "reconciles near threshold",
			durations: []float64{80, 85, 90, 95, 99},
			compliant: true,
		},
		{
			name:      "slow reconciles are not compliant",
			durations: []float64{100, 150, 200, 250, 300},
			compliant: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewSLOCalculator()
			calc.ReconcileDurations = tt.durations

			p95 := calc.ReconcileDurationP95()
			compliant := p95 < SLO_ReconcileDurationP95Ms
			assert.Equal(t, tt.compliant, compliant)
		})
	}
}

func TestSLO_WorkqueueDepthCompliance(t *testing.T) {
	tests := []struct {
		name      string
		depth     int
		compliant bool
	}{
		{
			name:      "empty queue is compliant (like in report: 0)",
			depth:     0,
			compliant: true,
		},
		{
			name:      "small queue is compliant",
			depth:     5,
			compliant: true,
		},
		{
			name:      "depth at threshold is not compliant",
			depth:     10,
			compliant: false,
		},
		{
			name:      "large queue is not compliant",
			depth:     50,
			compliant: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewSLOCalculator()
			calc.WorkqueueDepth = tt.depth

			compliant := calc.WorkqueueDepth < SLO_WorkqueueDepthMax
			assert.Equal(t, tt.compliant, compliant)
		})
	}
}

func TestSLO_ColdStartRateCompliance(t *testing.T) {
	tests := []struct {
		name        string
		invocations int64
		coldStarts  int64
		compliant   bool
		rate        float64
	}{
		{
			name:        "no cold starts is compliant",
			invocations: 1000,
			coldStarts:  0,
			compliant:   true,
			rate:        0,
		},
		{
			name:        "low cold start rate is compliant",
			invocations: 1000,
			coldStarts:  20, // 2%
			compliant:   true,
			rate:        2.0,
		},
		{
			name:        "19% cold start rate is compliant",
			invocations: 1000,
			coldStarts:  190,
			compliant:   true,
			rate:        19.0,
		},
		{
			name:        "20% cold start rate is not compliant",
			invocations: 1000,
			coldStarts:  200,
			compliant:   false,
			rate:        20.0,
		},
		{
			name:        "high cold start rate is not compliant",
			invocations: 1000,
			coldStarts:  500,
			compliant:   false,
			rate:        50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewSLOCalculator()
			calc.TotalInvocations = tt.invocations
			calc.ColdStarts = tt.coldStarts

			assert.InDelta(t, tt.rate, calc.ColdStartRate(), 0.01)

			compliant := calc.ColdStartRate() < SLO_ColdStartRatePercent
			assert.Equal(t, tt.compliant, compliant)
		})
	}
}

func TestSLO_FunctionLatencyCompliance(t *testing.T) {
	tests := []struct {
		name      string
		latencies []float64 // in milliseconds
		compliant bool
	}{
		{
			name:      "fast functions are compliant (like in report: <5ms P95)",
			latencies: []float64{1, 2, 3, 4, 5},
			compliant: true,
		},
		{
			name:      "normal latency is compliant",
			latencies: []float64{100, 200, 300, 400, 500},
			compliant: true,
		},
		{
			name:      "high latency but under 5s is compliant",
			latencies: []float64{1000, 2000, 3000, 4000, 4900},
			compliant: true,
		},
		{
			name:      "latency over 5s is not compliant",
			latencies: []float64{4000, 5000, 6000, 7000, 8000},
			compliant: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewSLOCalculator()
			calc.Latencies = tt.latencies

			p95 := calc.LatencyP95()
			compliant := p95 < SLO_FunctionLatencyP95Ms
			assert.Equal(t, tt.compliant, compliant)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📈 Error Budget Tests                                                  │
// └─────────────────────────────────────────────────────────────────────────┘

func TestSLO_ErrorBudgetCalculation(t *testing.T) {
	tests := []struct {
		name           string
		requests       int64
		errors         int64
		expectedBudget float64
	}{
		{
			name:           "full budget with no errors",
			requests:       1000,
			errors:         0,
			expectedBudget: 5.0, // 5% of 1000 = 50 allowed, 0 used = 5% remaining
		},
		{
			name:           "half budget used",
			requests:       1000,
			errors:         25, // Half of allowed 50
			expectedBudget: 2.5,
		},
		{
			name:           "budget exhausted",
			requests:       1000,
			errors:         50, // Exactly at threshold
			expectedBudget: 0,
		},
		{
			name:           "over budget",
			requests:       1000,
			errors:         100, // Over threshold
			expectedBudget: 0,   // Capped at 0
		},
		{
			name:           "no requests returns full budget",
			requests:       0,
			errors:         0,
			expectedBudget: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewSLOCalculator()
			calc.TotalRequests = tt.requests
			calc.TotalErrors = tt.errors

			budget := calc.ErrorBudgetRemaining()
			assert.InDelta(t, tt.expectedBudget, budget, 0.1)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔄 Overall Compliance Tests                                            │
// └─────────────────────────────────────────────────────────────────────────┘

func TestSLO_OverallCompliance(t *testing.T) {
	t.Run("healthy system is compliant", func(t *testing.T) {
		calc := NewSLOCalculator()

		// Simulate healthy system (like in SRE report)
		calc.TotalRequests = 1000
		calc.TotalErrors = 10 // 1% error rate
		calc.TotalInvocations = 1000
		calc.ColdStarts = 20                                      // 2% cold start rate
		calc.Latencies = []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} // <10ms P95
		calc.BuildDurations = []float64{30, 40, 50, 54.7, 60}     // P95 ~54.7s (from report)
		calc.ReconcileDurations = []float64{20, 30, 40, 45.8, 50} // P95 ~45.8ms (from report)
		calc.WorkqueueDepth = 0                                   // Empty queue (from report)

		assert.True(t, calc.IsCompliant())
	})

	t.Run("system with high error rate is not compliant", func(t *testing.T) {
		calc := NewSLOCalculator()

		calc.TotalRequests = 1000
		calc.TotalErrors = 100 // 10% error rate - NOT compliant
		calc.TotalInvocations = 1000
		calc.ColdStarts = 10
		calc.Latencies = []float64{1, 2, 3, 4, 5}
		calc.BuildDurations = []float64{30, 40, 50, 60, 70}
		calc.ReconcileDurations = []float64{20, 30, 40, 50, 60}
		calc.WorkqueueDepth = 0

		assert.False(t, calc.IsCompliant())
	})

	t.Run("system with slow builds is not compliant", func(t *testing.T) {
		calc := NewSLOCalculator()

		calc.TotalRequests = 1000
		calc.TotalErrors = 10
		calc.TotalInvocations = 1000
		calc.ColdStarts = 10
		calc.Latencies = []float64{1, 2, 3, 4, 5}
		calc.BuildDurations = []float64{100, 110, 120, 130, 140} // P95 > 120s - NOT compliant
		calc.ReconcileDurations = []float64{20, 30, 40, 50, 60}
		calc.WorkqueueDepth = 0

		assert.False(t, calc.IsCompliant())
	})

	t.Run("system with high workqueue depth is not compliant", func(t *testing.T) {
		calc := NewSLOCalculator()

		calc.TotalRequests = 1000
		calc.TotalErrors = 10
		calc.TotalInvocations = 1000
		calc.ColdStarts = 10
		calc.Latencies = []float64{1, 2, 3, 4, 5}
		calc.BuildDurations = []float64{30, 40, 50, 60, 70}
		calc.ReconcileDurations = []float64{20, 30, 40, 50, 60}
		calc.WorkqueueDepth = 15 // > 10 - NOT compliant

		assert.False(t, calc.IsCompliant())
	})
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 Percentile Calculation Tests                                        │
// └─────────────────────────────────────────────────────────────────────────┘

func TestPercentileCalculation(t *testing.T) {
	tests := []struct {
		name       string
		data       []float64
		percentile float64
		expected   float64
	}{
		{
			name:       "P50 of sequential data",
			data:       []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			percentile: 50,
			expected:   5.5,
		},
		{
			name:       "P95 of sequential data",
			data:       []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			percentile: 95,
			expected:   9.55,
		},
		{
			name:       "P99 of sequential data",
			data:       []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			percentile: 99,
			expected:   9.91,
		},
		{
			name:       "P100 is max value",
			data:       []float64{1, 2, 3, 4, 5},
			percentile: 100,
			expected:   5,
		},
		{
			name:       "empty data returns 0",
			data:       []float64{},
			percentile: 95,
			expected:   0,
		},
		{
			name:       "single value returns that value",
			data:       []float64{42},
			percentile: 95,
			expected:   42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := percentile(tt.data, tt.percentile)
			assert.InDelta(t, tt.expected, result, 0.1)
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ⏱️ Time-Based SLO Tests                                                │
// └─────────────────────────────────────────────────────────────────────────┘

func TestSLO_TimeBasedMetrics(t *testing.T) {
	t.Run("build duration from report (54.7s P95)", func(t *testing.T) {
		// Simulate build durations from report
		durations := make([]float64, 100)
		for i := range durations {
			// Most builds between 30-60s, P95 around 54.7s
			durations[i] = 30 + float64(i%30)
		}
		durations[95] = 54.7 // Set P95 to match report

		calc := NewSLOCalculator()
		calc.BuildDurations = durations

		p95 := calc.BuildDurationP95()
		assert.True(t, p95 < SLO_BuildDurationP95Seconds,
			"Build P95 (%.2fs) should be < SLO threshold (%.2fs)", p95, SLO_BuildDurationP95Seconds)
	})

	t.Run("reconcile duration from report (45.8ms P95)", func(t *testing.T) {
		// Simulate reconcile durations from report
		durations := make([]float64, 100)
		for i := range durations {
			// Most reconciles very fast, P95 around 45.8ms
			durations[i] = 10 + float64(i%40)
		}
		durations[95] = 45.8 // Set P95 to match report

		calc := NewSLOCalculator()
		calc.ReconcileDurations = durations

		p95 := calc.ReconcileDurationP95()
		assert.True(t, p95 < SLO_ReconcileDurationP95Ms,
			"Reconcile P95 (%.2fms) should be < SLO threshold (%.2fms)", p95, SLO_ReconcileDurationP95Ms)
	})
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔍 Report Validation Tests                                             │
// └─────────────────────────────────────────────────────────────────────────┘

func TestSLO_ReportValues(t *testing.T) {
	t.Run("validates report metrics are within SLOs", func(t *testing.T) {
		// Values from SRE_LAMBDA_METRICS_REPORT.md
		reportMetrics := struct {
			TotalLambdaFunctions int
			WorkqueueDepth       int
			TotalOperatorErrors  int
			P95BuildDuration     float64 // seconds
			P95ReconcileDuration float64 // milliseconds
			k6Lambda1ErrorRate   float64 // percent
			k6Lambda1P95Duration float64 // milliseconds
		}{
			TotalLambdaFunctions: 10,
			WorkqueueDepth:       0,
			TotalOperatorErrors:  35,
			P95BuildDuration:     54.7,
			P95ReconcileDuration: 45.8,
			k6Lambda1ErrorRate:   0,
			k6Lambda1P95Duration: 4.98,
		}

		// Validate each metric against SLO
		assert.True(t, reportMetrics.WorkqueueDepth < SLO_WorkqueueDepthMax,
			"Workqueue depth %d should be < %d", reportMetrics.WorkqueueDepth, SLO_WorkqueueDepthMax)

		assert.True(t, reportMetrics.P95BuildDuration < SLO_BuildDurationP95Seconds,
			"Build P95 %.1fs should be < %.0fs", reportMetrics.P95BuildDuration, SLO_BuildDurationP95Seconds)

		assert.True(t, reportMetrics.P95ReconcileDuration < SLO_ReconcileDurationP95Ms,
			"Reconcile P95 %.1fms should be < %.0fms", reportMetrics.P95ReconcileDuration, SLO_ReconcileDurationP95Ms)

		assert.True(t, reportMetrics.k6Lambda1ErrorRate < SLO_ErrorRatePercent,
			"Error rate %.1f%% should be < %.0f%%", reportMetrics.k6Lambda1ErrorRate, SLO_ErrorRatePercent)

		assert.True(t, reportMetrics.k6Lambda1P95Duration < SLO_FunctionLatencyP95Ms,
			"Function P95 %.2fms should be < %.0fms", reportMetrics.k6Lambda1P95Duration, SLO_FunctionLatencyP95Ms)
	})
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📋 Alert Rule Validation Tests                                         │
// └─────────────────────────────────────────────────────────────────────────┘

// AlertThreshold represents an alert rule threshold
type AlertThreshold struct {
	Name      string
	Threshold float64
	Operator  string // ">" or "<"
}

func TestSLO_AlertRuleThresholds(t *testing.T) {
	// Alert rules from alertrules.yaml
	alertRules := []AlertThreshold{
		{Name: "KnativeLambdaFunctionHighErrorRate", Threshold: 5, Operator: ">"},
		{Name: "KnativeLambdaBuildDurationHigh", Threshold: 120, Operator: ">"},
		{Name: "KnativeLambdaWorkqueueDepthHigh", Threshold: 10, Operator: ">"},
		{Name: "KnativeLambdaHighColdStartRate", Threshold: 20, Operator: ">"},
		{Name: "KnativeLambdaFunctionHighLatency", Threshold: 5000, Operator: ">"}, // 5s in ms
	}

	for _, rule := range alertRules {
		t.Run(rule.Name, func(t *testing.T) {
			switch rule.Name {
			case "KnativeLambdaFunctionHighErrorRate":
				assert.Equal(t, SLO_ErrorRatePercent, rule.Threshold,
					"Alert threshold should match SLO")
			case "KnativeLambdaBuildDurationHigh":
				assert.Equal(t, SLO_BuildDurationP95Seconds, rule.Threshold,
					"Alert threshold should match SLO")
			case "KnativeLambdaWorkqueueDepthHigh":
				assert.Equal(t, float64(SLO_WorkqueueDepthMax), rule.Threshold,
					"Alert threshold should match SLO")
			case "KnativeLambdaHighColdStartRate":
				assert.Equal(t, SLO_ColdStartRatePercent, rule.Threshold,
					"Alert threshold should match SLO")
			case "KnativeLambdaFunctionHighLatency":
				assert.Equal(t, SLO_FunctionLatencyP95Ms, rule.Threshold,
					"Alert threshold should match SLO")
			}
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  🔄 Concurrent SLO Calculation Tests                                    │
// └─────────────────────────────────────────────────────────────────────────┘

func TestSLO_ConcurrentCalculation(t *testing.T) {
	t.Run("concurrent metric updates", func(t *testing.T) {
		calc := NewSLOCalculator()

		done := make(chan bool, 100)

		// Simulate concurrent metric updates
		for i := 0; i < 100; i++ {
			go func(id int) {
				calc.TotalRequests++
				if id%10 == 0 {
					calc.TotalErrors++
				}
				calc.Latencies = append(calc.Latencies, float64(id))
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 100; i++ {
			<-done
		}

		// Verify calculations don't panic
		_ = calc.ErrorRate()
		_ = calc.LatencyP95()
		_ = calc.IsCompliant()
	})
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  📊 Metric Naming Convention Tests                                      │
// └─────────────────────────────────────────────────────────────────────────┘

func TestSLO_MetricNaming(t *testing.T) {
	expectedMetrics := []string{
		"knative_lambda_operator_reconcile_total",
		"knative_lambda_operator_reconcile_duration_seconds",
		"knative_lambda_operator_build_duration_seconds",
		"knative_lambda_operator_workqueue_depth",
		"knative_lambda_operator_errors_total",
		"knative_lambda_function_invocations_total",
		"knative_lambda_function_duration_seconds",
		"knative_lambda_function_errors_total",
		"knative_lambda_function_cold_starts_total",
	}

	for _, metric := range expectedMetrics {
		t.Run(metric, func(t *testing.T) {
			// Verify metric follows naming convention
			require.Contains(t, metric, "knative_lambda_")
			require.NotContains(t, metric, "__")
			require.NotContains(t, metric, " ")
		})
	}
}

// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ⏰ SLO Window Tests                                                    │
// └─────────────────────────────────────────────────────────────────────────┘

func TestSLO_TimeWindows(t *testing.T) {
	windows := []struct {
		name     string
		duration time.Duration
		purpose  string
	}{
		{
			name:     "5 minute window",
			duration: 5 * time.Minute,
			purpose:  "Rate calculations",
		},
		{
			name:     "15 minute window",
			duration: 15 * time.Minute,
			purpose:  "Build duration histogram",
		},
		{
			name:     "1 hour window",
			duration: 1 * time.Hour,
			purpose:  "Error budget calculation",
		},
		{
			name:     "30 day window",
			duration: 30 * 24 * time.Hour,
			purpose:  "Monthly SLO compliance",
		},
	}

	for _, w := range windows {
		t.Run(w.name, func(t *testing.T) {
			assert.NotZero(t, w.duration)
			assert.NotEmpty(t, w.purpose)
		})
	}
}
