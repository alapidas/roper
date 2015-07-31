package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//Repos
func TestCreateRepos(t *testing.T) {
	rs := NewRepos()
	assert.NotNil(t, rs)
}

func TestCreateRepo(t *testing.T) {
	// Dir Handling
	r := NewRepo("/myRoot", "myRepoName")
	assert.NotNil(t, r)
}
