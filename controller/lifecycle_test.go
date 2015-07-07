package controller

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestInitialize(t *testing.T) {
	// Dir Handling
	os.Setenv(reposEnvVar, "")
	err := Initialize()
	assert.Error(t, err)
	os.Setenv(reposEnvVar, "abcdefgh")
	err = Initialize()
	assert.Error(t, err)
	os.Setenv(reposEnvVar, os.TempDir())
	err = Initialize()
	assert.Nil(t, err)
	os.Setenv(reposEnvVar, fmt.Sprintf("%v,%v", os.TempDir(), "apples"))
	err = Initialize()
	assert.Nil(t, err)
}
