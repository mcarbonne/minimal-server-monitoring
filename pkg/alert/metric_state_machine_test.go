package alert

import (
	"strings"
	"testing"
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/notifier"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/scraping/provider"
	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/configmapper/customtypes"
	"gotest.tools/v3/assert"
)

func TestMetricStateMachine_SmartReminders(t *testing.T) {
	// Setup
	healthyThreshold := uint(1)
	unhealthyThreshold := uint(1)
	failureReminder := 1 * time.Hour
	failureReminderCount := uint(3)
	dailyReminder := customtypes.TimeOfDay{Hour: 8, Minute: 0}

	msm := MakeMetricStateMachine(healthyThreshold, unhealthyThreshold, failureReminder, failureReminderCount, dailyReminder)

	metricID := "test_metric"
	metricName := "Test Metric"
	// Start at 10:00 AM on Jan 1st
	start := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	// Helper to simulate update
	update := func(status provider.MetricStatus, timeOffset time.Duration) *notifier.Message {
		state := provider.MetricState{
			MetricID:    metricID,
			Name:        metricName,
			Status:      status,
			Description: "description",
		}
		return msm.Update(state, start.Add(timeOffset))
	}

	// 1. Initial Failure (10:00)
	msg := update(provider.Unhealthy, 0)
	assert.Assert(t, msg != nil)
	assert.Equal(t, notifier.Failure, msg.Type)
	assert.Assert(t, isContains(msg.Message, "failed"))

	// 2. Fast Reminder 1 (11:00)
	msg = update(provider.Unhealthy, 1*time.Hour)
	assert.Assert(t, msg != nil)
	assert.Assert(t, isContains(msg.Message, "reminder"))

	// 3. Fast Reminder 2 (12:00)
	msg = update(provider.Unhealthy, 2*time.Hour)
	assert.Assert(t, msg != nil)

	// 4. Fast Reminder 3 (13:00) - This is the 3rd reminder (failureReminderCount=3)
	msg = update(provider.Unhealthy, 3*time.Hour)
	assert.Assert(t, msg != nil)

	// 5. Check before next 08:00 (e.g. at 14:00, T+4h) -> Should be NO reminder
	msg = update(provider.Unhealthy, 4*time.Hour)
	assert.Assert(t, msg == nil)

	// 6. Check just before 08:00 next day (07:59 next day)
	// Start 10:00. Target 08:00 next day.
	// difference = 22h.
	msg = update(provider.Unhealthy, 22*time.Hour-1*time.Minute)
	assert.Assert(t, msg == nil)

	// 7. Slow Reminder 1 (08:00 next day / T+22h)
	msg = update(provider.Unhealthy, 22*time.Hour)
	assert.Assert(t, msg != nil, "Expected reminder at 08:00 next day")
	assert.Assert(t, isContains(msg.Message, "reminder"))

	// 8. Check before next day 08:00 (e.g. 12:00 next day, T+26h) -> Should be NO reminder
	msg = update(provider.Unhealthy, 26*time.Hour)
	assert.Assert(t, msg == nil)

	// 9. Slow Reminder 2 (08:00 day after next / T+22+24 => T+46h)
	msg = update(provider.Unhealthy, 46*time.Hour)
	assert.Assert(t, msg != nil)

	// 10. Recovery
	msg = update(provider.Healthy, 47*time.Hour)
	assert.Assert(t, msg != nil)
	assert.Equal(t, notifier.Recovery, msg.Type)

	// 11. New Failure - should reset counts
	msg = update(provider.Unhealthy, 48*time.Hour)
	assert.Assert(t, msg != nil)
	assert.Equal(t, notifier.Failure, msg.Type)

	// 12. Fast Reminder 1 (new cycle) - should be 1h after new failure
	// New failure at 48h. Next reminder at 49h.
	msg = update(provider.Unhealthy, 49*time.Hour)
	assert.Assert(t, msg != nil)
	assert.Assert(t, isContains(msg.Message, "reminder"))
}

func TestMetricStateMachine_Thresholds(t *testing.T) {
	// Setup: Unhealthy requires 3, Healthy requires 2
	msm := MakeMetricStateMachine(2, 3, 1*time.Hour, 3, customtypes.TimeOfDay{Hour: 8, Minute: 0})
	metricID := "test_metric"
	metricName := "Test Metric"
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	update := func(status provider.MetricStatus) *notifier.Message {
		return msm.Update(provider.MetricState{MetricID: metricID, Name: metricName, Status: status}, now)
	}

	// 1. Initial state: Healthy
	assert.Equal(t, true, msm.isHealthy)

	// 2. First failure: remains healthy
	assert.Assert(t, update(provider.Unhealthy) == nil)
	assert.Equal(t, true, msm.isHealthy)

	// 3. Second failure: remains healthy
	assert.Assert(t, update(provider.Unhealthy) == nil)
	assert.Equal(t, true, msm.isHealthy)

	// 4. Interrupted by a Success: counter resets
	assert.Assert(t, update(provider.Healthy) == nil)
	assert.Equal(t, true, msm.isHealthy)
	assert.Equal(t, uint(0), msm.oppositeInARow)

	// 5. Restart failures: 1, 2
	assert.Assert(t, update(provider.Unhealthy) == nil)
	assert.Assert(t, update(provider.Unhealthy) == nil)

	// 6. Third failure: becomes unhealthy
	msg := update(provider.Unhealthy)
	assert.Assert(t, msg != nil)
	assert.Equal(t, notifier.Failure, msg.Type)
	assert.Equal(t, false, msm.isHealthy)

	// 7. First recovery: remains unhealthy
	assert.Assert(t, update(provider.Healthy) == nil)
	assert.Equal(t, false, msm.isHealthy)

	// 8. Interrupted by a Failure: counter resets
	assert.Assert(t, update(provider.Unhealthy) == nil)
	assert.Equal(t, false, msm.isHealthy)
	assert.Equal(t, uint(0), msm.oppositeInARow)

	// 9. Second recovery attempt: 1, 2
	assert.Assert(t, update(provider.Healthy) == nil)
	msg = update(provider.Healthy)
	assert.Assert(t, msg != nil)
	assert.Equal(t, notifier.Recovery, msg.Type)
	assert.Equal(t, true, msm.isHealthy)
}

func TestMetricStateMachine_Removal(t *testing.T) {
	// Setup with high threshold
	msm := MakeMetricStateMachine(10, 1, 1*time.Hour, 3, customtypes.TimeOfDay{Hour: 8, Minute: 0})

	metricID := "test_metric"
	metricName := "Test Metric"
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// 1. Failure
	msm.Update(provider.MetricState{MetricID: metricID, Name: metricName, Status: provider.Unhealthy}, now)
	assert.Equal(t, false, msm.isHealthy)

	// 2. Removed - should force recovery even if threshold (10) is not met
	msg := msm.Update(provider.MetricState{MetricID: metricID, Name: metricName, Status: provider.Removed, Description: "service removed"}, now.Add(time.Minute))

	assert.Assert(t, msg != nil)
	assert.Equal(t, notifier.Recovery, msg.Type)
	assert.Assert(t, strings.Contains(msg.Message, "service removed"))
	assert.Equal(t, true, msm.isHealthy)
}

func isContains(s, substr string) bool {
	return strings.Contains(s, substr)
}
