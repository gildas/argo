package argo

import (
	"encoding/json"

	"github.com/gildas/go-errors"
)

// Unmarshal unmarshals the payload into a generic type
func Unmarshal[T any](payload []byte) (object T, err error) {
	if err = json.Unmarshal(payload, &object); errors.Is(err, errors.JSONUnmarshalError) {
		return object, err
	}
	return object, errors.JSONUnmarshalError.Wrap(err)
}
