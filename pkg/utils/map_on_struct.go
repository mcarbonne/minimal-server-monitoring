package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func getAs[Output any](input any) (Output, error) {
	value, ok := input.(Output)
	if !ok {
		return Dummy[Output](), fmt.Errorf("was expecting type %v, got %v (%v)", reflect.TypeFor[Output](), reflect.TypeOf(input), input)
	}
	return value, nil
}

func getDefaultValueInt(valueAsString string, type_ reflect.Type) (reflect.Value, error) {
	defaultValue := reflect.New(type_).Elem()
	intVal, err := strconv.ParseInt(valueAsString, 10, 64)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("unable to parse int: %v", err)
	}
	reflect.New(type_)
	defaultValue.SetInt(intVal)
	return defaultValue, nil
}

func getDefaultValueUint(valueAsString string, type_ reflect.Type) (reflect.Value, error) {
	defaultValue := reflect.New(type_).Elem()
	uintVal, err := strconv.ParseUint(valueAsString, 10, 64)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("unable to parse int: %v", err)
	}
	reflect.New(type_)
	defaultValue.SetUint(uintVal)
	return defaultValue, nil
}

func getDefaultValue(field reflect.StructField) (reflect.Value, error) {
	if tag, ok := field.Tag.Lookup("default"); ok {

		if field.Type == reflect.TypeOf(time.Second) {
			duration, err := time.ParseDuration(tag)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("unable to parse time.Duration: %v", err)
			}
			return reflect.ValueOf(duration), nil
		}

		switch kind := field.Type.Kind(); kind {
		case reflect.Slice, reflect.Array:
			if tag == "[]" {
				return reflect.MakeSlice(field.Type, 0, 0), nil
			} else {
				return reflect.Value{}, fmt.Errorf("unsupported default value for array: %v", tag)
			}
		case reflect.Map:
			if tag == "{}" {
				return reflect.MakeMap(field.Type), nil
			} else {
				return reflect.Value{}, fmt.Errorf("unsupported default value for map: %v", tag)
			}
		case reflect.Struct:
			if tag == "{}" {
				defaultValue, err := mapOnStruct(field.Type, map[string]any{}, "")
				if err != nil {
					return reflect.Value{}, fmt.Errorf("unable to default struct: %v", err)
				} else {
					return defaultValue, nil
				}
			} else {
				return reflect.Value{}, fmt.Errorf("unsupported default value for map: %v", tag)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return getDefaultValueInt(tag, field.Type)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return getDefaultValueUint(tag, field.Type)
		case reflect.String:
			return reflect.ValueOf(tag), nil
		default:
			return reflect.Value{}, fmt.Errorf("unsupported kind %v for default", kind)
		}

	}
	return reflect.Value{}, nil
}

func mapOnSlice(type_ reflect.Type, rawJson any, level string) (reflect.Value, error) {
	json, ok := rawJson.([]any)
	if !ok {
		return reflect.Value{}, fmt.Errorf("[%v] json does not match slice", level)
	}

	elemType := type_.Elem()
	slice := reflect.MakeSlice(type_, 0, 0)

	for i, elem := range json {
		value, err := mapOnAny(elemType, elem, fmt.Sprintf("%v[%v]", level, i))
		if err != nil {
			return reflect.Value{}, err
		}
		slice = reflect.Append(slice, value)
	}
	return slice, nil
}

func mapOnStruct(type_ reflect.Type, rawJson any, level string) (reflect.Value, error) {
	json, ok := rawJson.(map[string]any)
	if !ok {
		return reflect.Value{}, fmt.Errorf("[%v] json does not match struct", level)
	}

	if type_.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("[%v] work only on struct, %v provided", level, type_.Kind())
	}

	target := reflect.New(type_).Elem()

	for i := range type_.NumField() {
		field := type_.Field(i)
		if tag, ok := field.Tag.Lookup("json"); ok {
			jsonValue, ok := json[tag]
			if !ok {
				defaultValue, err := getDefaultValue(field)
				if err != nil {
					return reflect.Value{}, fmt.Errorf("[%v/%v] unable to parse default value: %v", level, tag, err)
				}
				if !defaultValue.IsValid() {
					return reflect.Value{}, fmt.Errorf("[%v/%v] missing field", level, tag)
				} else {
					target.Field(i).Set(defaultValue)
				}
			} else {
				value, err := mapOnAny(field.Type, jsonValue, level+"/"+tag)
				if err != nil {
					return reflect.Value{}, err
				} else {
					target.Field(i).Set(value)
				}

			}
		}
	}
	return target, nil
}

func mapOnMap(type_ reflect.Type, rawJson any, level string) (reflect.Value, error) {
	json, ok := rawJson.(map[string]any)
	if !ok {
		return reflect.Value{}, fmt.Errorf("[%v] json does not match map", level)
	}

	if type_.Kind() != reflect.Map {
		return reflect.Value{}, fmt.Errorf("[%v] work only on map", level)
	}

	elemType := type_.Elem()
	keyType := type_.Key()

	if keyType.Kind() != reflect.String {
		return reflect.Value{}, fmt.Errorf("[%v] only string is allowed for key (%v given)", level, keyType)
	}
	target := reflect.MakeMapWithSize(type_, 0)
	for k, v := range json {
		value, err := mapOnAny(elemType, v, level+"/"+k)
		if err != nil {
			return reflect.Value{}, err
		}
		target.SetMapIndex(reflect.ValueOf(k), value)
	}
	return target, nil
}

func mapOnInt(type_ reflect.Type, rawJson any) (reflect.Value, error) {
	value := reflect.New(type_).Elem()
	value.SetInt(int64(rawJson.(float64)))
	return value, nil
}

func mapOnUint(type_ reflect.Type, rawJson any) (reflect.Value, error) {
	value := reflect.New(type_).Elem()
	value.SetUint(uint64(rawJson.(float64)))
	return value, nil
}

func mapOnString(rawJson any) (reflect.Value, error) {
	value, err := getAs[string](rawJson)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(value), nil
}

func mapOnAny(type_ reflect.Type, rawJson any, level string) (reflect.Value, error) {
	if type_ == reflect.TypeOf(time.Second) {
		durationAsString, err := getAs[string](rawJson)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("[%v] unable to parse time.Duration: %v", level, err)
		}
		var duration time.Duration
		duration, err = time.ParseDuration(durationAsString)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("[%v] unable to parse time.Duration: %v", level, err)
		}
		return reflect.ValueOf(duration), nil
	}

	switch kind := type_.Kind(); kind {
	case reflect.String:
		return mapOnString(rawJson)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return mapOnUint(type_, rawJson)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return mapOnInt(type_, rawJson)
	case reflect.Struct:
		return mapOnStruct(type_, rawJson, level)
	case reflect.Map:
		return mapOnMap(type_, rawJson, level)
	case reflect.Slice:
		return mapOnSlice(type_, rawJson, level)
	case reflect.Interface:
		return reflect.ValueOf(rawJson), nil
	default:
		return reflect.Value{}, fmt.Errorf("[%v] mapOnAny: unsupported type %v", level, kind)
	}
}

func MapOnStruct[T any](rawJson map[string]any) (T, error) {
	if kind := reflect.TypeFor[T]().Kind(); kind != reflect.Struct {
		var output T
		return output, fmt.Errorf("require a struct type, %v provided", kind)
	}

	s, err := mapOnStruct(reflect.TypeFor[T](), rawJson, "")

	if err != nil {
		return Dummy[T](), err
	} else {
		return s.Interface().(T), nil
	}
}
