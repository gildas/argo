package argo

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
)

// TypeRegistry contains a map of identifier vs Type
//
// A TypeRegistry is used to unmarshal payloads with core.TypeCarriers.
//
// TypeRegistry is not thread safe.
type TypeRegistry[T core.TypeCarrier] struct {
	Typetags []string
	registry map[string]reflect.Type
	types    []string
}
// TypeRegistry[T TypeCarrier] is a TypeRegistry

// NewTypeRegistry creates a new TypeRegistry
func NewTypeRegistry[T core.TypeCarrier](tags ...string) *TypeRegistry[T] {
	return &TypeRegistry[T]{
		Typetags: tags,
		registry: make(map[string]reflect.Type),
		types:    []string{},
	}
}

// Size returns the number of TypeCarriers in the TypeRegistry
func (registry TypeRegistry[T]) Size() int {
	return len(registry.registry)
}

// Length returns the number of TypeCarriers in the TypeRegistry
func (registry TypeRegistry[T]) Length() int {
	return len(registry.registry)
}

// AddType adds a Type Tag to the TypeRegistry
func (registry *TypeRegistry[T]) AddTypeTag(tag string) *TypeRegistry[T] {
	registry.Typetags = append(registry.Typetags, tag)
	return registry
}

// AddTypeTags adds one or more Type Tags to the TypeRegistry
func (registry *TypeRegistry[T]) AddTypeTags(tags ...string) *TypeRegistry[T] {
	registry.Typetags = append(registry.Typetags, tags...)
	return registry
}

// Add adds one or more TypeCarriers to the TypeRegistry
func (registry *TypeRegistry[T]) Add(classes ...core.TypeCarrier) *TypeRegistry[T] {
	if registry.registry == nil {
		registry.registry = make(map[string]reflect.Type)
	}
	if registry.types == nil {
		registry.types = []string{}
	}
	for _, class := range classes {
		typename := class.GetType()
		registry.types = append(registry.types, typename)
		registry.registry[typename] = reflect.TypeOf(class)
	}
	return registry
}

// Append adds one or more TypeCarriers to the TypeRegistry
//
// This is a synonym for Add
func (registry *TypeRegistry[T]) Append(classes ...core.TypeCarrier) *TypeRegistry[T] {
	return registry.Add(classes...)
}

// Unmarshal unmarshal a payload into a Type Carrier
//
// The interface that is returned contains a pointer to the TypeCarrier structure.
//
// if the TypeRegistry has no Type Tags, "type" will be used.
func (registry TypeRegistry[T]) Unmarshal(payload []byte) (object T, err error) {
	var null T
	if len(registry.Typetags) == 0 {
		registry.Typetags = []string{"type"}
	}
	guts := map[string]json.RawMessage{}
	if err = json.Unmarshal(payload, &guts); err != nil {
		return null, errors.JSONUnmarshalError.Wrap(err)
	}
	objectType := ""
	for _, tag := range registry.Typetags {
		if value, found := guts[tag]; found {
			objectType = strings.Trim(string(value), "\"")
		}
	}
	if len(objectType) == 0 {
		return null, errors.JSONUnmarshalError.Wrap(errors.ArgumentMissing.With(registry.Typetags[0]))
	}

	if valueType, found := registry.registry[objectType]; found {
		value := reflect.New(valueType).Interface()
		if err = json.Unmarshal(payload, value); errors.Is(err, errors.JSONUnmarshalError) {
			return null, err
		} else if err != nil {
			return null, errors.JSONUnmarshalError.Wrap(err)
		}
		return value.(T), nil
	}
	return null, errors.JSONUnmarshalError.Wrap(errors.InvalidType.With(objectType, registry.types))
}
