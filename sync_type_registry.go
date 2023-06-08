package argo

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
)

// SyncTypeRegistry contains a map of identifier vs Type
//
// A SyncTypeRegistry is used to unmarshal payloads with core.TypeCarriers.
//
// SyncTypeRegistry is thread safe.
type SyncTypeRegistry[T core.TypeCarrier] struct {
	Typetags []string
	registry map[string]reflect.Type
	types    []string
	lock     sync.RWMutex
}

// NewSyncTypeRegistry creates a new TypeRegistry
func NewSyncTypeRegistry[T core.TypeCarrier](tags ...string) *SyncTypeRegistry[T] {
	return &SyncTypeRegistry[T]{
		Typetags: tags,
		registry: make(map[string]reflect.Type),
		types:    []string{},
		lock:     sync.RWMutex{},
	}
}

// Size returns the number of TypeCarriers in the SyncTypeRegistry
func (registry *SyncTypeRegistry[T]) Size() int {
	registry.lock.RLock()
	defer registry.lock.RUnlock()
	return len(registry.registry)
}

// Length returns the number of TypeCarriers in the SyncTypeRegistry
func (registry *SyncTypeRegistry[T]) Length() int {
	registry.lock.RLock()
	defer registry.lock.RUnlock()
	return len(registry.registry)
}

// AddType adds a Type Tag to the SyncTypeRegistry
func (registry *SyncTypeRegistry[T]) AddTypeTag(tag string) *SyncTypeRegistry[T] {
	registry.lock.Lock()
	defer registry.lock.Unlock()
	registry.Typetags = append(registry.Typetags, tag)
	return registry
}

// AddTypeTags adds one or more Type Tags to the SyncTypeRegistry
func (registry *SyncTypeRegistry[T]) AddTypeTags(tags ...string) *SyncTypeRegistry[T] {
	registry.lock.Lock()
	defer registry.lock.Unlock()
	registry.Typetags = append(registry.Typetags, tags...)
	return registry
}

// Add adds one or more TypeCarriers to the SyncTypeRegistry
func (registry *SyncTypeRegistry[T]) Add(classes ...core.TypeCarrier) *SyncTypeRegistry[T] {
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

// Append adds one or more TypeCarriers to the SyncTypeRegistry
//
// This is a synonym for Add
func (registry *SyncTypeRegistry[T]) Append(classes ...core.TypeCarrier) *SyncTypeRegistry[T] {
	return registry.Add(classes...)
}

// Unmarshal unmarshal a payload into a Type Carrier
//
// The interface that is returned contains a pointer to the TypeCarrier structure.
//
// if the SyncTypeRegistry has no Type Tags, "type" will be used.
func (registry *SyncTypeRegistry[T]) Unmarshal(payload []byte) (object T, err error) {
	registry.lock.RLock()
	defer registry.lock.RUnlock()
	var null T
	if len(registry.Typetags) == 0 {
		registry.Typetags = []string{"type"}
	}
	guts := map[string]json.RawMessage{}
	if err := json.Unmarshal(payload, &guts); err != nil {
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
		if err := json.Unmarshal(payload, value); errors.Is(err, errors.JSONUnmarshalError) {
			return null, err
		} else if err != nil {
			return null, errors.JSONUnmarshalError.Wrap(err)
		}
		return value.(T), nil
	}
	return null, errors.JSONUnmarshalError.Wrap(errors.InvalidType.With(objectType, registry.types))
}
