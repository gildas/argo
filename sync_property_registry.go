package argo

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
)

// SyncPropertyRegistry contains a map of identifier vs Type
//
// A SyncPropertyRegistry is used to unmarshal payloads that use a property to identify the TypeCarrier.
//
// SyncPropertyRegistry is thread safe.
type SyncPropertyRegistry[T core.TypeCarrier] struct {
	registry map[string]reflect.Type
	types    []string
	lock     sync.RWMutex
}

// NewSyncPropertyRegistry creates a new SyncPropertyRegistry
func NewSyncPropertyRegistry[T core.TypeCarrier](tags ...string) *SyncPropertyRegistry[T] {
	return &SyncPropertyRegistry[T]{
		registry: make(map[string]reflect.Type),
		types:    []string{},
		lock:     sync.RWMutex{},
	}
}

// Size returns the number of TypeCarriers in the TypeRegistry
func (registry *SyncPropertyRegistry[T]) Size() int {
	registry.lock.RLock()
	defer registry.lock.RUnlock()
	return len(registry.registry)
}

// Length returns the number of TypeCarriers in the TypeRegistry
func (registry *SyncPropertyRegistry[T]) Length() int {
	registry.lock.RLock()
	defer registry.lock.RUnlock()
	return len(registry.registry)
}

// Add adds one or more TypeCarriers to the SyncPropertyRegistry
func (registry *SyncPropertyRegistry[T]) Add(classes ...core.TypeCarrier) *SyncPropertyRegistry[T] {
	registry.lock.Lock()
	defer registry.lock.Unlock()
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

// Append adds one or more TypeCarriers to the SyncPropertyRegistry
//
// This is a synonym for Add
func (registry *SyncPropertyRegistry[T]) Append(classes ...core.TypeCarrier) *SyncPropertyRegistry[T] {
	return registry.Add(classes...)
}

// Unmarshal unmarshal a payload into a Type Carrier
//
// The interface that is returned contains a pointer to the TypeCarrier structure.
//
// Examples:
//  object, err := registry.Unmarshal(payload)
func (registry *SyncPropertyRegistry[T]) Unmarshal(payload []byte) (object T, err error) {
	registry.lock.RLock()
	defer registry.lock.RUnlock()
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
