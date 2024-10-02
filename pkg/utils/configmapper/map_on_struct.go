package configmapper

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/pkg/utils"
)

func getAs[Output any](input any) (Output, error) {
	value, ok := input.(Output)
	if !ok {
		return utils.Dummy[Output](), fmt.Errorf("was expecting type %v, got %v (%v)", reflect.TypeFor[Output](), reflect.TypeOf(input), input)
	}
	return value, nil
}

func stringToInt(type_ reflect.Type, valueAsString string) (reflect.Value, error) {
	defaultValue := reflect.New(type_).Elem()
	intVal, err := strconv.ParseInt(valueAsString, 10, 64)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("unable to parse int: %v", err)
	}
	reflect.New(type_)
	defaultValue.SetInt(intVal)
	return defaultValue, nil
}

func stringToUint(type_ reflect.Type, valueAsString string) (reflect.Value, error) {
	defaultValue := reflect.New(type_).Elem()
	uintVal, err := strconv.ParseUint(valueAsString, 10, 64)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("unable to parse int: %v", err)
	}
	reflect.New(type_)
	defaultValue.SetUint(uintVal)
	return defaultValue, nil
}

func getDefaultValue(ctx *Context, field reflect.StructField) (reflect.Value, error) {
	if tag, ok := field.Tag.Lookup("default"); ok {

		customParser, err := ctx.getCustomParserIfAny(&field)
		if err != nil {
			return reflect.Value{}, err
		} else if customParser != nil {
			return (*customParser)(tag)
		}

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
			} else if strings.HasPrefix(tag, "[") && strings.HasSuffix(tag, "]") {
				out := reflect.MakeSlice(field.Type, 0, 0)
				elements := strings.Split(tag[1:len(tag)-1], ",")
				for _, element := range elements {
					value, err := mapOnAny(ctx, field.Type.Elem(), strings.TrimSpace(element), "")
					if err != nil {
						return reflect.Value{}, err
					}
					out = reflect.Append(out, value)
				}
				return out, nil
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
				defaultValue, err := mapOnStruct(ctx, field.Type, map[string]any{}, "")
				if err != nil {
					return reflect.Value{}, fmt.Errorf("unable to default struct: %v", err)
				} else {
					return defaultValue, nil
				}
			} else {
				return reflect.Value{}, fmt.Errorf("unsupported default value for map: %v", tag)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return stringToInt(field.Type, tag)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return stringToUint(field.Type, tag)
		case reflect.String:
			return reflect.ValueOf(tag), nil
		default:
			return reflect.Value{}, fmt.Errorf("unsupported kind %v for default", kind)
		}

	}
	return reflect.Value{}, nil
}

func mapOnSlice(ctx *Context, type_ reflect.Type, raw any, level string) (reflect.Value, error) {
	asSlice, ok := raw.([]any)
	if !ok {
		return reflect.Value{}, fmt.Errorf("[%v] does not match slice", level)
	}

	elemType := type_.Elem()
	slice := reflect.MakeSlice(type_, 0, 0)

	for i, elem := range asSlice {
		value, err := mapOnAny(ctx, elemType, elem, fmt.Sprintf("%v[%v]", level, i))
		if err != nil {
			return reflect.Value{}, err
		}
		slice = reflect.Append(slice, value)
	}
	return slice, nil
}

