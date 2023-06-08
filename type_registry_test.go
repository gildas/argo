package argo_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gildas/argo"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/stretchr/testify/suite"
)

type TypeRegistrySuite struct {
	suite.Suite
	Name   string
	Start  time.Time
	Logger *logger.Logger
}

func TestTypeRegistrySuite(t *testing.T) {
	suite.Run(t, new(TypeRegistrySuite))
}

// *****************************************************************************
// Suite Tools

func (suite *TypeRegistrySuite) SetupSuite() {
	var err error
	suite.Name = strings.TrimSuffix(reflect.TypeOf(suite).Elem().Name(), "Suite")
	suite.Logger = logger.Create("test",
		&logger.FileStream{
			Path:        fmt.Sprintf("./log/test-%s.log", strings.ToLower(suite.Name)),
			Unbuffered:  true,
			FilterLevels: logger.NewLevelSet(logger.TRACE),
		},
	).Child("test", "test")
	suite.Logger.Infof("Suite Start: %s %s", suite.Name, strings.Repeat("=", 80-14-len(suite.Name)))
	suite.Assert().Nil(err)
}

func (suite *TypeRegistrySuite) TearDownSuite() {
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaning")
	} else {
		suite.Logger.Infof("All tests succeeded, we are cleaning")
	}
	suite.Logger.Infof("Suite End: %s %s", suite.Name, strings.Repeat("=", 80-12-len(suite.Name)))
	suite.Logger.Close()
}

func (suite *TypeRegistrySuite) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Test Start: %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()
}

