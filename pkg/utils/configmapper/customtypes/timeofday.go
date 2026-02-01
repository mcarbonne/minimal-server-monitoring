package customtypes

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidTimeFormat = errors.New("invalid time format (expected HH:MM)")
	ErrInvalidTimeValues = errors.New("invalid numeric values in time")
	ErrTimeOutOfRange    = errors.New("time out of range")
)

type TimeOfDay struct {
	Hour, Minute int
}

func ParseTimeOfDay(s string) (TimeOfDay, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return TimeOfDay{}, fmt.Errorf("%w: %s", ErrInvalidTimeFormat, s)
	}

	h, errH := strconv.Atoi(parts[0])
	m, errM := strconv.Atoi(parts[1])

	if errH != nil || errM != nil {
		return TimeOfDay{}, fmt.Errorf("%w: %s", ErrInvalidTimeValues, s)
	}

	// Validation logic
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return TimeOfDay{}, fmt.Errorf("%w: %s", ErrTimeOutOfRange, s)
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
