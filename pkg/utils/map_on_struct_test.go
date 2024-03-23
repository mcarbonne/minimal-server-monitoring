package utils_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils"
	"gotest.tools/v3/assert"
)

type subStruct struct {
	Int int `json:"int" default:"5"`
}

type testStruct struct {
	Int          int   `json:"int"`
	IntDefault   int   `json:"int_d" default:"-3"`
	Int8         int8  `json:"int8"`
	Int8Default  int8  `json:"int8_d" default:"-4"`
	Int16        int16 `json:"int16"`
	Int16Default int16 `json:"int16_d" default:"-5"`
	Int32        int32 `json:"int32"`
	Int32Default int32 `json:"int32_d" default:"-6"`
	Int64        int64 `json:"int64"`
	Int64Default int64 `json:"int64_d" default:"-7"`

	Uint          uint   `json:"uint"`
	UintDefault   uint   `json:"uint_d" default:"9"`
	Uint8         uint8  `json:"uint8"`
	Uint8Default  uint8  `json:"uint8_d" default:"10"`
	Uint16        uint16 `json:"uint16"`
	Uint16Default uint16 `json:"uint16_d" default:"11"`
	Uint32        uint32 `json:"uint32"`
	Uint32Default uint32 `json:"uint32_d" default:"12"`
	Uint64        uint64 `json:"uint64"`
	Uint64Default uint64 `json:"uint64_d" default:"13"`

	Str              string            `json:"str"`
	StrDefault       string            `json:"str_d" default:"default_str"`
	Slice            []int             `json:"slice_int"`
	SliceEmpty       []int             `json:"slice_int_empty" default:"[]"`
	Map              map[string]string `json:"map_str"`
	MapEmpty         map[string]string `json:"map_str_empty" default:"{}"`
	Struct           subStruct         `json:"struct"`
	StructNotPresent subStruct         `json:"struct_to_present" default:"{}"`
	Duration         time.Duration     `json:"duration"`
	DurationDefault  time.Duration     `json:"duration_d" default:"6s"`
}

func TestMapOnStruct1(t *testing.T) {
	var rawJson map[string]any
	myJsonString := `{"int":-5,
	"int8":-6,
	"int16":-7,
	"int32":-8,
	"int64":-9,
	"uint":7,
	"uint8":8,
	"uint16":9,
	"uint32":10,
	"uint64":11,
	"str":"str",
	"slice_int":[1,2,3],
	"map_str": {"a":"abc", "b":"def"},
	"struct": {"int":5},
	"duration":"5s"
	}`
	json.Unmarshal([]byte(myJsonString), &rawJson)
	data, err := utils.MapOnStruct[testStruct](rawJson)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, data.Int, -5)
	assert.Equal(t, data.IntDefault, -3)
	assert.Equal(t, data.Int8, int8(-6))
	assert.Equal(t, data.Int8Default, int8(-4))
	assert.Equal(t, data.Int16, int16(-7))
	assert.Equal(t, data.Int16Default, int16(-5))
	assert.Equal(t, data.Int32, int32(-8))
	assert.Equal(t, data.Int32Default, int32(-6))
	assert.Equal(t, data.Int64, int64(-9))
	assert.Equal(t, data.Int64Default, int64(-7))

	assert.Equal(t, data.Uint, uint(7))
	assert.Equal(t, data.UintDefault, uint(9))
	assert.Equal(t, data.Uint8, uint8(8))
	assert.Equal(t, data.Uint8Default, uint8(10))
	assert.Equal(t, data.Uint16, uint16(9))
	assert.Equal(t, data.Uint16Default, uint16(11))
	assert.Equal(t, data.Uint32, uint32(10))
	assert.Equal(t, data.Uint32Default, uint32(12))
	assert.Equal(t, data.Uint64, uint64(11))
	assert.Equal(t, data.Uint64Default, uint64(13))

	assert.Equal(t, data.Str, "str")
	assert.Equal(t, data.StrDefault, "default_str")
	assert.DeepEqual(t, data.Slice, []int{1, 2, 3})
	assert.DeepEqual(t, data.SliceEmpty, []int{})
	assert.DeepEqual(t, data.Map, map[string]string{"a": "abc", "b": "def"})
	assert.DeepEqual(t, data.MapEmpty, map[string]string{})
	assert.DeepEqual(t, data.Struct, subStruct{Int: 5})

	assert.Equal(t, data.Duration, time.Second*5)
	assert.Equal(t, data.DurationDefault, time.Second*6)
}