func mapOnStruct(ctx *Context, type_ reflect.Type, raw any, level string) (reflect.Value, error) {
	asMap, ok := raw.(map[string]any)
	if !ok {
		return reflect.Value{}, fmt.Errorf("[%v] does not match struct", level)
	}

	if type_.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("[%v] work only on struct, %v provided", level, type_.Kind())
	}

	target := reflect.New(type_).Elem()

	for i := range type_.NumField() {
		field := type_.Field(i)
		if tag, ok := field.Tag.Lookup("json"); ok {
			rawValue, ok := asMap[tag]
			if !ok {
				defaultValue, err := getDefaultValue(ctx, field)
				if err != nil {
					return reflect.Value{}, fmt.Errorf("[%v/%v] unable to parse default value: %v", level, tag, err)
				}
				if !defaultValue.IsValid() {
					return reflect.Value{}, fmt.Errorf("[%v/%v] missing field", level, tag)
				} else {
					target.Field(i).Set(defaultValue)
				}
			} else {
				customParser, err := ctx.getCustomParserIfAny(&field)
				var value reflect.Value
				if err != nil {
					return reflect.Value{}, err
				} else if customParser != nil {
					var stringValue string
					stringValue, err = getAs[string](rawValue)
					if err == nil {
						value, err = (*customParser)(stringValue)
					}
				} else {
					value, err = mapOnAny(ctx, field.Type, rawValue, level+"/"+tag)
				}

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

func mapOnMap(ctx *Context, type_ reflect.Type, raw any, level string) (reflect.Value, error) {
	asMap, ok := raw.(map[string]any)
	if !ok {
		return reflect.Value{}, fmt.Errorf("[%v] does not match map", level)
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
	for k, v := range asMap {
		value, err := mapOnAny(ctx, elemType, v, level+"/"+k)
		if err != nil {
			return reflect.Value{}, err
		}
		target.SetMapIndex(reflect.ValueOf(k), value)
	}
	return target, nil
}

func mapOnInt(type_ reflect.Type, raw any) (reflect.Value, error) {
	value := reflect.New(type_).Elem()
	switch intVal := raw.(type) {
	case int64:
		value.SetInt(int64(intVal))
	case uint64:
		value.SetInt(int64(intVal))
	case float64:
		value.SetInt(int64(intVal))
	case string:
		return stringToInt(type_, intVal)
	default:
		return reflect.Value{}, fmt.Errorf("unsupported type for int: %T", intVal)
	}

	return value, nil
}

func mapOnUint(type_ reflect.Type, raw any) (reflect.Value, error) {
	value := reflect.New(type_).Elem()

	switch intVal := raw.(type) {
	case uint64:
		value.SetUint(uint64(intVal))
	case float64:
		value.SetUint(uint64(intVal))
	case string:
		return stringToUint(type_, intVal)
	default:
		return reflect.Value{}, fmt.Errorf("unsupported type for uint: %T", intVal)
	}
	return value, nil
}

func mapOnString(raw any) (reflect.Value, error) {
	value, err := getAs[string](raw)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(value), nil
}

func mapOnAny(ctx *Context, type_ reflect.Type, raw any, level string) (reflect.Value, error) {
	if type_ == reflect.TypeOf(time.Second) {
		durationAsString, err := getAs[string](raw)
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
		return mapOnString(raw)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return mapOnUint(type_, raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return mapOnInt(type_, raw)
	case reflect.Struct:
		return mapOnStruct(ctx, type_, raw, level)
	case reflect.Map:
		return mapOnMap(ctx, type_, raw, level)
	case reflect.Slice:
		return mapOnSlice(ctx, type_, raw, level)
	case reflect.Interface:
		return reflect.ValueOf(raw), nil
	default:
		return reflect.Value{}, fmt.Errorf("[%v] mapOnAny: unsupported type %v", level, kind)
	}
}

func MapOnStructWithContext[T any](context *Context, raw map[string]any) (T, error) {
	if kind := reflect.TypeFor[T]().Kind(); kind != reflect.Struct {
		var output T
		return output, fmt.Errorf("require a struct type, %v provided", kind)
	}

	s, err := mapOnStruct(context, reflect.TypeFor[T](), raw, "")

	if err != nil {
		return utils.Dummy[T](), err
	} else {
		return s.Interface().(T), nil
	}
}

func MapOnStruct[T any](raw map[string]any) (T, error) {
	ctx := MakeContext()
	return MapOnStructWithContext[T](&ctx, raw)
}
