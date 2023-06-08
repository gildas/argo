package argo

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
)

// PropertyRegistry contains a map of identifier vs Type
//
// A PropertyRegistry is used to unmarshal payloads that use a property to identify the TypeCarrier.
//
// PropertyRegistry is not thread safe.
type PropertyRegistry[T core.TypeCarrier] struct {
	registry map[string]reflect.Type
	types    []string
}

// NewPropertyRegistry creates a new PropertyRegistry
func NewPropertyRegistry[T core.TypeCarrier](tags ...string) *PropertyRegistry[T] {
	return &PropertyRegistry[T]{
		registry: make(map[string]reflect.Type),
		types:    []string{},
	}
}

// Size returns the number of TypeCarriers in the TypeRegistry
func (registry PropertyRegistry[T]) Size() int {
	return len(registry.registry)
}

// Length returns the number of TypeCarriers in the TypeRegistry
func (registry PropertyRegistry[T]) Length() int {
	return len(registry.registry)
}

// Add adds one or more TypeCarriers to the PropertyRegistry
func (registry *PropertyRegistry[T]) Add(classes ...core.TypeCarrier) *PropertyRegistry[T] {
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

// Append adds one or more TypeCarriers to the PropertyRegistry
//
// This is a synonym for Add
func (registry *PropertyRegistry[T]) Append(classes ...core.TypeCarrier) *PropertyRegistry[T] {
	return registry.Add(classes...)
}

// Unmarshal unmarshal a payload into a Type Carrier
//
// The interface that is returned contains a pointer to the TypeCarrier structure.
//
// Examples:
//  object, err := registry.Unmarshal(payload)
func (registry PropertyRegistry[T]) Unmarshal(payload []byte) (object T, err error) {
	var null T
	guts := map[string]json.RawMessage{}
	if err := json.Unmarshal(payload, &guts); err != nil {
		return null, errors.JSONUnmarshalError.Wrap(err)
	}
	for property := range guts {
		valueType, found := registry.registry[property]
		if !found {
			continue
		}

		// Search for the property in the guts
		if value, found := guts[property]; found {
			instance := reflect.New(valueType).Interface()
			if err := json.Unmarshal(value, instance); errors.Is(err, errors.JSONUnmarshalError) {
				return null, err
			} else if err != nil {
				return null, errors.JSONUnmarshalError.Wrap(err)
			}
			// Return the instance
			return instance.(T), nil
		}
	}
	return null, errors.JSONUnmarshalError.Wrap(errors.ArgumentMissing.With(strings.Join(registry.types, ", ")))
}
