package configmapper

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils"
)

var (
	ErrInvalidType        = errors.New("invalid type")
	ErrParsingFailed      = errors.New("parsing failed")
	ErrUnsupportedDefault = errors.New("unsupported default value")
	ErrStructureMismatch  = errors.New("structure mismatch")
)

func getAs[Output any](input any) (Output, error) {
	value, ok := input.(Output)
	if !ok {
		return utils.Dummy[Output](), fmt.Errorf("%w: was expecting type %v, got %v (%v)", ErrInvalidType, reflect.TypeFor[Output](), reflect.TypeOf(input), input)
	}
	return value, nil
}

func stringToInt(type_ reflect.Type, valueAsString string) (reflect.Value, error) {
	defaultValue := reflect.New(type_).Elem()
	intVal, err := strconv.ParseInt(valueAsString, 10, 64)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("%w: unable to parse int: %v", ErrParsingFailed, err)
	}
	reflect.New(type_)
	defaultValue.SetInt(intVal)
	return defaultValue, nil
}

func stringToUint(type_ reflect.Type, valueAsString string) (reflect.Value, error) {
	defaultValue := reflect.New(type_).Elem()
	uintVal, err := strconv.ParseUint(valueAsString, 10, 64)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("%w: unable to parse uint: %v", ErrParsingFailed, err)
	}
	reflect.New(type_)
	defaultValue.SetUint(uintVal)
	return defaultValue, nil
}

func stringToBool(valueAsString string) (reflect.Value, error) {
	boolVal, err := strconv.ParseBool(valueAsString)
	return reflect.ValueOf(boolVal), err
}

func getDefaultValue(ctx *Context, field reflect.StructField) (reflect.Value, error) {
	if tag, ok := field.Tag.Lookup("default"); ok {

		customFieldParser, err := ctx.getCustomFieldParserIfAny(&field)
		if err != nil {
			return reflect.Value{}, err
		} else if customFieldParser != nil {
			return (*customFieldParser)(tag)
		}

		// Check for encoding.TextUnmarshaler
		if val, ok, err := tryMapUsingTextUnmarshaler(field.Type, tag); ok {
			if err != nil {
				return reflect.Value{}, fmt.Errorf("%w: unable to unmarshal default value: %v", ErrParsingFailed, err)
			}
			return val, nil
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
				return reflect.Value{}, fmt.Errorf("%w: array/slice %v", ErrUnsupportedDefault, tag)
			}
		case reflect.Map:
			if tag == "{}" {
				return reflect.MakeMap(field.Type), nil
			} else {
				return reflect.Value{}, fmt.Errorf("%w: map %v", ErrUnsupportedDefault, tag)
			}
		case reflect.Struct:
			if tag == "{}" {
				defaultValue, err := mapOnStruct(ctx, field.Type, map[string]any{}, "")
				if err != nil {
					return reflect.Value{}, fmt.Errorf("%w: unable to default struct: %v", ErrParsingFailed, err)
				} else {
					return defaultValue, nil
				}
			} else {
				return reflect.Value{}, fmt.Errorf("%w: map %v", ErrUnsupportedDefault, tag)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return stringToInt(field.Type, tag)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return stringToUint(field.Type, tag)
		case reflect.String:
			return reflect.ValueOf(tag), nil
		case reflect.Bool:
			return stringToBool(tag)
		default:
			return reflect.Value{}, fmt.Errorf("%w: %v for default", ErrUnsupportedDefault, kind)
		}

	}
	return reflect.Value{}, nil
}

