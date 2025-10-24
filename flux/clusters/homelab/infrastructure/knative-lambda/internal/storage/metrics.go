// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	📊 STORAGE METRICS - Metrics collection for storage operations
//
//	🎯 Purpose: Track performance and usage metrics for storage operations
//	💡 Features: Duration tracking, byte counting, operation counting
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package storage

import (
	"context"
	"time"

	"knative-lambda-new/internal/observability"
)

// 📊 MetricsRecorder - "Records metrics for storage operations"
type MetricsRecorder struct {
	obs      *observability.Observability
	provider StorageProvider
}

// 🏗️ NewMetricsRecorder - "Create new metrics recorder"
func NewMetricsRecorder(obs *observability.Observability, provider StorageProvider) *MetricsRecorder {
	return &MetricsRecorder{
		obs:      obs,
		provider: provider,
	}
}

// 📊 RecordUpload - "Record metrics for upload operation"
func (m *MetricsRecorder) RecordUpload(ctx context.Context, bucket, key string, size int64, duration time.Duration, err error) {
	// Log upload operation with metrics
	if err == nil {
		m.obs.Info(ctx, "Storage upload completed",
			"provider", string(m.provider),
			"bucket", bucket,
			"key", key,
			"size_bytes", size,
			"duration_seconds", duration.Seconds())
	} else {
		m.obs.Error(ctx, err, "Storage upload failed",
			"provider", string(m.provider),
			"bucket", bucket,
			"key", key,
			"duration_seconds", duration.Seconds(),
			"error_type", errorType(err))
	}
}

// 📊 RecordDownload - "Record metrics for download operation"
func (m *MetricsRecorder) RecordDownload(ctx context.Context, bucket, key string, size int64, duration time.Duration, err error) {
	// Log download operation with metrics
	if err == nil {
		m.obs.Info(ctx, "Storage download completed",
			"provider", string(m.provider),
			"bucket", bucket,
			"key", key,
			"size_bytes", size,
			"duration_seconds", duration.Seconds())
	} else {
		m.obs.Error(ctx, err, "Storage download failed",
			"provider", string(m.provider),
			"bucket", bucket,
			"key", key,
			"duration_seconds", duration.Seconds(),
			"error_type", errorType(err))
	}
}

// 📊 RecordDelete - "Record metrics for delete operation"
func (m *MetricsRecorder) RecordDelete(ctx context.Context, bucket, key string, duration time.Duration, err error) {
	// Log delete operation with metrics
	if err == nil {
		m.obs.Info(ctx, "Storage delete completed",
			"provider", string(m.provider),
			"bucket", bucket,
			"key", key,
			"duration_seconds", duration.Seconds())
	} else {
		m.obs.Error(ctx, err, "Storage delete failed",
			"provider", string(m.provider),
			"bucket", bucket,
			"key", key,
			"duration_seconds", duration.Seconds(),
			"error_type", errorType(err))
	}
}

// 📊 RecordExistsCheck - "Record metrics for exists check operation"
func (m *MetricsRecorder) RecordExistsCheck(ctx context.Context, bucket, key string, exists bool, duration time.Duration, err error) {
	// Log exists check operation with metrics
	if err == nil {
		m.obs.Info(ctx, "Storage exists check completed",
			"provider", string(m.provider),
			"bucket", bucket,
			"key", key,
			"exists", exists,
			"duration_seconds", duration.Seconds())
	} else {
		m.obs.Error(ctx, err, "Storage exists check failed",
			"provider", string(m.provider),
			"bucket", bucket,
			"key", key,
			"duration_seconds", duration.Seconds(),
			"error_type", errorType(err))
	}
}

// 🔧 errorType - "Classify error type for metrics"
func errorType(err error) string {
	if err == nil {
		return "none"
	}

	// Check for common error types
	errStr := err.Error()

	if IsRetryable(err) {
		return "transient"
	}

	// Classify based on error string (basic classification)
	switch {
	case contains(errStr, "NotFound", "NoSuchKey"):
		return "not_found"
	case contains(errStr, "AccessDenied", "Forbidden"):
		return "access_denied"
	case contains(errStr, "timeout", "context deadline exceeded"):
		return "timeout"
	case contains(errStr, "connection refused", "network"):
		return "network"
	default:
		return "other"
	}
}

// 🔧 contains - "Check if string contains any of the substrings"
func contains(str string, subs ...string) bool {
	for _, sub := range subs {
		if len(str) >= len(sub) {
			for i := 0; i <= len(str)-len(sub); i++ {
				if str[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// 📊 MeasuredOperation - "Helper to measure and record operation metrics"
type MeasuredOperation struct {
	recorder  *MetricsRecorder
	ctx       context.Context
	operation string
	bucket    string
	key       string
	size      int64
	startTime time.Time
}

// 🏗️ StartMeasuredOperation - "Start measuring an operation"
func (m *MetricsRecorder) StartMeasuredOperation(ctx context.Context, operation, bucket, key string, size int64) *MeasuredOperation {
	return &MeasuredOperation{
		recorder:  m,
		ctx:       ctx,
		operation: operation,
		bucket:    bucket,
		key:       key,
		size:      size,
		startTime: time.Now(),
	}
}

// ✅ Complete - "Complete the measured operation and record metrics"
func (mo *MeasuredOperation) Complete(err error) {
	duration := time.Since(mo.startTime)

	switch mo.operation {
	case "upload":
		mo.recorder.RecordUpload(mo.ctx, mo.bucket, mo.key, mo.size, duration, err)
	case "download":
		mo.recorder.RecordDownload(mo.ctx, mo.bucket, mo.key, mo.size, duration, err)
	case "delete":
		mo.recorder.RecordDelete(mo.ctx, mo.bucket, mo.key, duration, err)
	case "exists":
		mo.recorder.RecordExistsCheck(mo.ctx, mo.bucket, mo.key, mo.size > 0, duration, err)
	}
}
