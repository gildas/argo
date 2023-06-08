package argo_test

import (
	"fmt"
	"testing"

	"github.com/gildas/argo"
	"github.com/gildas/go-errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExampleUnmarshal() {
	object, err := argo.Unmarshal[Something1]([]byte(`{"data": "data"}`))
	if err != nil {
		return
	}
	fmt.Println(object.Data)
	// Output: data
}

func TestCanUnmarshal(t *testing.T) {
	something, err := argo.Unmarshal[Something1]([]byte(`{"type": "something1", "data": "data"}`))
	assert.NoError(t, err, "Failed to unmarshal payload")
	assert.Equal(t, "data", something.Data)
}

func TestShouldFailUnmarshalWithWrongPayload(t *testing.T) {
	_, err := argo.Unmarshal[Something1]([]byte(`{"type": "something1", "data": "notgood"}`))
	assert.Error(t, err, "Should have failed to unmarshal payload")
	assert.ErrorIs(t, err, errors.JSONUnmarshalError)
	assert.ErrorIs(t, err, errors.ArgumentInvalid)

	details := errors.ArgumentInvalid.Clone()
	require.ErrorAs(t, err, &details, "Error chain should contain an ArgumentInvalid")
	assert.Equal(t, "data", details.What)
	assert.Equal(t, "notgood", details.Value)
}

func TestShouldFailUnmarshalWithBogusPayload(t *testing.T) {
	_, err := argo.Unmarshal[Something1]([]byte(`{"type": , "data": "data"}`))
	assert.Error(t, err, "Should have failed to unmarshal payload")
	assert.ErrorIs(t, err, errors.JSONUnmarshalError)
}