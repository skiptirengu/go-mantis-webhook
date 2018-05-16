package util

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/pkg/errors"
)

type testStru struct {
	test string
}

func TestMapStringSlice(t *testing.T) {
	testStruct := []testStru{
		{"one"},
		{"two"},
	}
	ret := MapStringSlice(testStruct, func(val interface{}) string { return val.(testStru).test })
	assert.Len(t, ret, 2)
	assert.Equal(t, "one", ret[0])
	assert.Equal(t, "two", ret[1])
}

func TestPopError(t *testing.T) {
	errs := []error{errors.New("One"), errors.New("Two")}
	err := PopError(errs)
	assert.Equal(t, "Two", err.Error())
}
