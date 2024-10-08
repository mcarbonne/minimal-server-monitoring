package configmapper

import (
	"errors"
	"fmt"
	"reflect"
)

type CustomParserFunction func(string) (reflect.Value, error)

type Context struct {
	customParsers map[string]CustomParserFunction
}

func MakeContext() Context {
	return Context{
		customParsers: make(map[string]CustomParserFunction),
	}
}

func (c *Context) RegisterCustomParser(name string, lambda CustomParserFunction) error {
	_, ok := c.customParsers[name]
	if ok {
		return errors.New("Custom parser " + name + " already registered")
	} else {
		c.customParsers[name] = lambda
		return nil
	}
}

func (ctx *Context) getCustomParserIfAny(structField *reflect.StructField) (*CustomParserFunction, error) {
	if tag, ok := structField.Tag.Lookup("custom"); ok {
		parser, ok := ctx.customParsers[tag]
		if ok {
			return &parser, nil
		} else {
			return nil, fmt.Errorf("custom parser missing for '%v'", tag)
		}
	}
	return nil, nil
}
