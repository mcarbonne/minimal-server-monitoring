package configmapper

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrCustomParserMissing           = errors.New("custom parser missing")
	ErrCustomParserAlreadyRegistered = errors.New("custom parser already registered")
)

type CustomFieldsParserFunction func(string) (reflect.Value, error)

type Context struct {
	customFieldsParsers map[string]CustomFieldsParserFunction
}

func MakeContext() Context {
	return Context{
		customFieldsParsers: make(map[string]CustomFieldsParserFunction),
	}
}

func (c *Context) RegisterCustomFieldParser(name string, lambda CustomFieldsParserFunction) error {
	_, ok := c.customFieldsParsers[name]
	if ok {
		return fmt.Errorf("%w: %s", ErrCustomParserAlreadyRegistered, name)
	} else {
		c.customFieldsParsers[name] = lambda
		return nil
	}
}

func (ctx *Context) getCustomFieldParserIfAny(structField *reflect.StructField) (*CustomFieldsParserFunction, error) {
	if tag, ok := structField.Tag.Lookup("custom"); ok {
		parser, ok := ctx.customFieldsParsers[tag]
		if ok {
			return &parser, nil
		} else {
			return nil, fmt.Errorf("%w: '%v'", ErrCustomParserMissing, tag)
		}
	}
	return nil, nil
}
