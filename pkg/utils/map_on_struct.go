package utils

import (
	"fmt"
	"reflect"
	"strconv"
)

func getAs[Input any, Output any](input any) (reflect.Value, error) {
	value, ok := input.(Input)
	if !ok {
		return reflect.Value{}, fmt.Errorf("was expecting type %v, got %v (%v)", reflect.TypeFor[Input](), reflect.TypeOf(input), input)
	}
	targetType := reflect.TypeFor[Output]()
	output := reflect.ValueOf(value)
	if !output.CanConvert(targetType) {
		return reflect.Value{}, fmt.Errorf("unable to convert to %v", targetType)
	}
	return output.Convert(targetType), nil
}

func getDefaultValue(field reflect.StructField) (reflect.Value, error) {
	var defaultValue reflect.Value

	if tag, ok := field.Tag.Lookup("default"); ok {
		switch kind := field.Type.Kind(); kind {
		case reflect.Slice, reflect.Array:
			if tag == "[]" {
				defaultValue = reflect.MakeSlice(field.Type, 0, 0)
			} else {
				return reflect.Value{}, fmt.Errorf("unsupported default value for array: %v", tag)
			}
		case reflect.Map:
			if tag == "{}" {
				defaultValue = reflect.MakeMap(field.Type)
			} else {
				return reflect.Value{}, fmt.Errorf("unsupported default value for map: %v", tag)
			}
		case reflect.Struct:
			if tag == "{}" {
				var err error
				defaultValue, err = mapOnStruct(field.Type, map[string]any{}, "")
				if err != nil {
					return reflect.Value{}, fmt.Errorf("unable to default struct: %v", err)
				}
			} else {
				return reflect.Value{}, fmt.Errorf("unsupported default value for map: %v", tag)
			}
		case reflect.Int:
			intVal, err := strconv.ParseInt(tag, 10, 64)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("unable to parse int: %v", err)
			}
			defaultValue = reflect.ValueOf(int(intVal))
		case reflect.Uint:
			intVal, err := strconv.ParseUint(tag, 10, 64)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("unable to parse int: %v", err)
			}
			defaultValue = reflect.ValueOf(uint(intVal))
		case reflect.String:
			defaultValue = reflect.ValueOf(tag)
		default:
			return reflect.Value{}, fmt.Errorf("unsupported kind %v for default", kind)
		}

	}

	return defaultValue, nil
}

func mapOnSlice(type_ reflect.Type, rawJson any, level string) (reflect.Value, error) {
	json, ok := rawJson.([]any)
	if !ok {
		return reflect.Value{}, fmt.Errorf("json does not match slice")
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
		return reflect.Value{}, fmt.Errorf("json does not match struct")
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
					return reflect.Value{}, err
				}
				if !defaultValue.IsValid() {
					return reflect.Value{}, fmt.Errorf("missing field %v/%v", level, tag)
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
		return reflect.Value{}, fmt.Errorf("json does not match struct")
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

func mapOnAny(type_ reflect.Type, rawJson any, level string) (reflect.Value, error) {
	switch kind := type_.Kind(); kind {
	case reflect.String:
		return getAs[string, string](rawJson)
	case reflect.Uint:
		return getAs[float64, uint](rawJson)
	case reflect.Int:
		return getAs[float64, int](rawJson)
	case reflect.Struct:
		return mapOnStruct(type_, rawJson, level)
	case reflect.Map:
		return mapOnMap(type_, rawJson, level)
	case reflect.Slice:
		return mapOnSlice(type_, rawJson, level)
	case reflect.Interface:
		return reflect.ValueOf(rawJson), nil
	default:
		return reflect.Value{}, fmt.Errorf("mapOnAny: unsupported type %v", kind)
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
