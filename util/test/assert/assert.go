package assert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Contains(t *testing.T, s interface{}, contains interface{}, msgAndArgs ...interface{}) bool {
	return assert.Contains(t, s, contains, msgAndArgs...)
}

func Empty(t *testing.T, object interface{}, msgAndArgs ...interface{}) bool {
	return assert.Empty(t, object, msgAndArgs...)
}

func Equal(t *testing.T, expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool {
	return assert.Equal(t, expected, actual, msgAndArgs...)
}

func EqualError(t *testing.T, theError error, errString string, msgAndArgs ...interface{}) bool {
	return assert.EqualError(t, theError, errString, msgAndArgs...)
}

func EqualValues(t *testing.T, expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool {
	return assert.EqualValues(t, expected, actual, msgAndArgs...)
}

func False(t *testing.T, value bool, msgAndArgs ...interface{}) bool {
	return assert.False(t, value, msgAndArgs...)
}

func Nil(t *testing.T, object interface{}, msgAndArgs ...interface{}) bool {
	return assert.Nil(t, object, msgAndArgs)
}

func NotEmpty(t *testing.T, object interface{}, msgAndArgs ...interface{}) bool {
	return assert.NotEmpty(t, object, msgAndArgs...)
}

func NotEqual(t *testing.T, expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool {
	return assert.NotEqual(t, expected, actual, msgAndArgs...)
}

func NotNil(t *testing.T, object interface{}, msgAndArgs ...interface{}) bool {
	return assert.NotNil(t, object, msgAndArgs...)
}

func True(t *testing.T, value bool, msgAndArgs ...interface{}) bool {
	return assert.True(t, value, msgAndArgs...)
}