func (suite *TypeRegistrySuite) AfterTest(suiteName, testName string) {
	duration := time.Since(suite.Start)
	suite.Logger.Record("duration", duration.String()).Infof("Test End: %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
}

func (suite *TypeRegistrySuite) LoadTestData(filename string) []byte {
	payload, err := os.ReadFile(filepath.Join(".", "testdata", filename))
	suite.Require().Nilf(err, "Failed to load test data, error: %s", err)
	return payload
}

// *****************************************************************************
// Suite Tests
func ExampleTypeRegistry_Unmarshal() {
	// Typically, each struct would be declared in its own go file
	// and the argo.TypeRegistry.Add() func would be done in the init() func of each file
	registry := argo.NewTypeRegistry[Something]().Add(Something1{}, Something2{})
	object, err := registry.Unmarshal([]byte(`{"type": "something1", "data": "data"}`))
	if err != nil {
		fmt.Println(err)
		return
	}

	something1, ok := object.(*Something1)
	if !ok {
		fmt.Println("Object is not a Something1")
		return
	}
	fmt.Println(something1.Data)
	// Output: data
}

func ExampleTypeRegistry_Unmarshal_withTypeTag() {
	// Typically, each struct would be declared in its own go file
	// and the argo.TypeRegistry.Add() func would be done in the init() func of each file
	registry := argo.NewTypeRegistry[Something]("__type").Add(Something1{}, Something2{})
	object, err := registry.Unmarshal([]byte(`{"__type": "something1", "data": "data"}`))
	if err != nil {
		fmt.Println(err)
	}

	something1, ok := object.(*Something1)
	if !ok {
		fmt.Println("Object is not a Something1")
		return
	}
	fmt.Println(something1.Data)
	// Output: data
}

func (suite *TypeRegistrySuite) TestCanCreateTypeRegistry() {
	registry := argo.NewTypeRegistry[Something]()
	suite.Assert().NotNil(registry)
	suite.Assert().Len(registry.Typetags, 0)
}

func (suite *TypeRegistrySuite) TestCanCreateTypeRegistryWithTypeTags() {
	registry := argo.NewTypeRegistry[Something]("type1", "type2")
	suite.Assert().NotNil(registry)
	suite.Assert().Len(registry.Typetags, 2)
	suite.Assert().Contains(registry.Typetags, "type1")
	suite.Assert().Contains(registry.Typetags, "type2")

	registry = argo.NewTypeRegistry[Something]()
	suite.Require().NotNil(registry)
	registry.AddTypeTag("type1")
	registry.AddTypeTags("type2", "type3")
	suite.Assert().Len(registry.Typetags, 3)
	suite.Assert().Contains(registry.Typetags, "type1")
	suite.Assert().Contains(registry.Typetags, "type2")
	suite.Assert().Contains(registry.Typetags, "type3")
}

func (suite *TypeRegistrySuite) TestCanAddTypes() {
	// Here we do not use NewTypeRegistry, to test stuff works with default Go constructor
	registry := argo.TypeRegistry[Something]{}

	registry.Add(Something1{}, Something2{})
	suite.Assert().Equal(2, registry.Size())
	suite.Assert().Equal(2, registry.Length())
}

func (suite *TypeRegistrySuite) TestCanUnmarshal() {
	registry := argo.NewTypeRegistry[Something]().Add(Something1{}).Append(Something2{})

	payload := []byte(`{"type": "something1", "data": "data"}`)
	object, err := registry.Unmarshal(payload)
	suite.Require().NoErrorf(err, "Failed to unmarshall, Error: %s", err)
	suite.Require().NotNil(object, "Object should not be nil")

	something1, ok := object.(*Something1)
	suite.Require().Truef(ok, "Object should be a pointer to %s", reflect.TypeOf(Something1{}).Name())
	suite.Require().NotNil(something1, "Something1 should not be nil")
	suite.Assert().Equal("data", something1.Data)
}

func (suite *TypeRegistrySuite) TestCanUnmarshalWithTypeTag() {
	registry := argo.NewTypeRegistry[Something]("__type").Add(Something1{}).Append(Something2{})

	payload := []byte(`{"__type": "something1", "data": "data"}`)
	object, err := registry.Unmarshal(payload)
	suite.Require().NoErrorf(err, "Failed to unmarshall, Error: %s", err)
	suite.Require().NotNil(object, "Object should not be nil")

	something1, ok := object.(*Something1)
	suite.Require().Truef(ok, "Object should be a pointer to %s", reflect.TypeOf(Something1{}).Name())
	suite.Require().NotNil(something1, "Something1 should not be nil")
	suite.Assert().Equal("data", something1.Data)
}

func (suite *TypeRegistrySuite) TestShouldFailWithoutType() {
	registry := argo.NewTypeRegistry[Something]().Add(Something1{}).Append(Something2{})

	payload := []byte(`{"data": "data"}`)
	object, err := registry.Unmarshal(payload)
	suite.Require().Errorf(err, "Should have failed to unmarshall, Error: %s", err)
	suite.Require().Nil(object, "Object should be nil")
	suite.Assert().ErrorIs(err, errors.JSONUnmarshalError)
	suite.Assert().ErrorIs(err, errors.ArgumentMissing)

	details := errors.ArgumentMissing.Clone()
	suite.Require().ErrorAs(err, &details, "Error chain should contain an ArgumentMissing")
	suite.Assert().Equal("type", details.What)
}

func (suite *TypeRegistrySuite) TestShouldFailWithInvalidType() {
	registry := argo.NewTypeRegistry[Something]().Add(Something1{}).Append(Something2{})

	payload := []byte(`{"type": "something3", "data": "data"}`)
	object, err := registry.Unmarshal(payload)
	suite.Require().Errorf(err, "Should have failed to unmarshall, Error: %s", err)
	suite.Require().Nil(object, "Object should be nil")
	suite.Assert().ErrorIs(err, errors.JSONUnmarshalError)
	suite.Assert().ErrorIs(err, errors.InvalidType)

	details := errors.InvalidType.Clone()
	suite.Require().ErrorAs(err, &details, "Error chain should contain an InvalidType")
	suite.Assert().Equal("something3", details.What)
	suite.Assert().Contains(details.Value, "something1")
	suite.Assert().Contains(details.Value, "something2")
}

func (suite *TypeRegistrySuite) TestShouldFailWithInvalidJSON() {
	registry := argo.NewTypeRegistry[Something]().Add(Something1{}).Append(Something2{})

	payload := []byte(`{"type": 3", "data": "data"}`)
	object, err := registry.Unmarshal(payload)
	suite.Require().Errorf(err, "Should have failed to unmarshall, Error: %s", err)
	suite.Require().Nil(object, "Object should be nil")
	suite.Assert().ErrorIs(err, errors.JSONUnmarshalError)
	suite.Assert().Equal("invalid character '\"' after object key:value pair", errors.Unwrap(err).Error())
}

func (suite *TypeRegistrySuite) TestShouldFailWithInvalidDataType() {
	registry := argo.NewTypeRegistry[Something]().Add(Something1{}).Append(Something2{})

	payload := []byte(`{"type": "something1", "data": 2}`)
	object, err := registry.Unmarshal(payload)
	suite.Require().Errorf(err, "Should have failed to unmarshall, Error: %s", err)
	suite.Require().Nil(object, "Object should be nil")
	suite.Assert().ErrorIs(err, errors.JSONUnmarshalError)
	suite.Assert().Equal("json: cannot unmarshal number into Go struct field .data of type string", errors.Unwrap(err).Error())

	payload = []byte(`{"type": "something1", "data": "else"}`)
	object, err = registry.Unmarshal(payload)
	suite.Require().Errorf(err, "Should have failed to unmarshall, Error: %s", err)
	suite.Require().Nil(object, "Object should be nil")
	suite.Assert().ErrorIs(err, errors.JSONUnmarshalError)
	suite.Assert().ErrorIs(err, errors.ArgumentInvalid)

	details := errors.ArgumentInvalid.Clone()
	suite.Require().ErrorAs(err, &details, "Error chain should contain an ArgumentInvalid")
	suite.Assert().Equal("data", details.What)
	suite.Assert().Equal("else", details.Value)
}

func (suite *TypeRegistrySuite) TestShouldFailWithInvalidData() {
	registry := argo.NewTypeRegistry[Something]().Add(Something1{}).Append(Something2{})

	payload := []byte(`{"type": "something1", "data": "else"}`)
	object, err := registry.Unmarshal(payload)
	suite.Require().Errorf(err, "Should have failed to unmarshall, Error: %s", err)
	suite.Require().Nil(object, "Object should be nil")
	suite.Assert().ErrorIs(err, errors.JSONUnmarshalError)
	suite.Assert().ErrorIs(err, errors.ArgumentInvalid)

	details := errors.ArgumentInvalid.Clone()
	suite.Require().ErrorAs(err, &details, "Error chain should contain an ArgumentInvalid")
	suite.Assert().Equal("data", details.What)
	suite.Assert().Equal("else", details.Value)
}
