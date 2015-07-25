package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRepos(t *testing.T) {
	// Dir Handling
	r := Repos{}
	r.AddRepo(NewRepo("datName", "/datPath"))
	assert.Nil(t, nil)
}
