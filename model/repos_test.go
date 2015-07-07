package model

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitialize(t *testing.T) {
	// Dir Handling
	r := Repos{}
	r.AddRepo(Repo{"datName", "/datPath"})
	fmt.Printf("%#v\n", r)
	assert.Nil(t, fmt.Errorf("a"))
}
