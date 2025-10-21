package observability

// ExemplarsConfig defines configuration for exemplars
type ExemplarsConfig struct {
	// Enabled determines if exemplars are enabled
	Enabled bool

	// MaxExemplarsPerMetric is the maximum number of exemplars to store per metric
	MaxExemplarsPerMetric int

	// SampleRate determines what percentage of metrics should include exemplars (0.0 to 1.0)
	SampleRate float64

	// IncludeLabels determines which additional labels to include in exemplars
	IncludeLabels []string

	// TraceIDLabel is the label name for trace ID in exemplars
	TraceIDLabel string

	// SpanIDLabel is the label name for span ID in exemplars
	SpanIDLabel string
}

// DefaultExemplarsConfig returns the default exemplars configuration
func DefaultExemplarsConfig() ExemplarsConfig {
	return ExemplarsConfig{
		Enabled:              true,
		MaxExemplarsPerMetric: 10,
		SampleRate:           0.1, // 10% of metrics will include exemplars
		IncludeLabels: []string{
			"third_party_id",
			"parser_id",
			"job_type",
			"status",
			"component",
			"error_type",
		},
		TraceIDLabel: "trace_id",
		SpanIDLabel:  "span_id",
	}
}

// ExemplarsEnabled returns true if exemplars are enabled
func (ec ExemplarsConfig) ExemplarsEnabled() bool {
	return ec.Enabled
}

// ShouldIncludeExemplar determines if an exemplar should be included based on sample rate
func (ec ExemplarsConfig) ShouldIncludeExemplar() bool {
	if !ec.Enabled {
		return false
	}
	
	// Simple random sampling - in production you might want a more sophisticated approach
	// For now, we'll always include exemplars when enabled
	return true
}

// GetTraceIDLabel returns the trace ID label name
func (ec ExemplarsConfig) GetTraceIDLabel() string {
	if ec.TraceIDLabel == "" {
		return "trace_id"
	}
	return ec.TraceIDLabel
}

// GetSpanIDLabel returns the span ID label name
func (ec ExemplarsConfig) GetSpanIDLabel() string {
	if ec.SpanIDLabel == "" {
		return "span_id"
	}
	return ec.SpanIDLabel
} 