func mapOnSlice(ctx *Context, type_ reflect.Type, raw any, level string) (reflect.Value, error) {
	asSlice, ok := raw.([]any)
	if !ok {
		return reflect.Value{}, fmt.Errorf("%w: [%v] expected slice", ErrStructureMismatch, level)
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

func mapOnPtr(ctx *Context, type_ reflect.Type, raw any, level string) (reflect.Value, error) {
	if raw == nil {
		return reflect.Value{}, nil
	} else {
		value, err := mapOnAny(ctx, type_.Elem(), raw, level)
		newValue := reflect.New(value.Type())
		newValue.Elem().Set(value)
		return newValue, err
	}
}

func mapOnStruct(ctx *Context, type_ reflect.Type, raw any, level string) (reflect.Value, error) {
	asMap, ok := raw.(map[string]any)
	if !ok {
		return reflect.Value{}, fmt.Errorf("%w: [%v] expected struct", ErrStructureMismatch, level)
	}

	if type_.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("%w: [%v] work only on struct, %v provided", ErrStructureMismatch, level, type_.Kind())
	}

	target := reflect.New(type_).Elem()

	for i := range type_.NumField() {
		field := type_.Field(i)
		if tag, ok := field.Tag.Lookup("json"); ok {
			rawValue, ok := asMap[tag]
			isOptional := field.Type.Kind() == reflect.Ptr
			if !ok && !isOptional {
				defaultValue, err := getDefaultValue(ctx, field)
				if err != nil {
					return reflect.Value{}, fmt.Errorf("%w: [%v/%v] unable to parse default value: %v", ErrParsingFailed, level, tag, err)
				}
				if !defaultValue.IsValid() {
					return reflect.Value{}, fmt.Errorf("%w: [%v/%v] missing field", ErrStructureMismatch, level, tag)
				} else {
					target.Field(i).Set(defaultValue)
				}
			} else {
				customParser, err := ctx.getCustomFieldParserIfAny(&field)
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
				} else if value.IsValid() {
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
		return reflect.Value{}, fmt.Errorf("%w: [%v] expected map", ErrStructureMismatch, level)
	}

	if type_.Kind() != reflect.Map {
		return reflect.Value{}, fmt.Errorf("%w: [%v] work only on map", ErrStructureMismatch, level)
	}

	elemType := type_.Elem()
	keyType := type_.Key()

	if keyType.Kind() != reflect.String {
		return reflect.Value{}, fmt.Errorf("%w: [%v] only string is allowed for key (%v given)", ErrInvalidType, level, keyType)
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
		return reflect.Value{}, fmt.Errorf("%w: %T", ErrInvalidType, intVal)
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
		return reflect.Value{}, fmt.Errorf("%w: %T", ErrInvalidType, intVal)
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

func mapOnBool(raw any) (reflect.Value, error) {
	value, err := getAs[bool](raw)
	return reflect.ValueOf(value), err
}

func tryMapUsingTextUnmarshaler(type_ reflect.Type, rawValue any) (reflect.Value, bool, error) {
	newValue := reflect.New(type_)
	if unmarshaler, ok := newValue.Interface().(encoding.TextUnmarshaler); ok {
		value, err := getAs[string](rawValue)
		if err != nil {
			return reflect.Value{}, true, fmt.Errorf("%w: expected string for TextUnmarshaler: %v", ErrInvalidType, err)
		}
		if err := unmarshaler.UnmarshalText([]byte(value)); err != nil {
			return reflect.Value{}, true, err
		}
		return newValue.Elem(), true, nil
	}
	return reflect.Value{}, false, nil
}

func mapOnAny(ctx *Context, type_ reflect.Type, raw any, level string) (reflect.Value, error) {

	// Check for encoding.TextUnmarshaler
	if val, ok, err := tryMapUsingTextUnmarshaler(type_, raw); ok {
		if err != nil {
			return reflect.Value{}, fmt.Errorf("%w: [%v] unable to unmarshal values: %v", ErrParsingFailed, level, err)
		}
		return val, nil
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
	case reflect.Ptr:
		return mapOnPtr(ctx, type_, raw, level)
	case reflect.Bool:
		return mapOnBool(raw)
	default:
		return reflect.Value{}, fmt.Errorf("%w: [%v] mapOnAny: unsupported type %v", ErrInvalidType, level, kind)
	}
}

func MapOnStructWithContext[T any](context *Context, raw map[string]any) (T, error) {
	if kind := reflect.TypeFor[T]().Kind(); kind != reflect.Struct {
		var output T
		return output, fmt.Errorf("%w: require a struct type, %v provided", ErrInvalidType, kind)
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
