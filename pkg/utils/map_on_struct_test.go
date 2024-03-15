package utils_test

import (
	"encoding/json"
	"testing"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils"
	"gotest.tools/v3/assert"
)

type subStruct struct {
	Int int `json:"int" default:"5"`
}

type testStruct struct {
	Int              int               `json:"int"`
	IntDefault       int               `json:"int_d" default:"-3"`
	Uint             uint              `json:"uint"`
	UintDefault      uint              `json:"uint_d" default:"9"`
	Str              string            `json:"str"`
	StrDefault       string            `json:"str_d" default:"default_str"`
	Slice            []int             `json:"slice_int"`
	SliceEmpty       []int             `json:"slice_int_empty"`
	Map              map[string]string `json:"map_str"`
	MapEmpty         map[string]string `json:"map_str_empty"`
	Struct           subStruct         `json:"struct"`
	StructNotPresent subStruct         `json:"struct_to_present"`
}

func TestMapOnStruct1(t *testing.T) {
	var rawJson map[string]any
	myJsonString := `{"int":-5,
	"uint":7,
	"str":"str",
	"slice_int":[1,2,3],
	"map_str": {"a":"abc", "b":"def"},
	"struct": {"int":5}
	}`
	json.Unmarshal([]byte(myJsonString), &rawJson)
	data, err := utils.MapOnStruct[testStruct](rawJson)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, data.Int, -5)
	assert.Equal(t, data.IntDefault, -3)
	assert.Equal(t, data.Uint, uint(7))
	assert.Equal(t, data.UintDefault, uint(9))
	assert.Equal(t, data.Str, "str")
	assert.Equal(t, data.StrDefault, "default_str")
	assert.DeepEqual(t, data.Slice, []int{1, 2, 3})
	assert.DeepEqual(t, data.SliceEmpty, []int{})
	assert.DeepEqual(t, data.Map, map[string]string{"a": "abc", "b": "def"})
	assert.DeepEqual(t, data.MapEmpty, map[string]string{})
	assert.DeepEqual(t, data.Struct, subStruct{Int: 5})
}
