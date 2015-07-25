package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCli(t *testing.T) {
	app := makeApp()
	assert.NotNil(t, app)
	// How do I test that this exits 0?
	// err := app.Run([]string{""})
	// assert.Error(t, err)
}
