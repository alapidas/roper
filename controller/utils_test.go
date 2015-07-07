package controller

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPathExists(t *testing.T) {
	exists, err := pathExists("/tmpppp")
	assert.False(t, exists)
	assert.Nil(t, err)
	exists, err = pathExists("/tmp")
	assert.True(t, exists)
	assert.Nil(t, err)
}
