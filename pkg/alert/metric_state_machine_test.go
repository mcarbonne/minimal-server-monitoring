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
	update := func(isHealthy bool, timeOffset time.Duration) *notifier.Message {
		state := provider.MetricState{
			MetricID:    metricID,
			Name:        metricName,
			IsHealthy:   isHealthy,
			Description: "description",
		}
		return msm.Update(state, start.Add(timeOffset))
	}

	// 1. Initial Failure (10:00)
	msg := update(false, 0)
	assert.Assert(t, msg != nil)
	assert.Equal(t, notifier.Failure, msg.Type)
	assert.Assert(t, isContains(msg.Message, "failed"))

	// 2. Fast Reminder 1 (11:00)
	msg = update(false, 1*time.Hour)
	assert.Assert(t, msg != nil)
	assert.Assert(t, isContains(msg.Message, "reminder"))

	// 3. Fast Reminder 2 (12:00)
	msg = update(false, 2*time.Hour)
	assert.Assert(t, msg != nil)

	// 4. Fast Reminder 3 (13:00) - This is the 3rd reminder (failureReminderCount=3)
	msg = update(false, 3*time.Hour)
	assert.Assert(t, msg != nil)

	// 5. Check before next 08:00 (e.g. at 14:00, T+4h) -> Should be NO reminder
	msg = update(false, 4*time.Hour)
	assert.Assert(t, msg == nil)

	// 6. Check just before 08:00 next day (07:59 next day)
	// Start 10:00. Target 08:00 next day.
	// difference = 22h.
	msg = update(false, 22*time.Hour-1*time.Minute)
	assert.Assert(t, msg == nil)

	// 7. Slow Reminder 1 (08:00 next day / T+22h)
	msg = update(false, 22*time.Hour)
	assert.Assert(t, msg != nil, "Expected reminder at 08:00 next day")
	assert.Assert(t, isContains(msg.Message, "reminder"))

	// 8. Check before next day 08:00 (e.g. 12:00 next day, T+26h) -> Should be NO reminder
	msg = update(false, 26*time.Hour)
	assert.Assert(t, msg == nil)

	// 9. Slow Reminder 2 (08:00 day after next / T+22+24 => T+46h)
	msg = update(false, 46*time.Hour)
	assert.Assert(t, msg != nil)

	// 10. Recovery
	msg = update(true, 47*time.Hour)
	assert.Assert(t, msg != nil)
	assert.Equal(t, notifier.Recovery, msg.Type)

	// 11. New Failure - should reset counts
	msg = update(false, 48*time.Hour)
	assert.Assert(t, msg != nil)
	assert.Equal(t, notifier.Failure, msg.Type)

	// 12. Fast Reminder 1 (new cycle) - should be 1h after new failure
	// New failure at 48h. Next reminder at 49h.
	msg = update(false, 49*time.Hour)
	assert.Assert(t, msg != nil)
	assert.Assert(t, isContains(msg.Message, "reminder"))
}

func isContains(s, substr string) bool {
	return strings.Contains(s, substr)
}
