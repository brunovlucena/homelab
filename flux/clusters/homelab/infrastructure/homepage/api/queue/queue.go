package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// QueueManager handles background job processing (Golden Rule #5: Message Queue)
type QueueManager struct {
	client *redis.Client
	ctx    context.Context
}

// Job represents a background job
type Job struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Payload    map[string]interface{} `json:"payload"`
	CreatedAt  time.Time              `json:"created_at"`
	Attempts   int                    `json:"attempts"`
	MaxRetries int                    `json:"max_retries"`
}

// JobHandler handles specific job types
type JobHandler func(job *Job) error

// NewQueueManager creates a new queue manager
func NewQueueManager(client *redis.Client) *QueueManager {
	return &QueueManager{
		client: client,
		ctx:    context.Background(),
	}
}

// Queue names
const (
	AnalyticsQueue       = "analytics_queue"
	EmailQueue           = "email_queue"
	ImageProcessingQueue = "image_processing_queue"
	DefaultQueue         = "default_queue"
)

// EnqueueJob adds a job to the specified queue
func (q *QueueManager) EnqueueJob(queueName string, jobType string, payload map[string]interface{}) error {
	job := &Job{
		ID:         fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:       jobType,
		Payload:    payload,
		CreatedAt:  time.Now(),
		Attempts:   0,
		MaxRetries: 3,
	}

	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Add to queue with priority (FIFO)
	err = q.client.LPush(q.ctx, queueName, jobData).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	log.Printf("✅ Job %s enqueued to %s", job.ID, queueName)
	return nil
}

// DequeueJob retrieves and processes a job from the specified queue
func (q *QueueManager) DequeueJob(queueName string, handler JobHandler) error {
	// Block for 5 seconds waiting for a job
	result, err := q.client.BRPop(q.ctx, 5*time.Second, queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil // No jobs available
		}
		return fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) != 2 {
		return fmt.Errorf("unexpected result format")
	}

	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Process the job
	job.Attempts++
	if err := handler(&job); err != nil {
		log.Printf("❌ Job %s failed (attempt %d): %v", job.ID, job.Attempts, err)

		// Retry if under max attempts
		if job.Attempts < job.MaxRetries {
			jobData, _ := json.Marshal(job)
			q.client.LPush(q.ctx, queueName, jobData)
			log.Printf("🔄 Job %s requeued for retry", job.ID)
		} else {
			log.Printf("💀 Job %s exceeded max retries, moving to dead letter queue", job.ID)
			q.client.LPush(q.ctx, queueName+":failed", result[1])
		}
		return err
	}

	log.Printf("✅ Job %s completed successfully", job.ID)
	return nil
}

// StartWorker starts a background worker for the specified queue
func (q *QueueManager) StartWorker(queueName string, handler JobHandler) {
	log.Printf("🚀 Starting worker for queue: %s", queueName)

	for {
		if err := q.DequeueJob(queueName, handler); err != nil {
			log.Printf("❌ Worker error for queue %s: %v", queueName, err)
			time.Sleep(5 * time.Second) // Backoff on error
		}
	}
}

// GetQueueStats returns statistics about the queue
func (q *QueueManager) GetQueueStats(queueName string) (map[string]int64, error) {
	pending := q.client.LLen(q.ctx, queueName).Val()
	failed := q.client.LLen(q.ctx, queueName+":failed").Val()

	return map[string]int64{
		"pending": pending,
		"failed":  failed,
	}, nil
}

// ClearQueue clears all jobs from a queue
func (q *QueueManager) ClearQueue(queueName string) error {
	return q.client.Del(q.ctx, queueName, queueName+":failed").Err()
}

// Health checks if the queue system is healthy
func (q *QueueManager) Health() error {
	return q.client.Ping(q.ctx).Err()
}
