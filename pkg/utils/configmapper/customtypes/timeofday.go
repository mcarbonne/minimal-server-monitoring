package customtypes

import (
	"fmt"
	"strconv"
	"strings"
)

type TimeOfDay struct {
	Hour, Minute int
}

func ParseTimeOfDay(s string) (TimeOfDay, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return TimeOfDay{}, fmt.Errorf("invalid time format: %s (expected HH:MM)", s)
	}

	h, errH := strconv.Atoi(parts[0])
	m, errM := strconv.Atoi(parts[1])

	if errH != nil || errM != nil {
		return TimeOfDay{}, fmt.Errorf("invalid numeric values in time: %s", s)
	}

	// Validation logic
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return TimeOfDay{}, fmt.Errorf("time out of range: %02d:%02d", h, m)
	}

	return TimeOfDay{Hour: h, Minute: m}, nil
}

func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour, t.Minute)
}

func (t *TimeOfDay) UnmarshalText(text []byte) error {
	timeOfDay, err := ParseTimeOfDay(string(text))
	if err != nil {
		return err
	}
	*t = timeOfDay
	return nil
}
