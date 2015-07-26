package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRepos(t *testing.T) {
	// Dir Handling
	r := Repos{}
	err := r.AddRepo(&Repo{"datName", "/datPath"})
	assert.Nil(t, err)
}
