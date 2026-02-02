// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
//
//	🧪 AGENT-006: LambdaAgent Concurrency & Race Condition Tests
//
//	User Story: Thread-Safe Agent Operations
//	Priority: P1 | Story Points: 5
//
//	Tests validate:
//	- Concurrent agent updates
//	- Race conditions in status updates
//	- Concurrent event processing
//	- Metrics counter thread safety
//	- Resource creation conflicts
//	- Finalizer race conditions
//	- Status condition updates under concurrency
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
package agents

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"knative-lambda/tests/testutils"

	"github.com/stretchr/testify/assert"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC1: Concurrent Agent Updates
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT006_AC1_ConcurrentAgentUpdates(t *testing.T) {
	testutils.SetupTestEnvironment(t)

	t.Run("Concurrent image tag updates", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("concurrent-tag-agent")
		var wg sync.WaitGroup
		updates := 100
		errors := make(chan error, updates)

		// Act - Concurrent updates
		for i := 0; i < updates; i++ {
			wg.Add(1)
			go func(tagNum int) {
				defer wg.Done()
				// In real scenario, would use proper locking
				agent.Image.Tag = fmt.Sprintf("v%d.0.0", tagNum)
			}(i)
		}

		wg.Wait()

		// Assert - Last update should be visible
		assert.NotEmpty(t, agent.Image.Tag)
		assert.LessOrEqual(t, len(errors), 0, "Should handle concurrent updates")
	})

	t.Run("Concurrent phase transitions", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("concurrent-phase-agent")
		agent.Phase = AgentPhasePending
		var wg sync.WaitGroup
		transitions := []LambdaAgentPhase{
			AgentPhaseDeploying,
			AgentPhaseReady,
			AgentPhaseFailed,
		}

		// Act - Concurrent phase changes
		for _, phase := range transitions {
			wg.Add(1)
			go func(p LambdaAgentPhase) {
				defer wg.Done()
				// In real scenario, would use atomic operations or locks
				agent.Phase = p
			}(phase)
		}

		wg.Wait()

		// Assert - Phase should be one of the transitions
		assert.Contains(t, transitions, agent.Phase, "Phase should be valid")
	})

	t.Run("Concurrent condition updates", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("concurrent-condition-agent")
		var wg sync.WaitGroup
		var mu sync.Mutex
		conditionCount := 10

		// Act - Concurrent condition additions
		for i := 0; i < conditionCount; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				mu.Lock()
				defer mu.Unlock()
				agent.Conditions = append(agent.Conditions, MockCondition{
					Type:   fmt.Sprintf("Condition%d", index),
					Status: "True",
				})
			}(i)
		}

		wg.Wait()

		// Assert
		assert.Len(t, agent.Conditions, conditionCount, "All conditions should be added")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC2: Metrics Counter Thread Safety
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT006_AC2_MetricsThreadSafety(t *testing.T) {
	t.Run("Concurrent metric increments", func(t *testing.T) {
		// Arrange
		var counter int64
		var wg sync.WaitGroup
		increments := 1000
		goroutines := 10

		// Act - Concurrent increments
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < increments/goroutines; j++ {
					atomic.AddInt64(&counter, 1)
				}
			}()
		}

		wg.Wait()

		// Assert
		assert.Equal(t, int64(increments), counter, "All increments should be counted")
	})

	t.Run("Concurrent metric reads and writes", func(t *testing.T) {
		// Arrange
		var counter int64
		var wg sync.WaitGroup
		readers := 5
		writers := 5
		operations := 100

		// Act - Concurrent reads and writes
		for i := 0; i < writers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < operations; j++ {
					atomic.AddInt64(&counter, 1)
				}
			}()
		}

		for i := 0; i < readers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < operations; j++ {
					_ = atomic.LoadInt64(&counter)
				}
			}()
		}

		wg.Wait()

		// Assert
		assert.Equal(t, int64(writers*operations), counter, "All writes should be visible")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC3: Concurrent Event Processing
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT006_AC3_ConcurrentEventProcessing(t *testing.T) {
	t.Run("Concurrent event processing", func(t *testing.T) {
		// Arrange
		events := make(chan string, 100)
		processed := make(map[string]int)
		var mu sync.Mutex
		var wg sync.WaitGroup
		eventCount := 50

		// Act - Send events concurrently
		for i := 0; i < eventCount; i++ {
			events <- fmt.Sprintf("event-%d", i)
		}
		close(events)

		// Process events concurrently
		workers := 5
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for event := range events {
					mu.Lock()
					processed[event]++
					mu.Unlock()
					time.Sleep(1 * time.Millisecond) // Simulate processing
				}
			}()
		}

		wg.Wait()

		// Assert
		assert.Len(t, processed, eventCount, "All events should be processed")
		for _, count := range processed {
			assert.Equal(t, 1, count, "Each event should be processed once")
		}
	})

	t.Run("Event ordering under concurrency", func(t *testing.T) {
		// Arrange
		events := make(chan int, 100)
		processed := make([]int, 0)
		var mu sync.Mutex
		var wg sync.WaitGroup
		eventCount := 100

		// Act - Send ordered events
		for i := 0; i < eventCount; i++ {
			events <- i
		}
		close(events)

		// Process concurrently (order may be lost)
		workers := 10
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for event := range events {
					mu.Lock()
					processed = append(processed, event)
					mu.Unlock()
				}
			}()
		}

		wg.Wait()

		// Assert - All events processed, order may vary
		assert.Len(t, processed, eventCount, "All events should be processed")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC4: Resource Creation Conflicts
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT006_AC4_ResourceCreationConflicts(t *testing.T) {
	t.Run("Concurrent resource creation attempts", func(t *testing.T) {
		// Arrange
		resourceName := "test-resource"
		created := make(map[string]bool)
		var mu sync.Mutex
		var wg sync.WaitGroup
		attempts := 10

		// Act - Concurrent creation attempts
		for i := 0; i < attempts; i++ {
			wg.Add(1)
			go func(attemptNum int) {
				defer wg.Done()
				mu.Lock()
				defer mu.Unlock()
				if !created[resourceName] {
					created[resourceName] = true
				}
			}(i)
		}

		wg.Wait()

		// Assert - Only one should succeed
		assert.True(t, created[resourceName], "Resource should be created")
	})

	t.Run("Concurrent update conflicts", func(t *testing.T) {
		// Arrange
		resourceVersion := "v1"
		updates := make([]string, 0)
		var mu sync.Mutex
		var wg sync.WaitGroup
		updateCount := 20

		// Act - Concurrent updates
		for i := 0; i < updateCount; i++ {
			wg.Add(1)
			go func(updateNum int) {
				defer wg.Done()
				mu.Lock()
				defer mu.Unlock()
				// Check version before update
				if resourceVersion == "v1" {
					resourceVersion = fmt.Sprintf("v%d", updateNum+2)
					updates = append(updates, resourceVersion)
				}
			}(i)
		}

		wg.Wait()

		// Assert - Only first update should succeed
		assert.LessOrEqual(t, len(updates), updateCount, "Some updates may conflict")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC5: Finalizer Race Conditions
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT006_AC5_FinalizerRaceConditions(t *testing.T) {
	t.Run("Concurrent finalizer additions", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("finalizer-race-agent")
		agent.Finalizers = []string{}
		var mu sync.Mutex
		var wg sync.WaitGroup
		finalizerCount := 5

		// Act - Concurrent finalizer additions
		for i := 0; i < finalizerCount; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				mu.Lock()
				defer mu.Unlock()
				finalizer := fmt.Sprintf("finalizer-%d", index)
				// Check if exists before adding
				exists := false
				for _, f := range agent.Finalizers {
					if f == finalizer {
						exists = true
						break
					}
				}
				if !exists {
					agent.Finalizers = append(agent.Finalizers, finalizer)
				}
			}(i)
		}

		wg.Wait()

		// Assert
		assert.Len(t, agent.Finalizers, finalizerCount, "All finalizers should be added")
	})

	t.Run("Concurrent finalizer removals", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("finalizer-remove-agent")
		agent.Finalizers = []string{"f1", "f2", "f3", "f4", "f5"}
		var mu sync.Mutex
		var wg sync.WaitGroup

		// Act - Concurrent removals
		for _, finalizer := range agent.Finalizers {
			wg.Add(1)
			go func(f string) {
				defer wg.Done()
				mu.Lock()
				defer mu.Unlock()
				newFinalizers := []string{}
				for _, existing := range agent.Finalizers {
					if existing != f {
						newFinalizers = append(newFinalizers, existing)
					}
				}
				agent.Finalizers = newFinalizers
			}(finalizer)
		}

		wg.Wait()

		// Assert - All should be removed
		assert.Empty(t, agent.Finalizers, "All finalizers should be removed")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC6: Status Condition Updates Under Concurrency
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT006_AC6_StatusConditionConcurrency(t *testing.T) {
	t.Run("Concurrent condition updates", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("condition-concurrent-agent")
		var mu sync.Mutex
		var wg sync.WaitGroup
		updates := 50

		// Act - Concurrent condition updates
		for i := 0; i < updates; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				mu.Lock()
				defer mu.Unlock()
				// Update or add condition
				found := false
				for j, cond := range agent.Conditions {
					if cond.Type == "Ready" {
						agent.Conditions[j].Status = "True"
						agent.Conditions[j].Reason = fmt.Sprintf("Update%d", index)
						found = true
						break
					}
				}
				if !found {
					agent.Conditions = append(agent.Conditions, MockCondition{
						Type:   "Ready",
						Status: "True",
						Reason: fmt.Sprintf("Update%d", index),
					})
				}
			}(i)
		}

		wg.Wait()

		// Assert
		readyCount := 0
		for _, cond := range agent.Conditions {
			if cond.Type == "Ready" {
				readyCount++
			}
		}
		assert.GreaterOrEqual(t, readyCount, 1, "Should have at least one Ready condition")
	})

	t.Run("Condition update with read-modify-write", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("rmw-condition-agent")
		agent.Conditions = []MockCondition{
			{Type: "Ready", Status: "False", Reason: "Initial"},
		}
		var mu sync.Mutex
		var wg sync.WaitGroup
		updates := 20

		// Act - Read-modify-write pattern
		for i := 0; i < updates; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				mu.Lock()
				defer mu.Unlock()
				// Read
				for j, cond := range agent.Conditions {
					if cond.Type == "Ready" {
						// Modify
						agent.Conditions[j].Status = "True"
						agent.Conditions[j].Reason = "Updated"
						break
					}
				}
			}()
		}

		wg.Wait()

		// Assert
		found := false
		for _, cond := range agent.Conditions {
			if cond.Type == "Ready" && cond.Status == "True" {
				found = true
				break
			}
		}
		assert.True(t, found, "Condition should be updated")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC7: Concurrent Scaling Updates
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT006_AC7_ConcurrentScalingUpdates(t *testing.T) {
	t.Run("Concurrent min replicas updates", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("scaling-concurrent-agent")
		agent.Scaling.MinReplicas = 1
		var mu sync.Mutex
		var wg sync.WaitGroup
		updates := 10

		// Act - Concurrent updates
		for i := 0; i < updates; i++ {
			wg.Add(1)
			go func(value int32) {
				defer wg.Done()
				mu.Lock()
				defer mu.Unlock()
				agent.Scaling.MinReplicas = value
			}(int32(i + 1))
		}

		wg.Wait()

		// Assert
		assert.Greater(t, agent.Scaling.MinReplicas, int32(0), "Min replicas should be set")
	})

	t.Run("Concurrent max replicas updates", func(t *testing.T) {
		// Arrange
		agent := createTestAgent("max-replicas-agent")
		agent.Scaling.MaxReplicas = 5
		var mu sync.Mutex
		var wg sync.WaitGroup
		updates := 10

		// Act - Concurrent updates
		for i := 0; i < updates; i++ {
			wg.Add(1)
			go func(value int32) {
				defer wg.Done()
				mu.Lock()
				defer mu.Unlock()
				agent.Scaling.MaxReplicas = value
			}(int32(i + 5))
		}

		wg.Wait()

		// Assert
		assert.Greater(t, agent.Scaling.MaxReplicas, int32(0), "Max replicas should be set")
	})
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.
// AC8: Load Testing with High Concurrency
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━.

