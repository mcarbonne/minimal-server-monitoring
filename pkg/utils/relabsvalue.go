package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
)

type RelativeAbsoluteValue struct {
	absValue uint64
	relValue float64
	relative bool
}

func (relAbsVal *RelativeAbsoluteValue) GetValue(reference uint64) uint64 {
	if relAbsVal.relative {
		return uint64(relAbsVal.relValue * float64(reference))
	} else {
		return relAbsVal.absValue
	}
}

func RelativeAbsoluteValueFromString(value string) (RelativeAbsoluteValue, error) {
	var relAbsValue RelativeAbsoluteValue
	var err error
	if strings.HasSuffix(value, "%") {
		trimmed := strings.TrimSpace(value[0 : len(value)-1])
		relAbsValue.relValue, err = strconv.ParseFloat(trimmed, 64)
		if math.Signbit(relAbsValue.relValue) {
			err = fmt.Errorf("illegal relative value '%v'<0", relAbsValue.relValue)
			relAbsValue.relValue = 0
		}
		relAbsValue.relValue /= 100.
		relAbsValue.relative = true
	} else {
		relAbsValue.absValue, err = humanize.ParseBytes(strings.TrimSpace(value))
		relAbsValue.relative = false
	}
	return relAbsValue, err
}
