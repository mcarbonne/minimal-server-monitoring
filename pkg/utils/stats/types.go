package stats

import (
	"fmt"
	"time"
)

type TimestampedData[T any] struct {
	Timestamp time.Time
	Data      T
}

func (tsdata TimestampedData[T]) String() string {
	return fmt.Sprintf("%v: %v", tsdata.Timestamp.UnixMilli(), tsdata.Data)
}