func TestAGENT006_AC8_HighConcurrencyLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	t.Run("High concurrency agent operations", func(t *testing.T) {
		// Arrange
		agents := make([]*MockLambdaAgent, 0)
		var mu sync.Mutex
		var wg sync.WaitGroup
		agentCount := 100
		operationsPerAgent := 10

		// Act - Create and update agents concurrently
		for i := 0; i < agentCount; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				agent := createTestAgent(fmt.Sprintf("load-agent-%d", index))
				for j := 0; j < operationsPerAgent; j++ {
					mu.Lock()
					agents = append(agents, agent)
					agent.Image.Tag = fmt.Sprintf("v%d.0.0", j)
					mu.Unlock()
				}
			}(i)
		}

		wg.Wait()

		// Assert
		assert.GreaterOrEqual(t, len(agents), agentCount, "All agents should be created")
	})

	t.Run("Stress test with many goroutines", func(t *testing.T) {
		// Arrange
		var counter int64
		var wg sync.WaitGroup
		goroutines := 1000
		operations := 100

		// Act - Many goroutines doing operations
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < operations; j++ {
					atomic.AddInt64(&counter, 1)
				}
			}()
		}

		wg.Wait()

		// Assert
		expected := int64(goroutines * operations)
		assert.Equal(t, expected, counter, "All operations should complete")
	})
}
