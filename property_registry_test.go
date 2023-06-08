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

type PropertyRegistrySuite struct {
	suite.Suite
	Name   string
	Start  time.Time
	Logger *logger.Logger
}

func TestPropertyRegistrySuite(t *testing.T) {
	suite.Run(t, new(PropertyRegistrySuite))
}

// *****************************************************************************
// Suite Tools

func (suite *PropertyRegistrySuite) SetupSuite() {
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

func (suite *PropertyRegistrySuite) TearDownSuite() {
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaning")
	} else {
		suite.Logger.Infof("All tests succeeded, we are cleaning")
	}
	suite.Logger.Infof("Suite End: %s %s", suite.Name, strings.Repeat("=", 80-12-len(suite.Name)))
	suite.Logger.Close()
}

func (suite *PropertyRegistrySuite) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Test Start: %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()
}

func (suite *PropertyRegistrySuite) AfterTest(suiteName, testName string) {
	duration := time.Since(suite.Start)
	suite.Logger.Record("duration", duration.String()).Infof("Test End: %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
}

func (suite *PropertyRegistrySuite) LoadTestData(filename string) []byte {
	payload, err := os.ReadFile(filepath.Join(".", "testdata", filename))
	suite.Require().Nilf(err, "Failed to load test data, error: %s", err)
	return payload
}

// *****************************************************************************
// Suite Tests

func ExamplePropertyRegistry_Unmarshal() {
	registry := argo.NewPropertyRegistry[Something]().Add(Something1{}, Something2{})
	object, err := registry.Unmarshal([]byte(`{"something2":{"data":"data"}}`))
	if err != nil {
		fmt.Println(err)
	}

	something2, ok := object.(*Something2)
	if !ok {
		fmt.Println("Object is not a Something2")
		return
	}
	fmt.Println(something2.Data)
	// Output: data
}

func (suite *PropertyRegistrySuite) TestCanCreatePropertyRegistry() {
	registry := argo.NewPropertyRegistry[Something]()
	suite.Assert().NotNil(registry)
}

func (suite *PropertyRegistrySuite) TestCanAddTypes() {
	// Here we do not use NewPropertyRegistry, to test stuff works with default Go constructor
	registry := argo.PropertyRegistry[Something]{}

	registry.Add(Something1{}, Something2{})
	suite.Assert().Equal(2, registry.Size())
	suite.Assert().Equal(2, registry.Length())
}

func (suite *PropertyRegistrySuite) TestShouldFailWithoutProperty() {
	registry := argo.NewPropertyRegistry[Something]().Add(Something1{}).Append(Something2{})

	payload := []byte(`{"something3":{"data":"data"}}`)
	object, err := registry.Unmarshal(payload)
	suite.Require().Errorf(err, "Should have failed to unmarshall, Error: %s", err)
	suite.Require().Nil(object, "Object should be nil")
	suite.Assert().ErrorIs(err, errors.JSONUnmarshalError)
	suite.Assert().ErrorIs(err, errors.ArgumentMissing)

	details := errors.ArgumentMissing.Clone()
	suite.Require().ErrorAs(err, &details, "Error chain should contain an ArgumentMissing")
	suite.Assert().Equal("something1, something2", details.What)
}

func (suite *PropertyRegistrySuite) TestShouldFailWithInvalidJSON() {
	registry := argo.NewPropertyRegistry[Something]().Add(Something1{}).Append(Something2{})

	payload := []byte(`{"type": 3", "data": "data"}`)
	object, err := registry.Unmarshal(payload)
	suite.Require().Errorf(err, "Should have failed to unmarshall, Error: %s", err)
	suite.Require().Nil(object, "Object should be nil")
	suite.Assert().ErrorIs(err, errors.JSONUnmarshalError)
	suite.Assert().Equal("invalid character '\"' after object key:value pair", errors.Unwrap(err).Error())
}

func (suite *PropertyRegistrySuite) TestShouldFailWithInvalidData() {
	registry := argo.NewPropertyRegistry[Something]().Add(Something1{}).Append(Something2{})

	payload := []byte(`{"something1":{"data":2}}`)
	object, err := registry.Unmarshal(payload)
	suite.Require().Errorf(err, "Should have failed to unmarshall, Error: %s", err)
	suite.Require().Nil(object, "Object should be nil")
	suite.Assert().ErrorIs(err, errors.JSONUnmarshalError)
	suite.Assert().Equal("json: cannot unmarshal number into Go struct field .data of type string", errors.Unwrap(err).Error())

	payload = []byte(`{"something1":{"data":"else"}}`)
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
