package testutils

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Phase represents a timing phase in a process.
type Phase struct {
	Name     string
	Duration time.Duration
}

// RunTimingTest runs a common timing test pattern.
func RunTimingTest(t *testing.T, testName string, startTime time.Time, endTime time.Time, maxDuration time.Duration, phases []Phase) {
	t.Run(testName, func(t *testing.T) {
		// Act
		actualDuration := endTime.Sub(startTime)

		// Assert
		assert.LessOrEqual(t, actualDuration.Seconds(), maxDuration.Seconds(),
			fmt.Sprintf("%s should be under %v", testName, maxDuration))
	})

	// Only run breakdown test if phases are provided
	if len(phases) > 0 {
		t.Run(fmt.Sprintf("%s breakdown", testName), func(t *testing.T) {
			// Act
			var totalDuration time.Duration
			for _, phase := range phases {
				totalDuration += phase.Duration
			}

			// Assert
			assert.LessOrEqual(t, totalDuration.Seconds(), maxDuration.Seconds(),
				fmt.Sprintf("Total %s should be under %v", testName, maxDuration))
		})
	}
}

// IntegrationTestData represents data for integration tests.
type IntegrationTestData struct {
	Name        string
	Description string
	Value       bool
}

// RunIntegrationTest runs a common integration test pattern.
func RunIntegrationTest(t *testing.T, testName string, testData []IntegrationTestData, successMessage string) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run(testName, func(t *testing.T) {
		// Assert all criteria
		for _, data := range testData {
			assert.True(t, data.Value, data.Description)
		}

		t.Logf("🎯 %s", successMessage)
	})
}
