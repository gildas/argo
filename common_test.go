package argo_test

import (
	"encoding/json"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
)

type Something interface {
	core.TypeCarrier
}

type Something1 struct {
	Data string `json:"data"`
}

func (s Something1) GetType() string {
	return "something1"
}

type Something2 struct {
	Data string `json:"data"`
}

func (s Something2) GetType() string {
	return "something2"
}

type Something3 struct {
	Blob []byte `json:"blob"`
}

func (s Something3) GetType() string {
	return "something3"
}

func (s *Something1) UnmarshalJSON(payload []byte) (err error) {
	type surrogate Something1
	var inner struct {
		surrogate
	}

	if err := json.Unmarshal(payload, &inner); errors.Is(err, errors.JSONUnmarshalError) {
		return err
	} else if err != nil {
		return err
	}
	*s = Something1(inner.surrogate)
	if s.Data != "data" {
		return errors.JSONUnmarshalError.Wrap(errors.ArgumentInvalid.With("data", s.Data))
	}

	return nil
}
